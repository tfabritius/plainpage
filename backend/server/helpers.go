package server

import (
	"net/http"

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

func (app App) RetrieveContentMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlPath := r.PathValue("*")

		validUrl := isValidUrl(urlPath)

		var page *model.Page
		var folder *model.Folder

		if validUrl && app.Content.IsPage(urlPath) {
			p, err := app.Content.ReadPage(urlPath, nil)
			if err != nil {
				panic(err)
			}
			page = &p
		} else if validUrl && app.Content.IsFolder(urlPath) {
			f, err := app.Content.ReadFolder(urlPath)
			if err != nil {
				panic(err)
			}

			folder = &f
		}

		metas, err := app.Content.ReadAncestorsMeta(urlPath)
		if err != nil {
			panic(err)
		}

		// Store information in request context
		ctx := r.Context()
		ctx = ctxutil.WithContent(ctx, page, folder, metas)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (app App) RequireContentPermission(op model.AccessOp, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := ctxutil.UserID(r.Context())

		page := ctxutil.Page(r.Context())
		folder := ctxutil.Folder(r.Context())
		metas := ctxutil.AncestorsMeta(r.Context())

		var meta model.ContentMeta

		if page != nil {
			meta = page.Meta
		} else if folder != nil {
			meta = folder.Meta
		}

		acl := app.Content.GetEffectivePermissions(meta, metas)

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
