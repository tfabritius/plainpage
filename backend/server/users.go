package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/render"
	"github.com/tfabritius/plainpage/model"
	"github.com/tfabritius/plainpage/service"
	"github.com/tfabritius/plainpage/service/ctxutil"
)

func (app App) getUsers(w http.ResponseWriter, r *http.Request) {
	users, err := app.Users.ReadAll()
	if err != nil {
		panic(err)
	}

	render.JSON(w, r, users)
}

func (app App) getUser(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")

	user, err := app.Users.GetByUsername(username)
	if errors.Is(err, model.ErrNotFound) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if err != nil {
		panic(err)
	}

	render.JSON(w, r, user)
}

func (app App) postUser(w http.ResponseWriter, r *http.Request) {
	var body model.PostUserRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Read config
	cfg, err := app.Storage.ReadConfig()
	if err != nil {
		panic(err)
	}

	// Check authorization
	if !cfg.SetupMode {
		userID := ctxutil.UserID(r.Context())

		if err := app.Users.CheckAppPermissions(userID, model.AccessOpRegister); err != nil {
			var e *service.AccessDeniedError
			if errors.As(err, &e) {
				http.Error(w, http.StatusText(e.StatusCode), e.StatusCode)
				return
			}

			panic(err)
		}
	}

	// Create user
	user, err := app.Users.Create(body.Username, body.Password, body.DisplayName)
	if err != nil {
		if errors.Is(err, model.ErrInvalidUsername) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, model.ErrUserExistsAlready) {
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		}
		panic(err)
	}

	if cfg.SetupMode {
		// Terminate setup mode
		cfg.SetupMode = false

		// Grant admin rights
		cfg.ACL = append(cfg.ACL, model.AccessRule{Subject: "user:" + user.ID, Operations: []model.AccessOp{model.AccessOpAdmin}})

		// Save config
		if err := app.Storage.WriteConfig(cfg); err != nil {
			panic(err)
		}
	}

	render.JSON(w, r, user)
}

func (app App) patchUser(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	userID := ctxutil.UserID(r.Context())

	isAdmin := app.isAdmin(userID)

	user, err := app.Users.GetByUsername(username)
	userNotFound := errors.Is(err, model.ErrNotFound)
	if err != nil && !userNotFound {
		panic(err)
	}

	if isAdmin && userNotFound {
		// Admins can modify any user - if it exists
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	} else if !isAdmin {
		// Non-admins can only modify themselves
		if userNotFound || user.ID != userID {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
	}

	// Poor man's implementation of RFC 6902
	var operations []model.PatchOperation
	if err := json.NewDecoder(r.Body).Decode(&operations); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, operation := range operations {
		if operation.Op != "replace" {
			http.Error(w, "operation "+operation.Op+" not supported", http.StatusBadRequest)
			return
		}

		var value string
		if operation.Value == nil {
			http.Error(w, "value missing", http.StatusBadRequest)
			return
		} else {
			if err := json.Unmarshal([]byte(*operation.Value), &value); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		switch operation.Path {
		case "/username":
			if err := app.Users.SetUsername(&user, value); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		case "/displayName":
			user.DisplayName = value
		case "/password":
			if err := app.Users.SetPasswordHash(&user, value); err != nil {
				panic(err)
			}
		default:
			http.Error(w, "path "+operation.Path+" not supported", http.StatusBadRequest)
			return
		}
	}

	if err := app.Users.Save(user); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
}

func (app App) deleteUser(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	userID := ctxutil.UserID(r.Context())

	isAdmin := app.isAdmin(userID)

	user, err := app.Users.GetByUsername(username)
	userNotFound := errors.Is(err, model.ErrNotFound)
	if err != nil && !userNotFound {
		panic(err)
	}

	if isAdmin && userNotFound {
		// Admins can delete any user - if it exists
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	} else if !isAdmin {
		// Non-admins can only delete themselves
		if userNotFound || user.ID != userID {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
	}

	err = app.Users.DeleteByUsername(username)
	if errors.Is(err, model.ErrNotFound) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
}

func (app App) login(w http.ResponseWriter, r *http.Request) {
	var body model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := app.Users.VerifyCredentials(body.Username, body.Password)
	if err != nil {
		panic(err)
	}

	if user == nil {
		// Record failure (consume token) and return 401.
		if app.LoginLimiter != nil {
			ip := clientIPFromRequest(r)
			app.LoginLimiter.OnFailure(ip)
		}
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	token, err := app.Token.GenerateToken(*user)
	if err != nil {
		panic(err)
	}

	response := model.TokenUserResponse{
		Token: token,
		User:  *user,
	}

	render.JSON(w, r, response)
}

func (app App) refreshToken(w http.ResponseWriter, r *http.Request) {
	id := ctxutil.UserID(r.Context())

	user, err := app.Users.GetById(id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		panic(err)
	}

	token, err := app.Token.GenerateToken(user)
	if err != nil {
		panic(err)
	}

	response := model.TokenUserResponse{
		Token: token,
		User:  user,
	}

	render.JSON(w, r, response)
}
