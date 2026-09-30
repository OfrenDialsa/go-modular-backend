package session

import (
	"github/OfrenDialsa/go-gin-starter/database"
	"github/OfrenDialsa/go-gin-starter/internal/modules/session/repository"
	"github/OfrenDialsa/go-gin-starter/internal/modules/session/service"
)

type Module struct {
	SessionService service.SessionService
}

func NewModule(db *database.WrapDB) *Module {
	repo := repository.NewSessionRepository(db)
	sessionSvc := service.NewSessionService(repo)

	return &Module{
		SessionService: sessionSvc,
	}
}
