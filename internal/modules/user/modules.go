package user

import (
	"github/OfrenDialsa/go-gin-starter/config"
	"github/OfrenDialsa/go-gin-starter/database"
	"github/OfrenDialsa/go-gin-starter/internal/infra/mailer"
	"github/OfrenDialsa/go-gin-starter/internal/infra/storage"
	auditRepo "github/OfrenDialsa/go-gin-starter/internal/modules/audit_log/repository"
	auditSvc "github/OfrenDialsa/go-gin-starter/internal/modules/audit_log/service"
	sessionRepo "github/OfrenDialsa/go-gin-starter/internal/modules/session/repository"
	sessionSvc "github/OfrenDialsa/go-gin-starter/internal/modules/session/service"
	"github/OfrenDialsa/go-gin-starter/internal/modules/user/handler"
	"github/OfrenDialsa/go-gin-starter/internal/modules/user/repository"
	"github/OfrenDialsa/go-gin-starter/internal/modules/user/service"
	"github/OfrenDialsa/go-gin-starter/middleware"
	"sync"

	"github.com/gin-gonic/gin"
)

type Module struct {
	hdlr handler.UserHandler
	mw   middleware.Middleware
}

func NewModule(env *config.EnvironmentVariable, db *database.WrapDB, mailer mailer.MailerService, st storage.StorageService, wg *sync.WaitGroup, mw middleware.Middleware) *Module {
	txManager := database.NewTxManager(db.Database.Conn)
	repo := repository.NewUserRepository(db)
	sessionrepo := sessionRepo.NewSessionRepository(db)
	sessionSvc := sessionSvc.NewSessionService(sessionrepo)
	auditRepo := auditRepo.NewAuditLogRepository(db)
	auditSvc := auditSvc.NewAuditLogService(auditRepo)
	svc := service.NewUserService(env, txManager, repo, sessionSvc, auditSvc, st)
	hdlr := handler.NewUserHandler(svc)

	return &Module{
		hdlr: hdlr,
		mw:   mw,
	}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	users.Use(m.mw.Validate())
	{
		users.GET("/me", m.hdlr.GetMe)
		users.PUT("/me", m.mw.EmailVerified(), m.hdlr.UpdateProfile)
		users.PUT("/me/password", m.hdlr.ChangePassword)
		users.DELETE("/me/avatar", m.hdlr.DeleteAvatar)
		users.DELETE("/me", m.hdlr.DeleteAccount)
	}
}
