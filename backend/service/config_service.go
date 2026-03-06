package service

import (
	"fmt"
	"log"
	"sync"

	"github.com/tfabritius/plainpage/libs/utils"
	"github.com/tfabritius/plainpage/model"
	"gopkg.in/yaml.v3"
)

// ConfigService handles all config-related operations
type ConfigService struct {
	storage        model.Storage
	mu             sync.RWMutex
	jwtSecretCache string // Cached JWT secret
}

// NewConfigService creates a new ConfigService and initializes config if needed
func NewConfigService(storage model.Storage) *ConfigService {
	s := &ConfigService{
		storage: storage,
	}

	// Initialize config.yml if it doesn't exist
	if !storage.Exists("config.yml") {
		log.Println("Initializing config...")
		if err := s.initializeConfig(); err != nil {
			log.Fatalln("Could not initialize config:", err)
		}
	}

	// Load and cache the JWT secret
	cfg, err := s.readUnlocked()
	if err != nil {
		log.Fatalln("Could not read config:", err)
	}
	s.jwtSecretCache = cfg.JwtSecret

	return s
}

// initializeConfig creates the default config on first run
func (s *ConfigService) initializeConfig() error {
	jwtSecret, err := utils.GenerateRandomString(16)
	if err != nil {
		return fmt.Errorf("could not generate JWT secret: %w", err)
	}

	cfg := model.Config{
		AppTitle:  "PlainPage",
		JwtSecret: jwtSecret,
		SetupMode: true,
	}

	return s.writeUnlocked(cfg)
}

// Read returns the current config
func (s *ConfigService) Read() (model.Config, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.readUnlocked()
}

func (s *ConfigService) readUnlocked() (model.Config, error) {
	bytes, err := s.storage.ReadFile("config.yml")
	if err != nil {
		return model.Config{}, fmt.Errorf("could not read config.yml: %w", err)
	}

	var config model.Config
	if err := yaml.Unmarshal(bytes, &config); err != nil {
		return model.Config{}, fmt.Errorf("could not parse config YAML: %w", err)
	}

	return config, nil
}

// Write saves the config
func (s *ConfigService) Write(config model.Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.writeUnlocked(config)
}

func (s *ConfigService) writeUnlocked(config model.Config) error {
	bytes, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("could not marshal config: %w", err)
	}

	if err := s.storage.WriteFile("config.yml", bytes); err != nil {
		return fmt.Errorf("could not write config.yml: %w", err)
	}

	// Update the cache
	s.jwtSecretCache = config.JwtSecret

	return nil
}

// GetJwtSecret returns the cached JWT secret for token signing
func (s *ConfigService) GetJwtSecret() []byte {
	return []byte(s.jwtSecretCache)
}

// RegenerateJwtSecret generates and saves a new JWT secret.
// This invalidates all existing sessions/tokens.
func (s *ConfigService) RegenerateJwtSecret() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.readUnlocked()
	if err != nil {
		return err
	}

	newSecret, err := utils.GenerateRandomString(16)
	if err != nil {
		return fmt.Errorf("could not generate new JWT secret: %w", err)
	}

	cfg.JwtSecret = newSecret
	return s.writeUnlocked(cfg)
}

// ExportForBackup returns config YAML bytes with sensitive data stripped
func (s *ConfigService) ExportForBackup() ([]byte, error) {
	cfg, err := s.Read()
	if err != nil {
		return nil, err
	}

	// Strip sensitive data
	cfg.JwtSecret = ""

	return yaml.Marshal(&cfg)
}

// RestoreFromBackup restores config from backup bytes.
// If regenerateSecret is true, a new JWT secret is generated (use when users.yml is also restored).
// Otherwise, the existing JWT secret is preserved.
func (s *ConfigService) RestoreFromBackup(content []byte, regenerateSecret bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Parse the backup config
	var newConfig model.Config
	if err := yaml.Unmarshal(content, &newConfig); err != nil {
		return fmt.Errorf("could not parse config: %w", err)
	}

	// Handle JWT secret
	if regenerateSecret {
		// Generate new JWT secret (invalidates all sessions)
		newSecret, err := utils.GenerateRandomString(16)
		if err != nil {
			return fmt.Errorf("could not generate JWT secret: %w", err)
		}
		newConfig.JwtSecret = newSecret
	} else {
		// Keep existing JWT secret
		existingConfig, err := s.readUnlocked()
		if err != nil {
			// If no existing config, generate new secret
			newSecret, err := utils.GenerateRandomString(16)
			if err != nil {
				return fmt.Errorf("could not generate JWT secret: %w", err)
			}
			newConfig.JwtSecret = newSecret
		} else {
			newConfig.JwtSecret = existingConfig.JwtSecret
		}
	}

	return s.writeUnlocked(newConfig)
}

// EndSetupMode transitions out of setup mode, granting admin to the first user
func (s *ConfigService) EndSetupMode(firstUserID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.readUnlocked()
	if err != nil {
		return err
	}

	if !cfg.SetupMode {
		return nil // Already out of setup mode
	}

	cfg.SetupMode = false
	cfg.ACL = append(cfg.ACL, model.AccessRule{
		Subject:    "user:" + firstUserID,
		Operations: []model.AccessOp{model.AccessOpAdmin},
	})

	return s.writeUnlocked(cfg)
}

// IsSetupMode returns whether the app is in setup mode
func (s *ConfigService) IsSetupMode() (bool, error) {
	cfg, err := s.Read()
	if err != nil {
		return false, err
	}
	return cfg.SetupMode, nil
}
