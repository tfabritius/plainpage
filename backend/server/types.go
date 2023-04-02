package server

import (
	"encoding/json"

	"github.com/tfabritius/plainpage/storage"
)

type Breadcrumb struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type GetAppResponse struct {
	AppName string `json:"appName"`
}

type GetPageResponse struct {
	Page        *storage.Page   `json:"page"`
	Folder      *storage.Folder `json:"folder"`
	AllowCreate bool            `json:"allowCreate"`
	Breadcrumbs []Breadcrumb    `json:"breadcrumbs"`
}

type GetAtticListResponse struct {
	Entries     []storage.AtticEntry `json:"entries"`
	Breadcrumbs []Breadcrumb         `json:"breadcrumbs"`
}

type PutRequest struct {
	Page *storage.Page `json:"page"`
}

type PostUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	RealName string `json:"realName"`
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

type TokenUserResponse struct {
	Token string       `json:"token"`
	User  storage.User `json:"user"`
}
