package handler

import (
	"net/http"

	"github.com/fitraditya/useria/internal/service"
	"github.com/fitraditya/useria/internal/utils"
)

type AdminHandler struct {
	admin *service.AdminService
}

func NewAdminHandler(admin *service.AdminService) *AdminHandler {
	return &AdminHandler{admin: admin}
}

func (h *AdminHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.admin.Stats(r.Context())
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "failed to load stats")
		return
	}
	utils.JSON(w, http.StatusOK, stats)
}

func (h *AdminHandler) Users(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	users, err := h.admin.ListUsers(r.Context(), q.Get("company_id"), q.Get("name"), q.Get("email"))
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "failed to list users")
		return
	}
	utils.JSON(w, http.StatusOK, users)
}
