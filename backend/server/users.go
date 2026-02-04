package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/tfabritius/plainpage/model"
	"github.com/tfabritius/plainpage/service"
	"github.com/tfabritius/plainpage/service/ctxutil"
)

const (
	refreshTokenCookieName = "refresh_token"
	refreshTokenCookiePath = "/_api/auth"
)

// refreshTokenCookieMaxAge derives the cookie MaxAge from the service's RefreshTokenValidity
var refreshTokenCookieMaxAge = int(service.RefreshTokenValidity.Seconds())

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

	// Generate access token
	accessToken, err := app.AccessToken.Create(user.ID)
	if err != nil {
		panic(err)
	}

	// Generate refresh token and store it
	refreshToken, err := app.RefreshToken.Create(user.ID)
	if err != nil {
		panic(err)
	}

	// Set refresh token as httpOnly cookie
	app.setRefreshTokenCookie(w, r, refreshToken)

	response := model.LoginResponse{
		AccessToken: accessToken,
		User:        *user,
	}

	render.JSON(w, r, response)
}

func (app App) refreshToken(w http.ResponseWriter, r *http.Request) {
	// Get refresh token from cookie
	cookie, err := r.Cookie(refreshTokenCookieName)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	refreshTokenID := cookie.Value

	// Validate refresh token
	userID, err := app.RefreshToken.Validate(refreshTokenID)
	if err != nil {
		// Clear invalid cookie
		app.clearRefreshTokenCookie(w, r)
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// Get user
	user, err := app.Users.GetById(userID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			// User no longer exists, revoke token
			_ = app.RefreshToken.Delete(refreshTokenID)
			app.clearRefreshTokenCookie(w, r)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		panic(err)
	}

	// Update refresh token's lastUsedAt and expiresAt
	if err := app.RefreshToken.Refresh(refreshTokenID); err != nil {
		panic(err)
	}

	// Generate new access token
	accessToken, err := app.AccessToken.Create(user.ID)
	if err != nil {
		panic(err)
	}

	// Refresh the cookie expiration
	app.setRefreshTokenCookie(w, r, refreshTokenID)

	response := model.RefreshResponse{
		AccessToken: accessToken,
		User:        user,
	}

	render.JSON(w, r, response)
}

func (app App) logout(w http.ResponseWriter, r *http.Request) {
	// Get refresh token from cookie
	cookie, err := r.Cookie(refreshTokenCookieName)
	if err != nil {
		// No cookie, nothing to do
		w.WriteHeader(http.StatusOK)
		return
	}

	// Delete refresh token from storage
	_ = app.RefreshToken.Delete(cookie.Value)

	// Clear the cookie
	app.clearRefreshTokenCookie(w, r)

	w.WriteHeader(http.StatusOK)
}

func (app App) setRefreshTokenCookie(w http.ResponseWriter, r *http.Request, token string) {
	secure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"

	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    token,
		Path:     refreshTokenCookiePath,
		MaxAge:   refreshTokenCookieMaxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (app App) clearRefreshTokenCookie(w http.ResponseWriter, r *http.Request) {
	secure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"

	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    "",
		Path:     refreshTokenCookiePath,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (app App) changePassword(w http.ResponseWriter, r *http.Request) {
	userID := ctxutil.UserID(r.Context())
	username := r.PathValue("username")

	// Parse request body
	var body model.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the logged-in user for password verification
	loggedInUser, err := app.Users.GetById(userID)
	if err != nil {
		panic(err)
	}

	// Verify current password against the LOGGED-IN user's password
	valid, err := app.Users.VerifyPassword(&loggedInUser, body.CurrentPassword)
	if err != nil {
		panic(err)
	}
	if !valid {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	// Authorization check
	if username != loggedInUser.Username && !app.isAdmin(userID) {
		// Non-admins can only change their own password
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	// Get target user
	var targetUser model.User
	if username == loggedInUser.Username {
		// Optimize for the common case where users change their own password
		targetUser = loggedInUser
	} else {
		var err error
		targetUser, err = app.Users.GetByUsername(username)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			panic(err)
		}
	}

	// Set new password for target user and save
	if err := app.Users.SetPasswordHash(&targetUser, body.NewPassword); err != nil {
		panic(err)
	}
	if err := app.Users.Save(targetUser); err != nil {
		panic(err)
	}

	// Revoke all refresh tokens for the target user (security measure)
	if err := app.RefreshToken.DeleteAllForUser(targetUser.ID); err != nil {
		// Log error but don't fail the request
		// The password change itself succeeded
		log.Printf("[background] could not revoke refresh tokens for user %s: %v", targetUser.ID, err)
	}

	w.WriteHeader(http.StatusOK)
}
