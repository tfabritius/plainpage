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
	"github.com/tfabritius/plainpage/storage"
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

func (s *AppTestSuite) createUser(token *string, username, realName, password string) {
	s.T().Helper()
	r := s.Require()

	res := s.api("POST", "/_api/auth/users",
		model.PostUserRequest{Username: username, RealName: realName, Password: password},
		token)
	r.Equal(200, res.Code)

	body, _ := jsonbody[model.User](res)
	r.Equal(username, body.Username)
	r.Equal(realName, body.RealName)
	r.NotEmpty(body.ID)
	r.Empty(body.PasswordHash)
}

func (s *AppTestSuite) loginUser(username string, password string) string {
	r := s.Require()

	res := s.api("POST", "/_api/auth/login",
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

	res := s.api("PATCH", "/_api/config", []model.PatchOperation{{Op: "replace", Path: "/acl", Value: &aclJson}}, adminToken)
	r.Equal(200, res.Code)
}

func (s *AppTestSuite) setupInitialApp() {
	store := storage.NewFsStorage(s.T().TempDir())
	s.app = server.NewApp(http.FS(emptyFs{}), store)
	s.handler = s.app.GetHandler()

	r := s.Require()

	// Setup mode is enabled initially
	{
		body, res := jsonbody[model.GetAppResponse](s.api("GET", "/_api/app", nil, nil))
		r.Equal(200, res.Code)
		r.True(body.SetupMode)
	}

	// Register admin user
	s.createUser(nil, "admin", "Administrator", "secret")

	// Setup mode is disabled
	{
		body, res := jsonbody[model.GetAppResponse](s.api("GET", "/_api/app", nil, nil))
		r.Equal(200, res.Code)
		r.False(body.SetupMode)
	}

	adminToken := s.loginUser("admin", "secret")

	// Register another user
	s.createUser(&adminToken, "user", "User", "secret")
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

	req := httptest.NewRequest(method, target, bodyReader)
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
