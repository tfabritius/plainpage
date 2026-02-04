package service

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/tfabritius/plainpage/libs/utils"
	"github.com/tfabritius/plainpage/model"
	"gopkg.in/yaml.v3"
)

const (
	refreshTokenLength = 32
	// RefreshTokenValidity is the duration a refresh token remains valid (exported for use in cookie MaxAge)
	RefreshTokenValidity = 90 * 24 * time.Hour // 90 days
)

// RefreshTokenIndexEntry represents an entry in the refresh_tokens.yml index file
type RefreshTokenIndexEntry struct {
	ID     string `yaml:"id"`
	UserID string `yaml:"userId"`
}

// RefreshTokenData represents the data stored in individual token files
type RefreshTokenData struct {
	UserID     string    `yaml:"userId"`
	CreatedAt  time.Time `yaml:"createdAt"`
	LastUsedAt time.Time `yaml:"lastUsedAt"`
	ExpiresAt  time.Time `yaml:"expiresAt"`
}

// RefreshToken combines index entry and data for internal use
type RefreshToken struct {
	ID         string
	UserID     string
	CreatedAt  time.Time
	LastUsedAt time.Time
	ExpiresAt  time.Time
}

func NewRefreshTokenService(store model.Storage) *RefreshTokenService {
	s := &RefreshTokenService{
		storage: store,
	}

	// Initialize refresh_tokens.yml index if it doesn't exist
	if !s.storage.Exists("refresh_tokens.yml") {
		err := s.saveIndexUnlocked([]RefreshTokenIndexEntry{})
		if err != nil {
			log.Fatalln("Could not create refresh_tokens.yml:", err)
		}
	}

	// Ensure refresh_tokens directory exists
	if !s.storage.Exists("refresh_tokens") {
		err := s.storage.CreateDirectory("refresh_tokens")
		if err != nil {
			log.Fatalln("Could not create refresh_tokens directory:", err)
		}
	}

	return s
}

type RefreshTokenService struct {
	storage model.Storage
	mu      sync.RWMutex
}

// readIndexUnlocked reads the refresh_tokens.yml index file (caller must hold lock)
func (s *RefreshTokenService) readIndexUnlocked() ([]RefreshTokenIndexEntry, error) {
	bytes, err := s.storage.ReadFile("refresh_tokens.yml")
	if err != nil {
		return nil, fmt.Errorf("could not read refresh_tokens.yml: %w", err)
	}

	var index []RefreshTokenIndexEntry
	if err := yaml.Unmarshal(bytes, &index); err != nil {
		return nil, fmt.Errorf("could not parse refresh_tokens.yml: %w", err)
	}

	return index, nil
}

// saveIndexUnlocked writes the refresh_tokens.yml index file (caller must hold lock)
func (s *RefreshTokenService) saveIndexUnlocked(index []RefreshTokenIndexEntry) error {
	bytes, err := yaml.Marshal(&index)
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	if err := s.storage.WriteFile("refresh_tokens.yml", bytes); err != nil {
		return fmt.Errorf("could not write refresh_tokens.yml: %w", err)
	}

	return nil
}

// tokenFilePath returns the path to a token's data file
func (s *RefreshTokenService) tokenFilePath(tokenID string) string {
	return fmt.Sprintf("refresh_tokens/%s.yml", tokenID)
}

// readTokenData reads an individual token's data file
func (s *RefreshTokenService) readTokenData(tokenID string) (*RefreshTokenData, error) {
	bytes, err := s.storage.ReadFile(s.tokenFilePath(tokenID))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("could not read token file: %w", err)
	}

	var data RefreshTokenData
	if err := yaml.Unmarshal(bytes, &data); err != nil {
		return nil, fmt.Errorf("could not parse token file: %w", err)
	}

	return &data, nil
}

// saveTokenData writes an individual token's data file
func (s *RefreshTokenService) saveTokenData(tokenID string, data *RefreshTokenData) error {
	bytes, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal token data: %w", err)
	}

	if err := s.storage.WriteFile(s.tokenFilePath(tokenID), bytes); err != nil {
		return fmt.Errorf("could not write token file: %w", err)
	}

	return nil
}

// deleteTokenData deletes an individual token's data file
func (s *RefreshTokenService) deleteTokenData(tokenID string) error {
	err := s.storage.DeleteFile(s.tokenFilePath(tokenID))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not delete token file: %w", err)
	}
	return nil
}

// Create generates a new refresh token for the given user
func (s *RefreshTokenService) Create(userID string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate random token ID
	tokenID, err := utils.GenerateRandomString(refreshTokenLength)
	if err != nil {
		return "", fmt.Errorf("could not generate token ID: %w", err)
	}

	now := time.Now()

	// Create token data file
	data := &RefreshTokenData{
		UserID:     userID,
		CreatedAt:  now,
		LastUsedAt: now,
		ExpiresAt:  now.Add(RefreshTokenValidity),
	}

	if err := s.saveTokenData(tokenID, data); err != nil {
		return "", err
	}

	// Add to index
	index, err := s.readIndexUnlocked()
	if err != nil {
		// Cleanup token file on index read failure
		_ = s.deleteTokenData(tokenID)
		return "", err
	}

	index = append(index, RefreshTokenIndexEntry{
		ID:     tokenID,
		UserID: userID,
	})

	if err := s.saveIndexUnlocked(index); err != nil {
		// Cleanup token file on index write failure
		_ = s.deleteTokenData(tokenID)
		return "", err
	}

	return tokenID, nil
}

// Validate checks if a token is valid and returns the associated user ID
func (s *RefreshTokenService) Validate(tokenID string) (string, error) {
	// No lock needed for reading individual token file
	data, err := s.readTokenData(tokenID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return "", model.ErrNotFound
		}
		return "", err
	}

	// Check if token is expired
	if time.Now().After(data.ExpiresAt) {
		return "", model.ErrNotFound
	}

	return data.UserID, nil
}

// Refresh updates the token's lastUsedAt and expiresAt timestamps
func (s *RefreshTokenService) Refresh(tokenID string) error {
	// No lock needed for individual token file updates
	data, err := s.readTokenData(tokenID)
	if err != nil {
		return err
	}

	// Check if token is expired
	if time.Now().After(data.ExpiresAt) {
		return model.ErrNotFound
	}

	now := time.Now()
	data.LastUsedAt = now
	data.ExpiresAt = now.Add(RefreshTokenValidity)

	return s.saveTokenData(tokenID, data)
}

// Delete removes a specific refresh token
func (s *RefreshTokenService) Delete(tokenID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Delete token file
	if err := s.deleteTokenData(tokenID); err != nil {
		return err
	}

	// Remove from index
	index, err := s.readIndexUnlocked()
	if err != nil {
		return err
	}

	newIndex := make([]RefreshTokenIndexEntry, 0, len(index))
	for _, entry := range index {
		if entry.ID != tokenID {
			newIndex = append(newIndex, entry)
		}
	}

	return s.saveIndexUnlocked(newIndex)
}

// DeleteAllForUser removes all refresh tokens for a specific user
func (s *RefreshTokenService) DeleteAllForUser(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	index, err := s.readIndexUnlocked()
	if err != nil {
		return err
	}

	// Find and delete all tokens for the user
	newIndex := make([]RefreshTokenIndexEntry, 0, len(index))
	for _, entry := range index {
		if entry.UserID == userID {
			// Delete token file (ignore errors for individual files)
			if err := s.deleteTokenData(entry.ID); err != nil {
				return fmt.Errorf("could not delete all tokens: %w", err)
			}
		} else {
			newIndex = append(newIndex, entry)
		}
	}

	return s.saveIndexUnlocked(newIndex)
}

// GetTokensForUser returns all tokens for a specific user (for future session management)
func (s *RefreshTokenService) GetTokensForUser(userID string) ([]RefreshToken, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	index, err := s.readIndexUnlocked()
	if err != nil {
		return nil, err
	}

	var tokens []RefreshToken
	for _, entry := range index {
		if entry.UserID == userID {
			data, err := s.readTokenData(entry.ID)
			if err != nil {
				// Skip tokens with read errors
				continue
			}

			tokens = append(tokens, RefreshToken{
				ID:         entry.ID,
				UserID:     data.UserID,
				CreatedAt:  data.CreatedAt,
				LastUsedAt: data.LastUsedAt,
				ExpiresAt:  data.ExpiresAt,
			})
		}
	}

	return tokens, nil
}

// CleanupExpired removes all expired tokens (can be called periodically)
func (s *RefreshTokenService) CleanupExpired() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	index, err := s.readIndexUnlocked()
	if err != nil {
		return err
	}

	now := time.Now()
	newIndex := make([]RefreshTokenIndexEntry, 0, len(index))

	for _, entry := range index {
		data, err := s.readTokenData(entry.ID)
		if err != nil || now.After(data.ExpiresAt) {
			// Delete expired or unreadable tokens
			_ = s.deleteTokenData(entry.ID)
		} else {
			newIndex = append(newIndex, entry)
		}
	}

	return s.saveIndexUnlocked(newIndex)
}
