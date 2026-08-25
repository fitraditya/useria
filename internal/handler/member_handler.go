package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/fitraditya/useria/internal/service"
	"github.com/fitraditya/useria/internal/utils"
)

type MemberHandler struct {
	members *service.MemberService
}

func NewMemberHandler(members *service.MemberService) *MemberHandler {
	return &MemberHandler{members: members}
}

type inviteMemberRequest struct {
	Email  string `json:"email"`
	RoleID string `json:"role_id"`
}

type updateMemberRequest struct {
	RoleID string `json:"role_id"`
}

func (h *MemberHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := requireClaims(w, r)
	if !ok {
		return
	}

	members, err := h.members.List(r.Context(), claims.CompanyID)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "failed to list members")
		return
	}
	utils.JSON(w, http.StatusOK, members)
}

func (h *MemberHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims, ok := requireClaims(w, r)
	if !ok {
		return
	}

	id := chi.URLParam(r, "id")
	m, err := h.members.Get(r.Context(), claims.CompanyID, id)
	if errors.Is(err, service.ErrMemberNotFound) {
		utils.JSONError(w, http.StatusNotFound, "member not found")
		return
	}
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "failed to load member")
		return
	}
	utils.JSON(w, http.StatusOK, m)
}

func (h *MemberHandler) Invite(w http.ResponseWriter, r *http.Request) {
	claims, ok := requireClaims(w, r)
	if !ok {
		return
	}

	var req inviteMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" || req.RoleID == "" {
		utils.JSONError(w, http.StatusBadRequest, "email and role_id are required")
		return
	}

	m, err := h.members.Invite(r.Context(), claims.CompanyID, claims.UserID, req.Email, req.RoleID)
	switch {
	case errors.Is(err, service.ErrUserNotFound):
		utils.JSONError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrAlreadyMember):
		utils.JSONError(w, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrRoleNotAllowed):
		utils.JSONError(w, http.StatusBadRequest, err.Error())
	case err != nil:
		utils.JSONError(w, http.StatusInternalServerError, "failed to invite member")
	default:
		utils.JSON(w, http.StatusCreated, m)
	}
}

func (h *MemberHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	claims, ok := requireClaims(w, r)
	if !ok {
		return
	}

	id := chi.URLParam(r, "id")
	var req updateMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RoleID == "" {
		utils.JSONError(w, http.StatusBadRequest, "role_id is required")
		return
	}

	err := h.members.UpdateRole(r.Context(), claims.CompanyID, claims.UserID, id, req.RoleID)
	switch {
	case errors.Is(err, service.ErrMemberNotFound):
		utils.JSONError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrProtectedMember):
		utils.JSONError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, service.ErrRoleNotAllowed):
		utils.JSONError(w, http.StatusBadRequest, err.Error())
	case err != nil:
		utils.JSONError(w, http.StatusInternalServerError, "failed to update member")
	default:
		utils.JSON(w, http.StatusOK, map[string]bool{"updated": true})
	}
}

func (h *MemberHandler) Remove(w http.ResponseWriter, r *http.Request) {
	claims, ok := requireClaims(w, r)
	if !ok {
		return
	}

	id := chi.URLParam(r, "id")
	err := h.members.Remove(r.Context(), claims.CompanyID, claims.UserID, id)
	if errors.Is(err, service.ErrMemberNotFound) {
		utils.JSONError(w, http.StatusNotFound, err.Error())
		return
	}
	if errors.Is(err, service.ErrProtectedMember) {
		utils.JSONError(w, http.StatusForbidden, err.Error())
		return
	}
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "failed to remove member")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
