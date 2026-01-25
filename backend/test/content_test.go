package test

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tfabritius/plainpage/model"
)

type ContentTestSuite struct {
	AppTestSuite
}

func TestContentTestSuite(t *testing.T) {
	suite.Run(t, &ContentTestSuite{})
}

func (s *ContentTestSuite) SetupSuite() {
	s.setupInitialApp()
}

func (s *ContentTestSuite) SetupTest() {
	r := s.Require()

	// Create folders with ACL
	folders := []struct {
		Name string
		ACL  []model.AccessRule
	}{
		{
			// Read/write only for admins
			Name: "admin-only",
			ACL:  []model.AccessRule{},
		},
		{
			// All users can read
			Name: "read-only",
			ACL: []model.AccessRule{
				{
					Subject: "all",
					Operations: []model.AccessOp{
						model.AccessOpRead,
					}},
			},
		},
		{
			// Users can write, everybody can read
			Name: "published",
			ACL: []model.AccessRule{
				{
					Subject: "all",
					Operations: []model.AccessOp{
						model.AccessOpWrite,
						model.AccessOpDelete,
						model.AccessOpRead,
					}},
				{
					Subject: "anonymous",
					Operations: []model.AccessOp{
						model.AccessOpRead,
					}},
			},
		},
		{
			// Everybody can write/read
			Name: "public",
			ACL: []model.AccessRule{
				{
					Subject: "anonymous",
					Operations: []model.AccessOp{
						model.AccessOpWrite,
						model.AccessOpDelete,
						model.AccessOpRead,
					}},
			},
		},
	}

	for _, folder := range folders {
		r.NoError(s.app.Content.CreateFolder(
			folder.Name,
			model.ContentMeta{
				Title: folder.Name,
				ACL:   &folder.ACL,
			},
		))
	}
}

func (s *ContentTestSuite) TearDownTest() {
	r := s.Require()

	r.NoError(s.app.Content.DeleteAll())
}

func (s *ContentTestSuite) TestCreatePage() {
	tests := []struct {
		name         string
		token        *string
		url          string
		responseCode int
	}{
		// root
		{"admin:root", s.adminToken, "page", 200},
		{"user:root", s.userToken, "page", 200},
		{"anonymous:root", nil, "page", 401},
		// admin-only
		{"admin:adminOnly", s.adminToken, "admin-only/page", 200},
		{"user:adminOnly", s.userToken, "admin-only/page", 403},
		{"anonymous:adminOnly", nil, "admin-only/page", 401},
		// read-only
		{"admin:readOnly", s.adminToken, "read-only/page", 200},
		{"user:readOnly", s.userToken, "read-only/page", 403},
		{"anonymous:readOnly", nil, "read-only/page", 401},
		// published
		{"admin:published", s.adminToken, "published/page", 200},
		{"user:published", s.userToken, "published/page", 200},
		{"anonymous:published", nil, "published/page", 401},
		// public
		{"admin:public", s.adminToken, "public/page", 200},
		{"user:public", s.userToken, "public/page", 200},
		{"anonymous:public", nil, "public/page", 200},
		// nonexistent
		{"admin:nonexistent", s.adminToken, "nonexistent/page", 400},
		{"user:nonexistent", s.userToken, "nonexistent/page", 400},
		{"anonymous:nonexistent", nil, "nonexistent/page", 401},
		// invalid name
		{"admin:invalid", s.adminToken, "page!", 400},
		{"user:invalid", s.userToken, "page!", 400},
		{"anonymous:invalid", nil, "page!", 401},
		// conflict with folder with same name
		{"admin:conflict", s.adminToken, "admin-only", 400},
		{"user:conflict", s.userToken, "admin-only", 403},
		{"anonymous:conflict", nil, "admin-only", 401},
	}

	for _, tc := range tests {
		t := s.T()
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			res := s.api("PUT", "/pages/"+tc.url,
				model.PutRequest{Page: &model.Page{Meta: model.ContentMeta{Title: "Title"}}},
				tc.token)
			r.Equal(tc.responseCode, res.Code)

			if tc.responseCode == 200 {
				page, err := s.app.Content.ReadPage(tc.url, nil)
				r.NoError(err)
				r.Equal("Title", page.Meta.Title)

				r.NoError(s.app.Content.DeletePage(tc.url))
			} else {
				r.False(s.app.Content.IsPage(tc.url))
			}
		})
	}
}

func (s *ContentTestSuite) TestCreateFolder() {
	r := s.Require()
	r.NoError(s.app.Content.SavePage("existingpage", "", model.ContentMeta{}))

	tests := []struct {
		name         string
		token        *string
		url          string
		responseCode int
	}{
		// root
		{"admin:root", s.adminToken, "folder", 200},
		{"user:root", s.userToken, "folder", 200},
		{"anonymous:root", nil, "folder", 401},
		// admin-only
		{"admin:adminOnly", s.adminToken, "admin-only/folder", 200},
		{"user:adminOnly", s.userToken, "admin-only/folder", 403},
		{"anonymous:adminOnly", nil, "admin-only/folder", 401},
		// read-only
		{"admin:readOnly", s.adminToken, "read-only/folder", 200},
		{"user:readOnly", s.userToken, "read-only/folder", 403},
		{"anonymous:readOnly", nil, "read-only/folder", 401},
		// published
		{"admin:published", s.adminToken, "published/folder", 200},
		{"user:published", s.userToken, "published/folder", 200},
		{"anonymous:published", nil, "published/folder", 401},
		// public
		{"admin:public", s.adminToken, "public/folder", 200},
		{"user:public", s.userToken, "public/folder", 200},
		{"anonymous:public", nil, "public/folder", 200},
		// nonexistent
		{"admin:nonexistent", s.adminToken, "nonexistent/folder", 400},
		{"user:nonexistent", s.userToken, "nonexistent/folder", 400},
		{"anonymous:nonexistent", nil, "nonexistent/folder", 401},
		// invalid name
		{"admin:invalid", s.adminToken, "folder!", 400},
		{"user:invalid", s.userToken, "folder!", 400},
		{"anonymous:invalid", nil, "folder!", 401},
		// conflict with page with same name
		{"admin:conflict", s.adminToken, "existingpage", 400},
		{"user:conflict", s.userToken, "existingpage", 400},
		{"anonymous:conflict", nil, "existingpage", 401},
	}

	for _, tc := range tests {
		t := s.T()
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			res := s.api("PUT", "/pages/"+tc.url,
				model.PutRequest{Folder: &model.Folder{Meta: model.ContentMeta{Title: tc.name}}},
				tc.token)
			r.Equal(tc.responseCode, res.Code)

			if tc.responseCode == 200 {
				folder, err := s.app.Content.ReadFolder(tc.url)
				r.NoError(err)
				r.Len(folder.Content, 0)
				r.Equal(tc.name, folder.Meta.Title)

				r.NoError(s.app.Content.DeleteEmptyFolder(tc.url))
			} else {
				r.False(s.app.Content.IsFolder(tc.url))
			}
		})
	}
}

func (s *ContentTestSuite) TestCreateContentInvalid() {
	r := s.Require()

	res := s.api("PUT", "/pages/test",
		model.PutRequest{},
		s.adminToken)
	r.Equal(400, res.Code)
}

func (s *ContentTestSuite) TestReadPage() {
	tests := []struct {
		name         string
		token        *string
		url          string
		responseCode int
		showACL      bool
	}{
		// root
		{"admin:root", s.adminToken, "page", 200, true},
		{"user:root", s.userToken, "page", 200, false},
		{"anonymous:root", nil, "page", 401, false},
		// admin-only
		{"admin:adminOnly", s.adminToken, "admin-only/page", 200, true},
		{"user:adminOnly", s.userToken, "admin-only/page", 403, false},
		{"anonymous:adminOnly", nil, "admin-only/page", 401, false},
		// read-only
		{"admin:readOnly", s.adminToken, "read-only/page", 200, true},
		{"user:readOnly", s.userToken, "read-only/page", 200, false},
		{"anonymous:readOnly", nil, "read-only/page", 401, false},
		// published
		{"admin:published", s.adminToken, "published/page", 200, true},
		{"user:published", s.userToken, "published/page", 200, false},
		{"anonymous:published", nil, "published/page", 200, false},
		// public
		{"admin:public", s.adminToken, "public/page", 200, true},
		{"user:public", s.userToken, "public/page", 200, false},
		{"anonymous:public", nil, "public/page", 200, false},
	}

	for _, tc := range tests {
		t := s.T()
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			// Prepare
			r.NoError(s.app.Content.SavePage(tc.url, "Content", model.ContentMeta{Title: "Title"}))

			// Test
			res := s.api("GET", "/pages/"+tc.url,
				nil,
				tc.token)
			r.Equal(tc.responseCode, res.Code)

			if tc.responseCode == 200 {
				body, _ := jsonbody[model.GetContentResponse](res)
				r.Nil(body.Folder)
				r.NotNil(body.Page)
				r.Equal("Title", body.Page.Meta.Title)

				// Breadcrumbs
				r.Len(body.Breadcrumbs, strings.Count(tc.url, "/")+1)
				r.Equal(tc.url, body.Breadcrumbs[len(body.Breadcrumbs)-1].Url)
				r.Equal("page", body.Breadcrumbs[len(body.Breadcrumbs)-1].Name)
				r.Equal("Title", body.Breadcrumbs[len(body.Breadcrumbs)-1].Title)

				// Change ACL
				acl := []model.AccessRule{{Subject: "anonymous", Operations: []model.AccessOp{model.AccessOpRead}}}
				r.NoError(s.app.Content.SavePage(tc.url, "Content", model.ContentMeta{Title: "Title", ACL: &acl}))

				// Test ACL in output
				res := s.api("GET", "/pages/"+tc.url,
					nil,
					tc.token)
				r.Equal(tc.responseCode, res.Code)
				body, _ = jsonbody[model.GetContentResponse](res)
				if tc.showACL {
					r.Equal(acl, *body.Page.Meta.ACL)
				} else {
					r.Nil(body.Page.Meta.ACL)
				}
			}

			// Cleanup
			r.NoError(s.app.Content.DeletePage(tc.url))
		})
	}
}

func (s *ContentTestSuite) TestReadNonexistentPage() {
	tests := []struct {
		name         string
		token        *string
		url          string
		responseCode int
	}{
		// root
		{"admin:root", s.adminToken, "nonexistent", 404},
		{"user:root", s.userToken, "nonexistent", 404},
		{"anonymous:root", nil, "nonexistent", 401},
		// admin-only
		{"admin:adminOnly", s.adminToken, "admin-only/nonexistent", 404},
		{"user:adminOnly", s.userToken, "admin-only/nonexistent", 403},
		{"anonymous:adminOnly", nil, "admin-only/nonexistent", 401},
		// read-only
		{"admin:readOnly", s.adminToken, "read-only/nonexistent", 404},
		{"user:readOnly", s.userToken, "read-only/nonexistent", 404},
		{"anonymous:readOnly", nil, "read-only/nonexistent", 401},
		// published
		{"admin:published", s.adminToken, "published/nonexistent", 404},
		{"user:published", s.userToken, "published/nonexistent", 404},
		{"anonymous:published", nil, "published/nonexistent", 404},
		// public
		{"admin:public", s.adminToken, "public/nonexistent", 404},
		{"user:public", s.userToken, "public/nonexistent", 404},
		{"anonymous:public", nil, "public/nonexistent", 404},
		// invalid name
		{"admin:invalid", s.adminToken, "invalid!", 404},
		{"user:invalid", s.userToken, "invalid!", 404},
		{"anonymous:invalid", nil, "invalid!", 401},
	}

	for _, tc := range tests {
		t := s.T()
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			res := s.api("GET", "/pages/"+tc.url,
				nil,
				tc.token)
			r.Equal(tc.responseCode, res.Code)
		})
	}
}

func (s *ContentTestSuite) TestReadFolder() {
	tests := []struct {
		name         string
		token        *string
		url          string
		responseCode int
		entries      int
		showACL      bool
	}{
		// root
		{"admin:root", s.adminToken, "", 200, 4, true},
		{"user:root", s.userToken, "", 200, 4, false},
		{"anonymous:root", nil, "", 401, 0, false},
		// admin-only
		{"admin:adminOnly", s.adminToken, "/admin-only", 200, 0, true},
		{"user:adminOnly", s.userToken, "/admin-only", 403, 0, false},
		{"anonymous:adminOnly", nil, "/admin-only", 401, 0, false},
		// read-only
		{"admin:readOnly", s.adminToken, "/read-only", 200, 0, true},
		{"user:readOnly", s.userToken, "/read-only", 200, 0, false},
		{"anonymous:readOnly", nil, "/read-only", 401, 0, false},
		// published
		{"admin:published", s.adminToken, "/published", 200, 0, true},
		{"user:published", s.userToken, "/published", 200, 0, false},
		{"anonymous:published", nil, "/published", 200, 0, false},
		// public
		{"admin:public", s.adminToken, "/public", 200, 0, true},
		{"user:public", s.userToken, "/public", 200, 0, false},
		{"anonymous:public", nil, "/public", 200, 0, false},
		// nonexistent
		{"admin:nonexistent", s.adminToken, "/nonexistent", 404, 0, true},
		{"user:nonexistent", s.userToken, "/nonexistent", 404, 0, false},
		{"anonymous:nonexistent", nil, "/nonexistent", 401, 0, false},
		// invalid name
		{"admin:invalid", s.adminToken, "/folder!", 404, 0, true},
		{"user:invalid", s.userToken, "/folder!", 404, 0, false},
		{"anonymous:invalid", nil, "/folder!", 401, 0, false},
	}

	for _, tc := range tests {
		t := s.T()
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			res := s.api("GET", "/pages"+tc.url,
				nil,
				tc.token)
			r.Equal(tc.responseCode, res.Code)

			if tc.responseCode == 200 {
				body, _ := jsonbody[model.GetContentResponse](res)
				r.Nil(body.Page)
				r.NotNil(body.Folder)
				r.Len(body.Folder.Content, tc.entries)

				for _, c := range body.Folder.Content {
					r.Equal(c.Name, c.Title)
				}

				// Breadcrumbs
				r.Len(body.Breadcrumbs, strings.Count(tc.url, "/"))
				if len(body.Breadcrumbs) > 0 {
					r.Equal(strings.TrimPrefix(tc.url, "/"), body.Breadcrumbs[len(body.Breadcrumbs)-1].Url)
					r.Equal(strings.TrimPrefix(tc.url, "/"), body.Breadcrumbs[len(body.Breadcrumbs)-1].Name)
					r.Equal(strings.TrimPrefix(tc.url, "/"), body.Breadcrumbs[len(body.Breadcrumbs)-1].Title)
				}

				// Test ACL in output
				if tc.showACL {
					r.NotNil(body.Folder.Meta.ACL)
				} else {
					r.Nil(body.Folder.Meta.ACL)
				}
			}
		})
	}
}

func (s *ContentTestSuite) TestDeletePage() {
	tests := []struct {
		name         string
		token        *string
		url          string
		responseCode int
	}{
		// root
		{"admin:root", s.adminToken, "page", 200},
		{"user:root", s.userToken, "page", 200},
		{"anonymous:root", nil, "page", 401},
		// admin-only
		{"admin:adminOnly", s.adminToken, "admin-only/page", 200},
		{"user:adminOnly", s.userToken, "admin-only/page", 403},
		{"anonymous:adminOnly", nil, "admin-only/page", 401},
		// read-only
		{"admin:readOnly", s.adminToken, "read-only/page", 200},
		{"user:readOnly", s.userToken, "read-only/page", 403},
		{"anonymous:readOnly", nil, "read-only/page", 401},
		// published
		{"admin:published", s.adminToken, "published/page", 200},
		{"user:published", s.userToken, "published/page", 200},
		{"anonymous:published", nil, "published/page", 401},
		// public
		{"admin:public", s.adminToken, "public/page", 200},
		{"user:public", s.userToken, "public/page", 200},
		{"anonymous:public", nil, "public/page", 200},
	}

	for _, tc := range tests {
		t := s.T()
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			r.NoError(s.app.Content.SavePage(tc.url, "Content", model.ContentMeta{Title: "Title"}))

			res := s.api("DELETE", "/pages/"+tc.url,
				nil,
				tc.token)
			r.Equal(tc.responseCode, res.Code)

			if tc.responseCode == 200 {
				r.False(s.app.Content.IsPage(tc.url))
			} else {
				r.True(s.app.Content.IsPage(tc.url))
				r.NoError(s.app.Content.DeletePage(tc.url))
			}
		})
	}
}

func (s *ContentTestSuite) TestDeleteFolder() {
	tests := []struct {
		name         string
		token        *string
		url          string
		responseCode int
	}{
		// root
		{"admin:root", s.adminToken, "folder", 200},
		{"user:root", s.userToken, "folder", 200},
		{"anonymous:root", nil, "folder", 401},
		// admin-only
		{"admin:adminOnly", s.adminToken, "admin-only/folder", 200},
		{"user:adminOnly", s.userToken, "admin-only/folder", 403},
		{"anonymous:adminOnly", nil, "admin-only/folder", 401},
		// read-only
		{"admin:readOnly", s.adminToken, "read-only/folder", 200},
		{"user:readOnly", s.userToken, "read-only/folder", 403},
		{"anonymous:readOnly", nil, "read-only/folder", 401},
		// published
		{"admin:published", s.adminToken, "published/folder", 200},
		{"user:published", s.userToken, "published/folder", 200},
		{"anonymous:published", nil, "published/folder", 401},
		// public
		{"admin:public", s.adminToken, "public/folder", 200},
		{"user:public", s.userToken, "public/folder", 200},
		{"anonymous:public", nil, "public/folder", 200},
	}

	for _, tc := range tests {
		t := s.T()
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			r.NoError(s.app.Content.CreateFolder(tc.url, model.ContentMeta{}))

			res := s.api("DELETE", "/pages/"+tc.url,
				model.PutRequest{Folder: &model.Folder{}},
				tc.token)
			r.Equal(tc.responseCode, res.Code)

			if tc.responseCode == 200 {
				r.False(s.app.Content.IsFolder(tc.url))
			} else {
				r.True(s.app.Content.IsFolder(tc.url))
				r.NoError(s.app.Content.DeleteEmptyFolder(tc.url))
			}
		})
	}
}

func (s *ContentTestSuite) TestDeleteNonexistentPageOrFolder() {
	tests := []struct {
		name         string
		token        *string
		url          string
		responseCode int
	}{
		// root
		{"admin:root", s.adminToken, "nonexistent", 404},
		{"user:root", s.userToken, "nonexistent", 404},
		{"anonymous:root", nil, "nonexistent", 401},
		// admin-only
		{"admin:adminOnly", s.adminToken, "admin-only/nonexistent", 404},
		{"user:adminOnly", s.userToken, "admin-only/nonexistent", 403},
		{"anonymous:adminOnly", nil, "admin-only/nonexistent", 401},
		// read-only
		{"admin:readOnly", s.adminToken, "read-only/nonexistent", 404},
		{"user:readOnly", s.userToken, "read-only/nonexistent", 403},
		{"anonymous:readOnly", nil, "read-only/nonexistent", 401},
		// published
		{"admin:published", s.adminToken, "published/nonexistent", 404},
		{"user:published", s.userToken, "published/nonexistent", 404},
		{"anonymous:published", nil, "published/nonexistent", 401},
		// public
		{"admin:public", s.adminToken, "public/nonexistent", 404},
		{"user:public", s.userToken, "public/nonexistent", 404},
		{"anonymous:public", nil, "public/nonexistent", 404},
		// nonexistent
		{"admin:nonexistent", s.adminToken, "nonexistent/nonexistent", 404},
		{"user:nonexistent", s.userToken, "nonexistent/nonexistent", 404},
		{"anonymous:nonexistent", nil, "nonexistent/nonexistent", 401},
		// invalid name
		{"admin:invalid", s.adminToken, "invalid!", 404},
		{"user:invalid", s.userToken, "invalid!", 404},
		{"anonymous:invalid", nil, "invalid!", 401},
	}

	for _, tc := range tests {
		t := s.T()
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			res := s.api("DELETE", "/pages/"+tc.url,
				nil,
				tc.token)
			r.Equal(tc.responseCode, res.Code)
		})
	}
}

func (s *ContentTestSuite) TestDeleteNonemptyFolder() {
	r := s.Require()

	// Prepare
	r.NoError(s.app.Content.CreateFolder("folder", model.ContentMeta{}))
	r.NoError(s.app.Content.SavePage("folder/page", "", model.ContentMeta{}))

	// Test
	{
		res := s.api("DELETE", "/pages/folder",
			nil,
			s.adminToken)
		r.Equal(400, res.Code)
	}
	{
		res := s.api("DELETE", "/pages/folder",
			nil,
			s.userToken)
		r.Equal(400, res.Code)
	}
	{
		res := s.api("DELETE", "/pages/folder",
			nil,
			nil)
		r.Equal(401, res.Code)
	}
}

func (s *ContentTestSuite) TestUpdatePage() {
	tests := []struct {
		name         string
		token        *string
		url          string
		responseCode int
	}{
		// root
		{"admin:root", s.adminToken, "page", 200},
		{"user:root", s.userToken, "page", 200},
		{"anonymous:root", nil, "page", 401},
		// admin-only
		{"admin:adminOnly", s.adminToken, "admin-only/page", 200},
		{"user:adminOnly", s.userToken, "admin-only/page", 403},
		{"anonymous:adminOnly", nil, "admin-only/page", 401},
		// read-only
		{"admin:readOnly", s.adminToken, "read-only/page", 200},
		{"user:readOnly", s.userToken, "read-only/page", 403},
		{"anonymous:readOnly", nil, "read-only/page", 401},
		// published
		{"admin:published", s.adminToken, "published/page", 200},
		{"user:published", s.userToken, "published/page", 200},
		{"anonymous:published", nil, "published/page", 401},
		// public
		{"admin:public", s.adminToken, "public/page", 200},
		{"user:public", s.userToken, "public/page", 200},
		{"anonymous:public", nil, "public/page", 200},
	}

	for _, tc := range tests {
		t := s.T()
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			// Prepare
			r.NoError(s.app.Content.SavePage(
				tc.url,
				"Old content",
				model.ContentMeta{Title: "Old title", Tags: []string{"old tag"}},
			))

			// Test
			res := s.api("PUT", "/pages/"+tc.url,
				model.PutRequest{Page: &model.Page{
					Content: "New content",
					Meta: model.ContentMeta{
						Title: "New title",
						Tags:  []string{"new tag"},
						ACL:   &[]model.AccessRule{},
					},
				}},
				tc.token)
			r.Equal(tc.responseCode, res.Code)

			page, err := s.app.Content.ReadPage(tc.url, nil)
			r.NoError(err)
			if tc.responseCode == 200 {
				r.Equal("New content", page.Content)
				r.Equal("New title", page.Meta.Title)
				r.Len(page.Meta.Tags, 1)
				r.Equal("new tag", page.Meta.Tags[0])
			} else {
				r.Equal("Old content", page.Content)
				r.Equal("Old title", page.Meta.Title)
				r.Len(page.Meta.Tags, 1)
				r.Equal("old tag", page.Meta.Tags[0])
			}

			r.Nil(page.Meta.ACL) // ACL remains unchanged

			// Cleanup
			r.NoError(s.app.Content.DeletePage(tc.url))
		})
	}
}

func (s *ContentTestSuite) TestUpdatePageACL() {
	tests := []struct {
		name         string
		token        *string
		url          string
		responseCode int
	}{
		// root
		{"admin:root", s.adminToken, "page", 200},
		{"user:root", s.userToken, "page", 403},
		{"anonymous:root", nil, "page", 401},
		// admin-only
		{"admin:adminOnly", s.adminToken, "admin-only/page", 200},
		{"user:adminOnly", s.userToken, "admin-only/page", 403},
		{"anonymous:adminOnly", nil, "admin-only/page", 401},
		// read-only
		{"admin:readOnly", s.adminToken, "read-only/page", 200},
		{"user:readOnly", s.userToken, "read-only/page", 403},
		{"anonymous:readOnly", nil, "read-only/page", 401},
		// published
		{"admin:published", s.adminToken, "published/page", 200},
		{"user:published", s.userToken, "published/page", 403},
		{"anonymous:published", nil, "published/page", 401},
		// public
		{"admin:public", s.adminToken, "public/page", 200},
		{"user:public", s.userToken, "public/page", 403},
		{"anonymous:public", nil, "public/page", 401},
	}

	for _, tc := range tests {
		t := s.T()
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			// Prepare
			r.NoError(s.app.Content.SavePage(
				tc.url, "",
				model.ContentMeta{ACL: nil},
			))

			acl := []model.AccessRule{
				{Subject: "anonymous", Operations: []model.AccessOp{model.AccessOpRead}},
			}
			res := s.api("PATCH", "/pages/"+tc.url,
				[]model.PatchOperation{
					{Op: "replace", Path: "/page/meta/acl", Value: acl2json(acl)},
				},
				tc.token)
			r.Equal(tc.responseCode, res.Code)

			page, err := s.app.Content.ReadPage(tc.url, nil)
			r.NoError(err)
			if tc.responseCode == 200 {
				r.Equal(acl, *page.Meta.ACL)
			} else {
				r.Nil(page.Meta.ACL)
			}

			res = s.api("PATCH", "/pages/"+tc.url,
				[]model.PatchOperation{
					{Op: "replace", Path: "/page/meta/acl", Value: nil},
				},
				tc.token)
			r.Equal(tc.responseCode, res.Code)

			page, err = s.app.Content.ReadPage(tc.url, nil)
			r.NoError(err)
			r.Nil(page.Meta.ACL)

			// Cleanup
			r.NoError(s.app.Content.DeletePage(tc.url))
		})
	}
}

func (s *ContentTestSuite) TestUpdateFolder() {
	tests := []struct {
		name         string
		token        *string
		url          string
		responseCode int
	}{
		// root
		{"admin:root", s.adminToken, "folder", 200},
		{"user:root", s.userToken, "folder", 200},
		{"anonymous:root", nil, "folder", 401},
		// admin-only
		{"admin:adminOnly", s.adminToken, "admin-only/folder", 200},
		{"user:adminOnly", s.userToken, "admin-only/folder", 403},
		{"anonymous:adminOnly", nil, "admin-only/folder", 401},
		// read-only
		{"admin:readOnly", s.adminToken, "read-only/folder", 200},
		{"user:readOnly", s.userToken, "read-only/folder", 403},
		{"anonymous:readOnly", nil, "read-only/folder", 401},
		// published
		{"admin:published", s.adminToken, "published/folder", 200},
		{"user:published", s.userToken, "published/folder", 200},
		{"anonymous:published", nil, "published/folder", 401},
		// public
		{"admin:public", s.adminToken, "public/folder", 200},
		{"user:public", s.userToken, "public/folder", 200},
		{"anonymous:public", nil, "public/folder", 200},
	}

	for _, tc := range tests {
		t := s.T()
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			// Prepare
			r.NoError(s.app.Content.CreateFolder(tc.url, model.ContentMeta{Title: "Old Title"}))

			// Test
			res := s.api("PUT", "/pages/"+tc.url,
				model.PutRequest{Folder: &model.Folder{
					Meta: model.ContentMeta{
						Title: "New Title",
						ACL:   &[]model.AccessRule{},
					},
				}},
				tc.token)
			r.Equal(tc.responseCode, res.Code)

			folder, err := s.app.Content.ReadFolder(tc.url)
			r.NoError(err)
			if tc.responseCode == 200 {
				r.Equal("New Title", folder.Meta.Title)
			} else {
				r.Equal("Old Title", folder.Meta.Title)
			}
			r.Nil(folder.Meta.ACL) // ACL remains unchanged

			// Cleanup
			r.NoError(s.app.Content.DeleteEmptyFolder(tc.url))
		})
	}
}

func (s *ContentTestSuite) TestUpdateFolderACL() {
	tests := []struct {
		name         string
		token        *string
		url          string
		responseCode int
	}{
		// root
		{"admin:root", s.adminToken, "folder", 200},
		{"user:root", s.userToken, "folder", 403},
		{"anonymous:root", nil, "folder", 401},
		// admin-only
		{"admin:adminOnly", s.adminToken, "admin-only/folder", 200},
		{"user:adminOnly", s.userToken, "admin-only/folder", 403},
		{"anonymous:adminOnly", nil, "admin-only/folder", 401},
		// read-only
		{"admin:readOnly", s.adminToken, "read-only/folder", 200},
		{"user:readOnly", s.userToken, "read-only/folder", 403},
		{"anonymous:readOnly", nil, "read-only/folder", 401},
		// published
		{"admin:published", s.adminToken, "published/folder", 200},
		{"user:published", s.userToken, "published/folder", 403},
		{"anonymous:published", nil, "published/folder", 401},
		// public
		{"admin:public", s.adminToken, "public/folder", 200},
		{"user:public", s.userToken, "public/folder", 403},
		{"anonymous:public", nil, "public/folder", 401},
	}

	for _, tc := range tests {
		t := s.T()
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			// Prepare
			r.NoError(s.app.Content.CreateFolder(tc.url, model.ContentMeta{}))

			// Test
			acl := []model.AccessRule{
				{Subject: "anonymous", Operations: []model.AccessOp{model.AccessOpRead}},
			}
			res := s.api("PATCH", "/pages/"+tc.url,
				[]model.PatchOperation{
					{Op: "replace", Path: "/folder/meta/acl", Value: acl2json(acl)},
				},
				tc.token)
			r.Equal(tc.responseCode, res.Code)

			folder, err := s.app.Content.ReadFolder(tc.url)
			r.NoError(err)
			if tc.responseCode == 200 {
				r.Equal(acl, *folder.Meta.ACL)
			} else {
				r.Nil(folder.Meta.ACL)
			}

			res = s.api("PATCH", "/pages/"+tc.url,
				[]model.PatchOperation{
					{Op: "replace", Path: "/folder/meta/acl", Value: nil},
				},
				tc.token)
			r.Equal(tc.responseCode, res.Code)

			folder, err = s.app.Content.ReadFolder(tc.url)
			r.NoError(err)
			r.Nil(folder.Meta.ACL)

			// Cleanup
			r.NoError(s.app.Content.DeleteEmptyFolder(tc.url))
		})
	}
}

func (s *ContentTestSuite) TestAtticRevisions() {
	r := s.Require()

	// Prepare
	urls := []string{
		"page", "admin-only/page", "read-only/page", "published/page", "public/page",
	}
	for _, url := range urls {
		err := (s.app.Content.SavePage(
			url,
			"Old content",
			model.ContentMeta{Title: "Old title", Tags: []string{"old tag"}},
		))
		r.NoError(err)
	}
	time.Sleep(1050 * time.Millisecond) // Only one revision per second possible
	for _, url := range urls {
		err := (s.app.Content.SavePage(
			url,
			"New content",
			model.ContentMeta{Title: "New title", Tags: []string{"new tag"}},
		))
		r.NoError(err)
	}

	tests := []struct {
		name         string
		token        *string
		url          string
		responseCode int
	}{
		// root
		{"admin:root", s.adminToken, "page", 200},
		{"user:root", s.userToken, "page", 200},
		{"anonymous:root", nil, "page", 401},
		// admin-only
		{"admin:adminOnly", s.adminToken, "admin-only/page", 200},
		{"user:adminOnly", s.userToken, "admin-only/page", 403},
		{"anonymous:adminOnly", nil, "admin-only/page", 401},
		// read-only
		{"admin:readOnly", s.adminToken, "read-only/page", 200},
		{"user:readOnly", s.userToken, "read-only/page", 200},
		{"anonymous:readOnly", nil, "read-only/page", 401},
		// published
		{"admin:published", s.adminToken, "published/page", 200},
		{"user:published", s.userToken, "published/page", 200},
		{"anonymous:published", nil, "published/page", 200},
		// public
		{"admin:public", s.adminToken, "public/page", 200},
		{"user:public", s.userToken, "public/page", 200},
		{"anonymous:public", nil, "public/page", 200},
	}

	for _, tc := range tests {
		t := s.T()
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			res := s.api("GET", "/attic/"+tc.url,
				nil,
				tc.token)
			r.Equal(tc.responseCode, res.Code)

			var firstRev int64
			if tc.responseCode == 200 {
				body, _ := jsonbody[model.GetAtticListResponse](res)
				r.Len(body.Entries, 2)
				firstRev = body.Entries[0].Revision
			}

			res = s.api("GET", "/attic/"+tc.url+"?rev="+strconv.Itoa(int(firstRev)),
				nil,
				tc.token)
			r.Equal(tc.responseCode, res.Code)

			if tc.responseCode == 200 {
				body, _ := jsonbody[model.GetContentResponse](res)
				r.Nil(body.Folder)
				r.NotNil(body.Page)
				r.Equal("Old content", body.Page.Content)
				r.Equal("Old title", body.Page.Meta.Title)
				r.Len(body.Page.Meta.Tags, 1)
				r.Equal("old tag", body.Page.Meta.Tags[0])
			}
		})
	}
}

func (s *ContentTestSuite) TestSearch() {
	r := s.Require()

	// Prepare
	urls := []string{
		"page", "admin-only/page", "read-only/page", "published/page", "public/page",
	}
	for _, url := range urls {
		err := s.app.Content.SavePage(
			url,
			"Content",
			model.ContentMeta{Title: "Title", Tags: []string{"tag"}},
		)
		r.NoError(err)
	}

	// Search with different tokens
	tests := []struct {
		name     string
		token    *string
		q        string
		nResults int
	}{
		{"admin", s.adminToken, "title", 5},
		{"user", s.userToken, "title", 4},
		{"anonymous", nil, "title", 2},
	}
	for _, tc := range tests {
		t := s.T()
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			res := s.api("POST", "/search?q="+tc.q,
				nil,
				tc.token)
			r.Equal(200, res.Code)

			body, _ := jsonbody[[]model.SearchHit](res)
			r.Len(body, tc.nResults)

			for _, hit := range body {
				r.Nil(hit.EffectiveACL)
				r.Nil(hit.Meta.ACL)
				r.NotEmpty(hit.Url)
				r.Equal("Title", hit.Meta.Title)
				r.Len(hit.Meta.Tags, 1)
				r.Equal("tag", hit.Meta.Tags[0])
				r.NotEmpty(hit.Fragments["meta.title"])
			}
		})
	}

	// Search for different aspects of pages
	moreTests := []struct {
		name     string
		q        string
		nResults int
	}{
		{"url", "page", 0},
		{"content", "content", 5},
		{"meta.title", "title", 5},
		{"meta.tags", "tag", 5},
	}
	for _, tc := range moreTests {
		t := s.T()
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			res := s.api("POST", "/search?q="+tc.q,
				nil,
				s.adminToken)
			r.Equal(200, res.Code)

			body, _ := jsonbody[[]model.SearchHit](res)
			r.Len(body, tc.nResults)

			for _, hit := range body {
				r.Nil(hit.EffectiveACL)
				r.Nil(hit.Meta.ACL)
				r.NotEmpty(hit.Url)
				r.Equal("Title", hit.Meta.Title)
				r.Len(hit.Meta.Tags, 1)
				r.Equal("tag", hit.Meta.Tags[0])

				r.NotEmpty(hit.Fragments[tc.name])
				r.Len(hit.Fragments[tc.name], 1)
				if tc.name == "content" {
					r.Equal("<mark>Content</mark>", hit.Fragments[tc.name][0])
				} else if tc.name == "meta.title" {
					r.Equal("<mark>Title</mark>", hit.Fragments[tc.name][0])
				} else if tc.name == "meta.tags" {
					r.Equal("<mark>tag</mark>", hit.Fragments[tc.name][0])
				}
			}
		})
	}
}

func (s *ContentTestSuite) TestSearchFolder() {
	tests := []struct {
		name     string
		q        string
		nResults int
	}{
		{"meta.title", "published", 1},
	}
	for _, tc := range tests {
		t := s.T()
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			res := s.api("POST", "/search?q="+tc.q,
				nil,
				s.adminToken)
			r.Equal(200, res.Code)

			body, _ := jsonbody[[]model.SearchHit](res)
			r.Len(body, tc.nResults)

			for _, hit := range body {
				r.Nil(hit.EffectiveACL)
				r.Nil(hit.Meta.ACL)
				r.NotEmpty(hit.Url)
				r.Equal(tc.q, hit.Meta.Title)
				r.Len(hit.Meta.Tags, 0)

				r.NotEmpty(hit.Fragments["meta.title"])
				r.Len(hit.Fragments["meta.title"], 1)
				r.Equal("<mark>"+tc.q+"</mark>", hit.Fragments["meta.title"][0])
			}
		})
	}
}

func (s *ContentTestSuite) TestMovePage() {
	tests := []struct {
		name         string
		token        *string
		srcUrl       string
		destUrl      string
		responseCode int
	}{
		// Move within root (rename)
		{"admin:rename", s.adminToken, "page", "renamed-page", 200},
		{"user:rename", s.userToken, "page", "renamed-page", 200},
		{"anonymous:rename", nil, "page", "renamed-page", 401},

		// Move from root to public folder
		{"admin:toPublic", s.adminToken, "page", "public/page", 200},
		{"user:toPublic", s.userToken, "page", "public/page", 200},
		{"anonymous:toPublic", nil, "page", "public/page", 401}, // anonymous can't delete from root

		// Move from root to admin-only folder (need write on dest)
		{"admin:toAdminOnly", s.adminToken, "page", "admin-only/page", 200},
		{"user:toAdminOnly", s.userToken, "page", "admin-only/page", 403},
		{"anonymous:toAdminOnly", nil, "page", "admin-only/page", 401},

		// Move from admin-only to root (need write+delete on source, write on dest)
		{"admin:fromAdminOnly", s.adminToken, "admin-only/page", "page", 200},
		{"user:fromAdminOnly", s.userToken, "admin-only/page", "page", 403},
		{"anonymous:fromAdminOnly", nil, "admin-only/page", "page", 401},

		// Move from public to published
		{"admin:publicToPublished", s.adminToken, "public/page", "published/page", 200},
		{"user:publicToPublished", s.userToken, "public/page", "published/page", 200},
		{"anonymous:publicToPublished", nil, "public/page", "published/page", 401},
	}

	for _, tc := range tests {
		t := s.T()
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			// Prepare: create source page
			r.NoError(s.app.Content.SavePage(tc.srcUrl, "Content", model.ContentMeta{Title: "Title"}))

			// Test
			res := s.api("PATCH", "/pages/"+tc.srcUrl,
				[]model.PatchOperation{
					{Op: "replace", Path: "/page/url", Value: str2json(tc.destUrl)},
				},
				tc.token)
			r.Equal(tc.responseCode, res.Code)

			if tc.responseCode == 200 {
				// Source should not exist
				r.False(s.app.Content.IsPage(tc.srcUrl))
				// Destination should exist with same content
				r.True(s.app.Content.IsPage(tc.destUrl))
				page, err := s.app.Content.ReadPage(tc.destUrl, nil)
				r.NoError(err)
				r.Equal("Content", page.Content)
				r.Equal("Title", page.Meta.Title)
				// Cleanup
				r.NoError(s.app.Content.DeletePage(tc.destUrl))
			} else {
				// Source should still exist
				r.True(s.app.Content.IsPage(tc.srcUrl))
				// Destination should not exist
				r.False(s.app.Content.IsPage(tc.destUrl))
				// Cleanup
				r.NoError(s.app.Content.DeletePage(tc.srcUrl))
			}
		})
	}
}

func (s *ContentTestSuite) TestMovePageErrors() {
	r := s.Require()

	// Prepare
	r.NoError(s.app.Content.SavePage("page1", "Content1", model.ContentMeta{Title: "Page1"}))
	r.NoError(s.app.Content.SavePage("page2", "Content2", model.ContentMeta{Title: "Page2"}))
	r.NoError(s.app.Content.CreateFolder("folder", model.ContentMeta{Title: "Folder"}))

	// Invalid destination URL
	{
		res := s.api("PATCH", "/pages/page1",
			[]model.PatchOperation{
				{Op: "replace", Path: "/page/url", Value: str2json("invalid!")},
			},
			s.adminToken)
		r.Equal(400, res.Code)
		r.True(s.app.Content.IsPage("page1"))
	}

	// Destination already exists (page)
	{
		res := s.api("PATCH", "/pages/page1",
			[]model.PatchOperation{
				{Op: "replace", Path: "/page/url", Value: str2json("page2")},
			},
			s.adminToken)
		r.Equal(400, res.Code)
		r.True(s.app.Content.IsPage("page1"))
		r.True(s.app.Content.IsPage("page2"))
	}

	// Destination already exists (folder)
	{
		res := s.api("PATCH", "/pages/page1",
			[]model.PatchOperation{
				{Op: "replace", Path: "/page/url", Value: str2json("folder")},
			},
			s.adminToken)
		r.Equal(400, res.Code)
		r.True(s.app.Content.IsPage("page1"))
		r.True(s.app.Content.IsFolder("folder"))
	}

	// Destination parent folder doesn't exist
	{
		res := s.api("PATCH", "/pages/page1",
			[]model.PatchOperation{
				{Op: "replace", Path: "/page/url", Value: str2json("nonexistent/page")},
			},
			s.adminToken)
		r.Equal(400, res.Code)
		r.True(s.app.Content.IsPage("page1"))
	}

	// Move to same location (no-op, should succeed)
	{
		res := s.api("PATCH", "/pages/page1",
			[]model.PatchOperation{
				{Op: "replace", Path: "/page/url", Value: str2json("page1")},
			},
			s.adminToken)
		r.Equal(200, res.Code)
		r.True(s.app.Content.IsPage("page1"))
	}
}

func (s *ContentTestSuite) TestMovePageWithAttic() {
	r := s.Require()

	// Create page with multiple revisions
	r.NoError(s.app.Content.SavePage("page", "Content v1", model.ContentMeta{Title: "Title"}))
	time.Sleep(1050 * time.Millisecond) // Only one revision per second possible
	r.NoError(s.app.Content.SavePage("page", "Content v2", model.ContentMeta{Title: "Title"}))

	// Verify attic has 2 entries
	atticEntries, err := s.app.Content.ListAttic("page")
	r.NoError(err)
	r.Len(atticEntries, 2)
	rev1 := atticEntries[0].Revision

	// Move page
	res := s.api("PATCH", "/pages/page",
		[]model.PatchOperation{
			{Op: "replace", Path: "/page/url", Value: str2json("moved-page")},
		},
		s.adminToken)
	r.Equal(200, res.Code)

	// Verify page moved
	r.False(s.app.Content.IsPage("page"))
	r.True(s.app.Content.IsPage("moved-page"))

	// Verify attic entries moved
	oldAtticEntries, _ := s.app.Content.ListAttic("page")
	r.Len(oldAtticEntries, 0)

	newAtticEntries, err := s.app.Content.ListAttic("moved-page")
	r.NoError(err)
	r.Len(newAtticEntries, 2)

	// Verify old revision content is accessible at new location
	oldPage, err := s.app.Content.ReadPage("moved-page", &rev1)
	r.NoError(err)
	r.Equal("Content v1", oldPage.Content)
}

func (s *ContentTestSuite) TestMoveFolder() {
	tests := []struct {
		name         string
		token        *string
		srcUrl       string
		destUrl      string
		responseCode int
	}{
		// Move within root (rename)
		{"admin:rename", s.adminToken, "folder", "renamed-folder", 200},
		{"user:rename", s.userToken, "folder", "renamed-folder", 200},
		{"anonymous:rename", nil, "folder", "renamed-folder", 401},

		// Move from root to public folder
		{"admin:toPublic", s.adminToken, "folder", "public/folder", 200},
		{"user:toPublic", s.userToken, "folder", "public/folder", 200},
		{"anonymous:toPublic", nil, "folder", "public/folder", 401}, // anonymous can't delete from root

		// Move from root to admin-only folder
		{"admin:toAdminOnly", s.adminToken, "folder", "admin-only/folder", 200},
		{"user:toAdminOnly", s.userToken, "folder", "admin-only/folder", 403},
		{"anonymous:toAdminOnly", nil, "folder", "admin-only/folder", 401},
	}

	for _, tc := range tests {
		t := s.T()
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			// Prepare: create source folder
			r.NoError(s.app.Content.CreateFolder(tc.srcUrl, model.ContentMeta{Title: "Folder"}))

			// Test
			res := s.api("PATCH", "/pages/"+tc.srcUrl,
				[]model.PatchOperation{
					{Op: "replace", Path: "/folder/url", Value: str2json(tc.destUrl)},
				},
				tc.token)
			r.Equal(tc.responseCode, res.Code)

			if tc.responseCode == 200 {
				// Source should not exist
				r.False(s.app.Content.IsFolder(tc.srcUrl))
				// Destination should exist
				r.True(s.app.Content.IsFolder(tc.destUrl))
				folder, err := s.app.Content.ReadFolder(tc.destUrl)
				r.NoError(err)
				r.Equal("Folder", folder.Meta.Title)
				// Cleanup
				r.NoError(s.app.Content.DeleteEmptyFolder(tc.destUrl))
			} else {
				// Source should still exist
				r.True(s.app.Content.IsFolder(tc.srcUrl))
				// Destination should not exist
				r.False(s.app.Content.IsFolder(tc.destUrl))
				// Cleanup
				r.NoError(s.app.Content.DeleteEmptyFolder(tc.srcUrl))
			}
		})
	}
}

func (s *ContentTestSuite) TestMoveFolderErrors() {
	r := s.Require()

	// Prepare
	r.NoError(s.app.Content.CreateFolder("folder1", model.ContentMeta{Title: "Folder1"}))
	r.NoError(s.app.Content.CreateFolder("folder2", model.ContentMeta{Title: "Folder2"}))
	r.NoError(s.app.Content.SavePage("page", "Content", model.ContentMeta{Title: "Page"}))

	// Invalid destination URL
	{
		res := s.api("PATCH", "/pages/folder1",
			[]model.PatchOperation{
				{Op: "replace", Path: "/folder/url", Value: str2json("invalid!")},
			},
			s.adminToken)
		r.Equal(400, res.Code)
		r.True(s.app.Content.IsFolder("folder1"))
	}

	// Destination already exists (folder)
	{
		res := s.api("PATCH", "/pages/folder1",
			[]model.PatchOperation{
				{Op: "replace", Path: "/folder/url", Value: str2json("folder2")},
			},
			s.adminToken)
		r.Equal(400, res.Code)
		r.True(s.app.Content.IsFolder("folder1"))
		r.True(s.app.Content.IsFolder("folder2"))
	}

	// Destination already exists (page)
	{
		res := s.api("PATCH", "/pages/folder1",
			[]model.PatchOperation{
				{Op: "replace", Path: "/folder/url", Value: str2json("page")},
			},
			s.adminToken)
		r.Equal(400, res.Code)
		r.True(s.app.Content.IsFolder("folder1"))
		r.True(s.app.Content.IsPage("page"))
	}

	// Destination parent folder doesn't exist
	{
		res := s.api("PATCH", "/pages/folder1",
			[]model.PatchOperation{
				{Op: "replace", Path: "/folder/url", Value: str2json("nonexistent/folder")},
			},
			s.adminToken)
		r.Equal(400, res.Code)
		r.True(s.app.Content.IsFolder("folder1"))
	}

	// Move to same location (no-op, should succeed)
	{
		res := s.api("PATCH", "/pages/folder1",
			[]model.PatchOperation{
				{Op: "replace", Path: "/folder/url", Value: str2json("folder1")},
			},
			s.adminToken)
		r.Equal(200, res.Code)
		r.True(s.app.Content.IsFolder("folder1"))
	}
}

func (s *ContentTestSuite) TestMoveFolderWithContent() {
	r := s.Require()

	// Create folder with nested content
	r.NoError(s.app.Content.CreateFolder("folder", model.ContentMeta{Title: "Folder"}))
	r.NoError(s.app.Content.SavePage("folder/page1", "Content1", model.ContentMeta{Title: "Page1"}))
	r.NoError(s.app.Content.SavePage("folder/page2", "Content2", model.ContentMeta{Title: "Page2"}))
	r.NoError(s.app.Content.CreateFolder("folder/subfolder", model.ContentMeta{Title: "Subfolder"}))
	r.NoError(s.app.Content.SavePage("folder/subfolder/page3", "Content3", model.ContentMeta{Title: "Page3"}))

	// Move folder
	res := s.api("PATCH", "/pages/folder",
		[]model.PatchOperation{
			{Op: "replace", Path: "/folder/url", Value: str2json("moved-folder")},
		},
		s.adminToken)
	r.Equal(200, res.Code)

	// Verify old paths don't exist
	r.False(s.app.Content.IsFolder("folder"))
	r.False(s.app.Content.IsPage("folder/page1"))
	r.False(s.app.Content.IsPage("folder/page2"))
	r.False(s.app.Content.IsFolder("folder/subfolder"))
	r.False(s.app.Content.IsPage("folder/subfolder/page3"))

	// Verify new paths exist with correct content
	r.True(s.app.Content.IsFolder("moved-folder"))

	page1, err := s.app.Content.ReadPage("moved-folder/page1", nil)
	r.NoError(err)
	r.Equal("Content1", page1.Content)
	r.Equal("Page1", page1.Meta.Title)

	page2, err := s.app.Content.ReadPage("moved-folder/page2", nil)
	r.NoError(err)
	r.Equal("Content2", page2.Content)

	r.True(s.app.Content.IsFolder("moved-folder/subfolder"))

	page3, err := s.app.Content.ReadPage("moved-folder/subfolder/page3", nil)
	r.NoError(err)
	r.Equal("Content3", page3.Content)
}

func (s *ContentTestSuite) TestMoveFolderWithAttic() {
	r := s.Require()

	// Create folder with pages that have attic entries
	r.NoError(s.app.Content.CreateFolder("folder", model.ContentMeta{Title: "Folder"}))
	r.NoError(s.app.Content.SavePage("folder/page", "Content v1", model.ContentMeta{Title: "Page"}))
	time.Sleep(1050 * time.Millisecond)
	r.NoError(s.app.Content.SavePage("folder/page", "Content v2", model.ContentMeta{Title: "Page"}))

	// Verify attic has entries
	atticEntries, err := s.app.Content.ListAttic("folder/page")
	r.NoError(err)
	r.Len(atticEntries, 2)
	rev1 := atticEntries[0].Revision

	// Move folder
	res := s.api("PATCH", "/pages/folder",
		[]model.PatchOperation{
			{Op: "replace", Path: "/folder/url", Value: str2json("moved-folder")},
		},
		s.adminToken)
	r.Equal(200, res.Code)

	// Verify old attic doesn't exist
	oldAttic, _ := s.app.Content.ListAttic("folder/page")
	r.Len(oldAttic, 0)

	// Verify new attic exists with entries
	newAttic, err := s.app.Content.ListAttic("moved-folder/page")
	r.NoError(err)
	r.Len(newAttic, 2)

	// Verify old revision is accessible
	oldPage, err := s.app.Content.ReadPage("moved-folder/page", &rev1)
	r.NoError(err)
	r.Equal("Content v1", oldPage.Content)
}

func (s *ContentTestSuite) TestMoveFolderSearchIndex() {
	r := s.Require()

	// Create folder with a page
	r.NoError(s.app.Content.CreateFolder("folder", model.ContentMeta{Title: "UniqueFolder"}))
	r.NoError(s.app.Content.SavePage("folder/page", "UniqueContent", model.ContentMeta{Title: "UniquePage"}))

	// Verify search finds the page
	results, err := s.app.Content.Search("UniquePage")
	r.NoError(err)
	r.Len(results, 1)
	r.Equal("folder/page", results[0].Url)

	// Move folder
	res := s.api("PATCH", "/pages/folder",
		[]model.PatchOperation{
			{Op: "replace", Path: "/folder/url", Value: str2json("moved-folder")},
		},
		s.adminToken)
	r.Equal(200, res.Code)

	// Verify search finds page at new location
	results, err = s.app.Content.Search("UniquePage")
	r.NoError(err)
	r.Len(results, 1)
	r.Equal("moved-folder/page", results[0].Url)

	// Verify search finds folder at new location
	results, err = s.app.Content.Search("UniqueFolder")
	r.NoError(err)
	r.Len(results, 1)
	r.Equal("moved-folder", results[0].Url)
}
