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

	// ReadPage returns folder at given path
	ReadFolder(urlPath string) (Folder, error)

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

	GetAllUsers() ([]User, error)
	GetUserByUsername(username string) (User, error)
	SaveAllUsers(users []User) error
	AddUser(username, password, realName string) (User, error)
	SaveUser(user User) error
	DeleteUserByUsername(username string) error
}

var ErrNotFound = errors.New("not found")
var ErrParentFolderNotFound = errors.New("parent folder not found")
var ErrPageOrFolderExistsAlready = errors.New("page or folder exists already")
var ErrFolderNotEmpty = errors.New("folder is not empty")
var ErrInvalidUsername = errors.New("invalid username")
var ErrUserExistsAlready = errors.New("user already exists")

type Page struct {
	Url     string   `json:"url"`
	Content string   `json:"content"`
	Meta    PageMeta `json:"meta"`
}

type PageMeta struct {
	Title string        `json:"title" yaml:"title"`
	Tags  []string      `json:"tags" yaml:"tags"`
	ACLs  *[]AccessRule `json:"acls" yaml:"acls"`
}

type Folder struct {
	Content []FolderEntry `json:"content"`
	Meta    PageMeta      `json:"meta"`
}

type FolderEntry struct {
	Url      string `json:"url"`
	Name     string `json:"name"`
	IsFolder bool   `json:"isFolder"`
}

type AtticEntry struct {
	Revision int64 `json:"rev"`
}

type AccessRule struct {
	// The subject trying to access an object
	// e.g.
	// - user:xyz
	// - group:xyz
	// - all (all registered users)
	// - anonymous (unregistered users)
	Subject string `json:"subject" yaml:"subject"`

	// List of permitted operations
	Operations []AccessOp `json:"ops" yaml:"ops"`

	// Additional information about subject, if applicable
	User *User `json:"user" yaml:"-"`
}

type AccessOp string

const (
	AccessOpRead   AccessOp = "read"
	AccessOpWrite  AccessOp = "write"
	AccessOpDelete AccessOp = "delete"
)

type User struct {
	ID           string `json:"id" yaml:"id"`
	Username     string `json:"username" yaml:"username"`
	PasswordHash string `json:"-" yaml:"passwordHash"`
	RealName     string `json:"realName" yaml:"realName"`
}
