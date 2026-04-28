package auth

import (
	"context"

	corejwt "goadmin/core/auth/jwt"
)

type contextKey string

const claimsContextKey contextKey = "goadmin.auth.claims"

func ContextWithClaims(ctx context.Context, claims *corejwt.Claims) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, claimsContextKey, claims)
}

func ClaimsFromContext(ctx context.Context) (*corejwt.Claims, bool) {
	if ctx == nil {
		return nil, false
	}
	claims, ok := ctx.Value(claimsContextKey).(*corejwt.Claims)
	if !ok || claims == nil {
		return nil, false
	}
	return claims, true
}
