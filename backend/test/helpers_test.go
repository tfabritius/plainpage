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

	"github.com/stretchr/testify/suite"
	"github.com/tfabritius/plainpage/model"
	"github.com/tfabritius/plainpage/server"
	"github.com/tfabritius/plainpage/service"
)

type emptyFs struct {
}

func (fs emptyFs) Open(name string) (fs.File, error) {
	return nil, errors.New("empty fs doesn't contain any files")
}

type AppTestSuite struct {
	suite.Suite
	app        server.App
	handler    http.Handler
	adminToken *string
	userToken  *string
}

func (s *AppTestSuite) saveGlobalAcl(adminToken *string, acl []model.AccessRule) {
	r := s.Require()

	res := s.api("PATCH", "/config", []model.PatchOperation{{Op: "replace", Path: "/acl", Value: acl2json(acl)}}, adminToken)
	r.Equal(200, res.Code)
}

func (s *AppTestSuite) setupInitialApp() {
	store := service.NewFsStorage(s.T().TempDir())
	s.app = server.NewApp(http.FS(emptyFs{}), store)
	s.handler = s.app.GetHandler()

	r := s.Require()

	// Setup mode is enabled initially
	{
		body, res := jsonbody[model.GetAppResponse](s.api("GET", "/app", nil, nil))
		r.Equal(200, res.Code)
		r.True(body.SetupMode)
	}

	// Register first user that will become admin automatically
	{
		res := s.api("POST", "/auth/users",
			model.PostUserRequest{Username: "admin", DisplayName: "Administrator", Password: "secret"},
			nil)
		r.Equal(200, res.Code)

		user, _ := jsonbody[model.User](res)

		err := s.app.Users.CheckAppPermissions(user.ID, model.AccessOpAdmin)
		r.NoError(err)

		token, err := s.app.Token.GenerateToken(user)
		r.NoError(err)

		s.adminToken = &token
	}

	// Setup mode is disabled
	{
		body, res := jsonbody[model.GetAppResponse](s.api("GET", "/app", nil, nil))
		r.Equal(200, res.Code)
		r.False(body.SetupMode)
	}

	// Register another user that will not become admin automatically
	{
		res := s.api("POST", "/auth/users",
			model.PostUserRequest{Username: "user", DisplayName: "User", Password: "secret"},
			s.adminToken)
		r.Equal(200, res.Code)

		user, _ := jsonbody[model.User](res)

		err := s.app.Users.CheckAppPermissions(user.ID, model.AccessOpAdmin)
		r.Error(err)

		token, err := s.app.Token.GenerateToken(user)
		r.NoError(err)

		s.userToken = &token
	}
}

func (s *AppTestSuite) api(method, target string, body any, token *string) *httptest.ResponseRecorder {
	headers := map[string]string{}

	if token != nil {
		headers["Authorization"] = "Bearer " + *token
	}

	return s.apiWithHeaders(method, target, body, headers)
}

func (s *AppTestSuite) apiWithHeaders(method, target string, body any, headers map[string]string) *httptest.ResponseRecorder {
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

	req := httptest.NewRequest(method, "/_api"+target, bodyReader)

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	res := httptest.NewRecorder()
	s.handler.ServeHTTP(res, req)
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

func acl2json(acl []model.AccessRule) *json.RawMessage {
	bytes, err := json.Marshal(acl)
	if err != nil {
		panic(err)
	}
	jsonRawMsg := json.RawMessage(bytes)
	return &jsonRawMsg
}
