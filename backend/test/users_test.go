package test

import (
	"strconv"
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

	// Invalid username fails
	{
		invalidUsername := "email@example.com"

		res := s.api("POST", "/auth/users",
			model.PostUserRequest{Username: invalidUsername, DisplayName: displayName, Password: password},
			nil)
		r.Equal(400, res.Code)
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
		r.Empty(body.User.PasswordHash)
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

	user, err := s.app.Users.Create(username, password, displayName)
	r.NoError(err)
	token, err := s.app.Token.GenerateToken(user)
	r.NoError(err)

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

	// Updating nonexisting user as user fails
	{
		res := s.api("PATCH", "/auth/users/does-not-exist",
			[]map[string]string{{"op": "replace", "path": "/displayName", "value": "Changed Test User"}},
			s.userToken)
		r.Equal(403, res.Code)
	}

	// Updating nonexisting user as admin fails
	{
		res := s.api("PATCH", "/auth/users/does-not-exist",
			[]map[string]string{{"op": "replace", "path": "/displayName", "value": "Changed Test User"}},
			s.adminToken)
		r.Equal(404, res.Code)
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

	user, err := s.app.Users.Create(username, password, "Test User")
	r.NoError(err)
	token, err := s.app.Token.GenerateToken(user)
	r.NoError(err)

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
		user, err := s.app.Users.Create(username, password, "Test User")
		r.NoError(err)
		token, err := s.app.Token.GenerateToken(user)
		r.NoError(err)

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

	// Deleting nonexistent user as user fails
	{
		res := s.api("DELETE", "/auth/users/does-not-exist", nil, s.userToken)
		r.Equal(403, res.Code)
	}

	// Deleting nonexistent user as admin fails
	{
		res := s.api("DELETE", "/auth/users/does-not-exist", nil, s.adminToken)
		r.Equal(404, res.Code)
	}
}

func (s *UsersTestSuite) TestChangePassword() {
	r := s.Require()

	username := "testChangePassword"
	displayName := "Test User"
	password := "myPassword"
	newPassword := "myNewPassword"

	user, err := s.app.Users.Create(username, password, displayName)
	r.NoError(err)
	token, err := s.app.Token.GenerateToken(user)
	r.NoError(err)

	// Unauthenticated request fails
	{
		res := s.api("POST", "/auth/users/"+username+"/password",
			model.ChangePasswordRequest{CurrentPassword: password, NewPassword: newPassword},
			nil)
		r.Equal(401, res.Code)
	}

	// User cannot change password with wrong current password
	{
		res := s.api("POST", "/auth/users/"+username+"/password",
			model.ChangePasswordRequest{CurrentPassword: "wrongPassword", NewPassword: newPassword},
			&token)
		r.Equal(403, res.Code)
	}

	// User can change own password with correct current password
	{
		res := s.api("POST", "/auth/users/"+username+"/password",
			model.ChangePasswordRequest{CurrentPassword: password, NewPassword: newPassword},
			&token)
		r.Equal(200, res.Code)

		// Verify new password works
		res = s.api("POST", "/auth/login",
			model.LoginRequest{Username: username, Password: newPassword},
			nil)
		r.Equal(200, res.Code)

		// Verify old password no longer works
		res = s.api("POST", "/auth/login",
			model.LoginRequest{Username: username, Password: password},
			nil)
		r.Equal(401, res.Code)
	}

	// Non-admin cannot change another user's password even with correct own password
	{
		// Create another user
		otherUsername := "testChgPwdOther"
		otherPassword := "otherPassword"
		otherUser, err := s.app.Users.Create(otherUsername, otherPassword, "Other User")
		r.NoError(err)
		otherToken, err := s.app.Token.GenerateToken(otherUser)
		r.NoError(err)

		res := s.api("POST", "/auth/users/"+username+"/password",
			model.ChangePasswordRequest{CurrentPassword: otherPassword, NewPassword: "someNewPassword"},
			&otherToken)
		r.Equal(403, res.Code)

		// Cleanup
		r.NoError(s.app.Users.DeleteByUsername(otherUsername))
	}

	// Cleanup
	r.NoError(s.app.Users.DeleteByUsername(username))
}

func (s *UsersTestSuite) TestChangePasswordAsAdmin() {
	r := s.Require()

	targetUsername := "testChgPwdTarget"
	targetDisplayName := "Target User"
	targetPassword := "targetPassword"
	newTargetPassword := "newTargetPassword"
	adminPassword := "secret" // Admin password from setupInitialApp

	// Create target user
	_, err := s.app.Users.Create(targetUsername, targetPassword, targetDisplayName)
	r.NoError(err)

	// Admin cannot change another user's password with wrong admin password
	{
		res := s.api("POST", "/auth/users/"+targetUsername+"/password",
			model.ChangePasswordRequest{CurrentPassword: "wrongAdminPassword", NewPassword: newTargetPassword},
			s.adminToken)
		r.Equal(403, res.Code)
	}

	// Admin can change another user's password with correct admin password
	{
		res := s.api("POST", "/auth/users/"+targetUsername+"/password",
			model.ChangePasswordRequest{CurrentPassword: adminPassword, NewPassword: newTargetPassword},
			s.adminToken)
		r.Equal(200, res.Code)

		// Verify target user can login with new password
		res = s.api("POST", "/auth/login",
			model.LoginRequest{Username: targetUsername, Password: newTargetPassword},
			nil)
		r.Equal(200, res.Code)

		// Verify old password no longer works
		res = s.api("POST", "/auth/login",
			model.LoginRequest{Username: targetUsername, Password: targetPassword},
			nil)
		r.Equal(401, res.Code)
	}

	// Admin cannot change nonexistent user's password
	{
		res := s.api("POST", "/auth/users/nonexistent/password",
			model.ChangePasswordRequest{CurrentPassword: adminPassword, NewPassword: newTargetPassword},
			s.adminToken)
		r.Equal(404, res.Code)
	}

	// Cleanup
	r.NoError(s.app.Users.DeleteByUsername(targetUsername))
}

func (s *UsersTestSuite) TestLoginRateLimit() {
	r := s.Require()

	username := "testLoginRateLimit"
	displayName := "Test User"
	password := "myPassword"
	headers := map[string]string{"X-Forwarded-For": "127.0.0.2"}

	_, err := s.app.Users.Create(username, password, displayName)
	r.NoError(err)

	nRequestLimit := 5

	// Limiter does not count successful logins.
	for i := 0; i < nRequestLimit+1; i++ {
		res := s.apiWithHeaders("POST", "/auth/login",
			model.LoginRequest{Username: username, Password: password},
			headers)
		r.Equal(200, res.Code)
	}

	// First wrong attempts should be 401.
	for i := 0; i < nRequestLimit; i++ {
		res := s.apiWithHeaders("POST", "/auth/login",
			model.LoginRequest{Username: username, Password: "wrong"},
			headers)
		r.Equal(401, res.Code)
	}

	// The next wrong attempt should be rate-limited with 429 and Retry-After header + JSON.
	{
		res := s.apiWithHeaders("POST", "/auth/login",
			model.LoginRequest{Username: username, Password: "wrong"},
			headers)
		r.Equal(429, res.Code)

		retryAfter := res.Header().Get("Retry-After")
		r.NotEmpty(retryAfter)
		retryAfterInt, err := strconv.Atoi(retryAfter)
		r.NoError(err)
		r.Greater(retryAfterInt, 0)
	}

	// A correct login should still be blocked.
	{
		res := s.apiWithHeaders("POST", "/auth/login",
			model.LoginRequest{Username: username, Password: password},
			headers)
		r.Equal(429, res.Code)
	}

	// Cleanup
	r.NoError(s.app.Users.DeleteByUsername(username))
}
