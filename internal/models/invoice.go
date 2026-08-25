package models

import "time"

type InvoiceStatus string

const (
	InvoiceStatusDraft         InvoiceStatus = "draft"
	InvoiceStatusOpen          InvoiceStatus = "open"
	InvoiceStatusPaid          InvoiceStatus = "paid"
	InvoiceStatusVoid          InvoiceStatus = "void"
	InvoiceStatusUncollectible InvoiceStatus = "uncollectible"
)

// Invoice is a billing record for a company, optionally tied to a
// Subscription. Not yet wired to any service or handler — schema-only
// groundwork for recurring billing.
type Invoice struct {
	ID                string        `db:"id" json:"id"`
	CompanyID         string        `db:"company_id" json:"company_id"`
	SubscriptionID    *string       `db:"subscription_id" json:"subscription_id,omitempty"`
	ExternalInvoiceID *string       `db:"external_invoice_id" json:"external_invoice_id,omitempty"`
	AmountCents       int           `db:"amount_cents" json:"amount_cents"`
	Currency          string        `db:"currency" json:"currency"`
	Status            InvoiceStatus `db:"status" json:"status"`
	PeriodStart       *time.Time    `db:"period_start" json:"period_start,omitempty"`
	PeriodEnd         *time.Time    `db:"period_end" json:"period_end,omitempty"`
	PaidAt            *time.Time    `db:"paid_at" json:"paid_at,omitempty"`
	CreatedAt         time.Time     `db:"created_at" json:"created_at"`
}
