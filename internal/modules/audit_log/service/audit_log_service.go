package service

import (
	"context"
	"github/OfrenDialsa/go-gin-starter/internal/modules/audit_log/dto"
)

type AuditLogService interface {
	Create(ctx context.Context, req dto.CreateRequest) error
	DeleteById(ctx context.Context, id string) error
}
