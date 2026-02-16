package server

import (
	"errors"
	"net"
	"net/http"
	"strconv"

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
			var e *service.AccessDeniedError
			if errors.As(err, &e) {
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
		var accessDeniedErr *service.AccessDeniedError
		if errors.As(err, &accessDeniedErr) {
			return false
		}

		panic(err)
	}
	return true
}

func clientIPFromRequest(r *http.Request) string {
	ip := r.RemoteAddr

	// remove port if present
	if host, _, err := net.SplitHostPort(ip); err == nil {
		ip = host
	}

	return ip
}

// SearchRateLimitMiddleware applies rate limiting to search requests.
// Uses stricter limits for unauthenticated users (by IP) and more lenient limits for authenticated users (by userID).
func (app App) SearchRateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := ctxutil.UserID(r.Context())

		var allowed bool
		var retryAfter int

		if userID != "" {
			// Authenticated user: rate limit by user ID
			allowed, retryAfter = app.SearchLimiterByUser.Allow(userID)
		} else {
			// Unauthenticated user: rate limit by IP
			ip := clientIPFromRequest(r)
			allowed, retryAfter = app.SearchLimiterByIP.Allow(ip)
		}

		if !allowed {
			if retryAfter < 1 {
				retryAfter = 1
			}
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// populateModifiedByUserInfo populates ModifiedByUsername and ModifiedByDisplayName from ModifiedByUserID
func (app App) populateModifiedByUserInfo(meta *model.ContentMeta) {
	if meta.ModifiedByUserID == "" {
		meta.ModifiedByUsername = ""
		meta.ModifiedByDisplayName = ""
		return
	}

	user, err := app.Users.GetById(meta.ModifiedByUserID)
	if err != nil {
		// User not found (possibly deleted), clear the fields
		meta.ModifiedByUsername = ""
		meta.ModifiedByDisplayName = ""
		return
	}

	meta.ModifiedByUsername = user.Username
	meta.ModifiedByDisplayName = user.DisplayName
}
