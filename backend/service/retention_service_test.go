package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tfabritius/plainpage/model"
)

func TestCleanupTrash_Disabled(t *testing.T) {
	rs := &RetentionService{}

	// maxAgeDays = 0 means disabled
	policy := model.TrashRetention{MaxAgeDays: 0}
	deleted, err := rs.CleanupTrash(policy)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if deleted != 0 {
		t.Errorf("expected 0 deleted, got %d", deleted)
	}
}

func TestCleanupTrash_DeletesOldEntries(t *testing.T) {
	r := require.New(t)
	mock := newMockStorage()
	contentService := NewContentService(mock)
	retentionService := NewRetentionService(contentService, mock)

	// Create pages
	err := contentService.SavePage("old-page", "Content", model.ContentMeta{Title: "Old Page"}, "")
	r.NoError(err)
	err = contentService.SavePage("recent-page", "Content", model.ContentMeta{Title: "Recent Page"}, "")
	r.NoError(err)

	// Delete pages at different times
	oldDeleteTime := time.Now().Add(-10 * 24 * time.Hour)   // 10 days ago
	recentDeleteTime := time.Now().Add(-3 * 24 * time.Hour) // 3 days ago

	err = contentService.deletePageAt("old-page", oldDeleteTime)
	r.NoError(err)
	err = contentService.deletePageAt("recent-page", recentDeleteTime)
	r.NoError(err)

	// Verify both entries are in trash
	entries, err := contentService.ListTrash()
	r.NoError(err)
	r.Len(entries, 2)

	// Cleanup with 7-day retention (should delete old-page)
	deleted, err := retentionService.CleanupTrash(model.TrashRetention{MaxAgeDays: 7})
	r.NoError(err)
	r.Equal(1, deleted)

	// Verify only recent entry remains
	entries, err = contentService.ListTrash()
	r.NoError(err)
	r.Len(entries, 1)
	r.Equal("recent-page", entries[0].Url)
}

func TestCleanupTrash_KeepsAllWhenNoneExpired(t *testing.T) {
	r := require.New(t)
	mock := newMockStorage()
	contentService := NewContentService(mock)
	retentionService := NewRetentionService(contentService, mock)

	// Create and delete pages recently
	err := contentService.SavePage("page1", "Content", model.ContentMeta{Title: "Page 1"}, "")
	r.NoError(err)
	err = contentService.SavePage("page2", "Content", model.ContentMeta{Title: "Page 2"}, "")
	r.NoError(err)

	recentDeleteTime := time.Now().Add(-2 * 24 * time.Hour) // 2 days ago
	err = contentService.deletePageAt("page1", recentDeleteTime)
	r.NoError(err)
	err = contentService.deletePageAt("page2", recentDeleteTime)
	r.NoError(err)

	// Verify both entries are in trash
	entries, err := contentService.ListTrash()
	r.NoError(err)
	r.Len(entries, 2)

	// Cleanup with 7-day retention (nothing should be deleted)
	deleted, err := retentionService.CleanupTrash(model.TrashRetention{MaxAgeDays: 7})
	r.NoError(err)
	r.Equal(0, deleted)

	// Verify both entries remain
	entries, err = contentService.ListTrash()
	r.NoError(err)
	r.Len(entries, 2)
}

func TestCleanupAttic_Disabled(t *testing.T) {
	rs := &RetentionService{}

	// Both maxAgeDays and maxVersions = 0 means disabled
	policy := model.AtticRetention{MaxAgeDays: 0, MaxVersions: 0}
	deleted, err := rs.CleanupAttic(policy)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if deleted != 0 {
		t.Errorf("expected 0 deleted, got %d", deleted)
	}
}

func TestCleanupAttic_DeletesByAge(t *testing.T) {
	r := require.New(t)
	mock := newMockStorage()
	contentService := NewContentService(mock)
	retentionService := NewRetentionService(contentService, mock)

	// Create a page with versions at different times
	oldRevisionTime := time.Now().Add(-15 * 24 * time.Hour)   // 15 days ago
	recentRevisionTime := time.Now().Add(-3 * 24 * time.Hour) // 3 days ago

	err := contentService.SavePageAt("testpage", "Content v1", model.ContentMeta{Title: "Test Page"}, "", oldRevisionTime)
	r.NoError(err)
	err = contentService.SavePageAt("testpage", "Content v2", model.ContentMeta{Title: "Test Page"}, "", recentRevisionTime)
	r.NoError(err)

	// Verify both attic entries exist
	entries, err := contentService.ListAttic("testpage")
	r.NoError(err)
	r.Len(entries, 2)

	// Cleanup with 7-day age retention (should delete old entry)
	deleted, err := retentionService.CleanupAttic(model.AtticRetention{MaxAgeDays: 7, MaxVersions: 0})
	r.NoError(err)
	r.Equal(1, deleted)

	// Verify only recent entry remains
	entries, err = contentService.ListAttic("testpage")
	r.NoError(err)
	r.Len(entries, 1)
	r.Equal(recentRevisionTime.Unix(), entries[0].Revision)
}

func TestCleanupAttic_DeletesByCount(t *testing.T) {
	r := require.New(t)
	mock := newMockStorage()
	contentService := NewContentService(mock)
	retentionService := NewRetentionService(contentService, mock)

	now := time.Now()

	// Create a page with 5 versions (all recent, so age won't delete them)
	for i := 1; i <= 5; i++ {
		revisionTime := now.Add(-time.Duration(i) * time.Minute) // 1-5 minutes ago
		err := contentService.SavePageAt("testpage", "Content v"+string(rune('0'+i)), model.ContentMeta{Title: "Test Page"}, "", revisionTime)
		r.NoError(err)
	}

	// Verify all 5 attic entries exist
	entries, err := contentService.ListAttic("testpage")
	r.NoError(err)
	r.Len(entries, 5)

	// Cleanup keeping only 2 versions
	deleted, err := retentionService.CleanupAttic(model.AtticRetention{MaxAgeDays: 0, MaxVersions: 2})
	r.NoError(err)
	r.Equal(3, deleted)

	// Verify only 2 newest entries remain
	entries, err = contentService.ListAttic("testpage")
	r.NoError(err)
	r.Len(entries, 2)
}

func TestCleanupAttic_DeletesByAgeAndCount(t *testing.T) {
	r := require.New(t)
	mock := newMockStorage()
	contentService := NewContentService(mock)
	retentionService := NewRetentionService(contentService, mock)

	now := time.Now()
	oldRevisionTime := now.Add(-15 * 24 * time.Hour) // 15 days ago

	// Create 1 old entry (will be deleted by age)
	err := contentService.SavePageAt("testpage", "Old content", model.ContentMeta{Title: "Test Page"}, "", oldRevisionTime)
	r.NoError(err)

	// Create 4 recent entries
	for i := 1; i <= 4; i++ {
		revisionTime := now.Add(-time.Duration(i) * time.Minute) // 1-4 minutes ago
		err := contentService.SavePageAt("testpage", "Content v"+string(rune('0'+i)), model.ContentMeta{Title: "Test Page"}, "", revisionTime)
		r.NoError(err)
	}

	// Verify all 5 attic entries exist
	entries, err := contentService.ListAttic("testpage")
	r.NoError(err)
	r.Len(entries, 5)

	// Cleanup with 7-day age AND keep only 2 versions
	// Step 1: Delete old entry (1 deleted by age)
	// Step 2: 4 remain, delete 2 oldest (2 deleted by count)
	// Total: 3 deleted
	deleted, err := retentionService.CleanupAttic(model.AtticRetention{MaxAgeDays: 7, MaxVersions: 2})
	r.NoError(err)
	r.Equal(3, deleted)

	// Verify only 2 entries remain
	entries, err = contentService.ListAttic("testpage")
	r.NoError(err)
	r.Len(entries, 2)
}

func TestCleanupAttic_MultiplePages(t *testing.T) {
	r := require.New(t)
	mock := newMockStorage()
	contentService := NewContentService(mock)
	retentionService := NewRetentionService(contentService, mock)

	now := time.Now()

	// Create page1 with 4 versions
	for i := 1; i <= 4; i++ {
		revisionTime := now.Add(-time.Duration(i) * time.Minute)
		err := contentService.SavePageAt("page1", "Content", model.ContentMeta{Title: "Page 1"}, "", revisionTime)
		r.NoError(err)
	}

	// Create page2 with 3 versions
	for i := 1; i <= 3; i++ {
		revisionTime := now.Add(-time.Duration(i) * time.Minute)
		err := contentService.SavePageAt("page2", "Content", model.ContentMeta{Title: "Page 2"}, "", revisionTime)
		r.NoError(err)
	}

	// Verify attic entries
	entries1, _ := contentService.ListAttic("page1")
	entries2, _ := contentService.ListAttic("page2")
	r.Len(entries1, 4)
	r.Len(entries2, 3)

	// Cleanup keeping only 2 versions
	deleted, err := retentionService.CleanupAttic(model.AtticRetention{MaxAgeDays: 0, MaxVersions: 2})
	r.NoError(err)
	r.Equal(3, deleted) // 2 from page1, 1 from page2

	// Verify 2 entries remain for each page
	entries1, _ = contentService.ListAttic("page1")
	entries2, _ = contentService.ListAttic("page2")
	r.Len(entries1, 2)
	r.Len(entries2, 2)
}

func TestCleanupAttic_KeepsAllWhenBelowMaxVersions(t *testing.T) {
	r := require.New(t)
	mock := newMockStorage()
	contentService := NewContentService(mock)
	retentionService := NewRetentionService(contentService, mock)

	now := time.Now()

	// Create a page with 2 versions
	for i := 1; i <= 2; i++ {
		revisionTime := now.Add(-time.Duration(i) * time.Minute)
		err := contentService.SavePageAt("testpage", "Content", model.ContentMeta{Title: "Test Page"}, "", revisionTime)
		r.NoError(err)
	}

	// Verify 2 attic entries exist
	entries, err := contentService.ListAttic("testpage")
	r.NoError(err)
	r.Len(entries, 2)

	// Cleanup keeping up to 5 versions (nothing should be deleted)
	deleted, err := retentionService.CleanupAttic(model.AtticRetention{MaxAgeDays: 0, MaxVersions: 5})
	r.NoError(err)
	r.Equal(0, deleted)

	// Verify both entries remain
	entries, err = contentService.ListAttic("testpage")
	r.NoError(err)
	r.Len(entries, 2)
}

func TestCleanupAttic_PreservesMostRecentVersion(t *testing.T) {
	r := require.New(t)
	mock := newMockStorage()
	contentService := NewContentService(mock)
	retentionService := NewRetentionService(contentService, mock)

	// Create a page with all versions older than maxAgeDays
	// Both versions are old enough to be deleted by age policy
	veryOldRevisionTime := time.Now().Add(-30 * 24 * time.Hour) // 30 days ago
	oldRevisionTime := time.Now().Add(-15 * 24 * time.Hour)     // 15 days ago

	err := contentService.SavePageAt("testpage", "Content v1", model.ContentMeta{Title: "Test Page"}, "", veryOldRevisionTime)
	r.NoError(err)
	err = contentService.SavePageAt("testpage", "Content v2", model.ContentMeta{Title: "Test Page"}, "", oldRevisionTime)
	r.NoError(err)

	// Verify both attic entries exist
	entries, err := contentService.ListAttic("testpage")
	r.NoError(err)
	r.Len(entries, 2)

	// Cleanup with 7-day age retention
	deleted, err := retentionService.CleanupAttic(model.AtticRetention{MaxAgeDays: 7, MaxVersions: 0})
	r.NoError(err)
	r.Equal(1, deleted) // Only the oldest one should be deleted

	// Verify the most recent entry is still there
	entries, err = contentService.ListAttic("testpage")
	r.NoError(err)
	r.Len(entries, 1)
	r.Equal(oldRevisionTime.Unix(), entries[0].Revision) // The newest (but still old) version is preserved
}

func TestCleanupAttic_PreservesOnlyVersion(t *testing.T) {
	r := require.New(t)
	mock := newMockStorage()
	contentService := NewContentService(mock)
	retentionService := NewRetentionService(contentService, mock)

	// Create a page with only one version, and it's old
	oldRevisionTime := time.Now().Add(-30 * 24 * time.Hour) // 30 days ago

	err := contentService.SavePageAt("testpage", "Content", model.ContentMeta{Title: "Test Page"}, "", oldRevisionTime)
	r.NoError(err)

	// Verify one attic entry exists
	entries, err := contentService.ListAttic("testpage")
	r.NoError(err)
	r.Len(entries, 1)

	// Cleanup with 7-day age retention
	// The only version is older than 7 days, but it must be preserved
	deleted, err := retentionService.CleanupAttic(model.AtticRetention{MaxAgeDays: 7, MaxVersions: 0})
	r.NoError(err)
	r.Equal(0, deleted) // Nothing should be deleted - can't delete the only version

	// Verify the entry is still there
	entries, err = contentService.ListAttic("testpage")
	r.NoError(err)
	r.Len(entries, 1)
}
