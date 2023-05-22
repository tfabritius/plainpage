package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
)

//go:generate rm -rf ./static/*
//go:generate sh -c "cp -r ../frontend/.output/public/* ./static/"
//go:embed all:static
var frontendFs embed.FS

func getStaticFrontend() http.FileSystem {
	fsys, err := fs.Sub(frontendFs, "static")
	if err != nil {
		log.Fatal(err)
	}

	return http.FS(fsys)
}
