package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-chi/render"
	"github.com/tfabritius/plainpage/model"
	"github.com/tfabritius/plainpage/service"
	"github.com/tfabritius/plainpage/service/ctxutil"
)

// Pre-compiled regex for URL validation
var urlRegex = regexp.MustCompile("^[a-z0-9-][a-z0-9_-]*(/[a-z0-9-][a-z0-9_-]*)*$")

// isValidUrl checks if URL is a valid URL for a page or folder. Valid URL consist of parts separated by /.
// Each part may only contain letters, digits, dash and underscore, but must not start with underscore.
func isValidUrl(urlPath string) bool {
	return urlPath == "" || urlRegex.MatchString(urlPath)
}

func (App) getBreadcrumbs(urlPath string, page *model.Page, folder *model.Folder, ancestorsMeta []model.ContentMetaWithURL) []model.Breadcrumb {
	breadcrumbs := []model.Breadcrumb{}

	for i := len(ancestorsMeta) - 1; i >= 0; i-- {
		if ancestorsMeta[i].Url == "" {
			continue
		}

		breadcrumb := model.Breadcrumb{
			Url:   ancestorsMeta[i].Url,
			Title: ancestorsMeta[i].Title,
			Name:  path.Base(ancestorsMeta[i].Url),
		}

		breadcrumbs = append(breadcrumbs, breadcrumb)
	}

	if page != nil {
		breadcrumbs = append(breadcrumbs, model.Breadcrumb{
			Url:   urlPath,
			Title: page.Meta.Title,
			Name:  path.Base(urlPath),
		})
	} else if folder != nil && urlPath != "" {
		breadcrumbs = append(breadcrumbs, model.Breadcrumb{
			Url:   urlPath,
			Title: folder.Meta.Title,
			Name:  path.Base(urlPath),
		})
	}

	return breadcrumbs
}

func (app App) getContent(w http.ResponseWriter, r *http.Request) {
	urlPath := r.PathValue("*")

	userID := ctxutil.UserID(r.Context())
	page := ctxutil.Page(r.Context())
	folder := ctxutil.Folder(r.Context())
	metas := ctxutil.AncestorsMeta(r.Context())

	response := model.GetContentResponse{}

	// Get the content's own metadata for permission calculation
	var contentMeta model.ContentMeta
	if page != nil {
		contentMeta = page.Meta
	} else if folder != nil {
		contentMeta = folder.Meta
	}

	effectiveAcl := app.Content.GetEffectivePermissions(contentMeta, metas)

	response.AllowWrite = app.Users.CheckContentPermissions(effectiveAcl, userID, model.AccessOpWrite) == nil
	response.AllowDelete = app.Users.CheckContentPermissions(effectiveAcl, userID, model.AccessOpDelete) == nil

	response.Breadcrumbs = app.getBreadcrumbs(urlPath, page, folder, metas)

	if page != nil {
		if app.isAdmin(userID) {
			if err := app.Users.EnhanceACLWithUserInfo(page.Meta.ACL); err != nil {
				panic(err)
			}
		} else {
			page.Meta.ACL = nil // Hide ACL
		}

		// Populate user info from userId for API response
		app.populateModifiedByUserInfo(&page.Meta)

		response.Page = page
	} else if folder != nil {
		// Filter folder entries based on read access
		accessibleContent := []model.FolderEntry{}
		for _, entry := range folder.Content {
			// Determine the effective ACL for this entry
			var entryEffectiveAcl []model.AccessRule
			if entry.ACL != nil {
				// Entry has its own ACL
				entryEffectiveAcl = *entry.ACL
			} else {
				// Entry inherits from the folder's effective ACL
				entryEffectiveAcl = effectiveAcl
			}

			// Check read permission
			if err := app.Users.CheckContentPermissions(entryEffectiveAcl, userID, model.AccessOpRead); err == nil {
				accessibleContent = append(accessibleContent, entry)
			}
		}
		folder.Content = accessibleContent

		if app.isAdmin(userID) {
			if err := app.Users.EnhanceACLWithUserInfo(folder.Meta.ACL); err != nil {
				panic(err)
			}
		} else {
			folder.Meta.ACL = nil // Hide ACL
		}

		// Populate user info from userId for API response
		app.populateModifiedByUserInfo(&folder.Meta)

		response.Folder = folder
	} else {
		// Not found
		w.WriteHeader(http.StatusNotFound)

		response.AllowDelete = false

		if !isValidUrl(urlPath) {
			response.AllowWrite = false
		} else {
			parentUrl, err := url.JoinPath(urlPath, "..")
			if err != nil {
				panic(err)
			}

			if !app.Content.IsFolder(parentUrl) {
				response.AllowWrite = false
			}
		}

		if !response.AllowWrite {
			response.Breadcrumbs = nil
		}
	}

	render.JSON(w, r, response)
}

// PatchableContent is the wrapper for content PATCH operations
type PatchableContent struct {
	Page   *model.Page   `json:"page" patch:"allow"`
	Folder *model.Folder `json:"folder" patch:"allow"`
}

func (app App) patchContent(w http.ResponseWriter, r *http.Request) {
	urlPath := r.PathValue("*")
	userID := ctxutil.UserID(r.Context())

	page := ctxutil.Page(r.Context())
	folder := ctxutil.Folder(r.Context())

	if !isValidUrl(urlPath) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var operations []model.PatchOperation
	if err := json.NewDecoder(r.Body).Decode(&operations); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	isFolder := folder != nil

	if page == nil && folder == nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	// Check if any operation targets ACL - require admin permission
	aclPatched := false
	for _, op := range operations {
		if strings.Contains(op.Path, "/meta/acl") {
			aclPatched = true
			if userID == "" {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			if !app.isAdmin(userID) {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
			break
		}
	}

	// Initialize patchable struct with current values
	patchReq := PatchableContent{}

	if isFolder {
		patchReq.Folder = &model.Folder{
			Url: urlPath,
			Meta: model.ContentMeta{
				Title: folder.Meta.Title,
				ACL:   folder.Meta.ACL,
			},
		}
	} else {
		patchReq.Page = &model.Page{
			Url: urlPath,
			Meta: model.ContentMeta{
				Title: page.Meta.Title,
				ACL:   page.Meta.ACL,
			},
		}
	}

	// Apply patch operations to patchable struct
	if err := ApplyJSONPatch(&patchReq, operations); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	urlChanged := false
	metadataChanged := false

	var newUrl string

	if isFolder {
		if patchReq.Folder.Meta.Title != folder.Meta.Title {
			metadataChanged = true
			folder.Meta.Title = patchReq.Folder.Meta.Title
		}

		if aclPatched {
			// Validate ACL only if it's not nil (nil means "inherit")
			if patchReq.Folder.Meta.ACL != nil {
				if err := model.ValidateContentACL(*patchReq.Folder.Meta.ACL); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			}

			metadataChanged = true
			folder.Meta.ACL = patchReq.Folder.Meta.ACL
		}

		if patchReq.Folder.Url != folder.Url {
			urlChanged = true
			newUrl = patchReq.Folder.Url
		}
	} else {
		if patchReq.Page.Meta.Title != page.Meta.Title {
			metadataChanged = true
			page.Meta.Title = patchReq.Page.Meta.Title
		}

		if aclPatched {
			// Validate ACL only if it's not nil (nil means "inherit")
			if patchReq.Page.Meta.ACL != nil {
				if err := model.ValidateContentACL(*patchReq.Page.Meta.ACL); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			}

			metadataChanged = true
			page.Meta.ACL = patchReq.Page.Meta.ACL
		}

		if patchReq.Page.Url != page.Url {
			urlChanged = true
			newUrl = patchReq.Page.Url
		}
	}

	// Apply url change if requested
	if urlChanged {
		if err := app.moveContent(w, urlPath, newUrl, userID, isFolder); err != nil {
			return // Error already written to response
		}
		urlPath = newUrl
	}

	// Save metadata changes (ACL, etc.) only if changed
	if metadataChanged {
		var err error
		if isFolder {
			err = app.Content.SaveFolder(urlPath, folder.Meta)
		} else {
			// Metadata-only changes (ACL, title) should not create a new version
			err = app.Content.SavePageWithoutVersion(urlPath, page.Content, page.Meta, userID)
		}

		if err != nil {
			panic(err)
		}
	}

	w.WriteHeader(http.StatusOK)
}

// moveContent handles moving/renaming a page or folder.
// Returns the new urlPath and any error.
// If an error occurs, the HTTP response is already written and the returned error is non-nil.
func (app App) moveContent(w http.ResponseWriter, urlPath, destinationPath, userID string, isFolder bool) error {
	// Validate url format
	if !isValidUrl(destinationPath) {
		http.Error(w, "invalid url format", http.StatusBadRequest)
		return errors.New("invalid url format")
	}

	// Check if actually moving
	if destinationPath == urlPath {
		return nil
	}

	// Check delete permission on source parent folder (moving removes from source)
	sourceParent := path.Dir(urlPath)
	if sourceParent == "." {
		sourceParent = ""
	}

	sourceParentMeta := model.ContentMeta{}
	if app.Content.IsFolder(sourceParent) {
		srcFolder, err := app.Content.ReadFolder(sourceParent)
		if err != nil {
			panic(err)
		}
		sourceParentMeta = srcFolder.Meta
	}

	srcAncestorMetas, err := app.Content.ReadAncestorsMeta(sourceParent)
	if err != nil {
		panic(err)
	}
	srcAcl := app.Content.GetEffectivePermissions(sourceParentMeta, srcAncestorMetas)

	if err := app.Users.CheckContentPermissions(srcAcl, userID, model.AccessOpDelete); err != nil {
		var e *service.AccessDeniedError
		if errors.As(err, &e) {
			http.Error(w, "no delete permission on source folder", e.StatusCode)
			return err
		}
		panic(err)
	}

	// Check write permission on destination parent folder
	destinationParent := path.Dir(destinationPath)
	if destinationParent == "." {
		destinationParent = ""
	}

	if !app.Content.IsFolder(destinationParent) {
		http.Error(w, "destination parent folder does not exist", http.StatusBadRequest)
		return errors.New("destination parent folder does not exist")
	}
	destFolder, err := app.Content.ReadFolder(destinationParent)
	if err != nil {
		panic(err)
	}

	destAncestorMetas, err := app.Content.ReadAncestorsMeta(destinationParent)
	if err != nil {
		panic(err)
	}
	destAcl := app.Content.GetEffectivePermissions(destFolder.Meta, destAncestorMetas)

	if err := app.Users.CheckContentPermissions(destAcl, userID, model.AccessOpWrite); err != nil {
		var e *service.AccessDeniedError
		if errors.As(err, &e) {
			http.Error(w, "no write permission on destination folder", e.StatusCode)
			return err
		}
		panic(err)
	}

	// Perform the move
	var moveErr error
	if isFolder {
		moveErr = app.Content.MoveFolder(urlPath, destinationPath)
	} else {
		moveErr = app.Content.MovePage(urlPath, destinationPath)
	}

	if moveErr != nil {
		if errors.Is(moveErr, model.ErrNotFound) {
			http.Error(w, moveErr.Error(), http.StatusNotFound)
			return moveErr
		} else if errors.Is(moveErr, model.ErrParentFolderNotFound) {
			http.Error(w, moveErr.Error(), http.StatusBadRequest)
			return moveErr
		} else if errors.Is(moveErr, model.ErrDestinationExists) {
			http.Error(w, moveErr.Error(), http.StatusBadRequest)
			return moveErr
		} else if errors.Is(moveErr, model.ErrCannotMoveRoot) {
			http.Error(w, moveErr.Error(), http.StatusBadRequest)
			return moveErr
		}
		panic(moveErr)
	}

	return nil
}

func (app App) putContent(w http.ResponseWriter, r *http.Request) {
	urlPath := r.PathValue("*")

	userID := ctxutil.UserID(r.Context())
	page := ctxutil.Page(r.Context())
	folder := ctxutil.Folder(r.Context())

	if !isValidUrl(urlPath) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var body model.PutRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var err error
	if body.Page != nil {
		if page != nil {
			// if page exists already, take over ACL
			body.Page.Meta.ACL = page.Meta.ACL
		}

		err = app.Content.SavePage(urlPath, body.Page.Content, body.Page.Meta, userID)
	} else if body.Folder != nil {
		if folder != nil {
			// if folder exists already, take over ACL
			body.Folder.Meta.ACL = folder.Meta.ACL

			// and update
			err = app.Content.SaveFolder(urlPath, body.Folder.Meta)
		} else {
			// make sure ACLs are not set
			body.Folder.Meta.ACL = nil

			// and create
			err = app.Content.CreateFolder(urlPath, body.Folder.Meta)
		}
	} else {
		http.Error(w, "Content missing", http.StatusBadRequest)
		return
	}

	if err != nil {
		if errors.Is(err, model.ErrParentFolderNotFound) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if errors.Is(err, model.ErrPageOrFolderExistsAlready) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
}

func (app App) deleteContent(w http.ResponseWriter, r *http.Request) {
	urlPath := r.PathValue("*")

	page := ctxutil.Page(r.Context())
	folder := ctxutil.Folder(r.Context())

	if !isValidUrl(urlPath) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var err error
	if page != nil {
		err = app.Content.DeletePage(urlPath)

	} else if folder != nil {
		err = app.Content.DeleteFolder(urlPath)

	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if err != nil {
		if errors.Is(err, model.ErrCannotDeleteRoot) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
}

func (app App) getAttic(w http.ResponseWriter, r *http.Request) {
	urlPath := r.PathValue("*")
	queryRev := r.URL.Query().Get("rev")

	page := ctxutil.Page(r.Context())
	ancestorsMeta := ctxutil.AncestorsMeta(r.Context())

	if !isValidUrl(urlPath) || page == nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	breadcrumbs := app.getBreadcrumbs(urlPath, page, nil, ancestorsMeta)

	var revision int64
	if queryRev == "" {
		list, err := app.Content.ListAttic(urlPath)
		if err != nil {
			panic(err)
		}

		response := model.GetAtticListResponse{
			Entries:     list,
			Breadcrumbs: breadcrumbs,
		}
		render.JSON(w, r, response)
	} else {
		var err error
		revision, err = strconv.ParseInt(queryRev, 10, 64)
		if err != nil {
			http.Error(w, "Invalid query parameter: rev", http.StatusBadRequest)
			return
		}

		if !app.Content.IsAtticPage(urlPath, revision) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		page, err := app.Content.ReadPage(urlPath, &revision)
		if err != nil {
			panic(err)
		}

		// Populate user info from userId for API response
		app.populateModifiedByUserInfo(&page.Meta)

		response := model.GetContentResponse{Page: &page, Breadcrumbs: breadcrumbs}
		render.JSON(w, r, response)
	}
}

func (app App) searchContent(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	userID := ctxutil.UserID(r.Context())

	// Parse pagination parameters
	page := 1
	limit := 20
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Calculate how many results we need to skip and collect
	skip := (page - 1) * limit
	need := limit

	// Iteratively fetch from Bleve until we have enough accessible results
	const batchSize = 100

	accessibleResults := []model.SearchHit{}
	skipped := 0
	bleveOffset := 0
	bleveExhausted := false
	stoppedEarly := false // Set to true if we break out of the inner loop with results remaining

	for len(accessibleResults) < need && !bleveExhausted {
		results, totalHits, err := app.Content.SearchWithPagination(q, bleveOffset, batchSize)
		if err != nil {
			panic(err)
		}

		// Check if Bleve has more results beyond this batch
		noMoreBatches := len(results) == 0 || bleveOffset+len(results) >= int(totalHits)

		// Filter each result by access control
		processedAll := true
		for i, result := range results {
			if err := app.Users.CheckContentPermissions(result.EffectiveACL, userID, model.AccessOpRead); err != nil {
				var e *service.AccessDeniedError
				if errors.As(err, &e) {
					// Skip this result - user doesn't have access
					continue
				}
				panic(err)
			}

			// User has access to this result
			if skipped < skip {
				// This result belongs to a previous page
				skipped++
			} else {
				// This result belongs to the current page
				result.Meta.ACL = nil // Hide ACL
				app.populateModifiedByUserInfo(&result.Meta)
				accessibleResults = append(accessibleResults, result)

				if len(accessibleResults) >= need {
					// We have enough for this page
					// Check if there are more results we haven't processed
					if i < len(results)-1 || !noMoreBatches {
						stoppedEarly = true
					}
					processedAll = false
					break
				}
			}
		}

		if processedAll && noMoreBatches {
			bleveExhausted = true
		}

		bleveOffset += batchSize
	}

	// hasMore is true if we stopped early (didn't check all results)
	hasMore := stoppedEarly

	response := model.SearchResponse{
		Items:   accessibleResults,
		Page:    page,
		Limit:   limit,
		HasMore: hasMore,
	}

	render.JSON(w, r, response)
}
