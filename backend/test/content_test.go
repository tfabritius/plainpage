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

				// Verify modification tracking
				r.False(page.Meta.ModifiedAt.IsZero(), "ModifiedAt should be set")
				r.WithinDuration(time.Now(), page.Meta.ModifiedAt, 5*time.Second, "ModifiedAt should be recent")

				// Verify ModifiedByUserID matches the user who made the request
				switch tc.token {
				case s.adminToken:
					r.Equal(s.adminUserID, page.Meta.ModifiedByUserID, "ModifiedByUserID should be admin's ID")
				case s.userToken:
					r.Equal(s.userUserID, page.Meta.ModifiedByUserID, "ModifiedByUserID should be user's ID")
				default:
					r.Empty(page.Meta.ModifiedByUserID, "ModifiedByUserID should be empty for anonymous")
				}

				r.NoError(s.app.Content.DeletePage(tc.url))
			} else {
				r.False(s.app.Content.IsPage(tc.url))
			}
		})
	}
}

func (s *ContentTestSuite) TestCreateFolder() {
	r := s.Require()
	r.NoError(s.app.Content.SavePage("existingpage", "", model.ContentMeta{}, ""))

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
			r.NoError(s.app.Content.SavePage(tc.url, "Content", model.ContentMeta{Title: "Title"}, ""))

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
				r.NoError(s.app.Content.SavePage(tc.url, "Content", model.ContentMeta{Title: "Title", ACL: &acl}, ""))

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
		{"user:root", s.userToken, "", 200, 3, false}, // admin-only folder is filtered out
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

			// Verify trash is empty before test
			trashBefore, err := s.app.Content.ListTrash()
			r.NoError(err)
			r.Empty(trashBefore, "Trash should be empty before deletion")

			beforeTime := time.Now().Unix()
			r.NoError(s.app.Content.SavePage(tc.url, "Content", model.ContentMeta{Title: "Title"}, ""))

			res := s.api("DELETE", "/pages/"+tc.url,
				nil,
				tc.token)
			r.Equal(tc.responseCode, res.Code)

			afterTime := time.Now().Unix()

			if tc.responseCode == 200 {
				r.False(s.app.Content.IsPage(tc.url))

				// Verify page is in trash with correct timestamp
				trashAfter, err := s.app.Content.ListTrash()
				r.NoError(err)
				r.Len(trashAfter, 1, "Trash should have exactly one entry")
				r.Equal(tc.url, trashAfter[0].Url)
				r.GreaterOrEqual(trashAfter[0].DeletedAt, beforeTime, "DeletedAt should be >= before time")
				r.LessOrEqual(trashAfter[0].DeletedAt, afterTime, "DeletedAt should be <= after time")

				// Clean up trash
				r.NoError(s.app.Content.DeleteTrashEntry(trashAfter[0].Url, trashAfter[0].DeletedAt))
			} else {
				r.True(s.app.Content.IsPage(tc.url))
				r.NoError(s.app.Content.DeletePage(tc.url))

				// Clean up trash
				trashAfter, _ := s.app.Content.ListTrash()
				for _, entry := range trashAfter {
					r.NoError(s.app.Content.DeleteTrashEntry(entry.Url, entry.DeletedAt))
				}
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
	tests := []struct {
		name         string
		token        *string
		responseCode int
	}{
		{"admin", s.adminToken, 200},
		{"user", s.userToken, 200},
		{"anonymous", nil, 401},
	}

	for _, tc := range tests {
		t := s.T()
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			// Verify trash is empty before test
			trashBefore, err := s.app.Content.ListTrash()
			r.NoError(err)
			r.Empty(trashBefore, "Trash should be empty before deletion")

			// Setup: create folder and page for this subtest
			beforeTime := time.Now().Unix()
			r.NoError(s.app.Content.CreateFolder("folder", model.ContentMeta{}))
			r.NoError(s.app.Content.SavePage("folder/page", "", model.ContentMeta{}, ""))

			// Test
			res := s.api("DELETE", "/pages/folder", nil, tc.token)
			r.Equal(tc.responseCode, res.Code)

			afterTime := time.Now().Unix()

			if tc.responseCode == 200 {
				r.False(s.app.Content.IsFolder("folder"))
				r.False(s.app.Content.IsPage("folder/page"))

				// Verify page is in trash with correct timestamp
				trashAfter, err := s.app.Content.ListTrash()
				r.NoError(err)
				r.Len(trashAfter, 1, "Trash should have exactly one entry")
				r.Equal("folder/page", trashAfter[0].Url)
				r.GreaterOrEqual(trashAfter[0].DeletedAt, beforeTime, "DeletedAt should be >= before time")
				r.LessOrEqual(trashAfter[0].DeletedAt, afterTime, "DeletedAt should be <= after time")

				// Clean up trash
				r.NoError(s.app.Content.DeleteTrashEntry(trashAfter[0].Url, trashAfter[0].DeletedAt))
			} else {
				// Cleanup if deletion failed
				r.True(s.app.Content.IsFolder("folder"))
				r.NoError(s.app.Content.DeleteFolder("folder"))

				// Clean up trash
				trashAfter, _ := s.app.Content.ListTrash()
				for _, entry := range trashAfter {
					r.NoError(s.app.Content.DeleteTrashEntry(entry.Url, entry.DeletedAt))
				}
			}
		})
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
				"",
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

				// Verify modification tracking is updated
				r.False(page.Meta.ModifiedAt.IsZero(), "ModifiedAt should be set")
				r.WithinDuration(time.Now(), page.Meta.ModifiedAt, 5*time.Second, "ModifiedAt should be recent")

				// Verify ModifiedByUserID matches the user who made the request
				switch tc.token {
				case s.adminToken:
					r.Equal(s.adminUserID, page.Meta.ModifiedByUserID, "ModifiedByUserID should be admin's ID")
				case s.userToken:
					r.Equal(s.userUserID, page.Meta.ModifiedByUserID, "ModifiedByUserID should be user's ID")
				default:
					r.Empty(page.Meta.ModifiedByUserID, "ModifiedByUserID should be empty for anonymous")
				}
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
				"",
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

// TestUpdatePageTitle tests updating the page's title using PATCH operation
func (s *ContentTestSuite) TestUpdatePageTitle() {
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
			r.NoError(s.app.Content.SavePage(tc.url, "Content", model.ContentMeta{Title: "Old Title"}, ""))

			// Test
			res := s.api("PATCH", "/pages/"+tc.url,
				[]model.PatchOperation{
					{Op: "replace", Path: "/page/meta/title", Value: str2json("New Title")},
				},
				tc.token)
			r.Equal(tc.responseCode, res.Code)

			page, err := s.app.Content.ReadPage(tc.url, nil)
			r.NoError(err)
			if tc.responseCode == 200 {
				r.Equal("New Title", page.Meta.Title)

				// Verify modification tracking is updated (even for metadata-only PATCH)
				r.False(page.Meta.ModifiedAt.IsZero(), "ModifiedAt should be set")
				r.WithinDuration(time.Now(), page.Meta.ModifiedAt, 5*time.Second, "ModifiedAt should be recent")

				// Verify ModifiedByUserID matches the user who made the request
				switch tc.token {
				case s.adminToken:
					r.Equal(s.adminUserID, page.Meta.ModifiedByUserID, "ModifiedByUserID should be admin's ID")
				case s.userToken:
					r.Equal(s.userUserID, page.Meta.ModifiedByUserID, "ModifiedByUserID should be user's ID")
				default:
					r.Empty(page.Meta.ModifiedByUserID, "ModifiedByUserID should be empty for anonymous")
				}
			} else {
				r.Equal("Old Title", page.Meta.Title)
			}

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

func (s *ContentTestSuite) TestUpdateFolderTitle() {
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
			res := s.api("PATCH", "/pages/"+tc.url,
				[]model.PatchOperation{
					{Op: "replace", Path: "/folder/meta/title", Value: str2json("New Title")},
				},
				tc.token)
			r.Equal(tc.responseCode, res.Code)

			folder, err := s.app.Content.ReadFolder(tc.url)
			r.NoError(err)
			if tc.responseCode == 200 {
				r.Equal("New Title", folder.Meta.Title)
			} else {
				r.Equal("Old Title", folder.Meta.Title)
			}

			// Cleanup
			r.NoError(s.app.Content.DeleteEmptyFolder(tc.url))
		})
	}
}

func (s *ContentTestSuite) TestCombinedPatchOperations() {
	r := s.Require()

	// Test combining title change and rename for folder
	{
		r.NoError(s.app.Content.CreateFolder("folder", model.ContentMeta{Title: "Old Title"}))

		res := s.api("PATCH", "/pages/folder",
			[]model.PatchOperation{
				{Op: "replace", Path: "/folder/url", Value: str2json("renamed-folder")},
				{Op: "replace", Path: "/folder/meta/title", Value: str2json("New Title")},
			},
			s.adminToken)
		r.Equal(200, res.Code)

		// Old folder should not exist
		r.False(s.app.Content.IsFolder("folder"))

		// New folder should exist with new title
		r.True(s.app.Content.IsFolder("renamed-folder"))
		folder, err := s.app.Content.ReadFolder("renamed-folder")
		r.NoError(err)
		r.Equal("New Title", folder.Meta.Title)

		// Cleanup
		r.NoError(s.app.Content.DeleteEmptyFolder("renamed-folder"))
	}

	// Test combining title change and rename for page
	{
		r.NoError(s.app.Content.SavePage("page", "Content", model.ContentMeta{Title: "Old Title"}, ""))

		res := s.api("PATCH", "/pages/page",
			[]model.PatchOperation{
				{Op: "replace", Path: "/page/url", Value: str2json("renamed-page")},
				{Op: "replace", Path: "/page/meta/title", Value: str2json("New Title")},
			},
			s.adminToken)
		r.Equal(200, res.Code)

		// Old page should not exist
		r.False(s.app.Content.IsPage("page"))

		// New page should exist with new title and same content
		r.True(s.app.Content.IsPage("renamed-page"))
		page, err := s.app.Content.ReadPage("renamed-page", nil)
		r.NoError(err)
		r.Equal("New Title", page.Meta.Title)
		r.Equal("Content", page.Content)

		// Cleanup
		r.NoError(s.app.Content.DeletePage("renamed-page"))
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
			"",
		))
		r.NoError(err)
	}
	time.Sleep(1050 * time.Millisecond) // Only one revision per second possible
	for _, url := range urls {
		err := (s.app.Content.SavePage(
			url,
			"New content",
			model.ContentMeta{Title: "New title", Tags: []string{"new tag"}},
			"",
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
			"",
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

			body, _ := jsonbody[model.SearchResponse](res)
			r.Len(body.Items, tc.nResults)

			for _, hit := range body.Items {
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

			body, _ := jsonbody[model.SearchResponse](res)
			r.Len(body.Items, tc.nResults)

			for _, hit := range body.Items {
				r.Nil(hit.EffectiveACL)
				r.Nil(hit.Meta.ACL)
				r.NotEmpty(hit.Url)
				r.Equal("Title", hit.Meta.Title)
				r.Len(hit.Meta.Tags, 1)
				r.Equal("tag", hit.Meta.Tags[0])

				r.NotEmpty(hit.Fragments[tc.name])
				r.Len(hit.Fragments[tc.name], 1)
				switch tc.name {
				case "content":
					r.Equal("<mark>Content</mark>", hit.Fragments[tc.name][0])
				case "meta.title":
					r.Equal("<mark>Title</mark>", hit.Fragments[tc.name][0])
				case "meta.tags":
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

			body, _ := jsonbody[model.SearchResponse](res)
			r.Len(body.Items, tc.nResults)

			for _, hit := range body.Items {
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

// TestSearchPaginationWithACL tests pagination when many results are filtered by ACL.
// Creates 20 pages: 10 in admin-only folder (not accessible to users) and 10 in public folder.
// Verifies that regular users can paginate through only the 10 accessible pages.
func (s *ContentTestSuite) TestSearchPaginationWithACL() {
	r := s.Require()

	// Create 10 pages in admin-only folder (not accessible to regular users)
	for i := 0; i < 10; i++ {
		err := s.app.Content.SavePage(
			"admin-only/paginationtest-"+strconv.Itoa(i),
			"PaginationTest content",
			model.ContentMeta{Title: "PaginationTest " + strconv.Itoa(i)},
			"",
		)
		r.NoError(err)
	}

	// Create 10 pages in public folder (accessible to everyone)
	for i := 0; i < 10; i++ {
		err := s.app.Content.SavePage(
			"public/paginationtest-"+strconv.Itoa(i),
			"PaginationTest content",
			model.ContentMeta{Title: "PaginationTest " + strconv.Itoa(i)},
			"",
		)
		r.NoError(err)
	}

	// Test 1: Admin can see all 20 pages
	{
		res := s.api("POST", "/search?q=PaginationTest&page=1&limit=100", nil, s.adminToken)
		r.Equal(200, res.Code)
		body, _ := jsonbody[model.SearchResponse](res)
		r.Len(body.Items, 20, "Admin should get all 20 results")
		r.False(body.HasMore, "Admin should not have more pages when all results fit")
	}

	// Test 2: Admin pagination with limit 5
	{
		allUrls := make(map[string]bool)

		for page := 1; page <= 4; page++ {
			res := s.api("POST", "/search?q=PaginationTest&page="+strconv.Itoa(page)+"&limit=5", nil, s.adminToken)
			r.Equal(200, res.Code)
			body, _ := jsonbody[model.SearchResponse](res)
			r.Len(body.Items, 5, "Admin should get 5 results on page %d", page)
			r.Equal(page, body.Page)

			for _, item := range body.Items {
				r.False(allUrls[item.Url], "Should not have duplicate results across pages")
				allUrls[item.Url] = true
			}

			if page < 4 {
				r.True(body.HasMore, "Admin should have more pages on page %d", page)
			} else {
				r.False(body.HasMore, "Admin should not have more pages on page 4")
			}
		}

		r.Len(allUrls, 20, "Admin should see all 20 pages total")
	}

	// Test 3: Regular user can only see 10 public pages
	{
		res := s.api("POST", "/search?q=PaginationTest&page=1&limit=100", nil, s.userToken)
		r.Equal(200, res.Code)
		body, _ := jsonbody[model.SearchResponse](res)
		r.Len(body.Items, 10, "User should get only 10 results")
		r.False(body.HasMore, "User should not have more pages")

		// All results should be from public folder
		for _, item := range body.Items {
			r.Contains(item.Url, "public/", "User should only see public pages")
		}
	}

	// Test 4: User pagination with limit 3
	{
		allUrls := make(map[string]bool)

		// Page 1
		res := s.api("POST", "/search?q=PaginationTest&page=1&limit=3", nil, s.userToken)
		r.Equal(200, res.Code)
		body, _ := jsonbody[model.SearchResponse](res)
		r.Len(body.Items, 3, "User should get 3 results on page 1")
		r.True(body.HasMore, "User should have more pages")

		for _, item := range body.Items {
			r.Contains(item.Url, "public/", "User should only see public pages")
			allUrls[item.Url] = true
		}

		// Page 2
		res = s.api("POST", "/search?q=PaginationTest&page=2&limit=3", nil, s.userToken)
		r.Equal(200, res.Code)
		body, _ = jsonbody[model.SearchResponse](res)
		r.Len(body.Items, 3, "User should get 3 results on page 2")
		r.True(body.HasMore, "User should have more pages on page 2")

		for _, item := range body.Items {
			r.Contains(item.Url, "public/", "User should only see public pages")
			r.False(allUrls[item.Url], "Should not have duplicate results")
			allUrls[item.Url] = true
		}

		// Page 3
		res = s.api("POST", "/search?q=PaginationTest&page=3&limit=3", nil, s.userToken)
		r.Equal(200, res.Code)
		body, _ = jsonbody[model.SearchResponse](res)
		r.Len(body.Items, 3, "User should get 3 results on page 3")
		r.True(body.HasMore, "User should have more pages on page 3")

		for _, item := range body.Items {
			r.Contains(item.Url, "public/", "User should only see public pages")
			r.False(allUrls[item.Url], "Should not have duplicate results")
			allUrls[item.Url] = true
		}

		// Page 4 (last page with 1 result)
		res = s.api("POST", "/search?q=PaginationTest&page=4&limit=3", nil, s.userToken)
		r.Equal(200, res.Code)
		body, _ = jsonbody[model.SearchResponse](res)
		r.Len(body.Items, 1, "User should get 1 result on page 4")
		r.False(body.HasMore, "User should not have more pages on page 4")

		for _, item := range body.Items {
			r.Contains(item.Url, "public/", "User should only see public pages")
			r.False(allUrls[item.Url], "Should not have duplicate results")
			allUrls[item.Url] = true
		}

		r.Len(allUrls, 10, "User should see all 10 public pages total")
	}

	// Test 5: Anonymous user sees same as regular user (public folder is accessible)
	{
		res := s.api("POST", "/search?q=PaginationTest&page=1&limit=100", nil, nil)
		r.Equal(200, res.Code)
		body, _ := jsonbody[model.SearchResponse](res)
		r.Len(body.Items, 10, "Anonymous should get 10 results")
		r.False(body.HasMore, "Anonymous should not have more pages")

		for _, item := range body.Items {
			r.Contains(item.Url, "public/", "Anonymous should only see public pages")
		}
	}

	// Test 6: Response metadata
	{
		res := s.api("POST", "/search?q=PaginationTest&page=2&limit=4", nil, s.userToken)
		r.Equal(200, res.Code)
		body, _ := jsonbody[model.SearchResponse](res)
		r.Equal(2, body.Page, "Response should include correct page number")
		r.Equal(4, body.Limit, "Response should include correct limit")
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
			r.NoError(s.app.Content.SavePage(tc.srcUrl, "Content", model.ContentMeta{Title: "Title"}, ""))

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
	r.NoError(s.app.Content.SavePage("page1", "Content1", model.ContentMeta{Title: "Page1"}, ""))
	r.NoError(s.app.Content.SavePage("page2", "Content2", model.ContentMeta{Title: "Page2"}, ""))
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
	r.NoError(s.app.Content.SavePage("page", "Content v1", model.ContentMeta{Title: "Title"}, ""))
	time.Sleep(1050 * time.Millisecond) // Only one revision per second possible
	r.NoError(s.app.Content.SavePage("page", "Content v2", model.ContentMeta{Title: "Title"}, ""))

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
	r.NoError(s.app.Content.SavePage("page", "Content", model.ContentMeta{Title: "Page"}, ""))

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
	r.NoError(s.app.Content.SavePage("folder/page1", "Content1", model.ContentMeta{Title: "Page1"}, ""))
	r.NoError(s.app.Content.SavePage("folder/page2", "Content2", model.ContentMeta{Title: "Page2"}, ""))
	r.NoError(s.app.Content.CreateFolder("folder/subfolder", model.ContentMeta{Title: "Subfolder"}))
	r.NoError(s.app.Content.SavePage("folder/subfolder/page3", "Content3", model.ContentMeta{Title: "Page3"}, ""))

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
	r.NoError(s.app.Content.SavePage("folder/page", "Content v1", model.ContentMeta{Title: "Page"}, ""))
	time.Sleep(1050 * time.Millisecond)
	r.NoError(s.app.Content.SavePage("folder/page", "Content v2", model.ContentMeta{Title: "Page"}, ""))

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

// TestAllowWriteWithPageACL verifies that when a page has its own ACL granting write permission,
// the AllowWrite field in the response is true, even if the parent folder denies write access.
// This tests that effective permissions are calculated from the page's own ACL, not just the parent's.
func (s *ContentTestSuite) TestAllowWriteWithPageACL() {
	r := s.Require()

	// Create a page in admin-only folder (which denies write to normal users)
	url := "admin-only/page-with-acl"
	acl := []model.AccessRule{
		{Subject: "all", Operations: []model.AccessOp{model.AccessOpRead, model.AccessOpWrite}},
	}
	r.NoError(s.app.Content.SavePage(url, "Content", model.ContentMeta{Title: "Page with ACL", ACL: &acl}, ""))

	// Test: Admin should have write and delete access
	{
		res := s.api("GET", "/pages/"+url, nil, s.adminToken)
		r.Equal(200, res.Code)
		body, _ := jsonbody[model.GetContentResponse](res)
		r.NotNil(body.Page)
		r.True(body.AllowWrite, "Admin should have write access")
		r.True(body.AllowDelete, "Admin should have delete access")
	}

	// Test: Normal user should have only write access (due to page's ACL granting "all")
	{
		res := s.api("GET", "/pages/"+url, nil, s.userToken)
		r.Equal(200, res.Code)
		body, _ := jsonbody[model.GetContentResponse](res)
		r.NotNil(body.Page)
		r.True(body.AllowWrite, "User should have write access")
		r.False(body.AllowDelete, "User should not have delete access")
	}

	// Cleanup
	r.NoError(s.app.Content.DeletePage(url))
}

// TestAllowWriteWithFolderACL verifies that when a folder has its own ACL granting write permission,
// the AllowWrite field in the response is true, even if the parent folder denies write access.
func (s *ContentTestSuite) TestAllowWriteWithFolderACL() {
	r := s.Require()

	// Create a folder in admin-only folder with its own ACL
	url := "admin-only/folder-with-acl"
	acl := []model.AccessRule{
		{Subject: "all", Operations: []model.AccessOp{model.AccessOpRead, model.AccessOpWrite}},
	}
	r.NoError(s.app.Content.CreateFolder(url, model.ContentMeta{Title: "Folder with ACL", ACL: &acl}))

	// Test: Admin should have write access
	{
		res := s.api("GET", "/pages/"+url, nil, s.adminToken)
		r.Equal(200, res.Code)
		body, _ := jsonbody[model.GetContentResponse](res)
		r.NotNil(body.Folder)
		r.True(body.AllowWrite, "Admin should have write access")
		r.True(body.AllowDelete, "Admin should have delete access")
	}

	// Test: Normal user should have write access (due to folder's ACL granting "all")
	{
		res := s.api("GET", "/pages/"+url, nil, s.userToken)
		r.Equal(200, res.Code)
		body, _ := jsonbody[model.GetContentResponse](res)
		r.NotNil(body.Folder)
		r.True(body.AllowWrite, "User should have write access")
		r.False(body.AllowDelete, "User should not have delete access")
	}

	// Cleanup
	r.NoError(s.app.Content.DeleteEmptyFolder(url))
}

// TestNonexistentContentPermissions tests that AllowWrite and AllowDelete are correctly
// determined by the parent's ACL when accessing non-existent content.
func (s *ContentTestSuite) TestNonexistentContentPermissions() {
	tests := []struct {
		name         string
		token        *string
		url          string
		responseCode int
		allowWrite   bool
		allowDelete  bool
	}{
		// Parent doesn't exist
		{"admin:parentNotExist", s.adminToken, "nonexistent-parent/page", 404, false, false},
		{"user:parentNotExist", s.userToken, "nonexistent-parent/page", 404, false, false},
		{"anonymous:parentNotExist", nil, "nonexistent-parent/page", 401, false, false},

		// Invalid URL
		{"admin:invalidURL", s.adminToken, "public/_invalid", 404, false, false},
		{"user:invalidURL", s.userToken, "public/_invalid", 404, false, false},
		{"anonymous:invalidURL", nil, "public/_invalid", 404, false, false},

		// Parent exists but user can't write
		{"admin:adminOnly", s.adminToken, "admin-only/nonexistent", 404, true, false},
		{"user:adminOnly", s.userToken, "admin-only/nonexistent", 403, false, false},
		{"anonymous:adminOnly", nil, "admin-only/nonexistent", 401, false, false},

		// Parent exists but user can only read
		{"admin:readOnly", s.adminToken, "read-only/nonexistent", 404, true, false},
		{"user:readOnly", s.userToken, "read-only/nonexistent", 404, false, false},
		{"anonymous:readOnly", nil, "read-only/nonexistent", 401, false, false},

		// Parent exists and allows write/delete for all
		{"admin:public", s.adminToken, "public/nonexistent", 404, true, false},
		{"user:public", s.userToken, "public/nonexistent", 404, true, false},
		{"anonymous:public", nil, "public/nonexistent", 404, true, false},

		// Parent exists, users can write/delete
		{"admin:published", s.adminToken, "published/nonexistent", 404, true, false},
		{"user:published", s.userToken, "published/nonexistent", 404, true, false},
		{"anonymous:published", nil, "published/nonexistent", 404, false, false},
	}

	for _, tc := range tests {
		t := s.T()
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			res := s.api("GET", "/pages/"+tc.url, nil, tc.token)
			r.Equal(tc.responseCode, res.Code)

			// Only check AllowWrite/AllowDelete for 404 responses (non-existent content)
			if tc.responseCode == 404 {
				body, _ := jsonbody[model.GetContentResponse](res)
				r.Nil(body.Page, "Page should be nil for non-existent content")
				r.Nil(body.Folder, "Folder should be nil for non-existent content")
				r.Equal(tc.allowWrite, body.AllowWrite, "AllowWrite mismatch")
				r.Equal(tc.allowDelete, body.AllowDelete, "AllowDelete mismatch")
			}
		})
	}
}

func (s *ContentTestSuite) TestMoveFolderSearchIndex() {
	r := s.Require()

	// Create folder with a page
	r.NoError(s.app.Content.CreateFolder("folder", model.ContentMeta{Title: "UniqueFolder"}))
	r.NoError(s.app.Content.SavePage("folder/page", "UniqueContent", model.ContentMeta{Title: "UniquePage"}, ""))

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

// TestModifiedByReturnsUserInfo tests that the API returns the username and display name (not userId)
// in the modifiedByUsername and modifiedByDisplayName fields of the response.
func (s *ContentTestSuite) TestModifiedByReturnsUserInfo() {
	r := s.Require()

	// Create a page as admin user via API
	{
		res := s.api("PUT", "/pages/test-page",
			model.PutRequest{Page: &model.Page{Content: "Content", Meta: model.ContentMeta{Title: "Test Page"}}},
			s.adminToken)
		r.Equal(200, res.Code)
	}

	// Read the page via API and verify modifiedBy fields contain user info, not userId
	{
		res := s.api("GET", "/pages/test-page", nil, s.adminToken)
		r.Equal(200, res.Code)

		body, _ := jsonbody[model.GetContentResponse](res)
		r.NotNil(body.Page)

		// The API should return the username "admin" and display name "Administrator"
		r.Equal("admin", body.Page.Meta.ModifiedByUsername, "API should return username")
		r.Equal("Administrator", body.Page.Meta.ModifiedByDisplayName, "API should return display name")
		r.NotEqual(s.adminUserID, body.Page.Meta.ModifiedByUsername, "API should not return internal userId")
	}

	// Update the page as regular user via API
	{
		// First grant write access to users
		acl := []model.AccessRule{{Subject: "all", Operations: []model.AccessOp{model.AccessOpRead, model.AccessOpWrite, model.AccessOpDelete}}}
		res := s.api("PATCH", "/pages/test-page",
			[]model.PatchOperation{{Op: "replace", Path: "/page/meta/acl", Value: acl2json(acl)}},
			s.adminToken)
		r.Equal(200, res.Code)

		// Now update as regular user
		res = s.api("PUT", "/pages/test-page",
			model.PutRequest{Page: &model.Page{Content: "Updated content", Meta: model.ContentMeta{Title: "Test Page"}}},
			s.userToken)
		r.Equal(200, res.Code)
	}

	// Read the page via API and verify modifiedBy is now "user" with display name "User"
	{
		res := s.api("GET", "/pages/test-page", nil, s.userToken)
		r.Equal(200, res.Code)

		body, _ := jsonbody[model.GetContentResponse](res)
		r.NotNil(body.Page)

		// The API should return the username "user" and display name "User"
		r.Equal("user", body.Page.Meta.ModifiedByUsername, "API should return username 'user'")
		r.Equal("User", body.Page.Meta.ModifiedByDisplayName, "API should return display name 'User'")
		r.NotEqual(s.userUserID, body.Page.Meta.ModifiedByUsername, "API should not return internal userId")
	}

	// Verify internal storage still contains userId (not username)
	{
		page, err := s.app.Content.ReadPage("test-page", nil)
		r.NoError(err)
		r.Equal(s.userUserID, page.Meta.ModifiedByUserID, "Internal storage should contain userId")
		r.NotEqual("user", page.Meta.ModifiedByUserID, "Internal storage should not contain username")
	}

	// Cleanup
	r.NoError(s.app.Content.DeletePage("test-page"))
}

// ==================== TRASH API TESTS ====================

// TestTrashListAPI tests the GET /trash/ endpoint
func (s *ContentTestSuite) TestTrashListAPI() {
	r := s.Require()

	// Create pages
	r.NoError(s.app.Content.SavePage("page1", "Content1", model.ContentMeta{Title: "Page 1"}, ""))
	r.NoError(s.app.Content.SavePage("page2", "Content2", model.ContentMeta{Title: "Page 2"}, ""))

	// Record time range for timestamp validation
	beforeTime := time.Now().Unix()
	r.NoError(s.app.Content.DeletePage("page1"))
	time.Sleep(10 * time.Millisecond)
	r.NoError(s.app.Content.DeletePage("page2"))
	afterTime := time.Now().Unix()

	tests := []struct {
		name         string
		token        *string
		responseCode int
	}{
		{"admin", s.adminToken, 200},
		{"user", s.userToken, 403},
		{"anonymous", nil, 401},
	}

	for _, tc := range tests {
		t := s.T()
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			res := s.api("GET", "/trash/", nil, tc.token)
			r.Equal(tc.responseCode, res.Code)

			if tc.responseCode == 200 {
				body, _ := jsonbody[model.GetTrashListResponse](res)
				r.Len(body.Items, 2)
				r.Equal(2, body.TotalCount)

				// Verify DeletedAt timestamps are within expected range
				for _, item := range body.Items {
					r.GreaterOrEqual(item.DeletedAt, beforeTime, "DeletedAt should be >= beforeTime")
					r.LessOrEqual(item.DeletedAt, afterTime, "DeletedAt should be <= afterTime")
				}
			}
		})
	}

	// Test pagination
	{
		res := s.api("GET", "/trash/?page=1&limit=1", nil, s.adminToken)
		r.Equal(200, res.Code)
		body, _ := jsonbody[model.GetTrashListResponse](res)
		r.Len(body.Items, 1)
		r.Equal(2, body.TotalCount)
		r.Equal(1, body.Page)
		r.Equal(1, body.Limit)
	}

	// Test sorting by url
	{
		res := s.api("GET", "/trash/?sortBy=url&sortOrder=asc", nil, s.adminToken)
		r.Equal(200, res.Code)
		body, _ := jsonbody[model.GetTrashListResponse](res)
		r.Len(body.Items, 2)
		r.Equal("page1", body.Items[0].Url)
		r.Equal("page2", body.Items[1].Url)
	}
}

// TestTrashShowPageAPI tests the GET /trash/page endpoint
func (s *ContentTestSuite) TestTrashShowPageAPI() {
	r := s.Require()

	// Create, save, and delete a page
	r.NoError(s.app.Content.SavePage("page", "Content", model.ContentMeta{Title: "Test Page"}, ""))
	r.NoError(s.app.Content.DeletePage("page"))

	// Get the trash entry
	trashEntries, err := s.app.Content.ListTrash()
	r.NoError(err)
	r.Len(trashEntries, 1)
	entry := trashEntries[0]

	tests := []struct {
		name         string
		token        *string
		responseCode int
	}{
		{"admin", s.adminToken, 200},
		{"user", s.userToken, 403},
		{"anonymous", nil, 401},
	}

	for _, tc := range tests {
		t := s.T()
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			res := s.api("GET", "/trash/page?url="+entry.Url+"&deletedAt="+strconv.FormatInt(entry.DeletedAt, 10), nil, tc.token)
			r.Equal(tc.responseCode, res.Code)

			if tc.responseCode == 200 {
				body, _ := jsonbody[model.GetTrashPageResponse](res)
				r.Equal("Content", body.Page.Content)
				r.Equal("Test Page", body.Page.Meta.Title)
			}
		})
	}

	// Test 404 for non-existent item
	{
		res := s.api("GET", "/trash/page?url=nonexistent&deletedAt=12345", nil, s.adminToken)
		r.Equal(404, res.Code)
	}

	// Test missing parameters
	{
		res := s.api("GET", "/trash/page", nil, s.adminToken)
		r.Equal(400, res.Code)

		res = s.api("GET", "/trash/page?url=page", nil, s.adminToken)
		r.Equal(400, res.Code)
	}
}

// TestTrashDeleteAPI tests the POST /trash/delete endpoint
func (s *ContentTestSuite) TestTrashDeleteAPI() {
	tests := []struct {
		name         string
		token        *string
		responseCode int
	}{
		{"admin", s.adminToken, 200},
		{"user", s.userToken, 403},
		{"anonymous", nil, 401},
	}

	for _, tc := range tests {
		t := s.T()
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			// Setup: create and delete a page
			r.NoError(s.app.Content.SavePage("page", "Content", model.ContentMeta{Title: "Page"}, ""))
			r.NoError(s.app.Content.DeletePage("page"))

			trashBefore, err := s.app.Content.ListTrash()
			r.NoError(err)
			r.Len(trashBefore, 1)

			// Test
			res := s.api("POST", "/trash/delete",
				model.TrashActionRequest{Items: []model.TrashItemRef{{Url: trashBefore[0].Url, DeletedAt: trashBefore[0].DeletedAt}}},
				tc.token)
			r.Equal(tc.responseCode, res.Code)

			trashAfter, err := s.app.Content.ListTrash()
			r.NoError(err)

			if tc.responseCode == 200 {
				r.Len(trashAfter, 0, "Trash should be empty after permanent deletion")
			} else {
				r.Len(trashAfter, 1, "Trash should still have the entry")
				// Cleanup
				r.NoError(s.app.Content.DeleteTrashEntry(trashAfter[0].Url, trashAfter[0].DeletedAt))
			}
		})
	}
}

// TestTrashRestoreAPI tests the POST /trash/restore endpoint
func (s *ContentTestSuite) TestTrashRestoreAPI() {
	tests := []struct {
		name         string
		token        *string
		responseCode int
	}{
		{"admin", s.adminToken, 200},
		{"user", s.userToken, 403},
		{"anonymous", nil, 401},
	}

	for _, tc := range tests {
		t := s.T()
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			// Setup: create and delete a page
			r.NoError(s.app.Content.SavePage("page", "Content", model.ContentMeta{Title: "Restored Page"}, ""))
			r.NoError(s.app.Content.DeletePage("page"))

			trashBefore, err := s.app.Content.ListTrash()
			r.NoError(err)
			r.Len(trashBefore, 1)

			// Test
			res := s.api("POST", "/trash/restore",
				model.TrashActionRequest{Items: []model.TrashItemRef{{Url: trashBefore[0].Url, DeletedAt: trashBefore[0].DeletedAt}}},
				tc.token)
			r.Equal(tc.responseCode, res.Code)

			if tc.responseCode == 200 {
				r.True(s.app.Content.IsPage("page"), "Page should be restored")

				page, err := s.app.Content.ReadPage("page", nil)
				r.NoError(err)
				r.Equal("Content", page.Content)
				r.Equal("Restored Page", page.Meta.Title)

				trashAfter, err := s.app.Content.ListTrash()
				r.NoError(err)
				r.Len(trashAfter, 0, "Trash should be empty after restore")

				// Cleanup
				r.NoError(s.app.Content.DeletePage("page"))
				trashAfter, _ = s.app.Content.ListTrash()
				for _, e := range trashAfter {
					r.NoError(s.app.Content.DeleteTrashEntry(e.Url, e.DeletedAt))
				}
			} else {
				r.False(s.app.Content.IsPage("page"), "Page should not be restored")
				// Cleanup
				r.NoError(s.app.Content.DeleteTrashEntry(trashBefore[0].Url, trashBefore[0].DeletedAt))
			}
		})
	}
}

// TestTrashRestoreDeepNestedPage tests restoring a deeply nested page with auto-created parent folders
func (s *ContentTestSuite) TestTrashRestoreDeepNestedPage() {
	r := s.Require()

	// Create deeply nested page: a/b/c/d/page
	r.NoError(s.app.Content.CreateFolder("nested-a", model.ContentMeta{Title: "A"}))
	r.NoError(s.app.Content.CreateFolder("nested-a/b", model.ContentMeta{Title: "B"}))
	r.NoError(s.app.Content.CreateFolder("nested-a/b/c", model.ContentMeta{Title: "C"}))
	r.NoError(s.app.Content.CreateFolder("nested-a/b/c/d", model.ContentMeta{Title: "D"}))
	r.NoError(s.app.Content.SavePage("nested-a/b/c/d/page", "Deep Content", model.ContentMeta{Title: "Deep Page"}, ""))

	// Create multiple attic versions
	time.Sleep(1050 * time.Millisecond)
	r.NoError(s.app.Content.SavePage("nested-a/b/c/d/page", "Deep Content v2", model.ContentMeta{Title: "Deep Page v2"}, ""))
	time.Sleep(1050 * time.Millisecond)
	r.NoError(s.app.Content.SavePage("nested-a/b/c/d/page", "Deep Content v3", model.ContentMeta{Title: "Deep Page v3"}, ""))

	// Verify we have 3 attic entries
	atticBefore, err := s.app.Content.ListAttic("nested-a/b/c/d/page")
	r.NoError(err)
	r.Len(atticBefore, 3, "Should have 3 attic entries before deletion")

	// Delete the entire folder structure
	r.NoError(s.app.Content.DeleteFolder("nested-a"))

	// Verify page is in trash
	trashEntries, err := s.app.Content.ListTrash()
	r.NoError(err)
	r.Len(trashEntries, 1)
	r.Equal("nested-a/b/c/d/page", trashEntries[0].Url)

	// Verify all folders are gone
	r.False(s.app.Content.IsFolder("nested-a"))
	r.False(s.app.Content.IsFolder("nested-a/b"))
	r.False(s.app.Content.IsFolder("nested-a/b/c"))
	r.False(s.app.Content.IsFolder("nested-a/b/c/d"))

	// Restore via API
	res := s.api("POST", "/trash/restore",
		model.TrashActionRequest{Items: []model.TrashItemRef{{Url: trashEntries[0].Url, DeletedAt: trashEntries[0].DeletedAt}}},
		s.adminToken)
	r.Equal(200, res.Code)

	// Verify all parent folders were auto-created
	r.True(s.app.Content.IsFolder("nested-a"), "nested-a folder should be auto-created")
	r.True(s.app.Content.IsFolder("nested-a/b"), "nested-a/b folder should be auto-created")
	r.True(s.app.Content.IsFolder("nested-a/b/c"), "nested-a/b/c folder should be auto-created")
	r.True(s.app.Content.IsFolder("nested-a/b/c/d"), "nested-a/b/c/d folder should be auto-created")

	// Verify page was restored with latest content
	r.True(s.app.Content.IsPage("nested-a/b/c/d/page"))
	page, err := s.app.Content.ReadPage("nested-a/b/c/d/page", nil)
	r.NoError(err)
	r.Equal("Deep Content v3", page.Content)
	r.Equal("Deep Page v3", page.Meta.Title)

	// Verify all attic entries were restored
	atticAfter, err := s.app.Content.ListAttic("nested-a/b/c/d/page")
	r.NoError(err)
	r.Len(atticAfter, 3, "All 3 attic entries should be restored")

	// Verify we can read old revisions
	for i, entry := range atticAfter {
		oldPage, err := s.app.Content.ReadPage("nested-a/b/c/d/page", &entry.Revision)
		r.NoError(err)
		r.NotEmpty(oldPage.Content, "Revision %d (rev %d) should have content", i+1, entry.Revision)
	}

	// Verify trash is empty
	trashAfter, err := s.app.Content.ListTrash()
	r.NoError(err)
	r.Len(trashAfter, 0, "Trash should be empty after restore")

	// Verify page is searchable
	results, err := s.app.Content.Search("Deep")
	r.NoError(err)
	r.GreaterOrEqual(len(results), 1, "Restored page should be searchable")
}

// TestTrashRestoreConflict tests that restore fails when destination already exists
func (s *ContentTestSuite) TestTrashRestoreConflict() {
	r := s.Require()

	// Create and delete a page
	r.NoError(s.app.Content.SavePage("conflict-page", "Original", model.ContentMeta{Title: "Original"}, ""))
	r.NoError(s.app.Content.DeletePage("conflict-page"))

	trashEntries, err := s.app.Content.ListTrash()
	r.NoError(err)
	r.Len(trashEntries, 1)

	// Create a new page at the same location
	r.NoError(s.app.Content.SavePage("conflict-page", "New", model.ContentMeta{Title: "New"}, ""))

	// Try to restore - should fail with 409 Conflict
	res := s.api("POST", "/trash/restore",
		model.TrashActionRequest{Items: []model.TrashItemRef{{Url: trashEntries[0].Url, DeletedAt: trashEntries[0].DeletedAt}}},
		s.adminToken)
	r.Equal(409, res.Code)

	// Page should still have the new content
	page, err := s.app.Content.ReadPage("conflict-page", nil)
	r.NoError(err)
	r.Equal("New", page.Content)

	// Trash should still have the entry
	trashAfter, err := s.app.Content.ListTrash()
	r.NoError(err)
	r.Len(trashAfter, 1)
}

// TestFolderContentFiltering tests that folder entries (pages and subfolders) are filtered
// based on the user's read access permissions.
func (s *ContentTestSuite) TestFolderContentFiltering() {
	r := s.Require()

	adminOnlyACL := []model.AccessRule{}
	usersOnlyACL := []model.AccessRule{
		{Subject: "all", Operations: []model.AccessOp{model.AccessOpRead}},
	}
	publicACL := []model.AccessRule{
		{Subject: "anonymous", Operations: []model.AccessOp{model.AccessOpRead, model.AccessOpWrite, model.AccessOpDelete}},
	}

	// === Mixed content with different ACLs ===
	// Create a public folder with mixed content that has different permissions
	r.NoError(s.app.Content.CreateFolder("test-folder", model.ContentMeta{Title: "Test Folder", ACL: &publicACL}))

	// Create pages with different ACLs inside the folder
	r.NoError(s.app.Content.SavePage("test-folder/public-page", "Public content", model.ContentMeta{Title: "Public Page"}, ""))
	r.NoError(s.app.Content.SavePage("test-folder/admin-page", "Admin content", model.ContentMeta{Title: "Admin Page", ACL: &adminOnlyACL}, ""))
	r.NoError(s.app.Content.SavePage("test-folder/users-page", "Users content", model.ContentMeta{Title: "Users Page", ACL: &usersOnlyACL}, ""))

	// Create subfolders with different ACLs
	r.NoError(s.app.Content.CreateFolder("test-folder/public-subfolder", model.ContentMeta{Title: "Public Subfolder"}))
	r.NoError(s.app.Content.CreateFolder("test-folder/admin-subfolder", model.ContentMeta{Title: "Admin Subfolder", ACL: &adminOnlyACL}))

	// Test: Admin sees all 5 entries
	{
		res := s.api("GET", "/pages/test-folder", nil, s.adminToken)
		r.Equal(200, res.Code)
		body, _ := jsonbody[model.GetContentResponse](res)
		r.NotNil(body.Folder)
		r.Len(body.Folder.Content, 5, "Admin should see all 5 entries")

		names := make([]string, len(body.Folder.Content))
		for i, entry := range body.Folder.Content {
			names[i] = entry.Name
		}
		r.Contains(names, "public-page")
		r.Contains(names, "admin-page")
		r.Contains(names, "users-page")
		r.Contains(names, "public-subfolder")
		r.Contains(names, "admin-subfolder")
	}

	// Test: Regular user sees 3 entries (excludes admin-page and admin-subfolder)
	{
		res := s.api("GET", "/pages/test-folder", nil, s.userToken)
		r.Equal(200, res.Code)
		body, _ := jsonbody[model.GetContentResponse](res)
		r.NotNil(body.Folder)
		r.Len(body.Folder.Content, 3, "User should see 3 entries")

		names := make([]string, len(body.Folder.Content))
		for i, entry := range body.Folder.Content {
			names[i] = entry.Name
		}
		r.Contains(names, "public-page")
		r.Contains(names, "users-page")
		r.Contains(names, "public-subfolder")
		r.NotContains(names, "admin-page")
		r.NotContains(names, "admin-subfolder")
	}

	// Test: Anonymous user sees 2 entries (only public ones)
	{
		res := s.api("GET", "/pages/test-folder", nil, nil)
		r.Equal(200, res.Code)
		body, _ := jsonbody[model.GetContentResponse](res)
		r.NotNil(body.Folder)
		r.Len(body.Folder.Content, 2, "Anonymous should see 2 entries")

		names := make([]string, len(body.Folder.Content))
		for i, entry := range body.Folder.Content {
			names[i] = entry.Name
		}
		r.Contains(names, "public-page")
		r.Contains(names, "public-subfolder")
		r.NotContains(names, "admin-page")
		r.NotContains(names, "users-page")
		r.NotContains(names, "admin-subfolder")
	}

	// === Entries with their own ACLs overriding parent ===
	// Create a folder that allows all users to read the folder listing
	r.NoError(s.app.Content.CreateFolder("mixed-access-folder", model.ContentMeta{Title: "Mixed Access Folder", ACL: &usersOnlyACL}))

	// Create entries with varying access levels
	r.NoError(s.app.Content.SavePage("mixed-access-folder/admin-page", "Admin content", model.ContentMeta{Title: "Admin Page", ACL: &adminOnlyACL}, ""))
	r.NoError(s.app.Content.SavePage("mixed-access-folder/users-page", "Users content", model.ContentMeta{Title: "Users Page"}, ""))
	anonymousReadACL := []model.AccessRule{{Subject: "anonymous", Operations: []model.AccessOp{model.AccessOpRead}}}
	r.NoError(s.app.Content.SavePage("mixed-access-folder/public-page", "Public content", model.ContentMeta{Title: "Public Page", ACL: &anonymousReadACL}, ""))
	r.NoError(s.app.Content.CreateFolder("mixed-access-folder/admin-subfolder", model.ContentMeta{Title: "Admin Subfolder", ACL: &adminOnlyACL}))

	// Test: Admin sees all 4 entries
	{
		res := s.api("GET", "/pages/mixed-access-folder", nil, s.adminToken)
		r.Equal(200, res.Code)
		body, _ := jsonbody[model.GetContentResponse](res)
		r.NotNil(body.Folder)
		r.Len(body.Folder.Content, 4, "Admin should see all 4 entries")
	}

	// Test: Regular user sees 2 entries
	{
		res := s.api("GET", "/pages/mixed-access-folder", nil, s.userToken)
		r.Equal(200, res.Code)
		body, _ := jsonbody[model.GetContentResponse](res)
		r.NotNil(body.Folder)
		r.Len(body.Folder.Content, 2, "User should see 2 entries")

		names := make([]string, len(body.Folder.Content))
		for i, entry := range body.Folder.Content {
			names[i] = entry.Name
		}
		r.Contains(names, "users-page")
		r.Contains(names, "public-page")
		r.NotContains(names, "admin-page")
		r.NotContains(names, "admin-subfolder")
	}

	// Test: Anonymous user cannot access this folder (folder requires user)
	{
		res := s.api("GET", "/pages/mixed-access-folder", nil, nil)
		r.Equal(401, res.Code)
	}

	// === Empty folder result ===
	// Create a public folder with only admin-only content
	r.NoError(s.app.Content.CreateFolder("public-folder", model.ContentMeta{Title: "Public Folder", ACL: &publicACL}))
	r.NoError(s.app.Content.SavePage("public-folder/admin-page", "Admin content", model.ContentMeta{Title: "Admin Page", ACL: &adminOnlyACL}, ""))
	r.NoError(s.app.Content.CreateFolder("public-folder/admin-subfolder", model.ContentMeta{Title: "Admin Subfolder", ACL: &adminOnlyACL}))

	// Test: Admin sees all entries
	{
		res := s.api("GET", "/pages/public-folder", nil, s.adminToken)
		r.Equal(200, res.Code)
		body, _ := jsonbody[model.GetContentResponse](res)
		r.NotNil(body.Folder)
		r.Len(body.Folder.Content, 2, "Admin should see 2 entries")
	}

	// Test: Regular user sees empty folder
	{
		res := s.api("GET", "/pages/public-folder", nil, s.userToken)
		r.Equal(200, res.Code)
		body, _ := jsonbody[model.GetContentResponse](res)
		r.NotNil(body.Folder)
		r.Len(body.Folder.Content, 0, "User should see 0 entries")
	}

	// Test: Anonymous user sees empty folder
	{
		res := s.api("GET", "/pages/public-folder", nil, nil)
		r.Equal(200, res.Code)
		body, _ := jsonbody[model.GetContentResponse](res)
		r.NotNil(body.Folder)
		r.Len(body.Folder.Content, 0, "Anonymous should see 0 entries")
	}
}
