package model

import (
	"encoding/json"
)

type Breadcrumb struct {
	Name string `json:"name"`
	Url  string `json:"url"`
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
	Version       string `json:"version"`
	GitSha        string `json:"gitSha"`
}

type PutRequest struct {
	Page *Page `json:"page"`
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

type TokenUserResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
