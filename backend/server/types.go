package server

import "github.com/tfabritius/plainpage/storage"

type Breadcrumb struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type GetPageResponse struct {
	Page        *storage.Page         `json:"page"`
	Folder      []storage.FolderEntry `json:"folder"`
	AllowCreate bool                  `json:"allowCreate"`
	Breadcrumbs []Breadcrumb          `json:"breadcrumbs"`
}

type GetAtticListResponse struct {
	Entries     []storage.AtticEntry `json:"entries"`
	Breadcrumbs []Breadcrumb         `json:"breadcrumbs"`
}

type PutRequest struct {
	Page *storage.Page `json:"page"`
}
