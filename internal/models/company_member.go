package models

import "time"

type MemberStatus string

const (
	MemberStatusActive   MemberStatus = "active"
	MemberStatusInvited  MemberStatus = "invited"
	MemberStatusInactive MemberStatus = "inactive"
)

type CompanyMember struct {
	ID        string       `db:"id" json:"id"`
	CompanyID string       `db:"company_id" json:"company_id"`
	UserID    string       `db:"user_id" json:"user_id"`
	RoleID    string       `db:"role_id" json:"role_id"`
	Status    MemberStatus `db:"status" json:"status"`
	InvitedAt *time.Time   `db:"invited_at" json:"invited_at,omitempty"`
	JoinedAt  *time.Time   `db:"joined_at" json:"joined_at,omitempty"`
	CreatedAt time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt time.Time    `db:"updated_at" json:"updated_at"`
}
