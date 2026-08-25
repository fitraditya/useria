package service

import (
	"context"
	"errors"
	"time"

	"github.com/fitraditya/useria/internal/models"
	"github.com/fitraditya/useria/internal/repository"
	"github.com/fitraditya/useria/internal/utils"
)

var (
	ErrAlreadyMember   = errors.New("user is already a member of this company")
	ErrUserNotFound    = errors.New("no user with this email, ask them to register first")
	ErrRoleNotAllowed  = errors.New("role does not belong to this company")
	ErrMemberNotFound  = errors.New("member not found in this company")
	ErrInviteNotFound  = errors.New("no pending invite for this company")
	ErrProtectedMember = errors.New("this member's role is platform-managed and cannot be changed here")
)

type MemberService struct {
	members *repository.MemberRepository
	users   *repository.UserRepository
	roles   *repository.RoleRepository
	audit   AuditService
}

func NewMemberService(members *repository.MemberRepository, users *repository.UserRepository, roles *repository.RoleRepository, audit AuditService) *MemberService {
	return &MemberService{members: members, users: users, roles: roles, audit: audit}
}

func (s *MemberService) List(ctx context.Context, companyID string) ([]repository.MemberWithUser, error) {
	return s.members.ListByCompanyID(ctx, companyID)
}

func (s *MemberService) Get(ctx context.Context, companyID, memberID string) (*models.CompanyMember, error) {
	m, err := s.members.GetByID(ctx, memberID)
	if errors.Is(err, repository.ErrNotFound) || (err == nil && m.CompanyID != companyID) {
		return nil, ErrMemberNotFound
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (s *MemberService) Invite(ctx context.Context, companyID, actorUserID, email, roleID string) (*models.CompanyMember, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	if _, err := s.members.GetByCompanyAndUser(ctx, companyID, user.ID); err == nil {
		return nil, ErrAlreadyMember
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	if err := s.validateRole(ctx, companyID, roleID); err != nil {
		return nil, err
	}

	now := time.Now()
	m := &models.CompanyMember{
		ID:        utils.NewUUID(),
		CompanyID: companyID,
		UserID:    user.ID,
		RoleID:    roleID,
		Status:    models.MemberStatusInvited,
		InvitedAt: &now,
	}
	if err := s.members.Create(ctx, m); err != nil {
		return nil, err
	}
	s.audit.Log(ctx, actorUserID, "member.invite", "company_member", m.ID, companyID, "invited="+email)
	return s.members.GetByID(ctx, m.ID)
}

func (s *MemberService) UpdateRole(ctx context.Context, companyID, actorUserID, memberID, roleID string) error {
	m, err := s.Get(ctx, companyID, memberID)
	if err != nil {
		return err
	}
	if err := s.requireNotProtected(ctx, m.RoleID); err != nil {
		return err
	}
	if err := s.validateRole(ctx, companyID, roleID); err != nil {
		return err
	}
	if err := s.members.UpdateRole(ctx, m.ID, roleID); err != nil {
		return err
	}
	s.audit.Log(ctx, actorUserID, "member.role_update", "company_member", m.ID, companyID, "")
	return nil
}

func (s *MemberService) Remove(ctx context.Context, companyID, actorUserID, memberID string) error {
	m, err := s.Get(ctx, companyID, memberID)
	if err != nil {
		return err
	}
	if err := s.requireNotProtected(ctx, m.RoleID); err != nil {
		return err
	}
	if err := s.members.Delete(ctx, m.ID); err != nil {
		return err
	}
	s.audit.Log(ctx, actorUserID, "member.remove", "company_member", m.ID, companyID, "")
	return nil
}

// requireNotProtected blocks role changes and removal for members currently
// holding the platform-level SuperAdmin role, which only the seed-admin
// bootstrap may grant or revoke.
func (s *MemberService) requireNotProtected(ctx context.Context, currentRoleID string) error {
	role, err := s.roles.GetByID(ctx, currentRoleID)
	if err != nil {
		return err
	}
	if role.IsSystem && role.Name == "SuperAdmin" {
		return ErrProtectedMember
	}
	return nil
}

func (s *MemberService) ListPendingInvites(ctx context.Context, userID string) ([]repository.MembershipWithCompany, error) {
	return s.members.ListPendingByUserID(ctx, userID)
}

func (s *MemberService) AcceptInvite(ctx context.Context, userID, companyID string) error {
	m, err := s.members.GetByCompanyAndUser(ctx, companyID, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrInviteNotFound
	}
	if err != nil {
		return err
	}
	if m.Status != models.MemberStatusInvited {
		return ErrInviteNotFound
	}
	return s.members.Accept(ctx, m.ID)
}

func (s *MemberService) validateRole(ctx context.Context, companyID, roleID string) error {
	role, err := s.roles.GetByID(ctx, roleID)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrRoleNotAllowed
	}
	if err != nil {
		return err
	}
	if role.CompanyID != nil && *role.CompanyID != companyID {
		return ErrRoleNotAllowed
	}
	if role.IsSystem && role.Name == "SuperAdmin" {
		// SuperAdmin is platform-level and only granted via the seed-admin
		// bootstrap, never through company-scoped member management.
		return ErrRoleNotAllowed
	}
	return nil
}
