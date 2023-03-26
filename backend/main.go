package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/tfabritius/plainpage/server"
	"github.com/tfabritius/plainpage/storage"
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
	if err := godotenv.Load(); err != nil {
		log.Fatalln("Error loading .env file")
	}

	dataDir := getDataDir()
	store := storage.NewFsStorage(dataDir)

	frontend := getStaticFrontend()

	handler := server.NewApp(frontend, store).GetHandler()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Listening on port " + port)
	http.ListenAndServe(":"+port, handler)
}
