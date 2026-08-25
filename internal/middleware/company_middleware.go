package middleware

import (
	"net/http"

	"github.com/fitraditya/useria/internal/utils"
)

// RequireCompany ensures the token carries a selected company context.
// Must run after Auth.
func RequireCompany(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := ClaimsFromContext(r.Context())
		if !ok || claims.CompanyID == "" {
			utils.JSONError(w, http.StatusForbidden, "select a company first")
			return
		}
		next.ServeHTTP(w, r)
	})
}
