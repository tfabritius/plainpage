package model

import (
	"encoding/json"
	"time"
)

/* Types corresponding to frontend/app/types/api.ts */

type Breadcrumb struct {
	Name  string `json:"name"`
	Title string `json:"title"`
	Url   string `json:"url"`
}

type GetAppResponse struct {
	AppTitle      string `json:"appTitle"`
	SetupMode     bool   `json:"setupMode"`
	AllowRegister bool   `json:"allowRegister"`
	AllowAdmin    bool   `json:"allowAdmin"`
	Version       string `json:"version,omitempty"`
	GitSha        string `json:"gitSha,omitempty"`
}

type PutRequest struct {
	Page   *Page   `json:"page"`
	Folder *Folder `json:"folder"`
}

type GetContentResponse struct {
	Page        *Page        `json:"page"`
	Folder      *Folder      `json:"folder"`
	AllowWrite  bool         `json:"allowWrite"`
	AllowDelete bool         `json:"allowDelete"`
	Breadcrumbs []Breadcrumb `json:"breadcrumbs"`
}

type GetAtticListResponse struct {
	Entries     []AtticEntry `json:"entries"`
	Breadcrumbs []Breadcrumb `json:"breadcrumbs"`
}

type PatchOperation struct {
	Op    string           `json:"op"`
	Path  string           `json:"path"`
	Value *json.RawMessage `json:"value,omitempty"`
	From  *string          `json:"from,omitempty"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

type PostUserRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
}

type DeleteUserRequest struct {
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string `json:"accessToken"`
	User        User   `json:"user"`
}

type RefreshResponse struct {
	AccessToken string `json:"accessToken"`
	User        User   `json:"user"`
}

type AtticEntry struct {
	Revision int64 `json:"rev"`
}

// TrashEntry represents a deleted page in the trash
type TrashEntry struct {
	Url       string      `json:"url"`
	DeletedAt int64       `json:"deletedAt"`
	Meta      ContentMeta `json:"meta"`
}

type GetTrashListResponse struct {
	Items      []TrashEntry `json:"items"`
	TotalCount int          `json:"totalCount"`
	Page       int          `json:"page"`
	Limit      int          `json:"limit"`
}

type GetTrashPageResponse struct {
	Page Page `json:"page"`
}

type TrashActionRequest struct {
	Items []TrashItemRef `json:"items"`
}

type TrashItemRef struct {
	Url       string `json:"url"`
	DeletedAt int64  `json:"deletedAt"`
}

type Page struct {
	Url     string      `json:"url"`
	Content string      `json:"content"`
	Meta    ContentMeta `json:"meta"`
}

type Folder struct {
	Url     string        `json:"url"`
	Content []FolderEntry `json:"content"`
	Meta    ContentMeta   `json:"meta"`
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

type FolderEntry struct {
	Url      string `json:"url"`
	Name     string `json:"name"`
	Title    string `json:"title"`
	IsFolder bool   `json:"isFolder"`

	ACL *[]AccessRule `json:"-"`
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

// ValidContentOps are the allowed operations for content (page/folder) ACLs
var ValidContentOps = []AccessOp{AccessOpRead, AccessOpWrite, AccessOpDelete}

// ValidConfigOps are the allowed operations for global config ACLs
var ValidConfigOps = []AccessOp{AccessOpAdmin, AccessOpRegister}

type User struct {
	ID           string `json:"id" yaml:"id"`
	Username     string `json:"username" yaml:"username"`
	PasswordHash string `json:"-" yaml:"passwordHash"`
	DisplayName  string `json:"displayName" yaml:"displayName"`
}

type Config struct {
	ACL       []AccessRule    `json:"acl" yaml:"acl"`
	AppTitle  string          `json:"appTitle" yaml:"appTitle"`
	JwtSecret string          `json:"-" yaml:"jwtSecret"`
	SetupMode bool            `json:"setupMode" yaml:"setupMode"`
	Retention RetentionConfig `json:"retention" yaml:"retention"`
}

// RetentionConfig defines automatic cleanup policies for trash and version history
type RetentionConfig struct {
	Trash TrashRetention `json:"trash" yaml:"trash"`
	Attic AtticRetention `json:"attic" yaml:"attic"`
}

// TrashRetention defines the retention policy for deleted items in trash
type TrashRetention struct {
	// MaxAgeDays specifies the maximum age in days for trash items.
	// Items older than this will be permanently deleted.
	// 0 means disabled (keep forever).
	MaxAgeDays int `json:"maxAgeDays" yaml:"maxAgeDays"`
}

// AtticRetention defines the retention policy for version history
type AtticRetention struct {
	// MaxAgeDays specifies the maximum age in days for versions.
	// Versions older than this will be deleted.
	// 0 means disabled (keep forever).
	MaxAgeDays int `json:"maxAgeDays" yaml:"maxAgeDays"`

	// MaxVersions specifies the maximum number of versions to keep per page.
	// Older versions beyond this limit will be deleted.
	// 0 means unlimited.
	MaxVersions int `json:"maxVersions" yaml:"maxVersions"`
}

type SearchHit struct {
	Url          string              `json:"url"`
	Meta         ContentMeta         `json:"meta"`
	Fragments    map[string][]string `json:"fragments"`
	EffectiveACL []AccessRule        `json:"-"`
	IsFolder     bool                `json:"isFolder"`
}

type SearchResponse struct {
	Items   []SearchHit `json:"items"`
	Page    int         `json:"page"`
	Limit   int         `json:"limit"`
	HasMore bool        `json:"hasMore"`
}
