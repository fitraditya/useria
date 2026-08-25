package models

import "time"

type OAuthToken struct {
	ID           string    `db:"id" json:"id"`
	UserID       string    `db:"user_id" json:"user_id"`
	CompanyID    string    `db:"company_id" json:"company_id"`
	AccessToken  string    `db:"access_token" json:"-"`
	RefreshToken *string   `db:"refresh_token" json:"-"`
	TokenType    string    `db:"token_type" json:"token_type"`
	Scopes       *string   `db:"scopes" json:"scopes,omitempty"`
	ExpiresAt    time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}
