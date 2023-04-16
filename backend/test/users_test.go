package test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/tfabritius/plainpage/model"
)

type UsersTestSuite struct {
	AppTestSuite
	defaultAcl []model.AccessRule
}

func TestUsersTestSuite(t *testing.T) {
	suite.Run(t, &UsersTestSuite{})
}

func (s *UsersTestSuite) SetupSuite() {
	s.setupInitialApp()

	r := s.Require()

	// Get initial ACL
	{
		res := s.api("GET", "/_api/config", nil, s.adminToken)
		r.Equal(200, res.Code)
		body, _ := jsonbody[model.Config](res)
		r.NotNil(body.ACL)

		s.defaultAcl = body.ACL
	}
}

func (s *UsersTestSuite) TestCreateUser() {
	r := s.Require()

	// Use default ACL
	s.saveGlobalAcl(s.adminToken, s.defaultAcl)

	// Anonymous cannot register user
	{
		res := s.api("POST", "/_api/auth/users", model.PostUserRequest{
			Username:    "test",
			DisplayName: "test",
			Password:    "secret",
		}, nil)
		r.Equal(401, res.Code)
	}

	// User cannot register user
	{
		res := s.api("POST", "/_api/auth/users", model.PostUserRequest{
			Username:    "test",
			DisplayName: "test",
			Password:    "secret",
		}, nil)
		r.Equal(401, res.Code)
	}

	// Enable user registration by registered users
	{
		acl := s.defaultAcl
		acl = append(acl, model.AccessRule{Subject: "all", Operations: []model.AccessOp{model.AccessOpRegister}})

		s.saveGlobalAcl(s.adminToken, acl)
	}

	// Anonymous cannot register user
	{
		res := s.api("POST", "/_api/auth/users", model.PostUserRequest{
			Username:    "test",
			DisplayName: "test",
			Password:    "secret",
		}, nil)
		r.Equal(401, res.Code)
	}

	// User can register user
	s.createUser(s.userToken, "test1", "test1", "secret")

	// Enable anonymous user registration
	{
		acl := s.defaultAcl
		acl = append(acl, model.AccessRule{Subject: "anonymous", Operations: []model.AccessOp{model.AccessOpRegister}})

		s.saveGlobalAcl(s.adminToken, acl)
	}

	// Anonymous can register user
	s.createUser(nil, "test2", "test2", "secret")

	// User can register user
	s.createUser(s.userToken, "test3", "test3", "secret")

	// Duplicate username fails
	{
		s.createUser(nil, "duplicate-username", "Duplicate User", "secret")

		res := s.api("POST", "/_api/auth/users", model.PostUserRequest{
			Username:    "Duplicate-Username",
			DisplayName: "test",
			Password:    "secret",
		}, nil)
		r.Equal(409, res.Code)
	}
}
func (s *UsersTestSuite) TestLoginUser() {
	r := s.Require()

	s.createUser(s.adminToken, "test-user", "Test User", "myPassword")

	// Valid login returns user details and token
	{
		res := s.api("POST", "/_api/auth/login",
			model.LoginRequest{Username: "test-user", Password: "myPassword"},
			nil)
		r.Equal(200, res.Code)

		body, _ := jsonbody[model.TokenUserResponse](res)
		r.Equal("test-user", body.User.Username)
		r.Equal("Test User", body.User.DisplayName)
		r.NotEmpty(body.User.ID)
		r.NotEmpty(body.Token)
	}

	// Wrong password fails
	{
		res := s.api("POST", "/_api/auth/login",
			model.LoginRequest{Username: "test-user", Password: "wrongPassword"},
			nil)
		r.Equal(401, res.Code)
		r.Equal("Unauthorized", strings.TrimSpace(res.Body.String()))
	}
}

func (s *UsersTestSuite) TestPatchUser() {
	r := s.Require()

	s.createUser(s.adminToken, "patch-user", "Test User", "secret")
	token := s.loginUser("patch-user", "secret")

	// Updating user fails if not logged in
	{
		res := s.api("PATCH", "/_api/auth/users/patch-user",
			[]map[string]string{{"op": "replace", "path": "/displayName", "value": "Changed Test User"}},
			nil)
		r.Equal(401, res.Code)
	}

	// User updates own displayName
	{
		res := s.api("PATCH", "/_api/auth/users/patch-user",
			[]map[string]string{{"op": "replace", "path": "/displayName", "value": "Changed Test User"}},
			&token)
		r.Equal(200, res.Code)
	}

	// Updating other user fails
	{
		res := s.api("PATCH", "/_api/auth/users/patch-user",
			[]map[string]string{{"op": "replace", "path": "/displayName", "value": "Changed Test User"}},
			s.userToken)
		r.Equal(403, res.Code)
	}

	// Admin updates other user
	{
		res := s.api("PATCH", "/_api/auth/users/patch-user",
			[]map[string]string{{"op": "replace", "path": "/displayName", "value": "Changed Test User"}},
			s.adminToken)
		r.Equal(200, res.Code)
	}

}
