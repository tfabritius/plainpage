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

func (s *AppTestSuite) loginUser(username string, password string) string {
	r := s.Require()

	res := s.api("POST", "/auth/login",
		model.LoginRequest{Username: username, Password: password},
		nil)
	r.Equal(200, res.Code)

	body, _ := jsonbody[model.TokenUserResponse](res)
	r.Equal(username, body.User.Username)
	r.NotEmpty(body.User.ID)
	r.Empty(body.User.PasswordHash)
	r.NotEmpty(body.Token)

	return body.Token
}

func (s *AppTestSuite) saveGlobalAcl(adminToken *string, acl []model.AccessRule) {
	r := s.Require()

	aclBytes, err := json.Marshal(acl)
	r.Nil(err)
	aclJson := json.RawMessage(aclBytes)

	res := s.api("PATCH", "/config", []model.PatchOperation{{Op: "replace", Path: "/acl", Value: &aclJson}}, adminToken)
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

		body, _ := jsonbody[model.User](res)

		err := s.app.Users.CheckAppPermissions(body.ID, model.AccessOpAdmin)
		r.NoError(err)
	}

	// Setup mode is disabled
	{
		body, res := jsonbody[model.GetAppResponse](s.api("GET", "/app", nil, nil))
		r.Equal(200, res.Code)
		r.False(body.SetupMode)
	}

	adminToken := s.loginUser("admin", "secret")

	// Register another user that will not become admin automatically
	{
		res := s.api("POST", "/auth/users",
			model.PostUserRequest{Username: "user", DisplayName: "User", Password: "secret"},
			&adminToken)
		r.Equal(200, res.Code)

		body, _ := jsonbody[model.User](res)

		err := s.app.Users.CheckAppPermissions(body.ID, model.AccessOpAdmin)
		r.Error(err)
	}

	userToken := s.loginUser("user", "secret")

	s.adminToken = &adminToken
	s.userToken = &userToken
}

func (s *AppTestSuite) api(method, target string, body any, token *string) *httptest.ResponseRecorder {
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
	if token != nil {
		req.Header.Add("Authorization", "Bearer "+*token)
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
