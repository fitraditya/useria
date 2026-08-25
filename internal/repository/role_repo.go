package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/fitraditya/useria/internal/models"
)

type RoleRepository struct {
	db *sql.DB
}

func NewRoleRepository(db *sql.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) GetByID(ctx context.Context, id string) (*models.Role, error) {
	var role models.Role
	err := r.db.QueryRowContext(ctx, `
		SELECT id, company_id, name, description, is_system, is_custom, created_at, updated_at
		FROM roles WHERE id = ?
	`, id).Scan(&role.ID, &role.CompanyID, &role.Name, &role.Description, &role.IsSystem, &role.IsCustom, &role.CreatedAt, &role.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan role: %w", err)
	}
	return &role, nil
}

// GetSystemRoleByName looks up a platform-wide system role (company_id IS
// NULL) by its exact name, e.g. "Admin".
func (r *RoleRepository) GetSystemRoleByName(ctx context.Context, name string) (*models.Role, error) {
	var role models.Role
	err := r.db.QueryRowContext(ctx, `
		SELECT id, company_id, name, description, is_system, is_custom, created_at, updated_at
		FROM roles WHERE company_id IS NULL AND name = ?
	`, name).Scan(&role.ID, &role.CompanyID, &role.Name, &role.Description, &role.IsSystem, &role.IsCustom, &role.CreatedAt, &role.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan role: %w", err)
	}
	return &role, nil
}

// ListAssignable returns the system roles (excluding the platform-only
// SuperAdmin) plus any custom roles owned by companyID. These are the roles
// a company admin may hand out via invite or role update.
func (r *RoleRepository) ListAssignable(ctx context.Context, companyID string) ([]models.Role, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, company_id, name, description, is_system, is_custom, created_at, updated_at
		FROM roles
		WHERE (company_id IS NULL AND name != 'SuperAdmin') OR company_id = ?
		ORDER BY name ASC
	`, companyID)
	if err != nil {
		return nil, fmt.Errorf("list assignable roles: %w", err)
	}
	defer rows.Close()

	var out []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.CompanyID, &role.Name, &role.Description, &role.IsSystem, &role.IsCustom, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		out = append(out, role)
	}
	return out, rows.Err()
}

func (r *RoleRepository) PermissionCodes(ctx context.Context, roleID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.code FROM role_permissions rp
		JOIN permissions p ON p.id = rp.permission_id
		WHERE rp.role_id = ?
	`, roleID)
	if err != nil {
		return nil, fmt.Errorf("list permission codes: %w", err)
	}
	defer rows.Close()

	var codes []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, fmt.Errorf("scan permission code: %w", err)
		}
		codes = append(codes, code)
	}
	return codes, rows.Err()
}
