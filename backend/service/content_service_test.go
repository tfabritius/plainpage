package service

import (
	"archive/zip"
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tfabritius/plainpage/model"
)

// TestDeleteAtticEntry tests the DeleteAtticEntry method
func TestDeleteAtticEntry(t *testing.T) {
	r := require.New(t)
	mock := newMockStorage()
	contentService := NewContentService(mock)

	// Create a page with first version
	t1 := time.Now()
	err := contentService.SavePageAt("testpage", "Content v1", model.ContentMeta{Title: "Test Page"}, "", t1)
	r.NoError(err)

	// Save again with a different timestamp to create another version
	t2 := t1.Add(time.Hour)
	err = contentService.SavePageAt("testpage", "Content v2", model.ContentMeta{Title: "Test Page"}, "", t2)
	r.NoError(err)

	// Verify we have 2 attic entries
	entries, err := contentService.ListAttic("testpage")
	r.NoError(err)
	r.Len(entries, 2)

	rev1 := entries[0].Revision
	rev2 := entries[1].Revision

	// Delete the first attic entry
	err = contentService.DeleteAtticEntry("testpage", rev1)
	r.NoError(err)

	// Verify only 1 entry remains
	entries, err = contentService.ListAttic("testpage")
	r.NoError(err)
	r.Len(entries, 1)
	r.Equal(rev2, entries[0].Revision)

	// Verify reading the deleted revision fails
	r.False(contentService.IsAtticPage("testpage", rev1))

	// Verify reading the remaining revision works
	r.True(contentService.IsAtticPage("testpage", rev2))
}

// TestDeleteAtticEntry_NotFound tests deleting a non-existent attic entry
func TestDeleteAtticEntry_NotFound(t *testing.T) {
	r := require.New(t)
	mock := newMockStorage()
	contentService := NewContentService(mock)

	// Try to delete attic entry for non-existent page
	err := contentService.DeleteAtticEntry("nonexistent", 12345)
	r.Error(err)
	r.ErrorIs(err, model.ErrNotFound)

	// Create a page
	err = contentService.SavePage("testpage", "Content", model.ContentMeta{Title: "Test Page"}, "")
	r.NoError(err)

	// Try to delete attic entry with wrong revision
	err = contentService.DeleteAtticEntry("testpage", 99999)
	r.Error(err)
	r.ErrorIs(err, model.ErrNotFound)
}

// TestWriteBackup_ContentOnly tests backup of content directories without config/users
func TestWriteBackup_ContentOnly(t *testing.T) {
	r := require.New(t)
	mock := newMockStorage()
	contentService := NewContentService(mock)

	// Create some content
	err := contentService.SavePage("page1", "Content 1", model.ContentMeta{Title: "Page 1"}, "")
	r.NoError(err)

	err = contentService.CreateFolder("folder", model.ContentMeta{Title: "Folder"})
	r.NoError(err)

	err = contentService.SavePage("folder/page2", "Content 2", model.ContentMeta{Title: "Page 2"}, "")
	r.NoError(err)

	// Write backup to buffer
	var buf bytes.Buffer
	err = contentService.WriteBackup(&buf, BackupOptions{
		IncludeConfig: false,
		IncludeUsers:  false,
	})
	r.NoError(err)

	// Read and verify the backup ZIP
	zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	r.NoError(err)

	// Check that expected files exist in the backup
	fileNames := make(map[string]bool)
	for _, f := range zipReader.File {
		fileNames[f.Name] = true
	}

	// Should have pages
	r.True(fileNames["pages/_index.md"], "should have root index")
	r.True(fileNames["pages/page1.md"], "should have page1")
	r.True(fileNames["pages/folder/_index.md"], "should have folder index")
	r.True(fileNames["pages/folder/page2.md"], "should have page2")

	// Should have attic entries
	r.True(len(getFilesWithPrefix(fileNames, "attic/")) > 0, "should have attic entries")

	// Should NOT have config or users
	r.False(fileNames["config.yml"], "should not have config")
	r.False(fileNames["users.yml"], "should not have users")
}

// TestWriteBackup_WithConfigAndUsers tests backup including config and users
func TestWriteBackup_WithConfigAndUsers(t *testing.T) {
	r := require.New(t)
	mock := newMockStorage().(*mockStorage)

	// Setup config
	err := mock.WriteConfig(model.Config{
		AppTitle:  "Test Wiki",
		JwtSecret: "secret123",
	})
	r.NoError(err)

	// Setup users
	mock.files["users.yml"] = []byte("- id: user1\n  username: testuser\n")

	contentService := NewContentService(mock)

	// Create a page
	err = contentService.SavePage("testpage", "Content", model.ContentMeta{Title: "Test"}, "")
	r.NoError(err)

	// Write backup with config and users
	var buf bytes.Buffer
	err = contentService.WriteBackup(&buf, BackupOptions{
		IncludeConfig: true,
		IncludeUsers:  true,
	})
	r.NoError(err)

	// Read and verify
	zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	r.NoError(err)

	fileNames := make(map[string]bool)
	for _, f := range zipReader.File {
		fileNames[f.Name] = true
	}

	r.True(fileNames["config.yml"], "should have config")
	r.True(fileNames["users.yml"], "should have users")

	// Verify config has JWT secret stripped
	for _, f := range zipReader.File {
		if f.Name == "config.yml" {
			rc, err := f.Open()
			r.NoError(err)
			var configContent bytes.Buffer
			_, err = configContent.ReadFrom(rc)
			rc.Close()
			r.NoError(err)
			r.NotContains(configContent.String(), "secret123", "JWT secret should be stripped from backup")
		}
	}
}

// TestRestoreBackup_ContentOnly tests restoring content from backup
func TestRestoreBackup_ContentOnly(t *testing.T) {
	r := require.New(t)

	// Create source storage with content
	srcMock := newMockStorage()
	srcService := NewContentService(srcMock)

	err := srcService.SavePage("original-page", "Original Content", model.ContentMeta{Title: "Original"}, "")
	r.NoError(err)

	err = srcService.CreateFolder("original-folder", model.ContentMeta{Title: "Original Folder"})
	r.NoError(err)

	// Create backup
	var buf bytes.Buffer
	err = srcService.WriteBackup(&buf, BackupOptions{})
	r.NoError(err)

	// Create destination storage with different content
	dstMock := newMockStorage()
	dstService := NewContentService(dstMock)

	err = dstService.SavePage("other-page", "Other Content", model.ContentMeta{Title: "Other"}, "")
	r.NoError(err)

	// Restore backup
	zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	r.NoError(err)

	usersRestored, err := dstService.RestoreBackup(zipReader)
	r.NoError(err)
	r.False(usersRestored, "users were not in backup")

	// Verify original content was restored
	r.True(dstService.IsPage("original-page"), "original-page should exist")
	r.True(dstService.IsFolder("original-folder"), "original-folder should exist")

	// Verify other-page was removed (overwritten)
	r.False(dstService.IsPage("other-page"), "other-page should be gone")

	// Verify content is correct
	page, err := dstService.ReadPage("original-page", nil)
	r.NoError(err)
	r.Equal("Original Content", page.Content)
	r.Equal("Original", page.Meta.Title)
}

// TestRestoreBackup_WithUsers tests that restoring users returns usersRestored=true
func TestRestoreBackup_WithUsers(t *testing.T) {
	r := require.New(t)

	// Create source storage with users
	srcMock := newMockStorage().(*mockStorage)
	srcMock.files["users.yml"] = []byte("- id: user1\n  username: testuser\n")
	err := srcMock.WriteConfig(model.Config{AppTitle: "Test", JwtSecret: "old-secret"})
	r.NoError(err)

	srcService := NewContentService(srcMock)

	// Create backup with users
	var buf bytes.Buffer
	err = srcService.WriteBackup(&buf, BackupOptions{
		IncludeConfig: true,
		IncludeUsers:  true,
	})
	r.NoError(err)

	// Create destination storage
	dstMock := newMockStorage().(*mockStorage)
	err = dstMock.WriteConfig(model.Config{AppTitle: "Dest", JwtSecret: "dest-secret"})
	r.NoError(err)

	dstService := NewContentService(dstMock)

	// Restore backup
	zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	r.NoError(err)

	usersRestored, err := dstService.RestoreBackup(zipReader)
	r.NoError(err)
	r.True(usersRestored, "users were in backup")

	// Verify users.yml was restored
	usersData, err := dstMock.ReadFile("users.yml")
	r.NoError(err)
	r.Contains(string(usersData), "testuser")

	// Verify JWT secret was regenerated (not the old one from backup or the dest one)
	config, err := dstMock.ReadConfig()
	r.NoError(err)
	r.NotEqual("old-secret", config.JwtSecret, "should not use backup's JWT secret")
	r.NotEqual("dest-secret", config.JwtSecret, "should not keep dest's JWT secret")
	r.NotEmpty(config.JwtSecret, "should have a new JWT secret")
}

// TestRestoreBackup_InvalidZip tests error handling for invalid ZIP
func TestRestoreBackup_InvalidZip(t *testing.T) {
	r := require.New(t)
	mock := newMockStorage()
	contentService := NewContentService(mock)

	// Create an invalid/empty ZIP
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)
	zipWriter.Close()

	zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	r.NoError(err)

	// Should succeed but not restore anything meaningful
	_, err = contentService.RestoreBackup(zipReader)
	r.NoError(err)
}

// getFilesWithPrefix returns files from a map that have the given prefix
func getFilesWithPrefix(files map[string]bool, prefix string) []string {
	var result []string
	for name := range files {
		if len(name) >= len(prefix) && name[:len(prefix)] == prefix {
			result = append(result, name)
		}
	}
	return result
}

// TestListAllPages tests the ListAllPages method
func TestListAllPages(t *testing.T) {
	r := require.New(t)
	mock := newMockStorage()
	contentService := NewContentService(mock)

	// Initially no pages (only root folder exists)
	pages, err := contentService.ListAllPages()
	r.NoError(err)
	r.Empty(pages)

	// Create some pages at root level
	err = contentService.SavePage("page1", "Content 1", model.ContentMeta{Title: "Page 1"}, "")
	r.NoError(err)
	err = contentService.SavePage("page2", "Content 2", model.ContentMeta{Title: "Page 2"}, "")
	r.NoError(err)

	pages, err = contentService.ListAllPages()
	r.NoError(err)
	r.Len(pages, 2)
	r.Contains(pages, "page1")
	r.Contains(pages, "page2")

	// Create a folder with pages
	err = contentService.CreateFolder("folder", model.ContentMeta{Title: "Folder"})
	r.NoError(err)
	err = contentService.SavePage("folder/page3", "Content 3", model.ContentMeta{Title: "Page 3"}, "")
	r.NoError(err)

	pages, err = contentService.ListAllPages()
	r.NoError(err)
	r.Len(pages, 3)
	r.Contains(pages, "page1")
	r.Contains(pages, "page2")
	r.Contains(pages, "folder/page3")

	// Create nested folder with pages
	err = contentService.CreateFolder("folder/subfolder", model.ContentMeta{Title: "Subfolder"})
	r.NoError(err)
	err = contentService.SavePage("folder/subfolder/page4", "Content 4", model.ContentMeta{Title: "Page 4"}, "")
	r.NoError(err)

	pages, err = contentService.ListAllPages()
	r.NoError(err)
	r.Len(pages, 4)
	r.Contains(pages, "page1")
	r.Contains(pages, "page2")
	r.Contains(pages, "folder/page3")
	r.Contains(pages, "folder/subfolder/page4")
}
