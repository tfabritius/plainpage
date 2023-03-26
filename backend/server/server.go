package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"regexp"
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
	var breadcrumbs []Breadcrumb
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

	response := GetResponse{}

	if !isValidUrl(urlPath) {
		w.WriteHeader(http.StatusNotFound)
	} else {

		response.Breadcrumbs = getBreadcrumbs(urlPath)

		if app.Storage.IsPage(urlPath) {
			page, err := app.Storage.ReadPage(urlPath)
			if err != nil {
				panic(err)
			}
			response.Page = &page
		} else if app.Storage.IsFolder(urlPath) {
			folder, err := app.Storage.ReadFolder(urlPath)
			if err != nil {
				panic(err)
			}
			response.Folder = folder
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
