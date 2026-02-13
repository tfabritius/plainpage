package model

import (
	"encoding/json"
)

type Breadcrumb struct {
	Name  string `json:"name"`
	Title string `json:"title"`
	Url   string `json:"url"`
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

type PostUserRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
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
