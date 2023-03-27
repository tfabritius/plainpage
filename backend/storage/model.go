package storage

import "errors"

type Storage interface {
	// IsPage checks if page exists at given path
	IsPage(urlPath string) bool

	// IsAtticPage checks if page and revision exist
	IsAtticPage(urlPath string, revision int64) bool

	// IsFolder checks if folder exists at given path
	IsFolder(urlPath string) bool

	// ReadPage returns page at given path
	ReadPage(urlPath string, revision *int64) (Page, error)

	// ReadPage returns folder entries at given path
	ReadFolder(urlPath string) ([]FolderEntry, error)

	// ListAttic returns relevent entries in attic
	ListAttic(urlPath string) ([]AtticEntry, error)

	// CreateFolder creates new folder at given path
	CreateFolder(urlPath string) error

	// SavePage creates or updates page
	SavePage(urlPath, content string, meta PageMeta) error

	// DeletePage removes page
	DeletePage(urlPath string) error

	// DeleteEmptyFolder removes folder
	DeleteEmptyFolder(urlPath string) error
}

var ErrNotFound = errors.New("not found")
var ErrParentFolderNotFound = errors.New("parent folder not found")
var ErrPageOrFolderExistsAlready = errors.New("page or folder exists already")
var ErrFolderNotEmpty = errors.New("folder is not empty")

type Page struct {
	Url     string   `json:"url"`
	Content string   `json:"content"`
	Meta    PageMeta `json:"meta"`
}

type PageMeta struct {
	Title string   `json:"title" yaml:"title"`
	Tags  []string `json:"tags" yaml:"tags"`
}

type FolderEntry struct {
	Url      string `json:"url"`
	Name     string `json:"name"`
	IsFolder bool   `json:"isFolder"`
}

type AtticEntry struct {
	Revision int64 `json:"rev"`
}
