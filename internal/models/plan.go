package models

import "time"

type BillingInterval string

const (
	BillingIntervalMonth BillingInterval = "month"
	BillingIntervalYear  BillingInterval = "year"
)

// Plan is a billable tier (free, pro, enterprise, ...). Not yet wired to any
// service or handler — schema-only groundwork for recurring billing.
type Plan struct {
	ID              string          `db:"id" json:"id"`
	Code            string          `db:"code" json:"code"`
	Name            string          `db:"name" json:"name"`
	Description     *string         `db:"description" json:"description,omitempty"`
	PriceCents      int             `db:"price_cents" json:"price_cents"`
	Currency        string          `db:"currency" json:"currency"`
	BillingInterval BillingInterval `db:"billing_interval" json:"billing_interval"`
	IsActive        bool            `db:"is_active" json:"is_active"`
	CreatedAt       time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at" json:"updated_at"`
}
