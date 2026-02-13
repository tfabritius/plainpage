package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"

	"github.com/go-chi/render"
	"github.com/tfabritius/plainpage/model"
)

func (app App) getTrash(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	pageNum := 1
	limit := 20

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			pageNum = parsed
		}
	}

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	// Parse sort parameters
	sortBy := r.URL.Query().Get("sortBy")
	if sortBy != "url" && sortBy != "deletedAt" {
		sortBy = "deletedAt" // default
	}

	sortOrder := r.URL.Query().Get("sortOrder")
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc" // default
	}

	entries, err := app.Content.ListTrash()
	if err != nil {
		panic(err)
	}

	// Sort entries
	sort.Slice(entries, func(i, j int) bool {
		if sortBy == "url" {
			if sortOrder == "asc" {
				return entries[i].Url < entries[j].Url
			}
			return entries[i].Url > entries[j].Url
		}
		// sortBy == "deletedAt"
		if sortOrder == "asc" {
			return entries[i].DeletedAt < entries[j].DeletedAt
		}
		return entries[i].DeletedAt > entries[j].DeletedAt
	})

	totalCount := len(entries)

	// Apply pagination
	start := (pageNum - 1) * limit
	end := start + limit

	if start > len(entries) {
		entries = []model.TrashEntry{}
	} else {
		if end > len(entries) {
			end = len(entries)
		}
		entries = entries[start:end]
	}

	// Populate user info for metadata
	for i := range entries {
		app.populateModifiedByUserInfo(&entries[i].Meta)
	}

	response := model.GetTrashListResponse{
		Items:      entries,
		TotalCount: totalCount,
		Page:       pageNum,
		Limit:      limit,
	}

	render.JSON(w, r, response)
}

func (app App) getTrashPage(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Query().Get("url")
	deletedAtStr := r.URL.Query().Get("deletedAt")

	if urlPath == "" {
		http.Error(w, "url parameter is required", http.StatusBadRequest)
		return
	}

	if deletedAtStr == "" {
		http.Error(w, "deletedAt parameter is required", http.StatusBadRequest)
		return
	}

	deletedAt, err := strconv.ParseInt(deletedAtStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid deletedAt parameter", http.StatusBadRequest)
		return
	}

	page, err := app.Content.ReadTrashPage(urlPath, deletedAt)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		panic(err)
	}

	// Populate user info for metadata
	app.populateModifiedByUserInfo(&page.Meta)

	response := model.GetTrashPageResponse{
		Page: page,
	}

	render.JSON(w, r, response)
}

func (app App) deleteTrashItems(w http.ResponseWriter, r *http.Request) {
	var req model.TrashActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, item := range req.Items {
		if err := app.Content.DeleteTrashEntry(item.Url, item.DeletedAt); err != nil {
			if errors.Is(err, model.ErrNotFound) {
				http.Error(w, "item not found: "+item.Url, http.StatusNotFound)
				return
			}
			panic(err)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (app App) restoreTrashItems(w http.ResponseWriter, r *http.Request) {
	var req model.TrashActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, item := range req.Items {
		if err := app.Content.RestoreFromTrash(item.Url, item.DeletedAt); err != nil {
			if errors.Is(err, model.ErrNotFound) {
				http.Error(w, "item not found: "+item.Url, http.StatusNotFound)
				return
			}
			if errors.Is(err, model.ErrPageOrFolderExistsAlready) {
				http.Error(w, "destination already exists: "+item.Url, http.StatusConflict)
				return
			}
			panic(err)
		}
	}

	w.WriteHeader(http.StatusOK)
}
