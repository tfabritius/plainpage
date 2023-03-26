package test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/tfabritius/plainpage/server"
	"github.com/tfabritius/plainpage/storage"
)

var app http.Handler

type emptyFs struct {
}

func (fs emptyFs) Open(name string) (fs.File, error) {
	return nil, errors.New("empty fs doesn't contain any files")
}

func TestMain(m *testing.M) {
	store := storage.NewFsStorage("./data")

	app = server.NewApp(http.FS(emptyFs{}), store).GetHandler()

	// Run tests
	exitVal := m.Run()

	os.Exit(exitVal)
}

func api(method, target string, body any) *httptest.ResponseRecorder {
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			panic(err)
		}

		bodyReader = bytes.NewReader(bodyBytes)
	} else {
		bodyReader = nil
	}

	req := httptest.NewRequest(method, target, bodyReader)

	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	return res
}

func jsonbody[T any](res *httptest.ResponseRecorder) (T, *httptest.ResponseRecorder) {
	var body T
	err := json.Unmarshal(res.Body.Bytes(), &body)
	if err != nil {
		panic(fmt.Errorf("Could not parse body: %w, body: %s", err, res.Body.String()))
	}
	return body, res
}
