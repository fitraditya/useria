package models

import "time"

// AuditLog records a sensitive mutation for accountability. Only persisted
// under the MySQL driver — see service.AuditService.
type AuditLog struct {
	ID           string    `db:"id" json:"id"`
	ActorUserID  string    `db:"actor_user_id" json:"actor_user_id"`
	Action       string    `db:"action" json:"action"`
	ResourceType string    `db:"resource_type" json:"resource_type"`
	ResourceID   string    `db:"resource_id" json:"resource_id"`
	CompanyID    *string   `db:"company_id" json:"company_id,omitempty"`
	Metadata     *string   `db:"metadata" json:"metadata,omitempty"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}
