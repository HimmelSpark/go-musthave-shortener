package auth

import "context"

type ctxKey struct{}

type Info struct {
	UserID        string
	CookiePresent bool
	CookieValid   bool
}

func WithInfo(ctx context.Context, info Info) context.Context {
	return context.WithValue(ctx, ctxKey{}, info)
}

func FromContext(ctx context.Context) Info {
	info, _ := ctx.Value(ctxKey{}).(Info)
	return info
}

func UserIDFromContext(ctx context.Context) string {
	return FromContext(ctx).UserID
}
