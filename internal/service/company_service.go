package service

import (
	"context"
	"errors"

	"github.com/fitraditya/useria/internal/models"
	"github.com/fitraditya/useria/internal/repository"
	"github.com/fitraditya/useria/internal/utils"
)

var ErrSlugTaken = errors.New("company slug already taken")

type CompanyService struct {
	companies *repository.CompanyRepository
	audit     AuditService
}

func NewCompanyService(companies *repository.CompanyRepository, audit AuditService) *CompanyService {
	return &CompanyService{companies: companies, audit: audit}
}

func (s *CompanyService) Create(ctx context.Context, name, description, website, plan, createdBy string) (*models.Company, error) {
	slug := utils.Slugify(name)
	if _, err := s.companies.GetBySlug(ctx, slug); err == nil {
		return nil, ErrSlugTaken
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	if plan == "" {
		plan = "free"
	}

	c := &models.Company{
		ID:        utils.NewUUID(),
		Name:      name,
		Slug:      slug,
		Plan:      plan,
		Status:    models.CompanyStatusActive,
		CreatedBy: createdBy,
	}
	if description != "" {
		c.Description = &description
	}
	if website != "" {
		c.Website = &website
	}

	if err := s.companies.Create(ctx, c); err != nil {
		return nil, err
	}
	s.audit.Log(ctx, createdBy, "company.create", "company", c.ID, c.ID, "name="+name)
	return s.companies.GetByID(ctx, c.ID)
}

func (s *CompanyService) List(ctx context.Context, name, adminEmail, status string) ([]repository.CompanyWithAdmin, error) {
	return s.companies.List(ctx, name, adminEmail, status)
}

func (s *CompanyService) Get(ctx context.Context, id string) (*models.Company, error) {
	return s.companies.GetByID(ctx, id)
}

func (s *CompanyService) Update(ctx context.Context, id, actorUserID, name, description, website string, status models.CompanyStatus, plan string) (*models.Company, error) {
	c, err := s.companies.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	previousStatus := c.Status

	if name != "" {
		c.Name = name
	}
	if description != "" {
		c.Description = &description
	}
	if website != "" {
		c.Website = &website
	}
	if status != "" {
		c.Status = status
	}
	if plan != "" {
		c.Plan = plan
	}

	if err := s.companies.Update(ctx, c); err != nil {
		return nil, err
	}

	action := "company.update"
	if status != "" && status != previousStatus {
		switch status {
		case models.CompanyStatusSuspended:
			action = "company.suspend"
		case models.CompanyStatusActive:
			action = "company.activate"
		}
	}
	s.audit.Log(ctx, actorUserID, action, "company", id, id, "")

	return s.companies.GetByID(ctx, id)
}

func (s *CompanyService) Delete(ctx context.Context, id, actorUserID string) error {
	if err := s.companies.Delete(ctx, id); err != nil {
		return err
	}
	s.audit.Log(ctx, actorUserID, "company.delete", "company", id, id, "")
	return nil
}
