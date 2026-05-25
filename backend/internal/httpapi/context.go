package httpapi

import "context"

type ctxKey string

const (
	ctxKeyUserID  ctxKey = "userID"
	ctxKeyIsAdmin ctxKey = "isAdmin"
)

func withAuth(ctx context.Context, userID int64, isAdmin bool) context.Context {
	ctx = context.WithValue(ctx, ctxKeyUserID, userID)
	ctx = context.WithValue(ctx, ctxKeyIsAdmin, isAdmin)
	return ctx
}

func currentUserID(ctx context.Context) (int64, bool) {
	v := ctx.Value(ctxKeyUserID)
	id, ok := v.(int64)
	return id, ok
}

func currentIsAdmin(ctx context.Context) bool {
	v := ctx.Value(ctxKeyIsAdmin)
	isAdmin, ok := v.(bool)
	return ok && isAdmin
}
