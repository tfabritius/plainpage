package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tfabritius/plainpage/server"
	"github.com/tfabritius/plainpage/storage"
)

func TestCRUD(t *testing.T) {
	r := require.New(t)

	// Cleanup
	{
		res := api("DELETE", "/_api/pages/foo", nil)
		r.NotEqual(500, res.Code)
	}

	// Create page
	{
		res := api("PUT", "/_api/pages/foo", server.PutRequest{Page: &storage.Page{Meta: storage.PageMeta{Title: "Foo"}}})
		fmt.Println(string(res.Body.Bytes()))
		r.Equal(200, res.Code)
	}
	{
		body, res := jsonbody[server.GetPageResponse](api("GET", "/_api/pages/foo", nil))
		r.Equal(200, res.Code)
		r.Nil(body.Folder)
		r.Equal("Foo", body.Page.Meta.Title)
	}

	// Update page
	{
		res := api("PUT", "/_api/pages/foo", server.PutRequest{Page: &storage.Page{Meta: storage.PageMeta{Title: "Updated foo"}}})
		r.Equal(200, res.Code)
	}
	{
		body, res := jsonbody[server.GetPageResponse](api("GET", "/_api/pages/foo", nil))
		fmt.Println(body)
		r.Equal(200, res.Code)
		r.Equal("Updated foo", body.Page.Meta.Title)
	}

	// Delete page
	{
		res := api("DELETE", "/_api/pages/foo", nil)
		r.Equal(200, res.Code)
	}
	{
		res := api("GET", "/_api/pages/foo", nil)
		r.Equal(404, res.Code)
	}

	// Create folder
	{
		res := api("PUT", "/_api/pages/foo", server.PutRequest{Page: nil})
		r.Equal(200, res.Code)
	}
	{
		body, res := jsonbody[server.GetPageResponse](api("GET", "/_api/pages/foo", nil))
		fmt.Println(body)
		r.Equal(200, res.Code)
		r.Nil(body.Page)
		r.Len(body.Folder.Content, 0)
	}

	// List root folder
	{
		body, res := jsonbody[server.GetPageResponse](api("GET", "/_api/pages", nil))
		r.Equal(200, res.Code)
		r.Nil(body.Page)
		r.Len(body.Folder.Content, 1)
		r.Equal("foo", body.Folder.Content[0].Name)
		r.Equal("/foo", body.Folder.Content[0].Url)
		r.True(body.Folder.Content[0].IsFolder)
	}

	// Create page in folder
	{
		res := api("PUT", "/_api/pages/foo/bar", server.PutRequest{Page: &storage.Page{Meta: storage.PageMeta{Title: "Bar"}}})
		r.Equal(200, res.Code)
	}
	{
		body, res := jsonbody[server.GetPageResponse](api("GET", "/_api/pages/foo/bar", nil))
		fmt.Println(body)
		r.Equal(200, res.Code)
		r.Equal("Bar", body.Page.Meta.Title)
	}

	// List folder
	{
		body, res := jsonbody[server.GetPageResponse](api("GET", "/_api/pages/foo", nil))
		r.Equal(200, res.Code)
		r.Nil(body.Page)
		r.Len(body.Folder.Content, 1)
		r.Equal("bar", body.Folder.Content[0].Name)
		r.Equal("/foo/bar", body.Folder.Content[0].Url)
		r.False(body.Folder.Content[0].IsFolder)
	}

	// Delete non-empty folder
	{
		res := api("DELETE", "/_api/pages/foo", nil)
		r.Equal(400, res.Code)
	}

	// Delete page in folder
	{
		res := api("DELETE", "/_api/pages/foo/bar", nil)
		r.Equal(200, res.Code)
	}

	// Delete empty folder
	{
		res := api("DELETE", "/_api/pages/foo", nil)
		r.Equal(200, res.Code)
	}
	{
		res := api("GET", "/_api/pages/foo", nil)
		r.Equal(404, res.Code)
	}

	/*
	 *** ERRORS ***
	 */

	// Get nonexistent page
	{
		res := api("GET", "/_api/pages/foo", nil)
		r.Equal(404, res.Code)
	}

	// Delete nonexistent page
	{
		res := api("DELETE", "/_api/pages/foo", nil)
		r.Equal(404, res.Code)
	}

	// Create folder in nonexistent folder
	{
		res := api("PUT", "/_api/pages/foo/bar", server.PutRequest{Page: nil})
		r.Equal(400, res.Code)
	}

	// Create page in nonexistent folder
	{
		res := api("PUT", "/_api/pages/foo/bar", server.PutRequest{Page: &storage.Page{Meta: storage.PageMeta{Title: "Bar"}}})
		r.Equal(400, res.Code)
	}

	// Create page/folder where folder exists already
	{
		res := api("PUT", "/_api/pages/foo", server.PutRequest{Page: nil})
		r.Equal(200, res.Code)
	}
	{
		res := api("PUT", "/_api/pages/foo", server.PutRequest{Page: &storage.Page{Meta: storage.PageMeta{Title: "Foo"}}})
		r.Equal(400, res.Code)
	}
	{
		res := api("PUT", "/_api/pages/foo", server.PutRequest{Page: nil})
		r.Equal(400, res.Code)
	}
	{
		res := api("DELETE", "/_api/pages/foo", nil)
		r.Equal(200, res.Code)
	}

	// Create folder where page exists already
	{
		res := api("PUT", "/_api/pages/foo", server.PutRequest{Page: &storage.Page{Meta: storage.PageMeta{Title: "Foo"}}})
		r.Equal(200, res.Code)
	}
	{
		res := api("PUT", "/_api/pages/foo", server.PutRequest{Page: nil})
		r.Equal(400, res.Code)
	}
	{
		res := api("DELETE", "/_api/pages/foo", nil)
		r.Equal(200, res.Code)
	}
}
