package httpapi

import "context"

type ctxKey string

const (
	ctxKeyUserID           ctxKey = "userID"
	ctxKeyIsAdmin          ctxKey = "isAdmin"
	ctxKeyTerminalID       ctxKey = "terminalID"
	ctxKeyTerminalSerial   ctxKey = "terminalSerial"
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

func withTerminalAuth(ctx context.Context, terminalID int64, serial string) context.Context {
	ctx = context.WithValue(ctx, ctxKeyTerminalID, terminalID)
	ctx = context.WithValue(ctx, ctxKeyTerminalSerial, serial)
	return ctx
}

func currentTerminalID(ctx context.Context) (int64, bool) {
	v := ctx.Value(ctxKeyTerminalID)
	id, ok := v.(int64)
	return id, ok
}

func currentTerminalSerial(ctx context.Context) (string, bool) {
	v := ctx.Value(ctxKeyTerminalSerial)
	serial, ok := v.(string)
	return serial, ok
}
