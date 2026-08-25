package handler

import (
	"net/http"

	appmw "github.com/fitraditya/useria/internal/middleware"
	"github.com/fitraditya/useria/internal/utils"
)

// requireClaims fetches JWT claims from the request context, writing a 401
// and returning ok=false if they are missing.
func requireClaims(w http.ResponseWriter, r *http.Request) (*utils.Claims, bool) {
	claims, ok := appmw.ClaimsFromContext(r.Context())
	if !ok {
		utils.JSONError(w, http.StatusUnauthorized, "unauthenticated")
		return nil, false
	}
	return claims, true
}
