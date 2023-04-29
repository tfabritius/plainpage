package test

import (
	"strings"
	"testing"
	"time"

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
		res := s.api("GET", "/config", nil, s.adminToken)
		r.Equal(200, res.Code)
		body, _ := jsonbody[model.Config](res)
		r.NotNil(body.ACL)

		s.defaultAcl = body.ACL
	}
}

func (s *UsersTestSuite) TestCreateUser() {
	r := s.Require()

	username := "testCreateUser"
	password := "myPassword"
	displayName := "Test User"

	// Use default ACL
	s.saveGlobalAcl(s.adminToken, s.defaultAcl)

	// Endpoint returns user details
	{
		res := s.api("POST", "/auth/users",
			model.PostUserRequest{Username: username, DisplayName: displayName, Password: password},
			s.adminToken)
		r.Equal(200, res.Code)

		body, _ := jsonbody[model.User](res)
		r.Equal(username, body.Username)
		r.Equal(displayName, body.DisplayName)
		r.NotEmpty(body.ID)
		r.Empty(body.PasswordHash)

		r.NoError(s.app.Users.DeleteByUsername(username))
	}

	// Anonymous cannot register user
	{
		res := s.api("POST", "/auth/users",
			model.PostUserRequest{Username: username, DisplayName: displayName, Password: password},
			nil)
		r.Equal(401, res.Code)
	}

	// User cannot register user
	{
		res := s.api("POST", "/auth/users",
			model.PostUserRequest{Username: username, DisplayName: displayName, Password: password},
			nil)
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
		res := s.api("POST", "/auth/users",
			model.PostUserRequest{Username: username, DisplayName: displayName, Password: password},
			nil)
		r.Equal(401, res.Code)
	}

	// User can register user
	{
		res := s.api("POST", "/auth/users",
			model.PostUserRequest{Username: username, DisplayName: displayName, Password: password},
			s.userToken)
		r.Equal(200, res.Code)

		r.NoError(s.app.Users.DeleteByUsername(username))
	}

	// Enable anonymous user registration
	{
		acl := s.defaultAcl
		acl = append(acl, model.AccessRule{Subject: "anonymous", Operations: []model.AccessOp{model.AccessOpRegister}})

		s.saveGlobalAcl(s.adminToken, acl)
	}

	// Anonymous can register user
	{
		res := s.api("POST", "/auth/users",
			model.PostUserRequest{Username: username, DisplayName: displayName, Password: password},
			nil)
		r.Equal(200, res.Code)

		r.NoError(s.app.Users.DeleteByUsername(username))
	}

	// User can register user
	{
		res := s.api("POST", "/auth/users",
			model.PostUserRequest{Username: username, DisplayName: displayName, Password: password},
			s.userToken)
		r.Equal(200, res.Code)

		r.NoError(s.app.Users.DeleteByUsername(username))
	}

	// Duplicate username fails
	{
		_, err := s.app.Users.Create(strings.ToLower(username), password, displayName)
		r.NoError(err)

		res := s.api("POST", "/auth/users",
			model.PostUserRequest{Username: strings.ToUpper(username), DisplayName: displayName, Password: password},
			nil)
		r.Equal(409, res.Code)

		r.NoError(s.app.Users.DeleteByUsername(username))
	}
}
func (s *UsersTestSuite) TestLoginUser() {
	r := s.Require()

	username := "testLoginUser"
	displayName := "Test User"
	password := "myPassword"

	_, err := s.app.Users.Create(username, password, displayName)
	r.NoError(err)

	// Valid login returns user details and token
	{
		res := s.api("POST", "/auth/login",
			model.LoginRequest{Username: username, Password: password},
			nil)
		r.Equal(200, res.Code)

		body, _ := jsonbody[model.TokenUserResponse](res)
		r.Equal(username, body.User.Username)
		r.Equal(displayName, body.User.DisplayName)
		r.NotEmpty(body.User.ID)
		r.NotEmpty(body.Token)
	}

	// Wrong password fails
	{
		res := s.api("POST", "/auth/login",
			model.LoginRequest{Username: username, Password: "wrongPassword"},
			nil)
		r.Equal(401, res.Code)
		r.Equal("Unauthorized", strings.TrimSpace(res.Body.String()))
	}
}

func (s *UsersTestSuite) TestPatchUser() {
	r := s.Require()

	username := "testPatchUser"
	displayName := "Test User"
	password := "myPassword"

	_, err := s.app.Users.Create(username, password, displayName)
	r.NoError(err)
	token := s.loginUser(username, password)

	// Updating user fails if not logged in
	{
		res := s.api("PATCH", "/auth/users/"+username,
			[]map[string]string{{"op": "replace", "path": "/displayName", "value": "Changed Test User"}},
			nil)
		r.Equal(401, res.Code)
	}

	// User updates own displayName
	{
		res := s.api("PATCH", "/auth/users/"+username,
			[]map[string]string{{"op": "replace", "path": "/displayName", "value": "Changed Test User"}},
			&token)
		r.Equal(200, res.Code)
	}

	// Updating other user fails
	{
		res := s.api("PATCH", "/auth/users/"+username,
			[]map[string]string{{"op": "replace", "path": "/displayName", "value": "Changed Test User"}},
			s.userToken)
		r.Equal(403, res.Code)
	}

	// Admin updates other user
	{
		res := s.api("PATCH", "/auth/users/"+username,
			[]map[string]string{{"op": "replace", "path": "/displayName", "value": "Changed Test User"}},
			s.adminToken)
		r.Equal(200, res.Code)
	}
}

func (s *UsersTestSuite) TestRenewToken() {
	r := s.Require()

	username := "testRenewToken"
	password := "myPassword"

	_, err := s.app.Users.Create(username, password, "Test User")
	r.NoError(err)
	token := s.loginUser(username, password)

	// Renew token
	{
		time.Sleep(1050 * time.Millisecond) // Tokens should differ

		res := s.api("POST", "/auth/refresh", nil, &token)
		r.Equal(200, res.Code)
		body, _ := jsonbody[model.TokenUserResponse](res)
		r.Equal(username, body.User.Username)
		r.NotEqual(token, body.Token)

		token = body.Token
	}

	// Delete user with new token
	{
		err := s.app.Users.DeleteByUsername(username)
		r.NoError(err)
	}

	// Token is still valid as JWT cannot be revoked :-(
	{
		res := s.api("GET", "/pages", nil, &token)
		r.Equal(200, res.Code)
	}

	// Renew fails
	{
		res := s.api("POST", "/auth/refresh", nil, &token)
		r.Equal(401, res.Code)
	}
}

func (s *UsersTestSuite) TestDeleteUser() {
	r := s.Require()

	username := "testDeleteUser"
	password := "myPassword"

	// User deletes itself
	{
		_, err := s.app.Users.Create(username, password, "Test User")
		r.NoError(err)

		token := s.loginUser(username, password)

		res := s.api("DELETE", "/auth/users/"+username, nil, &token)
		r.Equal(200, res.Code)

		_, err = s.app.Users.GetByUsername(username)
		r.ErrorIs(err, model.ErrNotFound)
	}

	_, err := s.app.Users.Create(username, password, "Test User")
	r.NoError(err)

	// User cannot delete other user
	{
		res := s.api("DELETE", "/auth/users/"+username, nil, s.userToken)
		r.Equal(403, res.Code)
	}

	// Anonymous cannot delete user
	{
		res := s.api("DELETE", "/auth/users/"+username, nil, nil)
		r.Equal(401, res.Code)
	}

	// Admin can delete other user
	{
		res := s.api("DELETE", "/auth/users/"+username, nil, s.adminToken)
		r.Equal(200, res.Code)

		_, err := s.app.Users.GetByUsername(username)
		r.ErrorIs(err, model.ErrNotFound)
	}
}
