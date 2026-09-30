package service

import (
	"context"
	"github/OfrenDialsa/go-gin-starter/internal/modules/session/dto"
)

type SessionService interface {
	CreateSession(ctx context.Context, session dto.CreateSession) error
	GetBySessionId(ctx context.Context, sessionID string) (*dto.GetSession, error)
	GetByToken(ctx context.Context, token, typee string) (*dto.GetSession, error)
	ValidateToken(ctx context.Context, tokenHash, tokenType string) (*dto.GetSession, error)
	RevokeSession(ctx context.Context, sessionID string) error
	Update(ctx context.Context, req dto.UpdateSessionRequest) error
	RevokeAllUserSessions(ctx context.Context, userID string) error
	RevokeOtherSessions(ctx context.Context, userID string, currentSessionID string) error
	DeleteSession(ctx context.Context, sessionID string) error
	DeleteByType(ctx context.Context, userId, typee string) error
}
