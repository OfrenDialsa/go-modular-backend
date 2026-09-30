package repository

import (
	"context"
	"fmt"
	"github/OfrenDialsa/go-gin-starter/database"
	"github/OfrenDialsa/go-gin-starter/internal/modules/audit_log/model"

	"github.com/jackc/pgx/v5"
)

type auditLogRepositoryImpl struct {
	db *database.WrapDB
}

func NewAuditLogRepository(db *database.WrapDB) AuditLogRepository {
	return &auditLogRepositoryImpl{db: db}
}

func (a *auditLogRepositoryImpl) Insert(ctx context.Context, audit *model.AuditLog) error {

	query := `
		INSERT INTO audit_logs (
			user_id, action, resource_type, resource_id, 
			details, ip_address, user_agent, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	executor := database.GetExecutor(ctx, a.db.Database.Conn)

	err := executor.QueryRow(ctx, query,
		audit.UserID,
		audit.Action,
		audit.ResourceType,
		audit.ResourceID,
		audit.Details,
		audit.IPAddress,
		audit.UserAgent,
		audit.CreatedAt,
	).Scan(&audit.ID)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

func (a *auditLogRepositoryImpl) Delete(ctx context.Context, id string) error {
	query := `
		DELETE FROM audit_log
		WHERE id = $1
	`
	executor := database.GetExecutor(ctx, a.db.Database.Conn)

	commandTag, err := executor.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete audit log: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
