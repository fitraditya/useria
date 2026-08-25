package handler

import (
	"net/http"

	"github.com/fitraditya/useria/internal/repository"
	"github.com/fitraditya/useria/internal/utils"
)

type RoleHandler struct {
	roles *repository.RoleRepository
}

func NewRoleHandler(roles *repository.RoleRepository) *RoleHandler {
	return &RoleHandler{roles: roles}
}

func (h *RoleHandler) ListAssignable(w http.ResponseWriter, r *http.Request) {
	claims, ok := requireClaims(w, r)
	if !ok {
		return
	}

	roles, err := h.roles.ListAssignable(r.Context(), claims.CompanyID)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "failed to list roles")
		return
	}
	utils.JSON(w, http.StatusOK, roles)
}
