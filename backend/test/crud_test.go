package test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/tfabritius/plainpage/model"
)

type CrudTestSuite struct {
	AppTestSuite
}

func TestCrudTestSuite(t *testing.T) {
	suite.Run(t, &CrudTestSuite{})
}

func (s *CrudTestSuite) SetupSuite() {
	s.setupInitialApp()

	r := s.Require()

	// Allow access for anonymous
	{
		res := s.api("GET", "/_api/pages", nil, s.adminToken)
		r.Equal(200, res.Code)
		body, _ := jsonbody[model.GetContentResponse](res)
		r.NotNil(body.Folder)
		r.NotNil(body.Folder.Meta.ACL)

		acl := *body.Folder.Meta.ACL
		acl = append(acl, model.AccessRule{Subject: "anonymous", Operations: []model.AccessOp{model.AccessOpRead, model.AccessOpWrite, model.AccessOpDelete}})

		aclBytes, err := json.Marshal(acl)
		r.Nil(err)
		aclJson := json.RawMessage(aclBytes)

		res = s.api("PATCH", "/_api/pages", []model.PatchOperation{{Op: "replace", Path: "/folder/meta/acl", Value: &aclJson}}, s.adminToken)
		r.Equal(200, res.Code)
	}
}

func (s *CrudTestSuite) TestCRUD() {
	r := s.Require()

	// Cleanup
	{
		res := s.api("DELETE", "/_api/pages/foo", nil, nil)
		r.NotEqual(500, res.Code)
	}

	// Create page
	{
		res := s.api("PUT", "/_api/pages/foo", model.PutRequest{Page: &model.Page{Meta: model.PageMeta{Title: "Foo"}}}, nil)
		r.Equal(200, res.Code)
	}
	{
		body, res := jsonbody[model.GetContentResponse](s.api("GET", "/_api/pages/foo", nil, nil))
		r.Equal(200, res.Code)
		r.Nil(body.Folder)
		r.Equal("Foo", body.Page.Meta.Title)
	}

	// Update page
	{
		res := s.api("PUT", "/_api/pages/foo", model.PutRequest{Page: &model.Page{Meta: model.PageMeta{Title: "Updated foo"}}}, nil)
		r.Equal(200, res.Code)
	}
	{
		body, res := jsonbody[model.GetContentResponse](s.api("GET", "/_api/pages/foo", nil, nil))
		r.Equal(200, res.Code)
		r.Equal("Updated foo", body.Page.Meta.Title)
	}

	// Delete page
	{
		res := s.api("DELETE", "/_api/pages/foo", nil, nil)
		r.Equal(200, res.Code)
	}
	{
		res := s.api("GET", "/_api/pages/foo", nil, nil)
		r.Equal(404, res.Code)
	}

	// Create folder
	{
		res := s.api("PUT", "/_api/pages/foo", model.PutRequest{Page: nil}, nil)
		r.Equal(200, res.Code)
	}
	{
		body, res := jsonbody[model.GetContentResponse](s.api("GET", "/_api/pages/foo", nil, nil))
		r.Equal(200, res.Code)
		r.Nil(body.Page)
		r.Len(body.Folder.Content, 0)
	}

	// List root folder
	{
		body, res := jsonbody[model.GetContentResponse](s.api("GET", "/_api/pages", nil, nil))
		r.Equal(200, res.Code)
		r.Nil(body.Page)
		r.Len(body.Folder.Content, 1)
		r.Equal("foo", body.Folder.Content[0].Name)
		r.Equal("/foo", body.Folder.Content[0].Url)
		r.True(body.Folder.Content[0].IsFolder)
	}

	// Create page in folder
	{
		res := s.api("PUT", "/_api/pages/foo/bar", model.PutRequest{Page: &model.Page{Meta: model.PageMeta{Title: "Bar"}}}, nil)
		r.Equal(200, res.Code)
	}
	{
		body, res := jsonbody[model.GetContentResponse](s.api("GET", "/_api/pages/foo/bar", nil, nil))
		r.Equal(200, res.Code)
		r.Equal("Bar", body.Page.Meta.Title)
	}

	// List folder
	{
		body, res := jsonbody[model.GetContentResponse](s.api("GET", "/_api/pages/foo", nil, nil))
		r.Equal(200, res.Code)
		r.Nil(body.Page)
		r.Len(body.Folder.Content, 1)
		r.Equal("bar", body.Folder.Content[0].Name)
		r.Equal("/foo/bar", body.Folder.Content[0].Url)
		r.False(body.Folder.Content[0].IsFolder)
	}

	// Update page in folder
	{
		time.Sleep(1050 * time.Millisecond) // Only one revision per second possible
		res := s.api("PUT", "/_api/pages/foo/bar", model.PutRequest{Page: &model.Page{Meta: model.PageMeta{Title: "New Bar"}}}, nil)
		r.Equal(200, res.Code)
	}
	{
		body, res := jsonbody[model.GetContentResponse](s.api("GET", "/_api/pages/foo/bar", nil, nil))
		r.Equal(200, res.Code)
		r.Equal("New Bar", body.Page.Meta.Title)
	}

	// List revisions in attic
	var firstRev int64
	{
		body, res := jsonbody[model.GetAtticListResponse](s.api("GET", "/_api/attic/foo/bar", nil, nil))
		r.Equal(200, res.Code)
		fmt.Println(body)
		r.Len(body.Entries, 2)
		firstRev = body.Entries[0].Revision
	}

	// Retrieve old revision from attic
	{
		body, res := jsonbody[model.GetContentResponse](s.api("GET", "/_api/attic/foo/bar?rev="+strconv.Itoa(int(firstRev)), nil, nil))
		r.Equal(200, res.Code)
		r.Nil(body.Folder)
		r.NotNil(body.Page)
		r.Equal("Bar", body.Page.Meta.Title)
	}

	// Delete non-empty folder
	{
		res := s.api("DELETE", "/_api/pages/foo", nil, nil)
		r.Equal(400, res.Code)
	}

	// Delete page in folder
	{
		res := s.api("DELETE", "/_api/pages/foo/bar", nil, nil)
		r.Equal(200, res.Code)
	}

	// Delete empty folder
	{
		res := s.api("DELETE", "/_api/pages/foo", nil, nil)
		r.Equal(200, res.Code)
	}
	{
		res := s.api("GET", "/_api/pages/foo", nil, nil)
		r.Equal(404, res.Code)
	}

	/*
	 *** ERRORS ***
	 */

	// Get nonexistent page
	{
		res := s.api("GET", "/_api/pages/foo", nil, nil)
		r.Equal(404, res.Code)
	}

	// Delete nonexistent page
	{
		res := s.api("DELETE", "/_api/pages/foo", nil, nil)
		r.Equal(404, res.Code)
	}

	// Create folder in nonexistent folder
	{
		res := s.api("PUT", "/_api/pages/foo/bar", model.PutRequest{Page: nil}, nil)
		r.Equal(400, res.Code)
	}

	// Create page in nonexistent folder
	{
		res := s.api("PUT", "/_api/pages/foo/bar", model.PutRequest{Page: &model.Page{Meta: model.PageMeta{Title: "Bar"}}}, nil)
		r.Equal(400, res.Code)
	}

	// Create page/folder where folder exists already
	{
		res := s.api("PUT", "/_api/pages/foo", model.PutRequest{Page: nil}, nil)
		r.Equal(200, res.Code)
	}
	{
		res := s.api("PUT", "/_api/pages/foo", model.PutRequest{Page: &model.Page{Meta: model.PageMeta{Title: "Foo"}}}, nil)
		r.Equal(400, res.Code)
	}
	{
		res := s.api("PUT", "/_api/pages/foo", model.PutRequest{Page: nil}, nil)
		r.Equal(400, res.Code)
	}
	{
		res := s.api("DELETE", "/_api/pages/foo", nil, nil)
		r.Equal(200, res.Code)
	}

	// Create folder where page exists already
	{
		res := s.api("PUT", "/_api/pages/foo", model.PutRequest{Page: &model.Page{Meta: model.PageMeta{Title: "Foo"}}}, nil)
		r.Equal(200, res.Code)
	}
	{
		res := s.api("PUT", "/_api/pages/foo", model.PutRequest{Page: nil}, nil)
		r.Equal(400, res.Code)
	}
	{
		res := s.api("DELETE", "/_api/pages/foo", nil, nil)
		r.Equal(200, res.Code)
	}
}
