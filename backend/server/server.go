package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
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

	cfg.SetupMode = true

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

			r.With(app.RequireAppPermission(storage.AccessOpAdmin)).Get("/config", app.getConfig)
			r.With(app.RequireAppPermission(storage.AccessOpAdmin)).Patch("/config", app.patchConfig)

			r.Route("/pages", func(r chi.Router) {
				r.Get("/*", app.RequireContentPermission(storage.AccessOpRead, http.HandlerFunc(app.getPageOrFolder)).ServeHTTP)
				r.Put("/*", app.RequireContentPermission(storage.AccessOpWrite, http.HandlerFunc(app.putPageOrFolder)).ServeHTTP)
				r.Delete("/*", app.RequireContentPermission(storage.AccessOpDelete, http.HandlerFunc(app.deletePageOrFolder)).ServeHTTP)

				r.Patch("/*", app.RequireContentPermission(storage.AccessOpAdmin, http.HandlerFunc(app.patchPageOrFolder)).ServeHTTP)
			})

			r.Route("/attic", func(r chi.Router) {
				r.Get("/*", app.RequireContentPermission(storage.AccessOpRead, http.HandlerFunc(app.getAttic)).ServeHTTP)
			})

			r.Route("/auth", func(r chi.Router) {
				r.With(app.RequireAppPermission(storage.AccessOpAdmin)).Get("/users", app.getUsers)
				r.With(app.RequireAppPermission(storage.AccessOpAdmin)).Get("/users/{username:[a-zA-Z0-9_-]+}", app.getUser)
				r.Post("/users", app.postUser)
				r.With(app.RequireAppPermission(storage.AccessOpAdmin)).Patch("/users/{username:[a-zA-Z0-9_-]+}", app.patchUser)
				r.With(app.RequireAppPermission(storage.AccessOpAdmin)).Delete("/users/{username:[a-zA-Z0-9_-]+}", app.deleteUser)

				r.Post("/login", app.login)
				r.Post("/refresh", app.refreshToken)

			})

		})

	serveFallback := spa.ServeFileContents("index.html", app.Frontend)
	r.With(spa.Catch404Middleware(serveFallback)).
		Handle("/*", http.FileServer(app.Frontend))

	return r
}

func (app App) RequireAppPermission(op storage.AccessOp) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			userID := ctxutil.UserID(r.Context())

			if err := app.Users.CheckAppPermissions(userID, op); err != nil {
				if e, ok := err.(*service.AccessDeniedError); ok {
					http.Error(w, http.StatusText(e.StatusCode), e.StatusCode)
					return
				}

				panic(err)
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (app App) RequireContentPermission(op storage.AccessOp, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := ctxutil.UserID(r.Context())

		urlPath := chi.URLParam(r, "*")

		acl, err := app.Storage.GetEffectivePermissions(urlPath)
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
