package model

import "time"

type Page struct {
	Url     string      `json:"url"`
	Content string      `json:"content"`
	Meta    ContentMeta `json:"meta"`
}

func (Page) BleveType() string {
	return "page"
}

type Folder struct {
	Url     string        `json:"url"`
	Content []FolderEntry `json:"content"`
	Meta    ContentMeta   `json:"meta"`
}

func (Folder) BleveType() string {
	return "folder"
}

type ContentMeta struct {
	Title                 string        `json:"title" yaml:"title"`
	Tags                  []string      `json:"tags" yaml:"tags"`
	ACL                   *[]AccessRule `json:"acl" yaml:"acl"`
	ModifiedAt            time.Time     `json:"modifiedAt,omitempty" yaml:"modifiedAt"`
	ModifiedByUserID      string        `json:"-" yaml:"modifiedBy"`                      // Stored in YAML, not exposed in API
	ModifiedByUsername    string        `json:"modifiedByUsername,omitempty" yaml:"-"`    // Exposed in API, not stored in YAML
	ModifiedByDisplayName string        `json:"modifiedByDisplayName,omitempty" yaml:"-"` // Exposed in API, not stored in YAML
}

type ContentMetaWithURL struct {
	ContentMeta
	Url string
}

type FolderEntry struct {
	Url      string `json:"url"`
	Name     string `json:"name"`
	Title    string `json:"title"`
	IsFolder bool   `json:"isFolder"`

	ACL *[]AccessRule `json:"-"`
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
	// Operations on pages/folders
	AccessOpRead   AccessOp = "read"
	AccessOpWrite  AccessOp = "write"
	AccessOpDelete AccessOp = "delete"

	// Global operations/permissions
	AccessOpAdmin    AccessOp = "admin"
	AccessOpRegister AccessOp = "register"
)

type User struct {
	ID           string `json:"id" yaml:"id"`
	Username     string `json:"username" yaml:"username"`
	PasswordHash string `json:"-" yaml:"passwordHash"`
	DisplayName  string `json:"displayName" yaml:"displayName"`
}

type Config struct {
	ACL       []AccessRule `json:"acl" yaml:"acl"`
	AppTitle  string       `json:"appTitle" yaml:"appTitle"`
	JwtSecret string       `json:"-" yaml:"jwtSecret"`
	SetupMode bool         `json:"setupMode" yaml:"setupMode"`
}

type SearchHit struct {
	Url          string              `json:"url"`
	Meta         ContentMeta         `json:"meta"`
	Fragments    map[string][]string `json:"fragments"`
	EffectiveACL []AccessRule        `json:"-"`
	IsFolder     bool                `json:"isFolder"`
}
