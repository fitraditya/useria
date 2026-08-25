package models

import "time"

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
)

type User struct {
	ID            string     `db:"id" json:"id"`
	Email         string     `db:"email" json:"email"`
	PasswordHash  *string    `db:"password_hash" json:"-"`
	FirstName     *string    `db:"first_name" json:"first_name,omitempty"`
	LastName      *string    `db:"last_name" json:"last_name,omitempty"`
	AvatarURL     *string    `db:"avatar_url" json:"avatar_url,omitempty"`
	OAuthProvider *string    `db:"oauth_provider" json:"oauth_provider,omitempty"`
	OAuthID       *string    `db:"oauth_id" json:"-"`
	Status        UserStatus `db:"status" json:"status"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}
