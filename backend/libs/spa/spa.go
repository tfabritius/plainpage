package spa

import (
	"net/http"
	"os"
)

// ServeFileContents returns a HandlerFunc that will
// read file from file system and write its content to ResponseWriter
func ServeFileContents(fileName string, fileSystem http.FileSystem) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Open the file
		file, err := fileSystem.Open(fileName)
		if err != nil {
			if os.IsNotExist(err) {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			panic(err)
		}

		// Retrieve FileInfo
		fi, err := file.Stat()
		if err != nil {
			panic(err)
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeContent(w, r, fi.Name(), fi.ModTime(), file)
	}
}

type prevent404ResponseWriter struct {
	http.ResponseWriter
	got404 bool
}

func (prevent404rw *prevent404ResponseWriter) WriteHeader(status int) {
	if status == http.StatusNotFound {
		// Don't actually write the 404 header, just set the flag
		prevent404rw.got404 = true
	} else {
		prevent404rw.ResponseWriter.WriteHeader(status)
	}
}

func (prevent404rw *prevent404ResponseWriter) Write(p []byte) (int, error) {
	if prevent404rw.got404 {
		// Do nothing, but pretend that we wrote len(p) bytes
		return len(p), nil
	}

	return prevent404rw.ResponseWriter.Write(p)
}

func Catch404Middleware(NotFoundHandler http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			prevent404writer := &prevent404ResponseWriter{ResponseWriter: w}
			next.ServeHTTP(prevent404writer, r)

			if prevent404writer.got404 {
				NotFoundHandler.ServeHTTP(w, r)
			}
		})
	}
}
