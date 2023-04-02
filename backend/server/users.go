package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/tfabritius/plainpage/service/ctxutil"
	"github.com/tfabritius/plainpage/storage"
)

func (app App) getUsers(w http.ResponseWriter, r *http.Request) {
	users, err := app.Storage.GetAllUsers()
	if err != nil {
		panic(err)
	}

	render.JSON(w, r, users)
}

func (app App) getUser(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	user, err := app.Users.GetByUsername(username)
	if errors.Is(err, storage.ErrNotFound) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if err != nil {
		panic(err)
	}

	render.JSON(w, r, user)
}

func (app App) postUser(w http.ResponseWriter, r *http.Request) {
	var body PostUserRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := app.Users.Create(body.Username, body.Password, body.RealName)
	if err != nil {
		if errors.Is(err, storage.ErrInvalidUsername) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, storage.ErrUserExistsAlready) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		panic(err)
	}

	render.JSON(w, r, user)
}

func (app App) patchUser(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	// Poor man's implementation of RFC 6902
	var operations []PatchOperation
	if err := json.NewDecoder(r.Body).Decode(&operations); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := app.Users.GetByUsername(username)
	if errors.Is(err, storage.ErrNotFound) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if err != nil {
		panic(err)
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

		if operation.Path == "/username" {
			if err := app.Users.SetUsername(&user, value); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		} else if operation.Path == "/realName" {
			user.RealName = value
		} else if operation.Path == "/password" {
			if err := app.Users.SetPasswordHash(&user, value); err != nil {
				panic(err)
			}
		} else {
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
	username := chi.URLParam(r, "username")

	err := app.Users.DeleteByUsername(username)
	if errors.Is(err, storage.ErrNotFound) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
}

func (app App) login(w http.ResponseWriter, r *http.Request) {
	var body LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := app.Users.VerifyCredentials(body.Username, body.Password)
	if err != nil {
		panic(err)
	}

	if user == nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	}

	token, err := app.Token.GenerateToken(*user)
	if err != nil {
		panic(err)
	}

	response := TokenUserResponse{
		Token: token,
		User:  *user,
	}

	render.JSON(w, r, response)
}

func (app App) refreshToken(w http.ResponseWriter, r *http.Request) {
	id := ctxutil.UserID(r.Context())
	if id == "" {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	user, err := app.Users.GetById(id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		panic(err)
	}

	token, err := app.Token.GenerateToken(user)
	if err != nil {
		panic(err)
	}

	response := TokenUserResponse{
		Token: token,
		User:  user,
	}

	render.JSON(w, r, response)
}