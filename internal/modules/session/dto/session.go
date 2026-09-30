package dto

import "time"

type RevokeFilter struct {
	SessionID string
	UserID    string
	TokenHash string
	ExceptID  string
}

type DeleteFilter struct {
	SessionID   string
	UserID      string
	SessionType string
}

type CreateSession struct {
	SessionId string
	UserId    string
	TokenHash string
	IPAddress string
	UserAgent string
	Type      string
	ExpiresAt time.Time
}

type GetSession struct {
	SessionId       string     `json:"session_id"`
	UserId          string     `json:"user_id"`
	TokenHash       string     `json:"-"`
	Type            string     `json:"type"`
	IPAddress       string     `json:"ip_address"`
	UserAgent       string     `json:"user_agent"`
	Role            string     `json:"role"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	ExpiresAt       time.Time  `json:"expires_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	RevokedAt       *time.Time `json:"revoked_at,omitempty"`
}

type UpdateSessionRequest struct {
	TokenHash string     `json:"token_hash"`
	ExpiresAt time.Time  `json:"expires_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}
