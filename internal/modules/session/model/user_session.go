package model

import "time"

type UserSession struct {
	Id        int64      `json:"id"`
	SessionId string     `json:"session_id"`
	UserId    string     `json:"user_id"`
	TokenHash string     `json:"token_hash"`
	Type      string     `json:"type"`
	IPAddress string     `json:"ip_address"`
	UserAgent string     `json:"user_agent"`
	ExpiresAt time.Time  `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`

	// Field tambahan dari JOIN tabel users
	Role            string     `json:"role"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
}
