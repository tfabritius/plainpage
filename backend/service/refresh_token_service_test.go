package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tfabritius/plainpage/model"
)

func newTestRefreshTokenService(t *testing.T) *RefreshTokenService {
	tempDir := t.TempDir()
	storage := NewFsStorage(tempDir)
	return NewRefreshTokenService(storage)
}

func TestRefreshTokenCreate(t *testing.T) {
	r := require.New(t)
	service := newTestRefreshTokenService(t)

	userID := "test-user-123"

	// Create a token
	tokenID, err := service.Create(userID)
	r.NoError(err)
	r.NotEmpty(tokenID)
	r.Len(tokenID, refreshTokenLength)

	// Verify token can be validated
	validatedUserID, err := service.Validate(tokenID)
	r.NoError(err)
	r.Equal(userID, validatedUserID)

	// Verify token appears in user's tokens
	tokens, err := service.GetTokensForUser(userID)
	r.NoError(err)
	r.Len(tokens, 1)
	r.Equal(tokenID, tokens[0].ID)
	r.Equal(userID, tokens[0].UserID)
}

func TestRefreshTokenValidate(t *testing.T) {
	r := require.New(t)
	service := newTestRefreshTokenService(t)

	userID := "test-user-456"

	// Create a valid token
	tokenID, err := service.Create(userID)
	r.NoError(err)

	// Valid token returns correct user ID
	validatedUserID, err := service.Validate(tokenID)
	r.NoError(err)
	r.Equal(userID, validatedUserID)

	// Non-existent token returns error
	_, err = service.Validate("non-existent-token")
	r.Error(err)
}

func TestRefreshTokenValidateExpired(t *testing.T) {
	r := require.New(t)
	tempDir := t.TempDir()
	storage := NewFsStorage(tempDir)
	service := NewRefreshTokenService(storage)

	userID := "test-user-expired"

	// Create a token
	tokenID, err := service.Create(userID)
	r.NoError(err)

	// Manually expire the token by modifying the data file
	data, err := service.readTokenData(tokenID)
	r.NoError(err)
	data.ExpiresAt = time.Now().Add(-1 * time.Hour) // Expired 1 hour ago
	err = service.saveTokenData(tokenID, data)
	r.NoError(err)

	// Expired token returns ErrNotFound
	_, err = service.Validate(tokenID)
	r.ErrorIs(err, model.ErrNotFound)
}

func TestRefreshTokenRefresh(t *testing.T) {
	r := require.New(t)
	service := newTestRefreshTokenService(t)

	userID := "test-user-refresh"

	// Create a token
	tokenID, err := service.Create(userID)
	r.NoError(err)

	// Get initial token data
	tokens, err := service.GetTokensForUser(userID)
	r.NoError(err)
	r.Len(tokens, 1)
	initialLastUsedAt := tokens[0].LastUsedAt
	initialExpiresAt := tokens[0].ExpiresAt

	// Wait a bit to ensure time difference
	time.Sleep(10 * time.Millisecond)

	// Refresh the token
	err = service.Refresh(tokenID)
	r.NoError(err)

	// Verify timestamps were updated
	tokens, err = service.GetTokensForUser(userID)
	r.NoError(err)
	r.Len(tokens, 1)
	r.True(tokens[0].LastUsedAt.After(initialLastUsedAt))
	r.True(tokens[0].ExpiresAt.After(initialExpiresAt))
}

func TestRefreshTokenRefreshExpired(t *testing.T) {
	r := require.New(t)
	tempDir := t.TempDir()
	storage := NewFsStorage(tempDir)
	service := NewRefreshTokenService(storage)

	userID := "test-user-refresh-expired"

	// Create a token
	tokenID, err := service.Create(userID)
	r.NoError(err)

	// Manually expire the token
	data, err := service.readTokenData(tokenID)
	r.NoError(err)
	data.ExpiresAt = time.Now().Add(-1 * time.Hour)
	err = service.saveTokenData(tokenID, data)
	r.NoError(err)

	// Refreshing expired token returns ErrNotFound
	err = service.Refresh(tokenID)
	r.ErrorIs(err, model.ErrNotFound)
}

func TestRefreshTokenDelete(t *testing.T) {
	r := require.New(t)
	service := newTestRefreshTokenService(t)

	userID := "test-user-delete"

	// Create a token
	tokenID, err := service.Create(userID)
	r.NoError(err)

	// Verify token exists
	_, err = service.Validate(tokenID)
	r.NoError(err)

	// Delete the token
	err = service.Delete(tokenID)
	r.NoError(err)

	// Token no longer exists
	_, err = service.Validate(tokenID)
	r.Error(err)

	// Token no longer appears in user's tokens
	tokens, err := service.GetTokensForUser(userID)
	r.NoError(err)
	r.Len(tokens, 0)

	// Deleting non-existent token returns error (file not found)
	err = service.Delete("non-existent-token")
	r.Error(err)
}

func TestRefreshTokenDeleteAllForUser(t *testing.T) {
	r := require.New(t)
	service := newTestRefreshTokenService(t)

	user1 := "test-user-1"
	user2 := "test-user-2"

	// Create multiple tokens for user1
	token1a, err := service.Create(user1)
	r.NoError(err)
	token1b, err := service.Create(user1)
	r.NoError(err)

	// Create a token for user2
	token2, err := service.Create(user2)
	r.NoError(err)

	// Verify all tokens exist
	tokens1, err := service.GetTokensForUser(user1)
	r.NoError(err)
	r.Len(tokens1, 2)

	tokens2, err := service.GetTokensForUser(user2)
	r.NoError(err)
	r.Len(tokens2, 1)

	// Delete all tokens for user1
	err = service.DeleteAllForUser(user1)
	r.NoError(err)

	// User1's tokens are gone
	_, err = service.Validate(token1a)
	r.Error(err)
	_, err = service.Validate(token1b)
	r.Error(err)

	tokens1, err = service.GetTokensForUser(user1)
	r.NoError(err)
	r.Len(tokens1, 0)

	// User2's token still exists
	_, err = service.Validate(token2)
	r.NoError(err)

	tokens2, err = service.GetTokensForUser(user2)
	r.NoError(err)
	r.Len(tokens2, 1)
}

func TestRefreshTokenGetTokensForUser(t *testing.T) {
	r := require.New(t)
	service := newTestRefreshTokenService(t)

	user1 := "test-user-get-1"
	user2 := "test-user-get-2"

	// Empty initially
	tokens, err := service.GetTokensForUser(user1)
	r.NoError(err)
	r.Len(tokens, 0)

	// Create tokens
	_, err = service.Create(user1)
	r.NoError(err)
	_, err = service.Create(user1)
	r.NoError(err)
	_, err = service.Create(user2)
	r.NoError(err)

	// Check user1's tokens
	tokens, err = service.GetTokensForUser(user1)
	r.NoError(err)
	r.Len(tokens, 2)
	for _, token := range tokens {
		r.Equal(user1, token.UserID)
		r.NotEmpty(token.ID)
		r.False(token.CreatedAt.IsZero())
		r.False(token.LastUsedAt.IsZero())
		r.False(token.ExpiresAt.IsZero())
	}

	// Check user2's tokens
	tokens, err = service.GetTokensForUser(user2)
	r.NoError(err)
	r.Len(tokens, 1)
	r.Equal(user2, tokens[0].UserID)
}

func TestRefreshTokenCleanupExpired(t *testing.T) {
	r := require.New(t)
	tempDir := t.TempDir()
	storage := NewFsStorage(tempDir)
	service := NewRefreshTokenService(storage)

	userID := "test-user-cleanup"

	// Create tokens
	token1, err := service.Create(userID)
	r.NoError(err)
	token2, err := service.Create(userID)
	r.NoError(err)
	token3, err := service.Create(userID)
	r.NoError(err)

	// Expire token1 and token2
	data1, err := service.readTokenData(token1)
	r.NoError(err)
	data1.ExpiresAt = time.Now().Add(-1 * time.Hour)
	err = service.saveTokenData(token1, data1)
	r.NoError(err)

	data2, err := service.readTokenData(token2)
	r.NoError(err)
	data2.ExpiresAt = time.Now().Add(-2 * time.Hour)
	err = service.saveTokenData(token2, data2)
	r.NoError(err)

	// token3 remains valid

	// Verify all 3 tokens exist in index before cleanup
	tokens, err := service.GetTokensForUser(userID)
	r.NoError(err)
	r.Len(tokens, 3)

	// Run cleanup
	err = service.CleanupExpired()
	r.NoError(err)

	// Only token3 should remain
	tokens, err = service.GetTokensForUser(userID)
	r.NoError(err)
	r.Len(tokens, 1)
	r.Equal(token3, tokens[0].ID)

	// Expired tokens are gone
	_, err = service.Validate(token1)
	r.Error(err)
	_, err = service.Validate(token2)
	r.Error(err)

	// Valid token still works
	_, err = service.Validate(token3)
	r.NoError(err)
}

func TestRefreshTokenMultipleUsers(t *testing.T) {
	r := require.New(t)
	service := newTestRefreshTokenService(t)

	// Create tokens for multiple users
	users := []string{"alice", "bob", "charlie"}
	tokensByUser := make(map[string][]string)

	for _, user := range users {
		for i := 0; i < 3; i++ {
			tokenID, err := service.Create(user)
			r.NoError(err)
			tokensByUser[user] = append(tokensByUser[user], tokenID)
		}
	}

	// Verify each user has correct tokens
	for _, user := range users {
		tokens, err := service.GetTokensForUser(user)
		r.NoError(err)
		r.Len(tokens, 3)

		// Verify each token validates to correct user
		for _, token := range tokens {
			validatedUser, err := service.Validate(token.ID)
			r.NoError(err)
			r.Equal(user, validatedUser)
		}
	}

	// Delete all tokens for bob
	err := service.DeleteAllForUser("bob")
	r.NoError(err)

	// Alice and charlie still have tokens
	tokens, _ := service.GetTokensForUser("alice")
	r.Len(tokens, 3)
	tokens, _ = service.GetTokensForUser("charlie")
	r.Len(tokens, 3)

	// Bob has no tokens
	tokens, _ = service.GetTokensForUser("bob")
	r.Len(tokens, 0)
}
