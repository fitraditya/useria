package models

import "time"

type SubscriptionStatus string

const (
	SubscriptionStatusTrialing   SubscriptionStatus = "trialing"
	SubscriptionStatusActive     SubscriptionStatus = "active"
	SubscriptionStatusPastDue    SubscriptionStatus = "past_due"
	SubscriptionStatusCanceled   SubscriptionStatus = "canceled"
	SubscriptionStatusIncomplete SubscriptionStatus = "incomplete"
)

// Subscription links a company to a plan. ExternalProvider/ExternalSubscriptionID
// are placeholders for a future payment processor (e.g. Stripe) integration.
// Not yet wired to any service or handler — schema-only groundwork.
type Subscription struct {
	ID                     string             `db:"id" json:"id"`
	CompanyID              string             `db:"company_id" json:"company_id"`
	PlanID                 string             `db:"plan_id" json:"plan_id"`
	Status                 SubscriptionStatus `db:"status" json:"status"`
	ExternalProvider       *string            `db:"external_provider" json:"external_provider,omitempty"`
	ExternalSubscriptionID *string            `db:"external_subscription_id" json:"external_subscription_id,omitempty"`
	CurrentPeriodStart     *time.Time         `db:"current_period_start" json:"current_period_start,omitempty"`
	CurrentPeriodEnd       *time.Time         `db:"current_period_end" json:"current_period_end,omitempty"`
	CancelAtPeriodEnd      bool               `db:"cancel_at_period_end" json:"cancel_at_period_end"`
	TrialEndsAt            *time.Time         `db:"trial_ends_at" json:"trial_ends_at,omitempty"`
	CreatedAt              time.Time          `db:"created_at" json:"created_at"`
	UpdatedAt              time.Time          `db:"updated_at" json:"updated_at"`
}
