package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/tfabritius/plainpage/model"
	"github.com/tfabritius/plainpage/service"
	"github.com/tfabritius/plainpage/service/ctxutil"
)

/**
 * Checks URL is a valid URL for a page or folder. Valid URL consist of parts separated by /.
 * Each part may only contain letters, digits, dash and underscore, but must not start with underscore.
 */
func isValidUrl(urlPath string) bool {
	urlRegex := regexp.MustCompile("^[a-z0-9-][a-z0-9_-]*(/[a-z0-9-][a-z0-9_-]*)*$")
	return urlPath == "" || urlRegex.MatchString(urlPath)
}

func (App) getBreadcrumbs(urlPath string, page *model.Page, folder *model.Folder, ancestorsMeta []model.ContentMetaWithURL) []model.Breadcrumb {
	breadcrumbs := []model.Breadcrumb{}

	for i := len(ancestorsMeta) - 1; i >= 0; i-- {
		if ancestorsMeta[i].Url == "" {
			continue
		}

		breadcrumb := model.Breadcrumb{
			Url:   ancestorsMeta[i].Url,
			Title: ancestorsMeta[i].Title,
			Name:  path.Base(ancestorsMeta[i].Url),
		}

		breadcrumbs = append(breadcrumbs, breadcrumb)
	}

	if page != nil {
		breadcrumbs = append(breadcrumbs, model.Breadcrumb{
			Url:   urlPath,
			Title: page.Meta.Title,
			Name:  path.Base(urlPath),
		})
	} else if folder != nil && urlPath != "" {
		breadcrumbs = append(breadcrumbs, model.Breadcrumb{
			Url:   urlPath,
			Title: folder.Meta.Title,
			Name:  path.Base(urlPath),
		})
	}

	return breadcrumbs
}

func (app App) getContent(w http.ResponseWriter, r *http.Request) {
	urlPath := chi.URLParam(r, "*")

	userID := ctxutil.UserID(r.Context())
	page := ctxutil.Page(r.Context())
	folder := ctxutil.Folder(r.Context())
	metas := ctxutil.AncestorsMeta(r.Context())

	response := model.GetContentResponse{}

	parentAcl := app.Content.GetEffectivePermissions(model.ContentMeta{ACL: nil}, metas)

	response.AllowWrite = app.Users.CheckContentPermissions(parentAcl, userID, model.AccessOpWrite) == nil
	response.AllowDelete = app.Users.CheckContentPermissions(parentAcl, userID, model.AccessOpDelete) == nil

	response.Breadcrumbs = app.getBreadcrumbs(urlPath, page, folder, metas)

	if page != nil {
		if app.isAdmin(userID) {
			if err := app.Users.EnhanceACLWithUserInfo(page.Meta.ACL); err != nil {
				panic(err)
			}
		} else {
			page.Meta.ACL = nil // Hide ACL
		}

		response.Page = page
	} else if folder != nil {
		if app.isAdmin(userID) {
			if err := app.Users.EnhanceACLWithUserInfo(folder.Meta.ACL); err != nil {
				panic(err)
			}
		} else {
			folder.Meta.ACL = nil // Hide ACL
		}

		response.Folder = folder
	} else {
		// Not found
		w.WriteHeader(http.StatusNotFound)

		if !isValidUrl(urlPath) {
			response.AllowWrite = false
		} else {
			parentUrl, err := url.JoinPath(urlPath, "..")
			if err != nil {
				panic(err)
			}

			if !app.Content.IsFolder(parentUrl) {
				response.AllowWrite = false
			}
		}

		if !response.AllowWrite {
			response.Breadcrumbs = nil
		}
	}

	render.JSON(w, r, response)
}

func (app App) patchContent(w http.ResponseWriter, r *http.Request) {
	urlPath := chi.URLParam(r, "*")

	page := ctxutil.Page(r.Context())
	folder := ctxutil.Folder(r.Context())

	if !isValidUrl(urlPath) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Poor man's implementation of RFC 6902
	var operations []model.PatchOperation
	if err := json.NewDecoder(r.Body).Decode(&operations); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	isFolder := folder != nil

	if page != nil && folder != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	for _, operation := range operations {
		if operation.Op != "replace" {
			http.Error(w, "operation "+operation.Op+" not supported", http.StatusBadRequest)
			return
		}

		var acl []model.AccessRule
		if operation.Value != nil {
			err := json.Unmarshal([]byte(*operation.Value), &acl)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		if isFolder && operation.Path == "/folder/meta/acl" {
			if operation.Value == nil {
				folder.Meta.ACL = nil
			} else {
				folder.Meta.ACL = &acl
			}
		} else if !isFolder && operation.Path == "/page/meta/acl" {
			if operation.Value == nil {
				page.Meta.ACL = nil
			} else {
				page.Meta.ACL = &acl
			}
		} else {
			http.Error(w, "path "+operation.Path+" not supported", http.StatusBadRequest)
			return
		}
	}

	var err error
	if isFolder {
		err = app.Content.SaveFolder(urlPath, folder.Meta)
	} else {
		err = app.Content.SavePage(urlPath, page.Content, page.Meta)
	}

	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
}

func (app App) putContent(w http.ResponseWriter, r *http.Request) {
	urlPath := chi.URLParam(r, "*")

	page := ctxutil.Page(r.Context())
	folder := ctxutil.Folder(r.Context())

	if !isValidUrl(urlPath) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var body model.PutRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var err error
	if body.Page != nil {
		if page != nil {
			// if page exists already, take over ACL
			body.Page.Meta.ACL = page.Meta.ACL
		}

		err = app.Content.SavePage(urlPath, body.Page.Content, body.Page.Meta)
	} else if body.Folder != nil {
		if folder != nil {
			// if folder exists already, take over ACL
			body.Folder.Meta.ACL = folder.Meta.ACL

			// and update
			err = app.Content.SaveFolder(urlPath, body.Folder.Meta)
		} else {
			// make sure ACLs are not set
			body.Folder.Meta.ACL = nil

			// and create
			err = app.Content.CreateFolder(urlPath, body.Folder.Meta)
		}
	} else {
		http.Error(w, "Content missing", http.StatusBadRequest)
		return
	}

	if err != nil {
		if errors.Is(err, model.ErrParentFolderNotFound) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if errors.Is(err, model.ErrPageOrFolderExistsAlready) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
}

func (app App) deleteContent(w http.ResponseWriter, r *http.Request) {
	urlPath := chi.URLParam(r, "*")

	page := ctxutil.Page(r.Context())
	folder := ctxutil.Folder(r.Context())

	if !isValidUrl(urlPath) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var err error
	if page != nil {
		err = app.Content.DeletePage(urlPath)

	} else if folder != nil {
		err = app.Content.DeleteEmptyFolder(urlPath)

	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if err != nil {
		if errors.Is(err, model.ErrFolderNotEmpty) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
}

func (app App) getAttic(w http.ResponseWriter, r *http.Request) {
	urlPath := chi.URLParam(r, "*")
	queryRev := r.URL.Query().Get("rev")

	page := ctxutil.Page(r.Context())
	ancestorsMeta := ctxutil.AncestorsMeta(r.Context())

	if !isValidUrl(urlPath) || page == nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	breadcrumbs := app.getBreadcrumbs(urlPath, page, nil, ancestorsMeta)

	var revision int64
	if queryRev == "" {
		list, err := app.Content.ListAttic(urlPath)
		if err != nil {
			panic(err)
		}

		response := model.GetAtticListResponse{
			Entries:     list,
			Breadcrumbs: breadcrumbs,
		}
		render.JSON(w, r, response)
	} else {
		var err error
		revision, err = strconv.ParseInt(queryRev, 10, 64)
		if err != nil {
			http.Error(w, "Invalid query parameter: rev", http.StatusBadRequest)
			return
		}

		if !app.Content.IsAtticPage(urlPath, revision) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		page, err := app.Content.ReadPage(urlPath, &revision)
		if err != nil {
			panic(err)
		}

		response := model.GetContentResponse{Page: &page, Breadcrumbs: breadcrumbs}
		render.JSON(w, r, response)
	}
}

func (app App) searchContent(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	userID := ctxutil.UserID(r.Context())

	results, err := app.Content.Search(q)
	if err != nil {
		panic(err)
	}

	// Filter results to only those accessible to the user
	accessibleResults := []model.SearchHit{}
	for _, r := range results {

		if err := app.Users.CheckContentPermissions(r.EffectiveACL, userID, model.AccessOpRead); err != nil {
			if _, ok := err.(*service.AccessDeniedError); ok {
				// Skip this result
				continue
			}

			panic(err)
		}

		r.Meta.ACL = nil // Hide ACL
		accessibleResults = append(accessibleResults, r)
	}

	// Limit to 10 results
	if len(accessibleResults) > 10 {
		accessibleResults = accessibleResults[:10]
	}

	render.JSON(w, r, accessibleResults)
}
