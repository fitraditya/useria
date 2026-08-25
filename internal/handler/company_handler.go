package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/fitraditya/useria/internal/models"
	"github.com/fitraditya/useria/internal/repository"
	"github.com/fitraditya/useria/internal/service"
	"github.com/fitraditya/useria/internal/utils"
)

type CompanyHandler struct {
	companies *service.CompanyService
}

func NewCompanyHandler(companies *service.CompanyService) *CompanyHandler {
	return &CompanyHandler{companies: companies}
}

type createCompanyRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Website     string `json:"website"`
	Plan        string `json:"plan"`
}

type updateCompanyRequest struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Website     string               `json:"website"`
	Status      models.CompanyStatus `json:"status"`
	Plan        string               `json:"plan"`
}

func (h *CompanyHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	companies, err := h.companies.List(r.Context(), q.Get("name"), q.Get("admin_email"), q.Get("status"))
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "failed to list companies")
		return
	}
	utils.JSON(w, http.StatusOK, companies)
}

func (h *CompanyHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := requireClaims(w, r)
	if !ok {
		return
	}

	var req createCompanyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		utils.JSONError(w, http.StatusBadRequest, "name is required")
		return
	}

	c, err := h.companies.Create(r.Context(), req.Name, req.Description, req.Website, req.Plan, claims.UserID)
	if errors.Is(err, service.ErrSlugTaken) {
		utils.JSONError(w, http.StatusConflict, "a company with a similar name already exists")
		return
	}
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "failed to create company")
		return
	}
	utils.JSON(w, http.StatusCreated, c)
}

func (h *CompanyHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	c, err := h.companies.Get(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		utils.JSONError(w, http.StatusNotFound, "company not found")
		return
	}
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "failed to load company")
		return
	}
	utils.JSON(w, http.StatusOK, c)
}

func (h *CompanyHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := requireClaims(w, r)
	if !ok {
		return
	}
	id := chi.URLParam(r, "id")

	var req updateCompanyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	c, err := h.companies.Update(r.Context(), id, claims.UserID, req.Name, req.Description, req.Website, req.Status, req.Plan)
	if errors.Is(err, repository.ErrNotFound) {
		utils.JSONError(w, http.StatusNotFound, "company not found")
		return
	}
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "failed to update company")
		return
	}
	utils.JSON(w, http.StatusOK, c)
}

func (h *CompanyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := requireClaims(w, r)
	if !ok {
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.companies.Delete(r.Context(), id, claims.UserID); err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "failed to delete company")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
