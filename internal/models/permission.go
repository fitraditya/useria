package models

import "time"

type Permission struct {
	ID          string    `db:"id" json:"id"`
	Code        string    `db:"code" json:"code"`
	Name        *string   `db:"name" json:"name,omitempty"`
	Description *string   `db:"description" json:"description,omitempty"`
	Category    *string   `db:"category" json:"category,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type RolePermission struct {
	ID           string    `db:"id" json:"id"`
	RoleID       string    `db:"role_id" json:"role_id"`
	PermissionID string    `db:"permission_id" json:"permission_id"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}
