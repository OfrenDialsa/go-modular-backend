package api

import (
	"context"
	"sync"
	"time"

	"github/OfrenDialsa/go-gin-starter/config"
	"github/OfrenDialsa/go-gin-starter/database"
	"github/OfrenDialsa/go-gin-starter/internal/infra/mailer"
	"github/OfrenDialsa/go-gin-starter/internal/infra/storage"
	"github/OfrenDialsa/go-gin-starter/internal/modules/auth"
	sessRepo "github/OfrenDialsa/go-gin-starter/internal/modules/session/repository"
	sessSvc "github/OfrenDialsa/go-gin-starter/internal/modules/session/service"
	"github/OfrenDialsa/go-gin-starter/internal/modules/user"
	"github/OfrenDialsa/go-gin-starter/middleware"
	"github/OfrenDialsa/go-gin-starter/router"

	"github.com/gin-gonic/gin"
	"github.com/nsqio/go-nsq"
	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog/log"
)

type Setup struct {
	Router     *gin.Engine
	WrapDB     *database.WrapDB
	Producer   *nsq.Producer
	Mailer     mailer.MailerService
	AuthModule *auth.Module
	UserModule *user.Module
}

func Init(env *config.EnvironmentVariable, wrapDB *database.WrapDB) (*Setup, error) {
	mailerService := mailer.NewMailerService(env, env.Mail.From, env.Mail.FromName)
	storageService, err := storage.NewStorageService(context.Background(), env)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize storage service")
		return nil, err
	}

	wg := &sync.WaitGroup{}
	cCache := cache.New(5*time.Minute, 10*time.Minute)
	sRepo := sessRepo.NewSessionRepository(wrapDB)
	sSvc := sessSvc.NewSessionService(sRepo)
	mw := middleware.NewMiddleware(env, sSvc, cCache)

	authModule := auth.NewModule(env, wrapDB, mailerService, storageService, wg, mw)
	userModule := user.NewModule(env, wrapDB, mailerService, storageService, wg, mw)

	routes := router.NewRouter(router.Config{
		Env:        env,
		Middleware: mw,
		DB:         wrapDB,
		AuthModule: authModule,
		UserModule: userModule,
	})

	return &Setup{
		Router:     routes,
		WrapDB:     wrapDB,
		Mailer:     mailerService,
		AuthModule: authModule,
		UserModule: userModule,
	}, nil
}
