package test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tfabritius/plainpage/model"
)

type ConfigTestSuite struct {
	AppTestSuite
	defaultAcl []model.AccessRule
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, &ConfigTestSuite{})
}

func (s *ConfigTestSuite) SetupSuite() {
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

func (s *ConfigTestSuite) TearDownTest() {
	// Restore default config after each test
	s.api("PATCH", "/config", []model.PatchOperation{
		{Op: "replace", Path: "/appTitle", Value: str2json("PlainPage")},
	}, s.adminToken)
	s.saveGlobalAcl(s.adminToken, s.defaultAcl)
}

// TestExposeConfig tests the GET /app endpoint which exposes basic app configuration
func (s *ConfigTestSuite) TestExposeConfig() {
	r := s.Require()

	// Test: Anonymous user gets basic config but no version info
	{
		res := s.api("GET", "/app", nil, nil)
		r.Equal(200, res.Code)

		body, _ := jsonbody[model.GetAppResponse](res)
		r.Equal("PlainPage", body.AppTitle)
		r.False(body.SetupMode, "Setup mode should be disabled after first user")
		r.Empty(body.Version, "Anonymous should not see version")
		r.Empty(body.GitSha, "Anonymous should not see gitSha")
	}

	// Test: Regular user gets basic config plus version info
	{
		res := s.api("GET", "/app", nil, s.userToken)
		r.Equal(200, res.Code)

		body, _ := jsonbody[model.GetAppResponse](res)
		r.Equal("PlainPage", body.AppTitle)
		r.False(body.SetupMode)
		// Version and GitSha may be empty in tests (not built with ldflags), but they should be present in the response
		r.False(body.AllowAdmin, "Regular user should not have admin access")
	}

	// Test: Admin user gets basic config plus version info
	{
		res := s.api("GET", "/app", nil, s.adminToken)
		r.Equal(200, res.Code)

		body, _ := jsonbody[model.GetAppResponse](res)
		r.Equal("PlainPage", body.AppTitle)
		r.False(body.SetupMode)
		r.True(body.AllowAdmin, "Admin user should have admin access")
	}
}

// TestExposeConfigAllowRegister tests that the allowRegister field changes based on ACL
func (s *ConfigTestSuite) TestExposeConfigAllowRegister() {
	r := s.Require()

	// By default, registration is not allowed for anonymous or users
	{
		res := s.api("GET", "/app", nil, nil)
		r.Equal(200, res.Code)
		body, _ := jsonbody[model.GetAppResponse](res)
		r.False(body.AllowRegister, "Anonymous should not be able to register by default")

		res = s.api("GET", "/app", nil, s.userToken)
		r.Equal(200, res.Code)
		body, _ = jsonbody[model.GetAppResponse](res)
		r.False(body.AllowRegister, "User should not be able to register by default")
	}

	// Enable registration for all users
	{
		acl := append(s.defaultAcl, model.AccessRule{Subject: "all", Operations: []model.AccessOp{model.AccessOpRegister}})
		s.saveGlobalAcl(s.adminToken, acl)

		res := s.api("GET", "/app", nil, nil)
		r.Equal(200, res.Code)
		body, _ := jsonbody[model.GetAppResponse](res)
		r.False(body.AllowRegister, "Anonymous should still not be able to register")

		res = s.api("GET", "/app", nil, s.userToken)
		r.Equal(200, res.Code)
		body, _ = jsonbody[model.GetAppResponse](res)
		r.True(body.AllowRegister, "User should now be able to register")
	}

	// Enable registration for anonymous users
	{
		acl := append(s.defaultAcl, model.AccessRule{Subject: "anonymous", Operations: []model.AccessOp{model.AccessOpRegister}})
		s.saveGlobalAcl(s.adminToken, acl)

		res := s.api("GET", "/app", nil, nil)
		r.Equal(200, res.Code)
		body, _ := jsonbody[model.GetAppResponse](res)
		r.True(body.AllowRegister, "Anonymous should now be able to register")

		res = s.api("GET", "/app", nil, s.userToken)
		r.Equal(200, res.Code)
		body, _ = jsonbody[model.GetAppResponse](res)
		r.True(body.AllowRegister, "User should also be able to register when anonymous can")
	}
}

// TestGetConfig tests the GET /config endpoint which returns full configuration
func (s *ConfigTestSuite) TestGetConfig() {
	r := s.Require()

	// Test: Anonymous user returns 401 Unauthorized
	{
		res := s.api("GET", "/config", nil, nil)
		r.Equal(401, res.Code)
	}

	// Test: Regular user returns 403 Forbidden
	{
		res := s.api("GET", "/config", nil, s.userToken)
		r.Equal(403, res.Code)
	}

	// Test: Admin user returns 200 with full config
	{
		res := s.api("GET", "/config", nil, s.adminToken)
		r.Equal(200, res.Code)

		body, _ := jsonbody[model.Config](res)
		r.Equal("PlainPage", body.AppTitle)
		r.False(body.SetupMode)
		r.NotNil(body.ACL)
	}
}

// TestGetConfigACLUserInfo tests that ACL entries with user subjects have user info populated
func (s *ConfigTestSuite) TestGetConfigACLUserInfo() {
	r := s.Require()

	// Add a user-specific ACL entry
	acl := append(s.defaultAcl, model.AccessRule{Subject: "user:" + s.adminUserID, Operations: []model.AccessOp{model.AccessOpAdmin}})
	s.saveGlobalAcl(s.adminToken, acl)

	// Get config and verify user info is populated
	res := s.api("GET", "/config", nil, s.adminToken)
	r.Equal(200, res.Code)

	body, _ := jsonbody[model.Config](res)

	// Find the user-specific ACL entry
	var foundUserEntry bool
	for _, rule := range body.ACL {
		if rule.Subject == "user:"+s.adminUserID {
			foundUserEntry = true
			r.NotNil(rule.User, "User info should be populated for user subject")
			r.Equal(s.adminUserID, rule.User.ID)
			r.Equal(TestAdminUsername, rule.User.Username)
			r.Equal("Administrator", rule.User.DisplayName)
			break
		}
	}
	r.True(foundUserEntry, "Should find the user-specific ACL entry")
}

// TestPatchConfigAuthorization tests authorization for PATCH /config endpoint
func (s *ConfigTestSuite) TestPatchConfigAuthorization() {
	r := s.Require()

	operations := []model.PatchOperation{
		{Op: "replace", Path: "/appTitle", Value: str2json("Test Title")},
	}

	// Test: Anonymous user returns 401 Unauthorized
	{
		res := s.api("PATCH", "/config", operations, nil)
		r.Equal(401, res.Code)
	}

	// Test: Regular user returns 403 Forbidden
	{
		res := s.api("PATCH", "/config", operations, s.userToken)
		r.Equal(403, res.Code)
	}

	// Test: Admin user returns 200
	{
		res := s.api("PATCH", "/config", operations, s.adminToken)
		r.Equal(200, res.Code)
	}
}

// TestPatchConfigAppTitle tests updating the app title via PATCH /config
func (s *ConfigTestSuite) TestPatchConfigAppTitle() {
	r := s.Require()

	// Update app title
	newTitle := "My Custom Wiki"
	res := s.api("PATCH", "/config", []model.PatchOperation{
		{Op: "replace", Path: "/appTitle", Value: str2json(newTitle)},
	}, s.adminToken)
	r.Equal(200, res.Code)

	// Verify response contains new title
	body, _ := jsonbody[model.Config](res)
	r.Equal(newTitle, body.AppTitle)

	// Verify GET /config returns new title
	res = s.api("GET", "/config", nil, s.adminToken)
	r.Equal(200, res.Code)
	body, _ = jsonbody[model.Config](res)
	r.Equal(newTitle, body.AppTitle)

	// Verify GET /app returns new title
	res = s.api("GET", "/app", nil, nil)
	r.Equal(200, res.Code)
	appBody, _ := jsonbody[model.GetAppResponse](res)
	r.Equal(newTitle, appBody.AppTitle)
}

// TestPatchConfigErrors tests error handling for invalid PATCH requests
func (s *ConfigTestSuite) TestPatchConfigErrors() {
	r := s.Require()

	// Test: Unsupported operation (not "replace")
	{
		res := s.api("PATCH", "/config", []model.PatchOperation{
			{Op: "add", Path: "/appTitle", Value: str2json("Test")},
		}, s.adminToken)
		r.Equal(400, res.Code)
		r.Contains(res.Body.String(), "operation add not supported")
	}

	// Test: Missing value
	{
		res := s.api("PATCH", "/config", []model.PatchOperation{
			{Op: "replace", Path: "/appTitle"},
		}, s.adminToken)
		r.Equal(400, res.Code)
		r.Contains(res.Body.String(), "value missing")
	}

	// Test: Unsupported path
	{
		res := s.api("PATCH", "/config", []model.PatchOperation{
			{Op: "replace", Path: "/unknownField", Value: str2json("Test")},
		}, s.adminToken)
		r.Equal(400, res.Code)
		r.Contains(res.Body.String(), "path /unknownField not supported")
	}

	// Test: Invalid JSON for appTitle (expecting string, sending number)
	{
		invalid := json.RawMessage(`123`)
		res := s.api("PATCH", "/config", []model.PatchOperation{
			{Op: "replace", Path: "/appTitle", Value: &invalid},
		}, s.adminToken)
		r.Equal(400, res.Code)
	}

	// Test: Invalid JSON for ACL (expecting array, sending string)
	{
		res := s.api("PATCH", "/config", []model.PatchOperation{
			{Op: "replace", Path: "/acl", Value: str2json("not an array")},
		}, s.adminToken)
		r.Equal(400, res.Code)
	}
}

// TestPatchConfigMultipleOperations tests applying multiple operations in a single request
func (s *ConfigTestSuite) TestPatchConfigMultipleOperations() {
	r := s.Require()

	// Apply multiple operations
	newACL := append(s.defaultAcl, model.AccessRule{Subject: "anonymous", Operations: []model.AccessOp{model.AccessOpRegister}})
	res := s.api("PATCH", "/config", []model.PatchOperation{
		{Op: "replace", Path: "/appTitle", Value: str2json("Multi-Op Test")},
		{Op: "replace", Path: "/acl", Value: acl2json(newACL)},
	}, s.adminToken)
	r.Equal(200, res.Code)

	// Verify both changes were applied
	body, _ := jsonbody[model.Config](res)
	r.Equal("Multi-Op Test", body.AppTitle)

	// Verify via GET /app that allowRegister changed
	res = s.api("GET", "/app", nil, nil)
	r.Equal(200, res.Code)
	appBody, _ := jsonbody[model.GetAppResponse](res)
	r.Equal("Multi-Op Test", appBody.AppTitle)
	r.True(appBody.AllowRegister, "Anonymous should now be able to register")
}

// TestPatchConfigEmptyOperations tests sending an empty operations array
func (s *ConfigTestSuite) TestPatchConfigEmptyOperations() {
	r := s.Require()

	// Empty operations array should succeed (no-op)
	res := s.api("PATCH", "/config", []model.PatchOperation{}, s.adminToken)
	r.Equal(200, res.Code)

	// Config should remain unchanged
	body, _ := jsonbody[model.Config](res)
	r.Equal("PlainPage", body.AppTitle)
}

// TestConfigACLValidation tests that ACL values are properly validated when setting global config ACLs
func (s *ConfigTestSuite) TestConfigACLValidation() {
	tests := []struct {
		name         string
		acl          []model.AccessRule
		responseCode int
	}{
		// Valid ACLs
		{
			name:         "valid:user-admin",
			acl:          []model.AccessRule{{Subject: "user:" + s.adminUserID, Operations: []model.AccessOp{model.AccessOpAdmin}}},
			responseCode: 200,
		},
		{
			name: "valid:anonymous-register",
			acl: []model.AccessRule{
				{Subject: "user:" + s.adminUserID, Operations: []model.AccessOp{model.AccessOpAdmin}},
				{Subject: "anonymous", Operations: []model.AccessOp{model.AccessOpRegister}},
			},
			responseCode: 200,
		},
		{
			name: "valid:all-register",
			acl: []model.AccessRule{
				{Subject: "user:" + s.adminUserID, Operations: []model.AccessOp{model.AccessOpAdmin}},
				{Subject: "all", Operations: []model.AccessOp{model.AccessOpRegister}},
			},
			responseCode: 200,
		},
		{
			name:         "valid:user-admin-register",
			acl:          []model.AccessRule{{Subject: "user:" + s.adminUserID, Operations: []model.AccessOp{model.AccessOpAdmin, model.AccessOpRegister}}},
			responseCode: 200,
		},
		{
			name: "valid:empty-ops",
			acl: []model.AccessRule{
				{Subject: "user:" + s.adminUserID, Operations: []model.AccessOp{model.AccessOpAdmin}},
				{Subject: "anonymous", Operations: []model.AccessOp{}},
			},
			responseCode: 200,
		},
		// Invalid subjects
		{
			name:         "invalid:subject-empty",
			acl:          []model.AccessRule{{Subject: "", Operations: []model.AccessOp{model.AccessOpAdmin}}},
			responseCode: 400,
		},
		{
			name:         "invalid:subject-garbage",
			acl:          []model.AccessRule{{Subject: "garbage", Operations: []model.AccessOp{model.AccessOpAdmin}}},
			responseCode: 400,
		},
		{
			name:         "invalid:subject-group",
			acl:          []model.AccessRule{{Subject: "group:admins", Operations: []model.AccessOp{model.AccessOpAdmin}}},
			responseCode: 400,
		},
		{
			name:         "invalid:subject-user-empty-id",
			acl:          []model.AccessRule{{Subject: "user:", Operations: []model.AccessOp{model.AccessOpAdmin}}},
			responseCode: 400,
		},
		{
			name:         "invalid:subject-admin",
			acl:          []model.AccessRule{{Subject: "admin", Operations: []model.AccessOp{model.AccessOpAdmin}}},
			responseCode: 400,
		},
		// Invalid operations (content ops not allowed for config)
		{
			name:         "invalid:op-read",
			acl:          []model.AccessRule{{Subject: "all", Operations: []model.AccessOp{model.AccessOpRead}}},
			responseCode: 400,
		},
		{
			name:         "invalid:op-write",
			acl:          []model.AccessRule{{Subject: "all", Operations: []model.AccessOp{model.AccessOpWrite}}},
			responseCode: 400,
		},
		{
			name:         "invalid:op-delete",
			acl:          []model.AccessRule{{Subject: "all", Operations: []model.AccessOp{model.AccessOpDelete}}},
			responseCode: 400,
		},
		{
			name:         "invalid:op-unknown",
			acl:          []model.AccessRule{{Subject: "all", Operations: []model.AccessOp{"superadmin"}}},
			responseCode: 400,
		},
		{
			name:         "invalid:op-mixed-valid-invalid",
			acl:          []model.AccessRule{{Subject: "user:" + s.adminUserID, Operations: []model.AccessOp{model.AccessOpAdmin, model.AccessOpRead}}},
			responseCode: 400,
		},
		// Mixed valid and invalid rules
		{
			name: "invalid:mixed-rules",
			acl: []model.AccessRule{
				{Subject: "user:" + s.adminUserID, Operations: []model.AccessOp{model.AccessOpAdmin}},
				{Subject: "invalid", Operations: []model.AccessOp{model.AccessOpAdmin}},
			},
			responseCode: 400,
		},
	}

	for _, tc := range tests {
		t := s.T()
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			res := s.api("PATCH", "/config",
				[]model.PatchOperation{
					{Op: "replace", Path: "/acl", Value: acl2json(tc.acl)},
				},
				s.adminToken)
			r.Equal(tc.responseCode, res.Code, "ACL: %+v", tc.acl)
		})
	}
}
