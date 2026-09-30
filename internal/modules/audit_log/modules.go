package audit

import (
	"github/OfrenDialsa/go-gin-starter/database"
	"github/OfrenDialsa/go-gin-starter/internal/modules/audit_log/repository"
	"github/OfrenDialsa/go-gin-starter/internal/modules/audit_log/service"
)

type Module struct {
	Service service.AuditLogService
}

func NewModule(db *database.WrapDB) *Module {
	repo := repository.NewAuditLogRepository(db)
	svc := service.NewAuditLogService(repo)

	return &Module{
		Service: svc,
	}
}
