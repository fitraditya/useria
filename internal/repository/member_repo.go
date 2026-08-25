package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/fitraditya/useria/internal/models"
)

type MembershipWithCompany struct {
	models.CompanyMember
	CompanyName string `json:"company_name"`
	CompanySlug string `json:"company_slug"`
	RoleName    string `json:"role_name"`
}

type MemberWithUser struct {
	models.CompanyMember
	Email     string  `json:"email"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	RoleName  string  `json:"role_name"`
}

// MemberWithUserAndCompany is a membership row joined with both the user
// and the company it belongs to — used by the SuperAdmin cross-company
// users list.
type MemberWithUserAndCompany struct {
	models.CompanyMember
	Email       string  `json:"email"`
	FirstName   *string `json:"first_name,omitempty"`
	LastName    *string `json:"last_name,omitempty"`
	RoleName    string  `json:"role_name"`
	CompanyName string  `json:"company_name"`
	CompanySlug string  `json:"company_slug"`
}

type MemberRepository struct {
	db *sql.DB
}

func NewMemberRepository(db *sql.DB) *MemberRepository {
	return &MemberRepository{db: db}
}

func (r *MemberRepository) ListByUserID(ctx context.Context, userID string) ([]MembershipWithCompany, error) {
	return r.listByUserIDStatus(ctx, userID, "active")
}

func (r *MemberRepository) ListPendingByUserID(ctx context.Context, userID string) ([]MembershipWithCompany, error) {
	return r.listByUserIDStatus(ctx, userID, "invited")
}

func (r *MemberRepository) listByUserIDStatus(ctx context.Context, userID, status string) ([]MembershipWithCompany, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT cm.id, cm.company_id, cm.user_id, cm.role_id, cm.status, c.name, c.slug, ro.name
		FROM company_members cm
		JOIN companies c ON c.id = cm.company_id
		JOIN roles ro ON ro.id = cm.role_id
		WHERE cm.user_id = ? AND cm.status = ? AND cm.deleted_at IS NULL
	`, userID, status)
	if err != nil {
		return nil, fmt.Errorf("list memberships: %w", err)
	}
	defer rows.Close()

	var out []MembershipWithCompany
	for rows.Next() {
		var m MembershipWithCompany
		if err := rows.Scan(&m.ID, &m.CompanyID, &m.UserID, &m.RoleID, &m.Status, &m.CompanyName, &m.CompanySlug, &m.RoleName); err != nil {
			return nil, fmt.Errorf("scan membership: %w", err)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// GetByCompanyAndUser returns the active (non-deleted) membership, if any.
func (r *MemberRepository) GetByCompanyAndUser(ctx context.Context, companyID, userID string) (*models.CompanyMember, error) {
	var m models.CompanyMember
	err := r.db.QueryRowContext(ctx, `
		SELECT id, company_id, user_id, role_id, status, invited_at, joined_at, created_at, updated_at, deleted_at
		FROM company_members WHERE company_id = ? AND user_id = ? AND deleted_at IS NULL
	`, companyID, userID).Scan(&m.ID, &m.CompanyID, &m.UserID, &m.RoleID, &m.Status, &m.InvitedAt, &m.JoinedAt, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan membership: %w", err)
	}
	return &m, nil
}

// GetByCompanyAndUserAny returns the membership row for (companyID, userID)
// regardless of deletion state, since the (company_id, user_id) DB
// constraint is unique across soft-deleted rows too — callers re-inviting a
// previously-removed member need this to find the row to revive instead of
// inserting a duplicate.
func (r *MemberRepository) GetByCompanyAndUserAny(ctx context.Context, companyID, userID string) (*models.CompanyMember, error) {
	var m models.CompanyMember
	err := r.db.QueryRowContext(ctx, `
		SELECT id, company_id, user_id, role_id, status, invited_at, joined_at, created_at, updated_at, deleted_at
		FROM company_members WHERE company_id = ? AND user_id = ?
	`, companyID, userID).Scan(&m.ID, &m.CompanyID, &m.UserID, &m.RoleID, &m.Status, &m.InvitedAt, &m.JoinedAt, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan membership: %w", err)
	}
	return &m, nil
}

func (r *MemberRepository) GetByID(ctx context.Context, id string) (*models.CompanyMember, error) {
	var m models.CompanyMember
	err := r.db.QueryRowContext(ctx, `
		SELECT id, company_id, user_id, role_id, status, invited_at, joined_at, created_at, updated_at, deleted_at
		FROM company_members WHERE id = ? AND deleted_at IS NULL
	`, id).Scan(&m.ID, &m.CompanyID, &m.UserID, &m.RoleID, &m.Status, &m.InvitedAt, &m.JoinedAt, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan membership: %w", err)
	}
	return &m, nil
}

func (r *MemberRepository) ListByCompanyID(ctx context.Context, companyID string) ([]MemberWithUser, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT cm.id, cm.company_id, cm.user_id, cm.role_id, cm.status, cm.invited_at, cm.joined_at, cm.created_at, cm.updated_at,
		       u.email, u.first_name, u.last_name, ro.name
		FROM company_members cm
		JOIN users u ON u.id = cm.user_id
		JOIN roles ro ON ro.id = cm.role_id
		WHERE cm.company_id = ? AND cm.deleted_at IS NULL
		ORDER BY cm.created_at ASC
	`, companyID)
	if err != nil {
		return nil, fmt.Errorf("list company members: %w", err)
	}
	defer rows.Close()

	var out []MemberWithUser
	for rows.Next() {
		var m MemberWithUser
		if err := rows.Scan(
			&m.ID, &m.CompanyID, &m.UserID, &m.RoleID, &m.Status, &m.InvitedAt, &m.JoinedAt, &m.CreatedAt, &m.UpdatedAt,
			&m.Email, &m.FirstName, &m.LastName, &m.RoleName,
		); err != nil {
			return nil, fmt.Errorf("scan company member: %w", err)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// ListAcrossCompanies returns memberships spanning all companies, filtered
// by any combination of company, name (matched against first or last name),
// and email substring. Used by the SuperAdmin Users page, which is
// deliberately never called with all filters blank (see AdminService.ListUsers).
func (r *MemberRepository) ListAcrossCompanies(ctx context.Context, companyID, name, email string) ([]MemberWithUserAndCompany, error) {
	query := strings.Builder{}
	query.WriteString(`
		SELECT cm.id, cm.company_id, cm.user_id, cm.role_id, cm.status, cm.invited_at, cm.joined_at, cm.created_at, cm.updated_at,
		       u.email, u.first_name, u.last_name, ro.name, c.name, c.slug
		FROM company_members cm
		JOIN users u ON u.id = cm.user_id
		JOIN roles ro ON ro.id = cm.role_id
		JOIN companies c ON c.id = cm.company_id
		WHERE cm.deleted_at IS NULL AND c.deleted_at IS NULL`)
	var args []any
	if companyID != "" {
		query.WriteString(" AND cm.company_id = ?")
		args = append(args, companyID)
	}
	if name != "" {
		query.WriteString(" AND (u.first_name LIKE ? OR u.last_name LIKE ?)")
		args = append(args, "%"+name+"%", "%"+name+"%")
	}
	if email != "" {
		query.WriteString(" AND u.email LIKE ?")
		args = append(args, "%"+email+"%")
	}
	query.WriteString(" ORDER BY cm.created_at DESC LIMIT 200")

	rows, err := r.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("list members across companies: %w", err)
	}
	defer rows.Close()

	var out []MemberWithUserAndCompany
	for rows.Next() {
		var m MemberWithUserAndCompany
		if err := rows.Scan(
			&m.ID, &m.CompanyID, &m.UserID, &m.RoleID, &m.Status, &m.InvitedAt, &m.JoinedAt, &m.CreatedAt, &m.UpdatedAt,
			&m.Email, &m.FirstName, &m.LastName, &m.RoleName, &m.CompanyName, &m.CompanySlug,
		); err != nil {
			return nil, fmt.Errorf("scan member across companies: %w", err)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *MemberRepository) Create(ctx context.Context, m *models.CompanyMember) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO company_members (id, company_id, user_id, role_id, status, invited_at, joined_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		m.ID, m.CompanyID, m.UserID, m.RoleID, m.Status, m.InvitedAt, m.JoinedAt,
	)
	if err != nil {
		return fmt.Errorf("insert company member: %w", err)
	}
	return nil
}

// Revive un-deletes a previously-removed membership row for a fresh invite,
// resetting it to a pending invite under the given role rather than
// inserting a new row (which would violate the (company_id, user_id)
// uniqueness constraint the soft-deleted row still occupies).
func (r *MemberRepository) Revive(ctx context.Context, id, roleID string) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE company_members SET deleted_at = NULL, status = 'invited', role_id = ?, invited_at = ?, joined_at = NULL, updated_at = ? WHERE id = ?`,
		roleID, now, now, id,
	)
	if err != nil {
		return fmt.Errorf("revive company member: %w", err)
	}
	return nil
}

func (r *MemberRepository) UpdateRole(ctx context.Context, id, roleID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE company_members SET role_id = ?, updated_at = ? WHERE id = ?`,
		roleID, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("update member role: %w", err)
	}
	return nil
}

func (r *MemberRepository) Accept(ctx context.Context, id string) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE company_members SET status = 'active', joined_at = ?, updated_at = ? WHERE id = ?`,
		now, now, id,
	)
	if err != nil {
		return fmt.Errorf("accept invite: %w", err)
	}
	return nil
}

// Delete soft-deletes the membership (sets deleted_at) rather than removing
// the row, so historical/audit data stays intact.
func (r *MemberRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE company_members SET deleted_at = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL`,
		time.Now(), time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("delete company member: %w", err)
	}
	return nil
}
