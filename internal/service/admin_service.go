package service

import (
	"context"

	"github.com/fitraditya/useria/internal/repository"
)

type AdminStats struct {
	TotalCompanies int `json:"total_companies"`
	TotalUsers     int `json:"total_users"`
}

type AdminService struct {
	companies *repository.CompanyRepository
	users     *repository.UserRepository
	members   *repository.MemberRepository
}

func NewAdminService(companies *repository.CompanyRepository, users *repository.UserRepository, members *repository.MemberRepository) *AdminService {
	return &AdminService{companies: companies, users: users, members: members}
}

func (s *AdminService) Stats(ctx context.Context) (AdminStats, error) {
	companies, err := s.companies.Count(ctx)
	if err != nil {
		return AdminStats{}, err
	}
	users, err := s.users.Count(ctx)
	if err != nil {
		return AdminStats{}, err
	}
	return AdminStats{TotalCompanies: companies, TotalUsers: users}, nil
}

// ListUsers returns members across all companies matching the given
// filters. The Users page is filter-first — with nothing to filter by, this
// returns an empty result without querying, rather than dumping every
// membership on the platform.
func (s *AdminService) ListUsers(ctx context.Context, companyID, name, email string) ([]repository.MemberWithUserAndCompany, error) {
	if companyID == "" && name == "" && email == "" {
		return []repository.MemberWithUserAndCompany{}, nil
	}
	return s.members.ListAcrossCompanies(ctx, companyID, name, email)
}
