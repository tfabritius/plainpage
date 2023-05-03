package server

import (
	"encoding/json"
	"net/http"
	"runtime/debug"

	"github.com/go-chi/render"
	"github.com/tfabritius/plainpage/model"
	"github.com/tfabritius/plainpage/service/ctxutil"
)

func getRevision() string {
	info, _ := debug.ReadBuildInfo()

	for _, kv := range info.Settings {
		if kv.Key == "vcs.revision" {
			return kv.Value
		}
	}

	return ""
}

func (app App) exposeConfig(w http.ResponseWriter, r *http.Request) {
	userID := ctxutil.UserID(r.Context())

	allowRegister := app.Users.CheckAppPermissions(userID, model.AccessOpRegister) == nil
	allowAdmin := app.Users.CheckAppPermissions(userID, model.AccessOpAdmin) == nil

	cfg, err := app.Storage.ReadConfig()
	if err != nil {
		panic(err)
	}

	gitSha := getRevision()

	response := model.GetAppResponse{
		AppTitle:      cfg.AppTitle,
		SetupMode:     cfg.SetupMode,
		AllowRegister: allowRegister,
		AllowAdmin:    allowAdmin,
		GitSha:        gitSha,
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

	for _, operation := range operations {
		if operation.Op != "replace" {
			http.Error(w, "operation "+operation.Op+" not supported", http.StatusBadRequest)
			return
		}

		if operation.Value == nil {
			http.Error(w, "value missing", http.StatusBadRequest)
			return
		}

		if operation.Path == "/appTitle" {
			var value string
			if err := json.Unmarshal([]byte(*operation.Value), &value); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			cfg.AppTitle = value

		} else if operation.Path == "/acl" {
			var value []model.AccessRule

			if err := json.Unmarshal([]byte(*operation.Value), &value); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			cfg.ACL = value

		} else {
			http.Error(w, "path "+operation.Path+" not supported", http.StatusBadRequest)
			return
		}
	}

	if err := app.Storage.WriteConfig(cfg); err != nil {
		panic(err)
	}

	render.JSON(w, r, cfg)
}
