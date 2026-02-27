package service

import (
	"fmt"
	"log"
	"net/url"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/tfabritius/plainpage/model"
)

func NewContentService(store model.Storage) *ContentService {
	s := ContentService{
		storage: store,
	}

	if err := s.initializeStorage(); err != nil {
		log.Fatalln("Could not initialize storage:", err)
	}

	if err := s.RecreateIndex(); err != nil {
		log.Fatalln("Could not initialize search index:", err)
	}

	return &s
}

type ContentService struct {
	storage model.Storage
	index   bleve.Index
}

func (s *ContentService) initializeStorage() error {
	// Create pages, attic, and trash directories
	for _, dir := range []string{"pages", "attic", "trash"} {
		// Create directory, if it doesn't exist
		if !s.storage.Exists(dir) {
			if err := s.storage.CreateDirectory(dir); err != nil {
				return fmt.Errorf("could not create %s directory: %w", dir, err)
			}
		}
	}

	// Create _index.md with default ACL if it doesn't exist
	if !s.IsFolder("") {
		defaultACL := []model.AccessRule{
			{Subject: "all", Operations: []model.AccessOp{model.AccessOpRead, model.AccessOpWrite, model.AccessOpDelete}},
		}
		if err := s.SaveFolder("", model.ContentMeta{ACL: &defaultACL}); err != nil {
			return fmt.Errorf("could not create default ACL: %w", err)
		}
	}

	return nil
}

func (s *ContentService) RecreateIndex() error {
	// Create new in-memory index
	idx, err := bleve.NewMemOnly(s.createIndexMapping())
	if err != nil {
		return err
	}

	// (Re-)Index all documents
	log.Println("Creating search index...")
	if err := s.indexFolder("", &idx); err != nil {
		return err
	}
	cnt, err := idx.DocCount()
	if err != nil {
		return err
	}
	log.Printf("done (%v entries).", cnt)

	s.index = idx
	return nil
}

func (*ContentService) createIndexMapping() *mapping.IndexMappingImpl {
	metaMapping := bleve.NewDocumentMapping()
	metaMapping.AddSubDocumentMapping("acl", bleve.NewDocumentDisabledMapping())

	pageMapping := bleve.NewDocumentMapping()
	pageMapping.AddSubDocumentMapping("meta", metaMapping)
	pageMapping.AddSubDocumentMapping("url", bleve.NewDocumentDisabledMapping())

	folderMapping := bleve.NewDocumentMapping()
	folderMapping.AddSubDocumentMapping("meta", metaMapping)
	folderMapping.AddSubDocumentMapping("content", bleve.NewDocumentDisabledMapping())

	indexMapping := bleve.NewIndexMapping()
	indexMapping.TypeField = "BleveType"
	indexMapping.AddDocumentMapping("page", pageMapping)
	indexMapping.AddDocumentMapping("folder", folderMapping)
	return indexMapping
}

func (s *ContentService) indexFolder(urlPath string, idx *bleve.Index) error {
	if idx == nil {
		panic("index pointer is nil")
	}

	folder, err := s.ReadFolder(urlPath)
	if err != nil {
		return err
	}

	if urlPath != "" {
		// Index the folder itself
		if err := (*idx).Index(urlPath, folder); err != nil {
			return err
		}
	}

	for _, c := range folder.Content {
		if c.IsFolder {
			// Recursively index subfolder
			if err := s.indexFolder(c.Url, idx); err != nil {
				return err
			}
		} else {
			// Index page
			page, err := s.ReadPage(c.Url, nil)
			if err != nil {
				return err
			}
			if err := (*idx).Index(c.Url, page); err != nil {
				return err
			}
		}
	}

	return nil
}

// removeFolderFromIndex deletes index entries for a folder and all its contents
func (s *ContentService) removeFolderFromIndex(urlPath string) error {
	folder, err := s.ReadFolder(urlPath)
	if err != nil {
		return fmt.Errorf("could not read folder %s: %w", urlPath, err)
	}

	for _, entry := range folder.Content {
		if entry.IsFolder {
			if err := s.removeFolderFromIndex(entry.Url); err != nil {
				return err
			}
		} else {
			if err := s.index.Delete(entry.Url); err != nil {
				log.Printf("[INDEX] Could not delete page %s from index: %v", entry.Url, err)
			}
		}
	}

	// Delete the folder itself (if not root)
	if urlPath != "" {
		if err := s.index.Delete(urlPath); err != nil {
			log.Printf("[INDEX] Could not delete folder %s from index: %v", urlPath, err)
		}
	}

	return nil
}

// Search searches for content and returns all matching results (up to 10000).
func (s *ContentService) Search(q string) ([]model.SearchHit, error) {
	results, _, err := s.SearchWithPagination(q, 0, 10000)
	return results, err
}

// SearchWithPagination searches for content with pagination support.
// Returns the search hits, total number of hits in the index, and any error.
func (s *ContentService) SearchWithPagination(q string, offset, size int) ([]model.SearchHit, uint64, error) {
	query := bleve.NewMatchQuery(q)

	search := bleve.NewSearchRequest(query)
	search.Highlight = bleve.NewHighlight()
	search.From = offset
	search.Size = size

	results, err := s.index.Search(search)
	if err != nil {
		return nil, 0, err
	}

	ret := []model.SearchHit{}
	for _, r := range results.Hits {
		var meta model.ContentMeta
		isFolder := false

		if s.IsPage(r.ID) {
			page, err := s.ReadPage(r.ID, nil)
			if err != nil {
				return nil, 0, err
			}
			meta = page.Meta
		} else if s.IsFolder(r.ID) {
			isFolder = true
			folder, err := s.ReadFolder(r.ID)
			if err != nil {
				return nil, 0, err
			}
			meta = folder.Meta
		} else {
			continue
		}

		metas, err := s.ReadAncestorsMeta(r.ID)
		if err != nil {
			return nil, 0, err
		}
		acl := s.GetEffectivePermissions(meta, metas)

		ret = append(ret, model.SearchHit{
			Url:          r.ID,
			Meta:         meta,
			Fragments:    r.Fragments,
			EffectiveACL: acl,
			IsFolder:     isFolder,
		})
	}

	return ret, results.Total, nil
}

func (s *ContentService) IsPage(urlPath string) bool {
	fsPath := filepath.Join("pages", urlPath+".md")
	return s.storage.Exists(fsPath)
}

func (s *ContentService) IsFolder(urlPath string) bool {
	fsPath := filepath.Join("pages", urlPath, "_index.md")
	return s.storage.Exists(fsPath)
}

func (s *ContentService) IsAtticPage(urlPath string, revision int64) bool {
	revStr := strconv.FormatInt(revision, 10)
	fsPath := filepath.Join("attic", urlPath+"."+revStr+".md")
	return s.storage.Exists(fsPath)
}

// ListTrash returns all trash entries (deleted pages with their deletion timestamps and metadata)
func (s *ContentService) ListTrash() ([]model.TrashEntry, error) {
	entries := []model.TrashEntry{}
	return s.listTrashRecursive("", entries)
}

func (s *ContentService) listTrashRecursive(relativePath string, entries []model.TrashEntry) ([]model.TrashEntry, error) {
	dirPath := filepath.Join("trash", relativePath)
	if !s.storage.Exists(dirPath) {
		return entries, nil
	}

	fileInfos, err := s.storage.ReadDirectory(dirPath)
	if err != nil {
		return nil, err
	}

	for _, fi := range fileInfos {
		if !fi.IsDir() {
			continue
		}

		// Check if this is a timestamp directory (starts with _ followed by digits)
		if strings.HasPrefix(fi.Name(), "_") {
			timestampStr := strings.TrimPrefix(fi.Name(), "_")
			if deletedAt, err := strconv.ParseInt(timestampStr, 10, 64); err == nil {
				// This is a deletion timestamp directory
				urlPath := relativePath

				// Read metadata from the page file
				meta := model.ContentMeta{}
				pageName := path.Base(urlPath)
				pageFile := filepath.Join("trash", urlPath, fi.Name(), pageName+".md")
				if s.storage.Exists(pageFile) {
					bytes, err := s.storage.ReadFile(pageFile)
					if err == nil {
						fm, _, err := parseFrontMatter(string(bytes))
						if err == nil {
							meta = fm
						}
					}
				}

				entries = append(entries, model.TrashEntry{
					Url:       urlPath,
					DeletedAt: deletedAt,
					Meta:      meta,
				})
				continue
			}
		}

		// This is a path component, recurse
		subPath := fi.Name()
		if relativePath != "" {
			subPath = relativePath + "/" + fi.Name()
		}
		entries, err = s.listTrashRecursive(subPath, entries)
		if err != nil {
			return nil, err
		}
	}

	return entries, nil
}

// DeleteTrashEntry permanently deletes a specific trash entry
func (s *ContentService) DeleteTrashEntry(urlPath string, deletedAt int64) error {
	timestampStr := "_" + strconv.FormatInt(deletedAt, 10)
	trashDir := filepath.Join("trash", urlPath, timestampStr)

	if !s.storage.Exists(trashDir) {
		return model.ErrNotFound
	}

	return s.storage.DeleteDirectory(trashDir)
}

// ReadTrashPage reads a specific page from the trash
func (s *ContentService) ReadTrashPage(urlPath string, deletedAt int64) (model.Page, error) {
	timestampStr := "_" + strconv.FormatInt(deletedAt, 10)
	pageName := path.Base(urlPath)
	pageFile := filepath.Join("trash", urlPath, timestampStr, pageName+".md")

	if !s.storage.Exists(pageFile) {
		return model.Page{}, model.ErrNotFound
	}

	bytes, err := s.storage.ReadFile(pageFile)
	if err != nil {
		return model.Page{}, err
	}

	fm, content, err := parseFrontMatter(string(bytes))
	if err != nil {
		return model.Page{}, fmt.Errorf("could not parse frontmatter: %w", err)
	}

	page := model.Page{
		Url:     urlPath,
		Content: content,
		Meta:    fm,
	}
	return page, nil
}

// RestoreFromTrash restores a page from trash to its original location
func (s *ContentService) RestoreFromTrash(urlPath string, deletedAt int64) error {
	timestampStr := "_" + strconv.FormatInt(deletedAt, 10)
	pageName := path.Base(urlPath)
	trashDir := filepath.Join("trash", urlPath, timestampStr)

	if !s.storage.Exists(trashDir) {
		return model.ErrNotFound
	}

	// Check if destination already exists
	if s.IsPage(urlPath) || s.IsFolder(urlPath) {
		return model.ErrPageOrFolderExistsAlready
	}

	// Create parent folders if they don't exist
	if err := s.ensureParentFoldersExist(urlPath); err != nil {
		return fmt.Errorf("could not create parent folders: %w", err)
	}

	// Move the page file back
	srcPagePath := filepath.Join(trashDir, pageName+".md")
	destPagePath := filepath.Join("pages", urlPath+".md")
	if err := s.storage.Rename(srcPagePath, destPagePath); err != nil {
		return fmt.Errorf("could not restore page: %w", err)
	}

	// Move all attic entries back
	trashFiles, err := s.storage.ReadDirectory(trashDir)
	if err != nil {
		// Directory might be empty now or error reading it, continue with empty list
		trashFiles = nil
	}

	for _, fi := range trashFiles {
		if fi.IsDir() {
			continue
		}

		// Check if this is an attic file (has revision number)
		name := fi.Name()
		if !strings.HasPrefix(name, pageName+".") || !strings.HasSuffix(name, ".md") {
			continue
		}

		// Extract revision number
		revPart := strings.TrimPrefix(name, pageName+".")
		revPart = strings.TrimSuffix(revPart, ".md")
		if _, err := strconv.ParseInt(revPart, 10, 64); err != nil {
			continue // Not a revision file
		}

		srcAtticPath := filepath.Join(trashDir, name)
		destAtticPath := filepath.Join("attic", urlPath+"."+revPart+".md")
		if err := s.storage.Rename(srcAtticPath, destAtticPath); err != nil {
			return fmt.Errorf("could not restore attic entry: %w", err)
		}
	}

	// Delete the now-empty trash directory
	_ = s.storage.DeleteEmptyDirectory(trashDir)

	// Update search index
	page, err := s.ReadPage(urlPath, nil)
	if err != nil {
		return fmt.Errorf("could not read restored page: %w", err)
	}
	if err := s.index.Index(urlPath, page); err != nil {
		log.Printf("[INDEX] Could not index restored page %s: %v", urlPath, err)
	}

	return nil
}

func (s *ContentService) ReadPage(urlPath string, revision *int64) (model.Page, error) {
	var fsPath string
	if revision == nil {
		fsPath = filepath.Join("pages", urlPath+".md")
	} else {
		revStr := strconv.FormatInt(*revision, 10)
		fsPath = filepath.Join("attic", urlPath+"."+revStr+".md")
	}

	bytes, err := s.storage.ReadFile(fsPath)
	if err != nil {
		return model.Page{}, err
	}

	fm, content, err := parseFrontMatter(string(bytes))
	if err != nil {
		return model.Page{}, fmt.Errorf("could not parse frontmatter: %w", err)
	}

	page := model.Page{
		Url:     urlPath,
		Content: content,
		Meta:    fm,
	}
	return page, nil
}

// SavePage saves a page and creates a version in the attic.
func (s *ContentService) SavePage(urlPath, content string, meta model.ContentMeta, userID string) error {
	return s.savePageAtInternal(urlPath, content, meta, userID, true, time.Now())
}

// SavePageWithoutVersion saves a page without creating a version in the attic.
// Use this for metadata-only changes (e.g., ACL, title) that shouldn't create history entries.
func (s *ContentService) SavePageWithoutVersion(urlPath, content string, meta model.ContentMeta, userID string) error {
	return s.savePageAtInternal(urlPath, content, meta, userID, false, time.Now())
}

// SavePageAt saves a page with a specific timestamp.
// This is primarily useful for testing scenarios where you need to create multiple
// attic versions without waiting for time to pass (since revisions are stored with second precision).
func (s *ContentService) SavePageAt(urlPath, content string, meta model.ContentMeta, userID string, revisionTime time.Time) error {
	return s.savePageAtInternal(urlPath, content, meta, userID, true, revisionTime)
}

// savePageAtInternal saves a page with a specific timestamp (internal implementation).
func (s *ContentService) savePageAtInternal(urlPath, content string, meta model.ContentMeta, userID string, createVersion bool, revisionTime time.Time) error {
	if !s.IsFolder(path.Dir(urlPath)) {
		return model.ErrParentFolderNotFound
	}
	if s.IsFolder(urlPath) {
		return model.ErrPageOrFolderExistsAlready
	}

	// Set modification metadata
	meta.ModifiedAt = time.Now().UTC()
	meta.ModifiedByUserID = userID

	fsPath := filepath.Join("pages", urlPath+".md")

	serializedPage, err := serializeFrontMatter(meta, content)
	if err != nil {
		return fmt.Errorf("could not serialize frontmatter: %w", err)
	}

	if err := s.storage.WriteFile(fsPath, []byte(serializedPage)); err != nil {
		return fmt.Errorf("could not write file: %w", err)
	}

	if createVersion {
		revision := revisionTime.Unix()
		revStr := strconv.FormatInt(revision, 10)
		atticFile := filepath.Join("attic", urlPath+"."+revStr+".md")

		if err := s.storage.WriteFile(atticFile, []byte(serializedPage)); err != nil {
			return fmt.Errorf("could not save page to attic: %w", err)
		}
	}

	// Update search index
	page := model.Page{
		Url:     urlPath,
		Content: content,
		Meta:    meta,
	}
	if err := s.index.Index(urlPath, page); err != nil {
		log.Printf("[INDEX] Could not update page %s in index: %v", urlPath, err)
	}

	return nil
}

func (s *ContentService) DeletePage(urlPath string) error {
	return s.deletePageAt(urlPath, time.Now())
}

// deletePageAt deletes a page at the specified time (for testing with custom timestamps).
func (s *ContentService) deletePageAt(urlPath string, deletedAt time.Time) error {
	// Move page and attic entries to trash
	if err := s.movePageToTrashAt(urlPath, deletedAt); err != nil {
		return err
	}

	// Update search index
	if err := s.index.Delete(urlPath); err != nil {
		log.Printf("[INDEX] Could not delete page %s from index: %v", urlPath, err)
	}

	return nil
}

// movePageToTrashAt moves a page and its attic entries to the trash folder at the specified time.
// Trash structure: trash/{urlPath}/_{timestamp}/{filename}.md
func (s *ContentService) movePageToTrashAt(urlPath string, deletedAt time.Time) error {
	timestamp := deletedAt.Unix()
	timestampStr := "_" + strconv.FormatInt(timestamp, 10)
	pageName := path.Base(urlPath)

	trashDir := filepath.Join("trash", urlPath, timestampStr)

	// Move the page file
	srcPagePath := filepath.Join("pages", urlPath+".md")
	destPagePath := filepath.Join(trashDir, pageName+".md")
	if err := s.storage.Rename(srcPagePath, destPagePath); err != nil {
		return fmt.Errorf("could not move page to trash: %w", err)
	}

	// Move all attic entries for this page
	atticEntries, err := s.ListAttic(urlPath)
	if err != nil {
		// If attic directory doesn't exist or is empty, that's fine
		return nil
	}

	for _, entry := range atticEntries {
		revStr := strconv.FormatInt(entry.Revision, 10)
		srcAtticPath := filepath.Join("attic", urlPath+"."+revStr+".md")
		destAtticPath := filepath.Join(trashDir, pageName+"."+revStr+".md")

		if err := s.storage.Rename(srcAtticPath, destAtticPath); err != nil {
			return fmt.Errorf("could not move attic entry %d to trash: %w", entry.Revision, err)
		}
	}

	return nil
}

func (s *ContentService) CreateFolder(urlPath string, meta model.ContentMeta) error {
	if !s.IsFolder(path.Dir(urlPath)) {
		return model.ErrParentFolderNotFound
	}
	if s.IsPage(urlPath) || s.IsFolder(urlPath) {
		return model.ErrPageOrFolderExistsAlready
	}

	serialized, err := serializeFrontMatter(meta, "")
	if err != nil {
		return fmt.Errorf("could not serialize frontmatter: %w", err)
	}

	dirPath := filepath.Join("pages", urlPath)
	if err := s.storage.CreateDirectory(dirPath); err != nil {
		return err
	}

	indexPath := filepath.Join("pages", urlPath, "_index.md")
	if err := s.storage.WriteFile(indexPath, []byte(serialized)); err != nil {
		return err
	}

	// Update search index
	folder := model.Folder{
		Content: nil,
		Meta:    meta,
	}
	if err := s.index.Index(urlPath, folder); err != nil {
		log.Printf("[INDEX] Could not add folder %s to index: %v", urlPath, err)
	}

	return nil
}

// ReadFolderMeta reads only the folder's metadata from _index.md without enumerating contents.
// This is more efficient when only the metadata (e.g., ACL) is needed.
func (s *ContentService) ReadFolderMeta(urlPath string) (model.ContentMeta, error) {
	indexPath := filepath.Join("pages", urlPath, "_index.md")
	bytes, err := s.storage.ReadFile(indexPath)
	if err != nil {
		return model.ContentMeta{}, err
	}
	fm, _, err := parseFrontMatter(string(bytes))
	if err != nil {
		return model.ContentMeta{}, err
	}
	return fm, nil
}

func (s *ContentService) ReadFolder(urlPath string) (model.Folder, error) {
	dirPath := filepath.Join("pages", urlPath)

	// Get a list of all files in the directory
	fileInfos, err := s.storage.ReadDirectory(dirPath)
	if err != nil {
		return model.Folder{}, err
	}

	folderEntries := make([]model.FolderEntry, 0, len(fileInfos))
	for _, fi := range fileInfos {

		u, err := url.JoinPath(urlPath, fi.Name())
		if err != nil {
			return model.Folder{}, fmt.Errorf("could not join url: %w", err)
		}

		e := model.FolderEntry{
			Url:      u,
			Name:     fi.Name(),
			IsFolder: fi.IsDir(),
		}

		if e.IsFolder {
			meta, err := s.ReadFolderMeta(e.Url)
			if err != nil {
				return model.Folder{}, fmt.Errorf("could not read folder %s: %w", e.Url, err)
			}
			e.Title = meta.Title
			e.ACL = meta.ACL
		} else {
			if !strings.HasPrefix(e.Name, "_") && strings.HasSuffix(e.Name, ".md") {
				e.Name = strings.TrimSuffix(e.Name, ".md")
				e.Url = strings.TrimSuffix(e.Url, ".md")
			} else {
				continue
			}

			page, err := s.ReadPage(e.Url, nil)
			if err != nil {
				return model.Folder{}, fmt.Errorf("could not read page %s: %w", e.Url, err)
			}

			e.Title = page.Meta.Title
			e.ACL = page.Meta.ACL
		}

		folderEntries = append(folderEntries, e)
	}

	// Read folder's own metadata
	fm, err := s.ReadFolderMeta(urlPath)
	if err != nil {
		return model.Folder{}, err
	}

	folder := model.Folder{
		Url:     urlPath,
		Content: folderEntries,
		Meta:    fm,
	}

	return folder, nil
}

func (s *ContentService) SaveFolder(urlPath string, meta model.ContentMeta) error {
	indexPath := filepath.Join("pages", urlPath, "_index.md")

	serialized, err := serializeFrontMatter(meta, "")
	if err != nil {
		return fmt.Errorf("could not serialize frontmatter: %w", err)
	}

	if err := s.storage.WriteFile(indexPath, []byte(serialized)); err != nil {
		return fmt.Errorf("could not write index file: %w", err)
	}

	// Update search index
	if urlPath != "" {
		folder := model.Folder{
			Content: nil,
			Meta:    meta,
		}
		if err := s.index.Index(urlPath, folder); err != nil {
			log.Printf("[INDEX] Could not update folder %s in index: %v", urlPath, err)
		}
	}

	return nil
}

func (s *ContentService) DeleteEmptyFolder(urlPath string) error {
	dirPath := filepath.Join("pages", urlPath)
	indexPath := filepath.Join("pages", urlPath, "_index.md")

	if !s.folderIsEmpty(urlPath) {
		return model.ErrFolderNotEmpty
	}

	if err := s.storage.DeleteFile(indexPath); err != nil {
		return err
	}

	if err := s.storage.DeleteEmptyDirectory(dirPath); err != nil {
		return err
	}

	// Update search index
	if err := s.index.Delete(urlPath); err != nil {
		log.Printf("[INDEX] Could not delete folder %s from index: %v", urlPath, err)
	}

	return nil
}

// DeleteFolder deletes a folder and all its contents by moving all pages to trash.
// Folders and their metadata are not preserved in trash - only individual pages and their attic entries.
func (s *ContentService) DeleteFolder(urlPath string) error {
	// Cannot delete root folder
	if urlPath == "" {
		return model.ErrCannotDeleteRoot
	}

	// Remove from search index first
	if err := s.removeFolderFromIndex(urlPath); err != nil {
		return err
	}

	// Move all pages in the folder (and subfolders) to trash
	if err := s.moveFolderPagesToTrash(urlPath); err != nil {
		return err
	}

	// Delete the folder directory (now empty except for _index.md files)
	dirPath := filepath.Join("pages", urlPath)
	if err := s.storage.DeleteDirectory(dirPath); err != nil {
		return fmt.Errorf("could not delete folder directory: %w", err)
	}

	return nil
}

// moveFolderPagesToTrash recursively moves all pages in a folder to trash.
func (s *ContentService) moveFolderPagesToTrash(urlPath string) error {
	folder, err := s.ReadFolder(urlPath)
	if err != nil {
		return fmt.Errorf("could not read folder %s: %w", urlPath, err)
	}

	for _, entry := range folder.Content {
		if entry.IsFolder {
			// Recursively process subfolder
			if err := s.moveFolderPagesToTrash(entry.Url); err != nil {
				return err
			}
		} else {
			// Move page to trash
			if err := s.movePageToTrashAt(entry.Url, time.Now()); err != nil {
				return fmt.Errorf("could not move page %s to trash: %w", entry.Url, err)
			}
		}
	}

	return nil
}

func (s *ContentService) folderIsEmpty(urlPath string) bool {
	dirPath := filepath.Join("pages", urlPath)

	entries, err := s.storage.ReadDirectory(dirPath)
	if err != nil {
		return false
	}

	return len(entries) == 1 &&
		entries[0].Name() == "_index.md" &&
		!entries[0].IsDir()
}

func (s *ContentService) DeleteAll() error {
	for _, dir := range []string{"pages", "attic", "trash"} {
		if err := s.storage.DeleteDirectory(dir); err != nil {
			return err
		}
	}

	if err := s.initializeStorage(); err != nil {
		return err
	}

	// Update search index
	if err := s.RecreateIndex(); err != nil {
		log.Println("[INDEX] Could not recreate index:", err)
	}

	return nil
}

func (s *ContentService) ReadAncestorsMeta(urlPath string) ([]model.ContentMetaWithURL, error) {
	return s.addFolderMetaFromParent(urlPath, []model.ContentMetaWithURL{})
}

func (s *ContentService) addFolderMetaFromParent(urlPath string, metas []model.ContentMetaWithURL) ([]model.ContentMetaWithURL, error) {
	parentUrl, err := url.JoinPath(urlPath, "..")
	if err != nil {
		return nil, err
	}

	if s.IsFolder(parentUrl) {
		folderMeta, err := s.ReadFolderMeta(parentUrl)
		if err != nil {
			return nil, err
		}

		meta := model.ContentMetaWithURL{
			Url:         parentUrl,
			ContentMeta: folderMeta,
		}

		metas = append(metas, meta)
	}
	// if it doesn't exist, skip it

	if parentUrl == "" {
		return metas, nil
	}
	return s.addFolderMetaFromParent(parentUrl, metas)
}

// GetEffectivePermissions returns the effective ACL for content by checking the content's own ACL
// and falling back to ancestor ACLs. Returns an empty slice if no ACL is found (should never
// occur in reality).
func (s *ContentService) GetEffectivePermissions(meta model.ContentMeta, ancestorsMetas []model.ContentMetaWithURL) []model.AccessRule {
	if meta.ACL != nil {
		return *meta.ACL
	}

	for i := range ancestorsMetas {
		if ancestorsMetas[i].ACL != nil {
			return *ancestorsMetas[i].ACL
		}
	}

	return []model.AccessRule{}
}

// ListAttic lists all attic entries (revisions) for a given page, sorted by revision number ascending.
func (s *ContentService) ListAttic(urlPath string) ([]model.AtticEntry, error) {
	pageName := path.Base(urlPath)
	parentDir := filepath.Join("attic", filepath.Dir(urlPath))

	fileInfos, err := s.storage.ReadDirectory(parentDir)
	if err != nil {
		return nil, err
	}

	atticEntries := []model.AtticEntry{}
	for _, fi := range fileInfos {
		if fi.IsDir() {
			continue
		}

		// Check if name start with page name
		name, found := strings.CutPrefix(fi.Name(), pageName+".")
		if !found {
			continue
		}

		// Check if name end with file extension
		name, found = strings.CutSuffix(name, ".md")
		if !found {
			continue
		}

		rev, err := strconv.ParseInt(name, 10, 64)
		if err != nil {
			continue
		}

		atticEntries = append(atticEntries, model.AtticEntry{Revision: rev})
	}

	// Sort by revision
	sort.Slice(atticEntries, func(i, j int) bool {
		return atticEntries[i].Revision < atticEntries[j].Revision
	})

	return atticEntries, nil
}

// MovePage moves a page from sourcePath to destinationPath, including all attic entries.
func (s *ContentService) MovePage(sourcePath, destinationPath string) error {
	// Validate source exists
	if !s.IsPage(sourcePath) {
		return model.ErrNotFound
	}

	// Validate destination parent folder exists
	if !s.IsFolder(path.Dir(destinationPath)) {
		return model.ErrParentFolderNotFound
	}

	// Validate destination doesn't already exist
	if s.IsPage(destinationPath) || s.IsFolder(destinationPath) {
		return model.ErrDestinationExists
	}

	// Move the page file
	srcFsPath := filepath.Join("pages", sourcePath+".md")
	destFsPath := filepath.Join("pages", destinationPath+".md")
	if err := s.storage.Rename(srcFsPath, destFsPath); err != nil {
		return fmt.Errorf("could not move page file: %w", err)
	}

	// Move all attic entries for this page
	if err := s.moveAtticEntries(sourcePath, destinationPath); err != nil {
		return fmt.Errorf("could not move attic entries: %w", err)
	}

	// Update search index: delete old, add new
	if err := s.index.Delete(sourcePath); err != nil {
		log.Printf("[INDEX] Could not delete old page %s from index: %v", sourcePath, err)
	}
	page, err := s.ReadPage(destinationPath, nil)
	if err != nil {
		return fmt.Errorf("could not read moved page: %w", err)
	}
	if err := s.index.Index(destinationPath, page); err != nil {
		log.Printf("[INDEX] Could not index new page %s: %v", destinationPath, err)
	}

	return nil
}

// MoveFolder moves a folder from sourcePath to destinationPath, including all content and attic entries.
func (s *ContentService) MoveFolder(sourcePath, destinationPath string) error {
	// Validate source exists
	if !s.IsFolder(sourcePath) {
		return model.ErrNotFound
	}

	// Cannot move root folder
	if sourcePath == "" {
		return model.ErrCannotMoveRoot
	}

	// Validate destination doesn't already exist
	if s.IsPage(destinationPath) || s.IsFolder(destinationPath) {
		return model.ErrDestinationExists
	}

	// Delete index entries before moving
	if err := s.removeFolderFromIndex(sourcePath); err != nil {
		return fmt.Errorf("could not remove folder from index: %w", err)
	}

	// Move the folder directory
	srcFsPath := filepath.Join("pages", sourcePath)
	destFsPath := filepath.Join("pages", destinationPath)
	if err := s.storage.Rename(srcFsPath, destFsPath); err != nil {
		return fmt.Errorf("could not move folder: %w", err)
	}

	// Move the attic folder (if it exists)
	srcAtticPath := filepath.Join("attic", sourcePath)
	destAtticPath := filepath.Join("attic", destinationPath)
	if s.storage.Exists(srcAtticPath) {
		if err := s.storage.Rename(srcAtticPath, destAtticPath); err != nil {
			return fmt.Errorf("could not move attic folder: %w", err)
		}
	}

	// Index the new folder and all its contents
	if err := s.indexFolder(destinationPath, &s.index); err != nil {
		log.Println("[INDEX] Could not index new folder:", err)
	}

	return nil
}

// ensureParentFoldersExist creates all parent folders for a given path if they don't exist.
func (s *ContentService) ensureParentFoldersExist(urlPath string) error {
	parentPath := path.Dir(urlPath)
	if parentPath == "." || parentPath == "" {
		// Root folder always exists
		return nil
	}

	// If parent exists, we're done
	if s.IsFolder(parentPath) {
		return nil
	}

	// Recursively ensure parent's parent exists first
	if err := s.ensureParentFoldersExist(parentPath); err != nil {
		return err
	}

	// Create this folder with empty metadata
	if err := s.CreateFolder(parentPath, model.ContentMeta{}); err != nil {
		return fmt.Errorf("could not create folder %s: %w", parentPath, err)
	}

	return nil
}

// moveAtticEntries moves all attic entries for a page from oldPath to newPath.
func (s *ContentService) moveAtticEntries(oldPath, newPath string) error {
	atticEntries, err := s.ListAttic(oldPath)
	if err != nil {
		// If attic directory doesn't exist, that's fine - no entries to move
		return nil
	}

	for _, entry := range atticEntries {
		revStr := strconv.FormatInt(entry.Revision, 10)
		oldAtticPath := filepath.Join("attic", oldPath+"."+revStr+".md")
		newAtticPath := filepath.Join("attic", newPath+"."+revStr+".md")

		if err := s.storage.Rename(oldAtticPath, newAtticPath); err != nil {
			return fmt.Errorf("could not move attic entry %d: %w", entry.Revision, err)
		}
	}

	return nil
}

// DeleteAtticEntry deletes a single attic entry (version) for a page.
func (s *ContentService) DeleteAtticEntry(urlPath string, revision int64) error {
	revStr := strconv.FormatInt(revision, 10)
	atticPath := filepath.Join("attic", urlPath+"."+revStr+".md")

	if !s.storage.Exists(atticPath) {
		return model.ErrNotFound
	}

	return s.storage.DeleteFile(atticPath)
}

// ListAllPages returns the URLs of all pages in the wiki by recursively walking the pages directory.
func (s *ContentService) ListAllPages() ([]string, error) {
	return s.listAllPagesRecursive("")
}

func (s *ContentService) listAllPagesRecursive(urlPath string) ([]string, error) {
	var pages []string

	folder, err := s.ReadFolder(urlPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range folder.Content {
		if entry.IsFolder {
			// Recursively process subfolder
			subPages, err := s.listAllPagesRecursive(entry.Url)
			if err != nil {
				return nil, err
			}
			pages = append(pages, subPages...)
		} else {
			pages = append(pages, entry.Url)
		}
	}

	return pages, nil
}
