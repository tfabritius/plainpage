package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/render"
	"github.com/tfabritius/plainpage/build"
	"github.com/tfabritius/plainpage/model"
	"github.com/tfabritius/plainpage/service/ctxutil"
)

func (app App) exposeConfig(w http.ResponseWriter, r *http.Request) {
	userID := ctxutil.UserID(r.Context())

	allowRegister := app.Users.CheckAppPermissions(userID, model.AccessOpRegister) == nil
	allowAdmin := app.Users.CheckAppPermissions(userID, model.AccessOpAdmin) == nil

	cfg, err := app.Storage.ReadConfig()
	if err != nil {
		panic(err)
	}

	response := model.GetAppResponse{
		AppTitle:      cfg.AppTitle,
		SetupMode:     cfg.SetupMode,
		AllowRegister: allowRegister,
		AllowAdmin:    allowAdmin,
	}

	// Only expose version info to logged-in users
	if userID != "" {
		response.GitSha = build.GetRevision()
		response.Version = build.GetVersion()
	}

	render.JSON(w, r, response)
}

func (app App) getConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := app.Storage.ReadConfig()
	if err != nil {
		panic(err)
	}

	if err := app.Users.EnhanceACLWithUserInfo(&cfg.ACL); err != nil {
		panic(err)
	}

	render.JSON(w, r, cfg)
}

func (app App) patchConfig(w http.ResponseWriter, r *http.Request) {
	var operations []model.PatchOperation
	if err := json.NewDecoder(r.Body).Decode(&operations); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cfg, err := app.Storage.ReadConfig()
	if err != nil {
		panic(err)
	}

	if err := ApplyJSONPatch(&cfg, operations); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validation: ACL rules
	if err := model.ValidateConfigACL(cfg.ACL); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validation: Retention values must be non-negative
	cfg.Retention.Trash.MaxAgeDays = max(cfg.Retention.Trash.MaxAgeDays, 0)
	cfg.Retention.Attic.MaxAgeDays = max(cfg.Retention.Attic.MaxAgeDays, 0)
	cfg.Retention.Attic.MaxVersions = max(cfg.Retention.Attic.MaxVersions, 0)

	if err := app.Storage.WriteConfig(cfg); err != nil {
		panic(err)
	}

	render.JSON(w, r, cfg)
}
