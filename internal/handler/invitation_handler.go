package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/fitraditya/useria/internal/service"
	"github.com/fitraditya/useria/internal/utils"
)

type InvitationHandler struct {
	members *service.MemberService
}

func NewInvitationHandler(members *service.MemberService) *InvitationHandler {
	return &InvitationHandler{members: members}
}

type acceptInviteRequest struct {
	CompanyID string `json:"company_id"`
}

func (h *InvitationHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := requireClaims(w, r)
	if !ok {
		return
	}

	invites, err := h.members.ListPendingInvites(r.Context(), claims.UserID)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "failed to list invitations")
		return
	}
	utils.JSON(w, http.StatusOK, invites)
}

func (h *InvitationHandler) Accept(w http.ResponseWriter, r *http.Request) {
	claims, ok := requireClaims(w, r)
	if !ok {
		return
	}

	var req acceptInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.CompanyID == "" {
		utils.JSONError(w, http.StatusBadRequest, "company_id is required")
		return
	}

	err := h.members.AcceptInvite(r.Context(), claims.UserID, req.CompanyID)
	if errors.Is(err, service.ErrInviteNotFound) {
		utils.JSONError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "failed to accept invite")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]bool{"accepted": true})
}
