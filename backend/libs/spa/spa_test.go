package spa

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServeFileContents(t *testing.T) {
	t.Run("serves existing file", func(t *testing.T) {
		content := []byte("<html><body>Hello World</body></html>")
		fs := http.FS(fstest.MapFS{
			"index.html": &fstest.MapFile{
				Data:    content,
				ModTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		})

		handler := ServeFileContents("index.html", fs)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		handler(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))
		assert.Equal(t, string(content), rec.Body.String())
	})

	t.Run("returns 404 for non-existent file", func(t *testing.T) {
		fs := http.FS(fstest.MapFS{})

		handler := ServeFileContents("missing.html", fs)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		handler(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestPrevent404ResponseWriter_WriteHeader(t *testing.T) {
	t.Run("sets flag on 404 without writing header", func(t *testing.T) {
		rec := httptest.NewRecorder()
		prevent404rw := &prevent404ResponseWriter{ResponseWriter: rec}

		prevent404rw.WriteHeader(http.StatusNotFound)

		assert.True(t, prevent404rw.got404)
		// The underlying recorder should not have received the 404 status
		// We need to check if Write was called to see if headers were flushed
		assert.Equal(t, http.StatusOK, rec.Code) // Default is 200 if WriteHeader wasn't called
	})

	t.Run("passes through non-404 status", func(t *testing.T) {
		rec := httptest.NewRecorder()
		prevent404rw := &prevent404ResponseWriter{ResponseWriter: rec}

		prevent404rw.WriteHeader(http.StatusOK)

		assert.False(t, prevent404rw.got404)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("passes through other status codes", func(t *testing.T) {
		testCases := []int{
			http.StatusCreated,
			http.StatusBadRequest,
			http.StatusInternalServerError,
			http.StatusForbidden,
		}

		for _, status := range testCases {
			t.Run(http.StatusText(status), func(t *testing.T) {
				rec := httptest.NewRecorder()
				prevent404rw := &prevent404ResponseWriter{ResponseWriter: rec}

				prevent404rw.WriteHeader(status)

				assert.False(t, prevent404rw.got404)
				assert.Equal(t, status, rec.Code)
			})
		}
	})
}

func TestPrevent404ResponseWriter_Write(t *testing.T) {
	t.Run("suppresses output when got404 is true", func(t *testing.T) {
		rec := httptest.NewRecorder()
		prevent404rw := &prevent404ResponseWriter{ResponseWriter: rec, got404: true}

		data := []byte("Not Found")
		n, err := prevent404rw.Write(data)

		require.NoError(t, err)
		assert.Equal(t, len(data), n)
		assert.Empty(t, rec.Body.String())
	})

	t.Run("writes output when got404 is false", func(t *testing.T) {
		rec := httptest.NewRecorder()
		prevent404rw := &prevent404ResponseWriter{ResponseWriter: rec, got404: false}

		data := []byte("Hello World")
		n, err := prevent404rw.Write(data)

		require.NoError(t, err)
		assert.Equal(t, len(data), n)
		assert.Equal(t, "Hello World", rec.Body.String())
	})
}

func TestCatch404Middleware(t *testing.T) {
	t.Run("passes through normal responses", func(t *testing.T) {
		normalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, "Normal Response")
		})

		notFoundHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = io.WriteString(w, "Custom 404")
		})

		middleware := Catch404Middleware(notFoundHandler)
		handler := middleware(normalHandler)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Normal Response", rec.Body.String())
	})

	t.Run("catches 404 and serves NotFoundHandler", func(t *testing.T) {
		handler404 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Not Found", http.StatusNotFound)
		})

		notFoundHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, "Custom SPA Fallback")
		})

		middleware := Catch404Middleware(notFoundHandler)
		handler := middleware(handler404)

		req := httptest.NewRequest(http.MethodGet, "/some/path", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Custom SPA Fallback", rec.Body.String())
	})

	t.Run("passes through other error status codes", func(t *testing.T) {
		errorHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = io.WriteString(w, "Internal Server Error")
		})

		notFoundHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, "Custom 404")
		})

		middleware := Catch404Middleware(notFoundHandler)
		handler := middleware(errorHandler)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "Internal Server Error", rec.Body.String())
	})
}
