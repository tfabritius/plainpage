package test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/tfabritius/plainpage/model"
	"github.com/tfabritius/plainpage/storage"
)

type UsersTestSuite struct {
	AppTestSuite
	defaultAcl []storage.AccessRule
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
		body, _ := jsonbody[storage.Config](res)
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
			Username: "test",
			RealName: "test",
			Password: "secret",
		}, nil)
		r.Equal(401, res.Code)
	}

	// User cannot register user
	{
		res := s.api("POST", "/_api/auth/users", model.PostUserRequest{
			Username: "test",
			RealName: "test",
			Password: "secret",
		}, nil)
		r.Equal(401, res.Code)
	}

	// Enable user registration by registered users
	{
		acl := s.defaultAcl
		acl = append(acl, storage.AccessRule{Subject: "all", Operations: []storage.AccessOp{storage.AccessOpRegister}})

		s.saveGlobalAcl(s.adminToken, acl)
	}

	// Anonymous cannot register user
	{
		res := s.api("POST", "/_api/auth/users", model.PostUserRequest{
			Username: "test",
			RealName: "test",
			Password: "secret",
		}, nil)
		r.Equal(401, res.Code)
	}

	// User can register user
	s.createUser(s.userToken, "test1", "test1", "secret")

	// Enable anonymous user registration
	{
		acl := s.defaultAcl
		acl = append(acl, storage.AccessRule{Subject: "anonymous", Operations: []storage.AccessOp{storage.AccessOpRegister}})

		s.saveGlobalAcl(s.adminToken, acl)
	}

	// Anonymous can register user
	s.createUser(nil, "test2", "test2", "secret")

	// User can register user
	s.createUser(s.userToken, "test3", "test3", "secret")
}
