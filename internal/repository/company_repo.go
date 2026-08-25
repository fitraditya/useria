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

type CompanyRepository struct {
	db *sql.DB
}

// CompanyWithAdmin is a company plus the email of its first (oldest) Admin
// member, if it has one — used by the SuperAdmin companies list so it can
// filter/display by admin email without a separate lookup per row.
type CompanyWithAdmin struct {
	models.Company
	AdminEmail *string `json:"admin_email,omitempty"`
}

func NewCompanyRepository(db *sql.DB) *CompanyRepository {
	return &CompanyRepository{db: db}
}

func (r *CompanyRepository) GetByID(ctx context.Context, id string) (*models.Company, error) {
	var c models.Company
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, slug, description, logo_url, website, status, plan, created_by, created_at, updated_at, deleted_at
		 FROM companies WHERE id = ? AND deleted_at IS NULL`, id,
	).Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.LogoURL, &c.Website, &c.Status, &c.Plan, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan company: %w", err)
	}
	return &c, nil
}

func (r *CompanyRepository) GetBySlug(ctx context.Context, slug string) (*models.Company, error) {
	var c models.Company
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, slug, description, logo_url, website, status, plan, created_by, created_at, updated_at, deleted_at
		 FROM companies WHERE slug = ? AND deleted_at IS NULL`, slug,
	).Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.LogoURL, &c.Website, &c.Status, &c.Plan, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan company: %w", err)
	}
	return &c, nil
}

func (r *CompanyRepository) Create(ctx context.Context, c *models.Company) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO companies (id, name, slug, description, website, plan, status, created_by)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ID, c.Name, c.Slug, c.Description, c.Website, c.Plan, c.Status, c.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("insert company: %w", err)
	}
	return nil
}

// List returns non-deleted companies, optionally filtered by name substring,
// admin email substring, and/or exact status. Any empty filter is skipped.
func (r *CompanyRepository) List(ctx context.Context, name, adminEmail, status string) ([]CompanyWithAdmin, error) {
	query := strings.Builder{}
	query.WriteString(`
		SELECT c.id, c.name, c.slug, c.description, c.logo_url, c.website, c.status, c.plan, c.created_by, c.created_at, c.updated_at, c.deleted_at,
		       (SELECT u.email FROM company_members cm
		        JOIN users u ON u.id = cm.user_id
		        JOIN roles r ON r.id = cm.role_id
		        WHERE cm.company_id = c.id AND cm.deleted_at IS NULL AND r.company_id IS NULL AND r.name = 'Admin'
		        ORDER BY cm.created_at ASC LIMIT 1) AS admin_email
		FROM companies c
		WHERE c.deleted_at IS NULL`)
	var args []any
	if name != "" {
		query.WriteString(" AND c.name LIKE ?")
		args = append(args, "%"+name+"%")
	}
	if status != "" {
		query.WriteString(" AND c.status = ?")
		args = append(args, status)
	}
	if adminEmail != "" {
		query.WriteString(` AND EXISTS (
			SELECT 1 FROM company_members cm
			JOIN users u ON u.id = cm.user_id
			JOIN roles r ON r.id = cm.role_id
			WHERE cm.company_id = c.id AND cm.deleted_at IS NULL AND r.company_id IS NULL AND r.name = 'Admin' AND u.email LIKE ?
		)`)
		args = append(args, "%"+adminEmail+"%")
	}
	query.WriteString(" ORDER BY c.created_at DESC")

	rows, err := r.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("list companies: %w", err)
	}
	defer rows.Close()

	var out []CompanyWithAdmin
	for rows.Next() {
		var c CompanyWithAdmin
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.LogoURL, &c.Website, &c.Status, &c.Plan, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt, &c.AdminEmail); err != nil {
			return nil, fmt.Errorf("scan company: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// Count returns the number of non-deleted companies.
func (r *CompanyRepository) Count(ctx context.Context) (int, error) {
	var n int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM companies WHERE deleted_at IS NULL`).Scan(&n); err != nil {
		return 0, fmt.Errorf("count companies: %w", err)
	}
	return n, nil
}

func (r *CompanyRepository) Update(ctx context.Context, c *models.Company) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE companies SET name = ?, description = ?, website = ?, status = ?, plan = ?, updated_at = ?
		 WHERE id = ? AND deleted_at IS NULL`,
		c.Name, c.Description, c.Website, c.Status, c.Plan, time.Now(), c.ID,
	)
	if err != nil {
		return fmt.Errorf("update company: %w", err)
	}
	return nil
}

// Delete soft-deletes the company (sets deleted_at) rather than removing the
// row, so historical data (members, audit trail, billing records) stays
// intact. Excluded from GetByID/GetBySlug/List once deleted.
func (r *CompanyRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE companies SET deleted_at = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL`,
		time.Now(), time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("delete company: %w", err)
	}
	return nil
}
