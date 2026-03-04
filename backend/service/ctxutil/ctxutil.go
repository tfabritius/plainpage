package ctxutil

import (
	"context"

	"github.com/tfabritius/plainpage/model"
)

type contextKey int

const (
	_ contextKey = iota
	ctxKeyUserID
	ctxKeyPage
	ctxKeyFolder
	ctxKeyAncestors
	ctxKeyEffectiveACL
)

// WithUserID creates a new context that has username injected
func WithUserID(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, ctxKeyUserID, username)
}

// UserID tries to retrieve username from the given context
func UserID(ctx context.Context) string {
	if username, ok := ctx.Value(ctxKeyUserID).(string); ok {
		return username
	}
	return ""
}

// WithContent creates a new context that has content injected
func WithContent(ctx context.Context, page *model.Page, folder *model.Folder, ancestors []model.UrlAndMeta, effectiveAcl []model.AccessRule) context.Context {
	ctx = context.WithValue(ctx, ctxKeyPage, page)
	ctx = context.WithValue(ctx, ctxKeyFolder, folder)
	ctx = context.WithValue(ctx, ctxKeyAncestors, ancestors)
	ctx = context.WithValue(ctx, ctxKeyEffectiveACL, effectiveAcl)
	return ctx
}

// Page tries to retrieve page from the given context
func Page(ctx context.Context) *model.Page {
	if page, ok := ctx.Value(ctxKeyPage).(*model.Page); ok {
		return page
	}
	return nil
}

// Folder tries to retrieve folder from the given context
func Folder(ctx context.Context) *model.Folder {
	if folder, ok := ctx.Value(ctxKeyFolder).(*model.Folder); ok {
		return folder
	}
	return nil
}

// Ancestors tries to retrieve ancestors from the given context
func Ancestors(ctx context.Context) []model.UrlAndMeta {
	if metas, ok := ctx.Value(ctxKeyAncestors).([]model.UrlAndMeta); ok {
		return metas
	}
	return nil
}

// EffectiveACL tries to retrieve the effective ACL from the given context
func EffectiveACL(ctx context.Context) []model.AccessRule {
	if acl, ok := ctx.Value(ctxKeyEffectiveACL).([]model.AccessRule); ok {
		return acl
	}
	return nil
}
