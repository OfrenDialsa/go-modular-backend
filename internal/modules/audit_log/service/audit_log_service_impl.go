package service

import (
	"context"
	"errors"
	"fmt"
	"github/OfrenDialsa/go-gin-starter/internal/modules/audit_log/dto"
	"github/OfrenDialsa/go-gin-starter/internal/modules/audit_log/model"
	"github/OfrenDialsa/go-gin-starter/internal/modules/audit_log/repository"
	"github/OfrenDialsa/go-gin-starter/lib"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

type auditLogServiceImpl struct {
	auditRepo repository.AuditLogRepository
}

func NewAuditLogService(
	auditRepo repository.AuditLogRepository,
) AuditLogService {
	return &auditLogServiceImpl{
		auditRepo: auditRepo,
	}
}

func (a *auditLogServiceImpl) Create(ctx context.Context, req dto.CreateRequest) error {
	auditLog := &model.AuditLog{
		UserID:       req.UserID,
		Action:       req.Action,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		Details:      req.Details,
		CreatedAt:    time.Now(),
	}

	if err := a.auditRepo.Insert(ctx, auditLog); err != nil {
		log.Error().Err(err).Msg("failed to create audit log")
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

func (a *auditLogServiceImpl) DeleteById(ctx context.Context, id string) error {
	if err := a.auditRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return lib.ErrAuditLogNotFound
		}
		log.Error().Err(err).Msg("failed to delete audit log")
		return lib.ErrInternalServer
	}
	return nil
}
