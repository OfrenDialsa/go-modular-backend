package repository

import (
	"context"
	"github/OfrenDialsa/go-gin-starter/internal/modules/session/dto"
	"github/OfrenDialsa/go-gin-starter/internal/modules/session/model"
)

type SessionRepository interface {
	Create(ctx context.Context, session *model.UserSession) error
	Update(ctx context.Context, session *model.UserSession) error
	GetByID(ctx context.Context, sessionID string) (*model.UserSession, error)
	GetByToken(ctx context.Context, tokenHash string, tokenType string) (*model.UserSession, error)
	Revoke(ctx context.Context, filter dto.RevokeFilter) error
	Delete(ctx context.Context, filter dto.DeleteFilter) error
}
