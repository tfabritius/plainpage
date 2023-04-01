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
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/tfabritius/plainpage/libs/spa"
	"github.com/tfabritius/plainpage/storage"
)

type App struct {
	Frontend http.FileSystem
	Storage  storage.Storage
}

func NewApp(staticFrontendFiles http.FileSystem, storage storage.Storage) App {
	return App{
		Frontend: staticFrontendFiles,
		Storage:  storage,
	}
}

func (app App) GetHandler() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.RedirectSlashes)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
	}))

	r.Route("/_api/pages", func(r chi.Router) {
		r.Get("/*", app.getPageOrFolder)
		r.Put("/*", app.putPageOrFolder)
		r.Delete("/*", app.deletePageOrFolder)

		r.Patch("/*", app.patchPageOrFolder)
	})

	r.Route("/_api/attic", func(r chi.Router) {
		r.Get("/*", app.getAttic)
	})

	r.Route("/_api/auth", func(r chi.Router) {
		r.Get("/users", app.getUsers)
		r.Get("/users/{username:[a-zA-Z0-9_-]+}", app.getUser)
		r.Post("/users", app.postUser)
		r.Patch("/users/{username:[a-zA-Z0-9_-]+}", app.patchUser)
		r.Delete("/users/{username:[a-zA-Z0-9_-]+}", app.deleteUser)
	})

	serveFallback := spa.ServeFileContents("index.html", app.Frontend)
	r.With(spa.Catch404Middleware(serveFallback)).
		Handle("/*", http.FileServer(app.Frontend))

	return r
}

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

	response := GetPageResponse{}

	if !isValidUrl(urlPath) {
		w.WriteHeader(http.StatusNotFound)
	} else {

		response.Breadcrumbs = getBreadcrumbs(urlPath)

		if app.Storage.IsPage(urlPath) {
			page, err := app.Storage.ReadPage(urlPath, nil)
			if err != nil {
				panic(err)
			}
			response.Page = &page
		} else if app.Storage.IsFolder(urlPath) {
			folder, err := app.Storage.ReadFolder(urlPath)
			if err != nil {
				panic(err)
			}
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

		var acls []storage.AccessRule
		if operation.Value != nil {
			err = json.Unmarshal([]byte(*operation.Value), &acls)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		if isFolder && operation.Path == "/folder/meta/acls" {
			if operation.Value == nil {
				folder.Meta.ACLs = nil
			} else {
				folder.Meta.ACLs = &acls
			}
		} else if !isFolder && operation.Path == "/page/meta/acls" {
			if operation.Value == nil {
				page.Meta.ACLs = nil
			} else {
				page.Meta.ACLs = &acls
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
			// if page exists already, take over ACLs
			oldPage, err := app.Storage.ReadPage(urlPath, nil)
			if err != nil {
				panic(err)
			}
			body.Page.Meta.ACLs = oldPage.Meta.ACLs
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

func (app App) getUsers(w http.ResponseWriter, r *http.Request) {
	users, err := app.Storage.GetAllUsers()
	if err != nil {
		panic(err)
	}

	render.JSON(w, r, users)
}

func (app App) getUser(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	user, err := app.Storage.GetUserByUsername(username)
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

	user, err := app.Storage.AddUser(body.Username, body.Password, body.RealName)
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

	user, err := app.Storage.GetUserByUsername(username)
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
			user.Username = value
		} else if operation.Path == "/realName" {
			user.RealName = value
		} else if operation.Path == "/password" {
			user.PasswordHash = "plain:" + value
		} else {
			http.Error(w, "path "+operation.Path+" not supported", http.StatusBadRequest)
			return
		}
	}

	if err := app.Storage.SaveUser(user); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
}

func (app App) deleteUser(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	err := app.Storage.DeleteUserByUsername(username)
	if errors.Is(err, storage.ErrNotFound) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
}
