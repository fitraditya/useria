package middleware

import (
	"net/http"
	"strings"

	"github.com/fitraditya/useria/internal/utils"
)

// Auth validates the Bearer JWT on the request and injects its claims into
// the request context. It does not require a company to be selected.
func Auth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				utils.JSONError(w, http.StatusUnauthorized, "missing bearer token")
				return
			}

			claims, err := utils.ParseJWT(secret, parts[1])
			if err != nil {
				utils.JSONError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			next.ServeHTTP(w, r.WithContext(WithClaims(r.Context(), claims)))
		})
	}
}
