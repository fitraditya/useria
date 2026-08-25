package models

import "time"

type CompanyStatus string

const (
	CompanyStatusActive    CompanyStatus = "active"
	CompanyStatusInactive  CompanyStatus = "inactive"
	CompanyStatusSuspended CompanyStatus = "suspended"
)

type Company struct {
	ID          string        `db:"id" json:"id"`
	Name        string        `db:"name" json:"name"`
	Slug        string        `db:"slug" json:"slug"`
	Description *string       `db:"description" json:"description,omitempty"`
	LogoURL     *string       `db:"logo_url" json:"logo_url,omitempty"`
	Website     *string       `db:"website" json:"website,omitempty"`
	Status      CompanyStatus `db:"status" json:"status"`
	Plan        string        `db:"plan" json:"plan"`
	CreatedBy   string        `db:"created_by" json:"created_by"`
	CreatedAt   time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time     `db:"updated_at" json:"updated_at"`
	DeletedAt   *time.Time    `db:"deleted_at" json:"deleted_at,omitempty"`
}
