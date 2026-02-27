package service

import (
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

	// Create a page
	err := contentService.SavePage("testpage", "Content v1", model.ContentMeta{Title: "Test Page"}, "")
	r.NoError(err)

	// Wait and save again to create another version
	time.Sleep(1100 * time.Millisecond)
	err = contentService.SavePage("testpage", "Content v2", model.ContentMeta{Title: "Test Page"}, "")
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
