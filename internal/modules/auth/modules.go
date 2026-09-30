package auth

import (
	"github/OfrenDialsa/go-gin-starter/config"
	"github/OfrenDialsa/go-gin-starter/database"
	"github/OfrenDialsa/go-gin-starter/internal/infra/mailer"
	"github/OfrenDialsa/go-gin-starter/internal/infra/storage"
	auditRepo "github/OfrenDialsa/go-gin-starter/internal/modules/audit_log/repository"
	auditSvc "github/OfrenDialsa/go-gin-starter/internal/modules/audit_log/service"
	"github/OfrenDialsa/go-gin-starter/internal/modules/auth/handler"
	"github/OfrenDialsa/go-gin-starter/internal/modules/auth/service"
	sessRepo "github/OfrenDialsa/go-gin-starter/internal/modules/session/repository"
	sessSvc "github/OfrenDialsa/go-gin-starter/internal/modules/session/service"
	userRepo "github/OfrenDialsa/go-gin-starter/internal/modules/user/repository"
	userSvc "github/OfrenDialsa/go-gin-starter/internal/modules/user/service"
	"github/OfrenDialsa/go-gin-starter/middleware"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Module struct {
	h  handler.AuthHandler
	mw middleware.Middleware
}

func NewModule(env *config.EnvironmentVariable, db *database.WrapDB, mailer mailer.MailerService, st storage.StorageService, wg *sync.WaitGroup, mw middleware.Middleware) *Module {
	txManager := database.NewTxManager(db.Database.Conn)
	repo := sessRepo.NewSessionRepository(db)
	auditRepo := auditRepo.NewAuditLogRepository(db)
	auditSvc := auditSvc.NewAuditLogService(auditRepo)
	sessionSvc := sessSvc.NewSessionService(repo)
	userRepo := userRepo.NewUserRepository(db)
	userSvc := userSvc.NewUserService(env, txManager, userRepo, sessionSvc, auditSvc, st)
	svc := service.NewAuthService(env, txManager, mailer, userSvc, sessionSvc, wg)

	hdlr := handler.NewAuthHandler(env, svc)

	return &Module{
		h:  hdlr,
		mw: mw,
	}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	{
		auth.POST("/check-availability", m.mw.RateLimit(30, time.Minute), m.h.CheckAvailability)
		auth.POST("/register", m.mw.RateLimit(10, time.Minute), m.h.Register)
		auth.POST("/resend-verification", m.mw.Validate(), m.mw.RateLimit(10, time.Minute), m.h.ResendVerification)
		auth.GET("/verify-email", m.mw.RateLimit(10, time.Minute), m.h.VerifyEmail)
		auth.POST("/login", m.mw.RateLimit(10, time.Minute), m.h.Login)
		auth.POST("/forgot-password", m.mw.RateLimit(3, 15*time.Minute), m.h.ForgotPassword)
		auth.POST("/reset-password", m.h.ResetPassword)
		auth.POST("/refresh", m.h.RefreshToken)
		auth.POST("/logout", m.mw.Validate(), m.h.Logout)
	}
}
