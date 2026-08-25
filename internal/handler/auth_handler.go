package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	appmw "github.com/fitraditya/useria/internal/middleware"
	"github.com/fitraditya/useria/internal/models"
	"github.com/fitraditya/useria/internal/service"
	"github.com/fitraditya/useria/internal/utils"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type registerRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	CompanyName string `json:"company_name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if !strings.Contains(req.Email, "@") {
		utils.JSONError(w, http.StatusBadRequest, "invalid email")
		return
	}
	if len(req.Password) < 8 {
		utils.JSONError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}
	if strings.TrimSpace(req.FirstName) == "" || strings.TrimSpace(req.LastName) == "" {
		utils.JSONError(w, http.StatusBadRequest, "first name and last name are required")
		return
	}
	if strings.TrimSpace(req.CompanyName) == "" {
		utils.JSONError(w, http.StatusBadRequest, "company name is required")
		return
	}

	user, token, err := h.auth.Register(r.Context(), req.Email, req.Password, req.FirstName, req.LastName, req.CompanyName)
	if errors.Is(err, service.ErrEmailTaken) {
		utils.JSONError(w, http.StatusConflict, "email already registered")
		return
	}
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "registration failed")
		return
	}

	utils.JSON(w, http.StatusCreated, authResponse{Token: token, User: user})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, token, err := h.auth.Login(r.Context(), req.Email, req.Password)
	if errors.Is(err, service.ErrInvalidCredentials) {
		utils.JSONError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "login failed")
		return
	}

	utils.JSON(w, http.StatusOK, authResponse{Token: token, User: user})
}

type selectCompanyRequest struct {
	CompanyID string `json:"company_id"`
}

func (h *AuthHandler) ListCompanies(w http.ResponseWriter, r *http.Request) {
	claims, ok := appmw.ClaimsFromContext(r.Context())
	if !ok {
		utils.JSONError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	memberships, err := h.auth.ListMemberships(r.Context(), claims.UserID)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "failed to list companies")
		return
	}

	utils.JSON(w, http.StatusOK, memberships)
}

func (h *AuthHandler) SelectCompany(w http.ResponseWriter, r *http.Request) {
	claims, ok := appmw.ClaimsFromContext(r.Context())
	if !ok {
		utils.JSONError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	var req selectCompanyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.CompanyID == "" {
		utils.JSONError(w, http.StatusBadRequest, "company_id is required")
		return
	}

	token, err := h.auth.SelectCompany(r.Context(), claims.UserID, req.CompanyID)
	if errors.Is(err, service.ErrNoMembership) {
		utils.JSONError(w, http.StatusForbidden, "not a member of this company")
		return
	}
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "failed to select company")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"token": token})
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Email) == "" {
		utils.JSONError(w, http.StatusBadRequest, "email is required")
		return
	}

	if err := h.auth.ForgotPassword(r.Context(), req.Email); err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "failed to process request")
		return
	}

	// Always the same response, whether or not the email exists.
	utils.JSON(w, http.StatusOK, map[string]string{"message": "if that email is registered, a reset link has been issued"})
}

type resetPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Token == "" {
		utils.JSONError(w, http.StatusBadRequest, "token is required")
		return
	}
	if len(req.Password) < 8 {
		utils.JSONError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	err := h.auth.ResetPassword(r.Context(), req.Token, req.Password)
	if errors.Is(err, service.ErrInvalidResetToken) {
		utils.JSONError(w, http.StatusBadRequest, "invalid or expired reset link")
		return
	}
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "failed to reset password")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]bool{"reset": true})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	claims, ok := appmw.ClaimsFromContext(r.Context())
	if !ok {
		utils.JSONError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	token, err := h.auth.Refresh(claims)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "failed to refresh token")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"token": token})
}
