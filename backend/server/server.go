package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
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
	"github.com/tfabritius/plainpage/libs/utils"
	"github.com/tfabritius/plainpage/service"
	"github.com/tfabritius/plainpage/service/ctxutil"
	"github.com/tfabritius/plainpage/storage"
)

type App struct {
	Frontend http.FileSystem
	Storage  storage.Storage
	Users    service.UserService
	Token    service.TokenService
}

func NewApp(staticFrontendFiles http.FileSystem, store storage.Storage) App {
	cfg, err := store.ReadConfig()
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			log.Println("Initializing config...")
			cfg = initializeConfig()

			if err := store.WriteConfig(cfg); err != nil {
				panic(err)
			}
		} else {
			panic(fmt.Errorf("could not load config: %w", err))
		}
	}

	userService := service.NewUserService(store)
	tokenService := service.NewTokenService(cfg.JwtSecret)

	return App{
		Frontend: staticFrontendFiles,
		Storage:  store,
		Users:    userService,
		Token:    tokenService,
	}
}

// initializeConfig creates default configuration on first start
func initializeConfig() storage.Config {
	cfg := storage.Config{}
	var err error

	cfg.AppName = "PlainPage"

	cfg.JwtSecret, err = utils.GenerateRandomString(16)
	if err != nil {
		panic(err)
	}

	return cfg
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

	r.
		With(app.Token.Token2ContextMiddleware).
		Route("/_api", func(r chi.Router) {
			r.Get("/app", app.exposeConfig)

			r.Get("/config", app.getConfig)
			r.Patch("/config", app.patchConfig)

			r.Route("/pages", func(r chi.Router) {
				r.Get("/*", app.getPageOrFolder)
				r.Put("/*", app.putPageOrFolder)
				r.Delete("/*", app.deletePageOrFolder)

				r.Patch("/*", app.patchPageOrFolder)
			})

			r.Route("/attic", func(r chi.Router) {
				r.Get("/*", app.getAttic)
			})

			r.Route("/auth", func(r chi.Router) {
				r.Get("/users", app.getUsers)
				r.Get("/users/{username:[a-zA-Z0-9_-]+}", app.getUser)
				r.Post("/users", app.postUser)
				r.Patch("/users/{username:[a-zA-Z0-9_-]+}", app.patchUser)
				r.Delete("/users/{username:[a-zA-Z0-9_-]+}", app.deleteUser)

				r.Post("/login", app.login)
				r.Post("/refresh", app.refreshToken)

			})

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
