package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	appmw "github.com/fitraditya/useria/internal/middleware"
	"github.com/fitraditya/useria/internal/repository"
	"github.com/fitraditya/useria/internal/utils"
)

type updateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type ProfileHandler struct {
	users *repository.UserRepository
}

func NewProfileHandler(users *repository.UserRepository) *ProfileHandler {
	return &ProfileHandler{users: users}
}

func (h *ProfileHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims, ok := appmw.ClaimsFromContext(r.Context())
	if !ok {
		utils.JSONError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	user, err := h.users.GetByID(r.Context(), claims.UserID)
	if errors.Is(err, repository.ErrNotFound) {
		utils.JSONError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "failed to load profile")
		return
	}

	utils.JSON(w, http.StatusOK, user)
}

func (h *ProfileHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := appmw.ClaimsFromContext(r.Context())
	if !ok {
		utils.JSONError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.users.UpdateProfile(r.Context(), claims.UserID, req.FirstName, req.LastName); err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "failed to update profile")
		return
	}

	user, err := h.users.GetByID(r.Context(), claims.UserID)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "failed to load profile")
		return
	}

	utils.JSON(w, http.StatusOK, user)
}
