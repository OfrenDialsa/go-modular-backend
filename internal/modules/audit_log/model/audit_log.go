package model

import (
	"encoding/json"
	"time"
)

type AuditLog struct {
	ID           int64           `json:"id"`
	UserID       *string         `json:"user_id"`
	Action       string          `json:"action"`
	ResourceType string          `json:"resource_type"`
	ResourceID   string          `json:"resource_id"`
	Details      json.RawMessage `json:"details"`
	IPAddress    *string         `json:"ip_address"`
	UserAgent    *string         `json:"user_agent"`
	CreatedAt    time.Time       `json:"created_at"`
}
