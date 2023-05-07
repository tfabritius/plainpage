package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tfabritius/plainpage/model"
	"github.com/tfabritius/plainpage/service"
	"github.com/tfabritius/plainpage/service/ctxutil"
)

// RequireAuth middleware only allows access for authenticated requests
func (app App) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := ctxutil.UserID(r.Context())

		if userID == "" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireAdminPermission middleware only allows access for users with admin privileges
func (app App) RequireAdminPermission(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		userID := ctxutil.UserID(r.Context())

		if !app.isAdmin(userID) {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}

	return app.RequireAuth(http.HandlerFunc(fn))
}

func (app App) RequireContentPermission(op model.AccessOp, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := ctxutil.UserID(r.Context())

		urlPath := chi.URLParam(r, "*")

		acl, err := app.Content.GetEffectivePermissions(urlPath)
		if err != nil {
			panic(err)
		}

		if err := app.Users.CheckContentPermissions(acl, userID, op); err != nil {
			if e, ok := err.(*service.AccessDeniedError); ok {
				http.Error(w, http.StatusText(e.StatusCode), e.StatusCode)
				return
			}

			panic(err)
		}

		next.ServeHTTP(w, r)
	})
}

// isAdmin checks if the user has admin privileges. Panics on errors.
func (app App) isAdmin(userID string) bool {
	err := app.Users.CheckAppPermissions(userID, model.AccessOpAdmin)

	if err != nil {
		if _, ok := err.(*service.AccessDeniedError); ok {
			return false
		}

		panic(err)
	}
	return true
}
