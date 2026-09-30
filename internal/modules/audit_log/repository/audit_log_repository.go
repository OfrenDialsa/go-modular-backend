package repository

import (
	"context"
	"github/OfrenDialsa/go-gin-starter/internal/modules/audit_log/model"
)

type AuditLogRepository interface {
	Insert(ctx context.Context, audit *model.AuditLog) error
	Delete(ctx context.Context, id string) error
}
