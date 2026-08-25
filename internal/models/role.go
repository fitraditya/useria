package models

import "time"

type Role struct {
	ID          string    `db:"id" json:"id"`
	CompanyID   *string   `db:"company_id" json:"company_id,omitempty"`
	Name        string    `db:"name" json:"name"`
	Description *string   `db:"description" json:"description,omitempty"`
	IsSystem    bool      `db:"is_system" json:"is_system"`
	IsCustom    bool      `db:"is_custom" json:"is_custom"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
