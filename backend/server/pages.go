package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/tfabritius/plainpage/service"
	"github.com/tfabritius/plainpage/service/ctxutil"
	"github.com/tfabritius/plainpage/storage"
)

/**
 * Checks URL is a valid URL for a page or folder. Valid URL consist of parts separated by /.
 * Each part may only contain letters, digits, dash and underscore, but must not start with underscore.
 */
func isValidUrl(urlPath string) bool {
	urlRegex := regexp.MustCompile("^[a-z0-9-][a-z0-9_-]*(/[a-z0-9-][a-z0-9_-]*)*$")
	return urlPath == "" || urlRegex.MatchString(urlPath)
}

func getBreadcrumbs(urlPath string) []Breadcrumb {
	breadcrumbs := []Breadcrumb{}
	paths := strings.Split(urlPath, "/")
	currentPath := ""
	for _, path := range paths {
		if path != "" {
			currentPath += "/" + path
			breadcrumb := Breadcrumb{
				Name: path,
				Url:  currentPath,
			}
			breadcrumbs = append(breadcrumbs, breadcrumb)
		}
	}
	return breadcrumbs
}

func (app App) getPageOrFolder(w http.ResponseWriter, r *http.Request) {
	urlPath := chi.URLParam(r, "*")
	userID := ctxutil.UserID(r.Context())

	response := GetPageResponse{}

	if !isValidUrl(urlPath) {
		w.WriteHeader(http.StatusNotFound)
	} else {

		// Get ACL
		acl, err := app.Storage.GetEffectivePermissions(urlPath)
		if err != nil {
			panic(err)
		}
		// Check ACL
		if err := app.Users.CheckPermissions(acl, userID, storage.AccessOpRead); err != nil {
			if e, ok := err.(*service.AccessDeniedError); ok {
				http.Error(w, http.StatusText(e.StatusCode), e.StatusCode)
				return
			}

			panic(err)
		}

		response.Breadcrumbs = getBreadcrumbs(urlPath)

		if app.Storage.IsPage(urlPath) {
			page, err := app.Storage.ReadPage(urlPath, nil)
			if err != nil {
				panic(err)
			}

			app.Users.EnhanceACLWithUserInfo(page.Meta.ACL)

			response.Page = &page
		} else if app.Storage.IsFolder(urlPath) {
			folder, err := app.Storage.ReadFolder(urlPath)
			if err != nil {
				panic(err)
			}

			app.Users.EnhanceACLWithUserInfo(folder.Meta.ACL)

			response.Folder = &folder
		} else {
			// Not found
			w.WriteHeader(http.StatusNotFound)

			if !isValidUrl(urlPath) {
				response.AllowCreate = false
			} else {
				parentUrl, err := url.JoinPath(urlPath, "..")
				if err != nil {
					panic(err)
				}

				response.AllowCreate = app.Storage.IsFolder(parentUrl)
			}

			if !response.AllowCreate {
				response.Breadcrumbs = nil
			}
		}
	}

	render.JSON(w, r, response)
}

func (app App) patchPageOrFolder(w http.ResponseWriter, r *http.Request) {
	urlPath := chi.URLParam(r, "*")

	if !isValidUrl(urlPath) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Poor man's implementation of RFC 6902
	var operations []PatchOperation
	if err := json.NewDecoder(r.Body).Decode(&operations); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var page storage.Page
	var folder storage.Folder
	isFolder := false
	var err error
	if app.Storage.IsPage(urlPath) {
		page, err = app.Storage.ReadPage(urlPath, nil)
		if err != nil {
			panic(err)
		}
	} else if app.Storage.IsFolder(urlPath) {
		folder, err = app.Storage.ReadFolder(urlPath)
		if err != nil {
			panic(err)
		}
		isFolder = true
	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	for _, operation := range operations {
		if operation.Op != "replace" {
			http.Error(w, "operation "+operation.Op+" not supported", http.StatusBadRequest)
			return
		}

		var acl []storage.AccessRule
		if operation.Value != nil {
			err = json.Unmarshal([]byte(*operation.Value), &acl)
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

	if isFolder {
		err = app.Storage.SaveFolder(urlPath, folder.Meta)
	} else {
		err = app.Storage.SavePage(urlPath, page.Content, page.Meta)
	}

	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
}

func (app App) putPageOrFolder(w http.ResponseWriter, r *http.Request) {
	urlPath := chi.URLParam(r, "*")

	if !isValidUrl(urlPath) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var body PutRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var err error
	if body.Page != nil {
		if app.Storage.IsPage(urlPath) {
			// if page exists already, take over ACL
			oldPage, err := app.Storage.ReadPage(urlPath, nil)
			if err != nil {
				panic(err)
			}
			body.Page.Meta.ACL = oldPage.Meta.ACL
		}

		err = app.Storage.SavePage(urlPath, body.Page.Content, body.Page.Meta)
	} else {
		err = app.Storage.CreateFolder(urlPath)
	}
	if err != nil {
		if errors.Is(err, storage.ErrParentFolderNotFound) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if errors.Is(err, storage.ErrPageOrFolderExistsAlready) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
}

func (app App) deletePageOrFolder(w http.ResponseWriter, r *http.Request) {
	urlPath := chi.URLParam(r, "*")

	if !isValidUrl(urlPath) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var err error
	if app.Storage.IsPage(urlPath) {
		err = app.Storage.DeletePage(urlPath)

	} else if app.Storage.IsFolder(urlPath) {
		err = app.Storage.DeleteEmptyFolder(urlPath)

	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if err != nil {
		if errors.Is(err, storage.ErrFolderNotEmpty) {
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

	if !isValidUrl(urlPath) || !app.Storage.IsPage(urlPath) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	breadcrumbs := getBreadcrumbs(urlPath)

	var revision int64
	if queryRev == "" {
		list, err := app.Storage.ListAttic(urlPath)
		if err != nil {
			panic(err)
		}

		response := GetAtticListResponse{
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

		if !app.Storage.IsAtticPage(urlPath, revision) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		page, err := app.Storage.ReadPage(urlPath, &revision)
		if err != nil {
			panic(err)
		}

		response := GetPageResponse{Page: &page, Breadcrumbs: breadcrumbs}
		render.JSON(w, r, response)
	}
}
