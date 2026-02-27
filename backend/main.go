package main

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/tfabritius/plainpage/build"
	"github.com/tfabritius/plainpage/server"
	"github.com/tfabritius/plainpage/service"
)

func getPathToExecutable() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalln("Error getting path to executable:", err)
	}
	return exePath[:len(exePath)-len(filepath.Base(exePath))]
}

func getDataDir() string {
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = filepath.Join(getPathToExecutable(), "data")
	} else {
		var err error
		dataDir, err = filepath.Abs(dataDir)
		if err != nil {
			panic(err)
		}
	}

	return dataDir
}

func main() {
	log.Printf("📄 Plainpage %s\n", build.GetVersion())

	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		log.Fatalln("Error loading .env file:", err)
	}

	dataDir := getDataDir()
	store := service.NewFsStorage(dataDir)

	frontend := getStaticFrontend()

	app := server.NewApp(frontend, store)

	// Start background schedulers
	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
	app.RefreshToken.StartCleanupScheduler(cleanupCtx, 24*time.Hour)
	app.Retention.StartCleanupScheduler(cleanupCtx, 24*time.Hour)

	handler := app.GetHandler()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		log.Println("Listening on port " + port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalln(err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Received signal to stop, shutting down server...")

	// Stop background tasks
	cleanupCancel()

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Failed to shutdown gracefully: ", err)
	}

	log.Println("Goodbye.")
}
