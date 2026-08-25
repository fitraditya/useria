package middleware

import (
	"net/http"
	"slices"

	"github.com/fitraditya/useria/internal/utils"
)

// RequireScope ensures the token's scopes include the given permission code.
// Must run after Auth and RequireCompany.
func RequireScope(scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				utils.JSONError(w, http.StatusUnauthorized, "unauthenticated")
				return
			}
			if !slices.Contains(claims.Scopes, scope) {
				utils.JSONError(w, http.StatusForbidden, "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
