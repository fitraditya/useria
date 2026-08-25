package service

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/fitraditya/useria/internal/models"
	"github.com/fitraditya/useria/internal/repository"
	"github.com/fitraditya/useria/internal/utils"
)

var (
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrNoMembership       = errors.New("not a member of this company")
	ErrInvalidResetToken  = errors.New("invalid or expired reset link")
)

const passwordResetTTL = time.Hour

type AuthService struct {
	users     *repository.UserRepository
	members   *repository.MemberRepository
	roles     *repository.RoleRepository
	companies *repository.CompanyRepository
	resets    *repository.PasswordResetRepository

	jwtSecret     string
	jwtExpiration int
}

func NewAuthService(
	users *repository.UserRepository,
	members *repository.MemberRepository,
	roles *repository.RoleRepository,
	companies *repository.CompanyRepository,
	resets *repository.PasswordResetRepository,
	jwtSecret string,
	jwtExpiration int,
) *AuthService {
	return &AuthService{
		users:         users,
		members:       members,
		roles:         roles,
		companies:     companies,
		resets:        resets,
		jwtSecret:     jwtSecret,
		jwtExpiration: jwtExpiration,
	}
}

// Register creates a new user and a brand-new company for them, granting
// them the Admin role over that company. Self-signup always creates a
// company — there is no path to an unaffiliated user account.
func (s *AuthService) Register(ctx context.Context, email, password, firstName, lastName, companyName string) (*models.User, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	_, err := s.users.GetByEmail(ctx, email)
	if err == nil {
		return nil, "", ErrEmailTaken
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, "", err
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		return nil, "", err
	}

	local := "local"
	user := &models.User{
		ID:            utils.NewUUID(),
		Email:         email,
		PasswordHash:  &hash,
		FirstName:     &firstName,
		LastName:      &lastName,
		OAuthProvider: &local,
		Status:        models.UserStatusActive,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, "", err
	}

	if err := s.createOwnedCompany(ctx, user.ID, companyName); err != nil {
		return nil, "", err
	}

	user, err = s.users.GetByID(ctx, user.ID)
	if err != nil {
		return nil, "", err
	}

	token, err := s.issueToken(user)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil
}

// createOwnedCompany creates a new company for userID and makes them its
// Admin. Slug collisions (e.g. two "Acme" signups) are resolved by
// appending a short random suffix.
func (s *AuthService) createOwnedCompany(ctx context.Context, userID, companyName string) error {
	slug := utils.Slugify(companyName)
	if _, err := s.companies.GetBySlug(ctx, slug); err == nil {
		suffix, err := utils.GenerateToken()
		if err != nil {
			return err
		}
		slug = slug + "-" + suffix[:6]
	} else if !errors.Is(err, repository.ErrNotFound) {
		return err
	}

	company := &models.Company{
		ID:        utils.NewUUID(),
		Name:      companyName,
		Slug:      slug,
		Plan:      "free",
		Status:    models.CompanyStatusActive,
		CreatedBy: userID,
	}
	if err := s.companies.Create(ctx, company); err != nil {
		return err
	}

	role, err := s.roles.GetSystemRoleByName(ctx, "Admin")
	if err != nil {
		return err
	}

	now := time.Now()
	member := &models.CompanyMember{
		ID:        utils.NewUUID(),
		CompanyID: company.ID,
		UserID:    userID,
		RoleID:    role.ID,
		Status:    models.MemberStatusActive,
		JoinedAt:  &now,
	}
	return s.members.Create(ctx, member)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*models.User, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.users.GetByEmail(ctx, email)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, "", ErrInvalidCredentials
	}
	if err != nil {
		return nil, "", err
	}
	if user.PasswordHash == nil || !utils.CheckPassword(password, *user.PasswordHash) {
		return nil, "", ErrInvalidCredentials
	}

	// SuperAdmin's authority isn't scoped to any one company — skip the
	// company-selection step and hand back a fully-scoped token right away.
	memberships, err := s.members.ListByUserID(ctx, user.ID)
	if err != nil {
		return nil, "", err
	}
	for _, m := range memberships {
		if m.RoleName == "SuperAdmin" {
			token, err := s.buildScopedToken(ctx, user.ID, m.CompanyID, m.RoleID)
			if err != nil {
				return nil, "", err
			}
			return user, token, nil
		}
	}

	token, err := s.issueToken(user)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil
}

// issueToken produces a pre-company-selection token carrying only the user
// identity. Company context, roles, and scopes are added by select-company.
func (s *AuthService) issueToken(user *models.User) (string, error) {
	return utils.GenerateJWT(s.jwtSecret, s.jwtExpiration, utils.Claims{UserID: user.ID})
}

func (s *AuthService) ListMemberships(ctx context.Context, userID string) ([]repository.MembershipWithCompany, error) {
	return s.members.ListByUserID(ctx, userID)
}

// SelectCompany verifies the user belongs to the company and issues a full
// token carrying company context, role, and permission scopes.
func (s *AuthService) SelectCompany(ctx context.Context, userID, companyID string) (string, error) {
	membership, err := s.members.GetByCompanyAndUser(ctx, companyID, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return "", ErrNoMembership
	}
	if err != nil {
		return "", err
	}
	if membership.Status != models.MemberStatusActive {
		return "", ErrNoMembership
	}

	// GetByID excludes soft-deleted companies, so a deleted company also
	// falls through here as ErrNotFound. Suspended companies exist but must
	// stop granting access to their own members, not just be hidden from
	// the admin list.
	company, err := s.companies.GetByID(ctx, companyID)
	if errors.Is(err, repository.ErrNotFound) {
		return "", ErrNoMembership
	}
	if err != nil {
		return "", err
	}
	if company.Status != models.CompanyStatusActive {
		return "", ErrNoMembership
	}

	return s.buildScopedToken(ctx, userID, companyID, membership.RoleID)
}

// buildScopedToken issues a full token carrying company context, role name,
// and permission scopes for that role.
func (s *AuthService) buildScopedToken(ctx context.Context, userID, companyID, roleID string) (string, error) {
	role, err := s.roles.GetByID(ctx, roleID)
	if err != nil {
		return "", err
	}
	scopes, err := s.roles.PermissionCodes(ctx, roleID)
	if err != nil {
		return "", err
	}

	return utils.GenerateJWT(s.jwtSecret, s.jwtExpiration, utils.Claims{
		UserID:    userID,
		CompanyID: companyID,
		Roles:     []string{role.Name},
		Scopes:    scopes,
	})
}

// Refresh reissues a token carrying the same claims with a fresh expiry.
func (s *AuthService) Refresh(claims *utils.Claims) (string, error) {
	return utils.GenerateJWT(s.jwtSecret, s.jwtExpiration, utils.Claims{
		UserID:    claims.UserID,
		CompanyID: claims.CompanyID,
		Roles:     claims.Roles,
		Scopes:    claims.Scopes,
	})
}

// ForgotPassword issues a reset token for the given email if an account
// exists. It never reports whether the email was found, to avoid leaking
// account existence. There is no email delivery wired up yet, so the reset
// link is logged server-side for an operator to relay manually.
func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.users.GetByEmail(ctx, email)
	if errors.Is(err, repository.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	rawToken, err := utils.GenerateToken()
	if err != nil {
		return err
	}

	reset := &models.PasswordResetToken{
		ID:        utils.NewUUID(),
		UserID:    user.ID,
		TokenHash: utils.HashToken(rawToken),
		ExpiresAt: time.Now().Add(passwordResetTTL),
	}
	if err := s.resets.Create(ctx, reset); err != nil {
		return err
	}

	log.Printf("[DEV] password reset requested for %s — no email service configured, share this link manually: /reset-password?token=%s (valid %s)", email, rawToken, passwordResetTTL)
	return nil
}

// ResetPassword consumes a reset token and sets a new password. Tokens are
// single-use and expire after passwordResetTTL.
func (s *AuthService) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	reset, err := s.resets.GetByTokenHash(ctx, utils.HashToken(rawToken))
	if errors.Is(err, repository.ErrNotFound) {
		return ErrInvalidResetToken
	}
	if err != nil {
		return err
	}
	if reset.UsedAt != nil || time.Now().After(reset.ExpiresAt) {
		return ErrInvalidResetToken
	}

	hash, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}
	if err := s.users.UpdatePassword(ctx, reset.UserID, hash); err != nil {
		return err
	}

	return s.resets.MarkUsed(ctx, reset.ID)
}
