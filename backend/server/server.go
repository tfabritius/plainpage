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
