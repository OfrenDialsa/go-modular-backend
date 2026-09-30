package service

import (
	"context"
	"errors"
	"fmt"
	"github/OfrenDialsa/go-gin-starter/internal/modules/session/dto"
	"github/OfrenDialsa/go-gin-starter/internal/modules/session/model"
	"github/OfrenDialsa/go-gin-starter/internal/modules/session/repository"
	"github/OfrenDialsa/go-gin-starter/lib"
	"github/OfrenDialsa/go-gin-starter/utils"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

type sessionServiceImpl struct {
	sessionRepo repository.SessionRepository
}

func (s *sessionServiceImpl) Update(ctx context.Context, req dto.UpdateSessionRequest) error {
	sess := &model.UserSession{
		TokenHash: req.TokenHash,
		ExpiresAt: req.ExpiresAt,
		RevokedAt: req.RevokedAt,
		UpdatedAt: time.Now(),
	}

	err := s.sessionRepo.Update(ctx, sess)
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to update session in repository")
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

func (s *sessionServiceImpl) GetByToken(ctx context.Context, token string, typee string) (*dto.GetSession, error) {
	if token == "" || typee == "" {
		return nil, lib.ErrInvalidInput
	}

	sess, err := s.sessionRepo.GetByToken(ctx, token, typee)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, lib.ErrRecordNotFound
		}
		return nil, fmt.Errorf("failed to get user session: %w", err)
	}

	res := &dto.GetSession{
		SessionId:       sess.SessionId,
		UserId:          sess.UserId,
		TokenHash:       sess.TokenHash,
		Type:            sess.Type,
		IPAddress:       sess.IPAddress,
		UserAgent:       sess.UserAgent,
		Role:            sess.Role,
		EmailVerifiedAt: sess.EmailVerifiedAt,
		ExpiresAt:       sess.ExpiresAt,
		CreatedAt:       sess.CreatedAt,
		UpdatedAt:       sess.UpdatedAt,
		RevokedAt:       sess.RevokedAt,
	}

	return res, nil
}

func (s *sessionServiceImpl) DeleteByType(ctx context.Context, userId string, typee string) error {
	if userId == "" || typee == "" {
		return lib.ErrInvalidInput
	}

	filter := dto.DeleteFilter{
		UserID:      userId,
		SessionType: typee,
	}

	if err := s.sessionRepo.Delete(ctx, filter); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

func NewSessionService(sessionRepo repository.SessionRepository) SessionService {
	return &sessionServiceImpl{
		sessionRepo: sessionRepo,
	}
}

func (s *sessionServiceImpl) CreateSession(ctx context.Context, session dto.CreateSession) error {
	now := time.Now()
	hashedToken := utils.HashTokenSHA256(session.TokenHash)

	sess := &model.UserSession{
		SessionId: session.SessionId,
		UserId:    session.UserId,
		TokenHash: hashedToken,
		IPAddress: session.IPAddress,
		UserAgent: session.UserAgent,
		Type:      session.Type,
		ExpiresAt: session.ExpiresAt,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.sessionRepo.Create(ctx, sess); err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

func (s *sessionServiceImpl) GetBySessionId(ctx context.Context, sessionID string) (*dto.GetSession, error) {
	if sessionID == "" {
		return nil, lib.ErrInvalidInput
	}

	sess, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, lib.ErrRecordNotFound
		}
		return nil, fmt.Errorf("failed to get user session: %w", err)
	}

	res := &dto.GetSession{
		SessionId:       sess.SessionId,
		UserId:          sess.UserId,
		TokenHash:       sess.TokenHash,
		Type:            sess.Type,
		IPAddress:       sess.IPAddress,
		UserAgent:       sess.UserAgent,
		Role:            sess.Role,
		EmailVerifiedAt: sess.EmailVerifiedAt,
		ExpiresAt:       sess.ExpiresAt,
		CreatedAt:       sess.CreatedAt,
		UpdatedAt:       sess.UpdatedAt,
		RevokedAt:       sess.RevokedAt,
	}

	return res, nil
}

func (s *sessionServiceImpl) RevokeAllUserSessions(ctx context.Context, userID string) error {
	if userID == "" {
		return lib.ErrInvalidInput
	}

	filter := dto.RevokeFilter{
		UserID: userID,
	}

	if err := s.sessionRepo.Revoke(ctx, filter); err != nil {
		return fmt.Errorf("failed to revoke all sessions for user %s: %w", userID, err)
	}

	return nil
}

func (s *sessionServiceImpl) RevokeOtherSessions(ctx context.Context, userID string, currentSessionID string) error {
	if userID == "" || currentSessionID == "" {
		return lib.ErrInvalidInput
	}

	filter := dto.RevokeFilter{
		UserID:   userID,
		ExceptID: currentSessionID,
	}

	if err := s.sessionRepo.Revoke(ctx, filter); err != nil {
		return fmt.Errorf("failed to revoke all sessions for user %s: %w", userID, err)
	}

	return nil
}

func (s *sessionServiceImpl) RevokeSession(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return lib.ErrInvalidInput
	}

	filter := dto.RevokeFilter{
		SessionID: sessionID,
	}

	if err := s.sessionRepo.Revoke(ctx, filter); err != nil {
		return fmt.Errorf("failed to revoke sessions %s: %w", sessionID, err)
	}

	return nil
}

func (s *sessionServiceImpl) ValidateToken(ctx context.Context, tokenHash string, tokenType string) (*dto.GetSession, error) {
	if tokenHash == "" || tokenType == "" {
		return nil, lib.ErrInvalidInput
	}

	sess, err := s.sessionRepo.GetByToken(ctx, tokenHash, tokenType)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, lib.ErrRecordNotFound
		}

		return nil, fmt.Errorf("failed to get session by token: %w", err)

	}

	res := &dto.GetSession{
		SessionId: sess.SessionId,
		UserId:    sess.UserId,
		TokenHash: sess.TokenHash,
		Type:      sess.Type,
		IPAddress: sess.IPAddress,
		UserAgent: sess.UserAgent,
		ExpiresAt: sess.ExpiresAt,
		CreatedAt: sess.CreatedAt,
		UpdatedAt: sess.UpdatedAt,
		RevokedAt: sess.RevokedAt,
	}

	return res, nil
}

func (s *sessionServiceImpl) DeleteSession(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return lib.ErrInvalidInput
	}

	filter := dto.DeleteFilter{
		SessionID: sessionID,
	}
	if err := s.sessionRepo.Delete(ctx, filter); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}
