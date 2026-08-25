package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/fitraditya/useria/internal/models"
)

type CompanyRepository struct {
	db *sql.DB
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

// List returns non-deleted companies. Soft-deleted rows are excluded.
func (r *CompanyRepository) List(ctx context.Context) ([]models.Company, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, slug, description, logo_url, website, status, plan, created_by, created_at, updated_at, deleted_at
		 FROM companies WHERE deleted_at IS NULL ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list companies: %w", err)
	}
	defer rows.Close()

	var out []models.Company
	for rows.Next() {
		var c models.Company
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.LogoURL, &c.Website, &c.Status, &c.Plan, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt); err != nil {
			return nil, fmt.Errorf("scan company: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
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
