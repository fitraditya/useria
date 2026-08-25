package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/fitraditya/useria/internal/models"
)

// AuditRepository writes to (and, under mysql, reads from) the audit_logs
// table. Only used under the mysql driver — audit_logs isn't created by the
// sqlite migration, see service.AuditService.
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

// AuditLogWithActor is an audit entry joined with the acting user and (if
// the event was company-scoped) the company name, for display on the
// SuperAdmin activity log.
type AuditLogWithActor struct {
	models.AuditLog
	ActorEmail     string  `json:"actor_email"`
	ActorFirstName *string `json:"actor_first_name,omitempty"`
	ActorLastName  *string `json:"actor_last_name,omitempty"`
	CompanyName    *string `json:"company_name,omitempty"`
}

// List returns audit entries newest-first, optionally filtered by company,
// actor name, actor email, action, and a created_at date range (dateFrom/
// dateTo are inclusive "YYYY-MM-DD" strings). Any empty filter is skipped.
func (r *AuditRepository) List(ctx context.Context, companyID, name, email, action, dateFrom, dateTo string) ([]AuditLogWithActor, error) {
	query := strings.Builder{}
	query.WriteString(`
		SELECT al.id, al.actor_user_id, al.action, al.resource_type, al.resource_id, al.company_id, al.metadata, al.created_at,
		       u.email, u.first_name, u.last_name, c.name
		FROM audit_logs al
		JOIN users u ON u.id = al.actor_user_id
		LEFT JOIN companies c ON c.id = al.company_id
		WHERE 1 = 1`)
	var args []any
	if companyID != "" {
		query.WriteString(" AND al.company_id = ?")
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
	if action != "" {
		query.WriteString(" AND al.action = ?")
		args = append(args, action)
	}
	if dateFrom != "" {
		query.WriteString(" AND al.created_at >= ?")
		args = append(args, dateFrom+" 00:00:00")
	}
	if dateTo != "" {
		query.WriteString(" AND al.created_at <= ?")
		args = append(args, dateTo+" 23:59:59")
	}
	query.WriteString(" ORDER BY al.created_at DESC LIMIT 300")

	rows, err := r.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("list audit logs: %w", err)
	}
	defer rows.Close()

	var out []AuditLogWithActor
	for rows.Next() {
		var a AuditLogWithActor
		if err := rows.Scan(&a.ID, &a.ActorUserID, &a.Action, &a.ResourceType, &a.ResourceID, &a.CompanyID, &a.Metadata, &a.CreatedAt,
			&a.ActorEmail, &a.ActorFirstName, &a.ActorLastName, &a.CompanyName); err != nil {
			return nil, fmt.Errorf("scan audit log: %w", err)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
