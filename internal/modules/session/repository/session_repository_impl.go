package repository

import (
	"context"
	"fmt"
	"github/OfrenDialsa/go-gin-starter/database"
	"github/OfrenDialsa/go-gin-starter/internal/modules/session/dto"
	"github/OfrenDialsa/go-gin-starter/internal/modules/session/model"

	"github.com/jackc/pgx/v5"
)

type sessionRepositoryImpl struct {
	db *database.WrapDB
}

func NewSessionRepository(db *database.WrapDB) SessionRepository {
	return &sessionRepositoryImpl{db: db}
}

func (r *sessionRepositoryImpl) Create(ctx context.Context, session *model.UserSession) error {
	query := `
        INSERT INTO user_sessions (
            session_id, user_id, token_hash, type, 
            ip_address, user_agent, expires_at, created_at, updated_at
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        RETURNING id
    `

	executor := database.GetExecutor(ctx, r.db.Database.Conn)

	err := executor.QueryRow(ctx, query,
		session.SessionId,
		session.UserId,
		session.TokenHash,
		session.Type,
		session.IPAddress,
		session.UserAgent,
		session.ExpiresAt,
		session.CreatedAt,
		session.UpdatedAt,
	).Scan(&session.Id)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return err
}

func (r *sessionRepositoryImpl) Update(ctx context.Context, session *model.UserSession) error {
	query := `
		UPDATE user_sessions 
		SET token_hash = $1, 
		    expires_at = $2, 
		    updated_at = $3, 
		    ip_address = $4, 
		    user_agent = $5
		WHERE session_id = $6 AND revoked_at IS NULL
	`
	executor := database.GetExecutor(ctx, r.db.Database.Conn)

	_, err := executor.Exec(ctx, query,
		session.TokenHash,
		session.ExpiresAt,
		session.UpdatedAt,
		session.IPAddress,
		session.UserAgent,
		session.SessionId,
	)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return err
}
func (r *sessionRepositoryImpl) GetByToken(ctx context.Context, tokenHash string, tokenType string) (*model.UserSession, error) {
	query := `
        SELECT id, session_id, user_id, token_hash, type, ip_address, user_agent, expires_at, created_at, updated_at, revoked_at
        FROM user_sessions
        WHERE token_hash = $1 AND type = $2
    `

	var session model.UserSession

	executor := database.GetExecutor(ctx, r.db.Database.Conn)
	err := executor.QueryRow(ctx, query, tokenHash, tokenType).Scan(
		&session.Id,
		&session.SessionId,
		&session.UserId,
		&session.TokenHash,
		&session.Type,
		&session.IPAddress,
		&session.UserAgent,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.UpdatedAt,
		&session.RevokedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get session by token: %w", err)
	}
	return &session, nil
}

func (r *sessionRepositoryImpl) GetByID(ctx context.Context, sessionID string) (*model.UserSession, error) {
	query := `
        SELECT 
            s.id, s.session_id, s.user_id, s.token_hash, s.type, s.ip_address, 
            s.user_agent, s.expires_at, s.created_at, s.updated_at, s.revoked_at,
            u.role, u.email_verified_at
        FROM user_sessions s
        JOIN users u ON s.user_id = u.user_id
        WHERE s.session_id = $1 
          AND s.revoked_at IS NULL 
          AND s.expires_at > NOW()
    `

	var session model.UserSession

	executor := database.GetExecutor(ctx, r.db.Database.Conn)
	err := executor.QueryRow(ctx, query, sessionID).Scan(
		&session.Id, &session.SessionId, &session.UserId, &session.TokenHash,
		&session.Type, &session.IPAddress, &session.UserAgent, &session.ExpiresAt,
		&session.CreatedAt, &session.UpdatedAt, &session.RevokedAt,
		&session.Role, &session.EmailVerifiedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepositoryImpl) Revoke(ctx context.Context, filter dto.RevokeFilter) error {
	query := `UPDATE user_sessions SET is_revoked = true, updated_at = NOW() WHERE 1=1`
	args := []any{}
	argIdx := 1

	if filter.SessionID != "" {
		query += fmt.Sprintf(" AND session_id = $%d", argIdx)
		args = append(args, filter.SessionID)
		argIdx++
	}

	if filter.TokenHash != "" {
		query += fmt.Sprintf(" AND token_hash = $%d", argIdx)
		args = append(args, filter.TokenHash)
		argIdx++
	}

	if filter.UserID != "" {
		query += fmt.Sprintf(" AND user_id = $%d", argIdx)
		args = append(args, filter.UserID)
		argIdx++
	}

	if filter.ExceptID != "" {
		query += fmt.Sprintf(" AND session_id != $%d", argIdx)
		args = append(args, filter.ExceptID)
		argIdx++
	}

	executor := database.GetExecutor(ctx, r.db.Database.Conn)
	_, err := executor.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	return nil
}

func (r *sessionRepositoryImpl) Delete(ctx context.Context, filter dto.DeleteFilter) error {
	query := `
		DELETE FROM user_sessions
		WHERE session_id = $1
	`
	args := []any{}
	argIdx := 1

	if filter.SessionID != "" {
		query += fmt.Sprintf(" AND session_id = $%d", argIdx)
		args = append(args, filter.SessionID)
		argIdx++
	}

	if filter.UserID != "" {
		query += fmt.Sprintf(" AND user_id = $%d", argIdx)
		args = append(args, filter.UserID)
		argIdx++
	}

	if filter.SessionType != "" {
		query += fmt.Sprintf(" AND session_type != $%d", argIdx)
		args = append(args, filter.SessionType)
	}

	executor := database.GetExecutor(ctx, r.db.Database.Conn)
	_, err := executor.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}
