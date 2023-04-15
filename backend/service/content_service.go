package service

import (
	"fmt"
	"log"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/tfabritius/plainpage/model"
)

func NewContentService(store model.Storage) ContentService {
	s := ContentService{
		storage: store,
	}

	// Create pages and attic directories
	for _, dir := range []string{"pages", "attic"} {
		// Create directory, if it doesn't exist
		if !s.storage.Exists(dir) {
			if err := s.storage.CreateDirectory(dir); err != nil {
				log.Fatalln("Could not create "+dir+" folder:", err)
			}
		}
	}

	// Create _index.md with default ACL if it doesn't exist
	if !s.IsFolder("/") {
		defaultACL := []model.AccessRule{
			{Subject: "all", Operations: []model.AccessOp{model.AccessOpRead, model.AccessOpWrite, model.AccessOpDelete}},
		}
		s.SaveFolder("/", model.PageMeta{ACL: &defaultACL})
	}

	return s
}

type ContentService struct {
	storage model.Storage
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

	u, err := url.JoinPath("/", urlPath)
	if err != nil {
		return model.Page{}, fmt.Errorf("could not join url: %w", err)
	}

	page := model.Page{
		Url:     u,
		Content: content,
		Meta:    fm,
	}
	return page, nil
}

func (s *ContentService) SavePage(urlPath, content string, meta model.PageMeta) error {
	if !s.IsFolder(path.Dir(urlPath)) {
		return model.ErrParentFolderNotFound
	}
	if s.IsFolder(urlPath) {
		return model.ErrPageOrFolderExistsAlready
	}

	fsPath := filepath.Join("pages", urlPath+".md")

	serializedPage, err := serializeFrontMatter(meta, content)
	if err != nil {
		return fmt.Errorf("could not serialize frontmatter: %w", err)
	}

	if err := s.storage.WriteFile(fsPath, []byte(serializedPage)); err != nil {
		return fmt.Errorf("could not write file: %w", err)
	}

	revision := time.Now().Unix()
	revStr := strconv.FormatInt(revision, 10)
	atticFile := filepath.Join("attic", urlPath+"."+revStr+".md")

	if err := s.storage.WriteFile(atticFile, []byte(serializedPage)); err != nil {
		return fmt.Errorf("could not save page to attic: %w", err)
	}

	return nil
}

func (s *ContentService) DeletePage(urlPath string) error {
	fsPath := filepath.Join("pages", urlPath+".md")
	return s.storage.DeleteFile(fsPath)
}

func (s *ContentService) CreateFolder(urlPath string) error {
	if !s.IsFolder(path.Dir(urlPath)) {
		return model.ErrParentFolderNotFound
	}
	if s.IsPage(urlPath) || s.IsFolder(urlPath) {
		return model.ErrPageOrFolderExistsAlready
	}

	dirPath := filepath.Join("pages", urlPath)
	if err := s.storage.CreateDirectory(dirPath); err != nil {
		return err
	}

	indexPath := filepath.Join("pages", urlPath, "_index.md")
	if err := s.storage.WriteFile(indexPath, nil); err != nil {
		return err
	}

	return nil
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

		u, err := url.JoinPath("/", urlPath, fi.Name())
		if err != nil {
			return model.Folder{}, fmt.Errorf("could not join url: %w", err)
		}

		e := model.FolderEntry{
			Url:      u,
			Name:     fi.Name(),
			IsFolder: fi.IsDir(),
		}
		if !e.IsFolder {
			if !strings.HasPrefix(e.Name, "_") && strings.HasSuffix(e.Name, ".md") {
				e.Name = strings.TrimSuffix(e.Name, ".md")
				e.Url = strings.TrimSuffix(e.Url, ".md")
			} else {
				continue
			}
		}

		folderEntries = append(folderEntries, e)
	}

	// Read _index.md
	indexPath := filepath.Join("pages", urlPath, "_index.md")
	bytes, err := s.storage.ReadFile(indexPath)
	if err != nil {
		return model.Folder{}, err
	}
	fm, _, err := parseFrontMatter(string(bytes))
	if err != nil {
		return model.Folder{}, err
	}

	folder := model.Folder{
		Content: folderEntries,
		Meta:    fm,
	}

	return folder, nil
}

func (s *ContentService) SaveFolder(urlPath string, meta model.PageMeta) error {
	indexPath := filepath.Join("pages", urlPath, "_index.md")

	serialized, err := serializeFrontMatter(meta, "")
	if err != nil {
		return fmt.Errorf("could not serialize frontmatter: %w", err)
	}

	if err := s.storage.WriteFile(indexPath, []byte(serialized)); err != nil {
		return fmt.Errorf("could not write index file: %w", err)
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

func (s *ContentService) GetEffectivePermissions(urlPath string) (*[]model.AccessRule, error) {
	if s.IsPage(urlPath) {
		page, err := s.ReadPage(urlPath, nil)
		if err != nil {
			return nil, err
		}

		if page.Meta.ACL != nil {
			return page.Meta.ACL, nil
		}

	} else if s.IsFolder(urlPath) {
		folder, err := s.ReadFolder(urlPath)
		if err != nil {
			return nil, err
		}

		if folder.Meta.ACL != nil {
			return folder.Meta.ACL, nil
		}
	}

	if urlPath == "" {
		return nil, nil
	}

	parentUrl, err := url.JoinPath(urlPath, "..")
	if err != nil {
		return nil, err
	}

	return s.GetEffectivePermissions(parentUrl)
}

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

	return atticEntries, nil
}
