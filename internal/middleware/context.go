package middleware

import (
	"context"

	"github.com/fitraditya/useria/internal/utils"
)

type contextKey string

const claimsContextKey contextKey = "claims"

func WithClaims(ctx context.Context, claims *utils.Claims) context.Context {
	return context.WithValue(ctx, claimsContextKey, claims)
}

func ClaimsFromContext(ctx context.Context) (*utils.Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(*utils.Claims)
	return claims, ok
}
