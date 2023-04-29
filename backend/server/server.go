package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/tfabritius/plainpage/libs/spa"
	"github.com/tfabritius/plainpage/libs/utils"
	"github.com/tfabritius/plainpage/model"
	"github.com/tfabritius/plainpage/service"
)

type App struct {
	Frontend http.FileSystem
	Storage  model.Storage
	Content  service.ContentService
	Users    service.UserService
	Token    service.TokenService
}

func NewApp(staticFrontendFiles http.FileSystem, store model.Storage) App {
	if !store.Exists("config.yml") {
		log.Println("Initializing config...")
		cfg := initializeConfig()

		if err := store.WriteConfig(cfg); err != nil {
			panic(err)
		}
	}

	cfg, err := store.ReadConfig()
	if err != nil {
		panic(fmt.Errorf("could not load config: %w", err))
	}

	contentService := service.NewContentService(store)
	userService := service.NewUserService(store)
	tokenService := service.NewTokenService(cfg.JwtSecret)

	return App{
		Frontend: staticFrontendFiles,
		Storage:  store,
		Content:  contentService,
		Users:    userService,
		Token:    tokenService,
	}
}

// initializeConfig creates default configuration on first start
func initializeConfig() model.Config {
	cfg := model.Config{}
	var err error

	cfg.AppTitle = "PlainPage"

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

			r.With(app.RequireAdminPermission).Get("/config", app.getConfig)
			r.With(app.RequireAdminPermission).Patch("/config", app.patchConfig)

			r.Route("/pages", func(r chi.Router) {
				r.Get("/*", app.RequireContentPermission(model.AccessOpRead, http.HandlerFunc(app.getContent)).ServeHTTP)
				r.Put("/*", app.RequireContentPermission(model.AccessOpWrite, http.HandlerFunc(app.putContent)).ServeHTTP)
				r.Delete("/*", app.RequireContentPermission(model.AccessOpDelete, http.HandlerFunc(app.deleteContent)).ServeHTTP)

				r.Patch("/*", app.RequireContentPermission(model.AccessOpAdmin, http.HandlerFunc(app.patchContent)).ServeHTTP)
			})

			r.Route("/attic", func(r chi.Router) {
				r.Get("/*", app.RequireContentPermission(model.AccessOpRead, http.HandlerFunc(app.getAttic)).ServeHTTP)
			})

			r.Route("/auth", func(r chi.Router) {
				r.With(app.RequireAdminPermission).Get("/users", app.getUsers)
				r.With(app.RequireAdminPermission).Get("/users/{username:[a-zA-Z0-9_-]+}", app.getUser)
				r.Post("/users", app.postUser)
				r.With(app.RequireAuth).Patch("/users/{username:[a-zA-Z0-9_-]+}", app.patchUser)
				r.With(app.RequireAuth).Delete("/users/{username:[a-zA-Z0-9_-]+}", app.deleteUser)

				r.Post("/login", app.login)
				r.Post("/refresh", app.refreshToken)

			})

		})

	serveFallback := spa.ServeFileContents("index.html", app.Frontend)
	r.With(spa.Catch404Middleware(serveFallback)).
		Handle("/*", http.FileServer(app.Frontend))

	return r
}
