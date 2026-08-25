package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/fitraditya/useria/internal/models"
)

// AuditRepository writes to the audit_logs table. Only used under the mysql
// driver — audit_logs isn't created by the sqlite migration.
type AuditRepository struct {
	db *sql.DB
}

func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Create(ctx context.Context, a *models.AuditLog) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO audit_logs (id, actor_user_id, action, resource_type, resource_id, company_id, metadata)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		a.ID, a.ActorUserID, a.Action, a.ResourceType, a.ResourceID, a.CompanyID, a.Metadata,
	)
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	return nil
}
