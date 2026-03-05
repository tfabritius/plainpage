package server

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"

	"github.com/go-chi/render"
	"github.com/tfabritius/plainpage/model"
	"github.com/tfabritius/plainpage/service"
)

// downloadStorage creates a ZIP archive of all storage data and streams it to the response
func (app App) downloadStorage(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters for optional inclusions
	opts := service.BackupOptions{
		IncludeConfig: r.URL.Query().Has("includeConfig"),
		IncludeUsers:  r.URL.Query().Has("includeUsers"),
	}

	w.Header().Set("Content-Type", "application/zip")

	// Delegate to content service
	if err := app.Content.WriteBackup(w, opts); err != nil {
		http.Error(w, "Failed to create backup", http.StatusInternalServerError)
	}
}

// restoreStorage restores a backup from an uploaded ZIP file
func (app App) restoreStorage(w http.ResponseWriter, r *http.Request) {
	// Read the entire body as raw bytes
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Could not read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(body) == 0 {
		http.Error(w, "No backup file provided", http.StatusBadRequest)
		return
	}

	// Create zip reader from bytes
	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		http.Error(w, "Invalid ZIP file", http.StatusBadRequest)
		return
	}

	// Restore backup
	usersRestored, err := app.Content.RestoreBackup(zipReader)
	if err != nil {
		http.Error(w, "Failed to restore backup: "+err.Error(), http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, model.RestoreBackupResponse{
		UsersRestored: usersRestored,
	})
}
