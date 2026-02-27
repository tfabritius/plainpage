package service

import (
	"context"
	"log"
	"time"

	"github.com/tfabritius/plainpage/model"
)

// RetentionService handles automatic cleanup of trash and attic based on retention policies
type RetentionService struct {
	content *ContentService
	storage model.Storage
}

// NewRetentionService creates a new retention service
func NewRetentionService(content *ContentService, storage model.Storage) *RetentionService {
	return &RetentionService{
		content: content,
		storage: storage,
	}
}

// CleanupTrash removes trash items older than the configured maxAgeDays
// Returns the number of deleted items and any error encountered.
func (s *RetentionService) CleanupTrash(policy model.TrashRetention) (int, error) {
	if policy.MaxAgeDays <= 0 {
		return 0, nil // Disabled
	}

	cutoff := time.Now().Add(-time.Duration(policy.MaxAgeDays) * 24 * time.Hour).Unix()
	deleted := 0

	entries, err := s.content.ListTrash()
	if err != nil {
		return 0, err
	}

	for _, entry := range entries {
		if entry.DeletedAt < cutoff {
			if err := s.content.DeleteTrashEntry(entry.Url, entry.DeletedAt); err != nil {
				log.Printf("[retention] Failed to delete trash entry %s (deleted at %d): %v", entry.Url, entry.DeletedAt, err)
				continue
			}
			log.Printf("[retention] Deleted trash entry %s (deleted at %d)", entry.Url, entry.DeletedAt)
			deleted++
		}
	}

	return deleted, nil
}

// CleanupAttic removes old versions based on the configured retention policy.
// If both maxAgeDays and maxVersions are set, a version is deleted if either condition is met.
// Returns the number of deleted versions and any error encountered.
func (s *RetentionService) CleanupAttic(policy model.AtticRetention) (int, error) {
	if policy.MaxAgeDays <= 0 && policy.MaxVersions <= 0 {
		return 0, nil // Disabled
	}

	deleted := 0

	pages, err := s.content.ListAllPages()
	if err != nil {
		return 0, err
	}

	for _, pageUrl := range pages {
		n, err := s.cleanupPageAttic(pageUrl, policy)
		if err != nil {
			log.Printf("[retention] Failed to cleanup attic for %s: %v", pageUrl, err)
			continue
		}
		deleted += n
	}

	return deleted, nil
}

// cleanupPageAttic cleans up attic entries for a single page.
func (s *RetentionService) cleanupPageAttic(pageUrl string, policy model.AtticRetention) (int, error) {
	// ListAttic returns entries sorted oldest to newest (by revision/timestamp)
	entries, err := s.content.ListAttic(pageUrl)
	if err != nil {
		return 0, err
	}

	// Always keep at least the most recent version (which is a copy of the current version)
	if len(entries) <= 1 {
		return 0, nil
	}

	deleted := 0

	// Step 1: Delete versions older than maxAgeDays (but not the most recent one)
	if policy.MaxAgeDays > 0 {
		cutoff := time.Now().Add(-time.Duration(policy.MaxAgeDays) * 24 * time.Hour).Unix()

		remaining := []model.AtticEntry{}
		for i, entry := range entries {
			isNewest := i == len(entries)-1 // Last entry is the newest
			if entry.Revision < cutoff && !isNewest {
				if err := s.content.DeleteAtticEntry(pageUrl, entry.Revision); err != nil {
					log.Printf("[retention] Failed to delete attic entry %s rev %d: %v", pageUrl, entry.Revision, err)
					remaining = append(remaining, entry) // Keep in list if deletion failed
				} else {
					log.Printf("[retention] Deleted attic entry %s rev %d (too old)", pageUrl, entry.Revision)
					deleted++
				}
			} else {
				remaining = append(remaining, entry)
			}
		}
		entries = remaining
	}

	// Step 2: Delete oldest versions if count exceeds maxVersions
	if policy.MaxVersions > 0 && len(entries) > policy.MaxVersions {
		// Entries are sorted oldest to newest, delete from the beginning
		excessCount := len(entries) - policy.MaxVersions
		for i := 0; i < excessCount; i++ {
			if err := s.content.DeleteAtticEntry(pageUrl, entries[i].Revision); err != nil {
				log.Printf("[retention] Failed to delete attic entry %s rev %d: %v", pageUrl, entries[i].Revision, err)
			} else {
				log.Printf("[retention] Deleted attic entry %s rev %d (exceeds max versions)", pageUrl, entries[i].Revision)
				deleted++
			}
		}
	}

	return deleted, nil
}

// Cleanup runs both trash and attic cleanup based on current configuration
func (s *RetentionService) Cleanup() error {
	cfg, err := s.storage.ReadConfig()
	if err != nil {
		return err
	}

	trashDeleted, err := s.CleanupTrash(cfg.Retention.Trash)
	if err != nil {
		log.Printf("[retention] Trash cleanup error: %v", err)
	} else if trashDeleted > 0 {
		log.Printf("[retention] Trash cleanup: deleted %d items", trashDeleted)
	}

	atticDeleted, err := s.CleanupAttic(cfg.Retention.Attic)
	if err != nil {
		log.Printf("[retention] Attic cleanup error: %v", err)
	} else if atticDeleted > 0 {
		log.Printf("[retention] Attic cleanup: deleted %d versions", atticDeleted)
	}

	return nil
}

// StartCleanupScheduler starts a background goroutine that periodically runs cleanup
func (s *RetentionService) StartCleanupScheduler(ctx context.Context, interval time.Duration) {
	// Run cleanup immediately at startup
	if err := s.Cleanup(); err != nil {
		log.Printf("[retention] Initial cleanup error: %v", err)
	} else {
		log.Println("[retention] Initial cleanup completed")
	}

	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				log.Println("[retention] Cleanup scheduler stopped")
				return
			case <-ticker.C:
				if err := s.Cleanup(); err != nil {
					log.Printf("[retention] Scheduled cleanup error: %v", err)
				} else {
					log.Println("[retention] Scheduled cleanup completed")
				}
			}
		}
	}()
}
