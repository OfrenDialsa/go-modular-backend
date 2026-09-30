package router

import (
	"context"
	"github/OfrenDialsa/go-gin-starter/config"
	"github/OfrenDialsa/go-gin-starter/database"
	_ "github/OfrenDialsa/go-gin-starter/docs"
	"github/OfrenDialsa/go-gin-starter/internal/modules/auth"
	"github/OfrenDialsa/go-gin-starter/internal/modules/user"
	"github/OfrenDialsa/go-gin-starter/middleware"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

type Config struct {
	Env        *config.EnvironmentVariable
	Middleware middleware.Middleware
	DB         *database.WrapDB
	AuthModule *auth.Module
	UserModule *user.Module
}

func NewRouter(cfg Config) *gin.Engine {
	r := gin.Default()
	r.ForwardedByClientIP = true
	r.SetTrustedProxies(nil)

	r.Use(cors.Default())

	if cfg.Middleware != nil {
		r.Use(cfg.Middleware.Prometheus())
	}

	base := r.Group("/")
	{
		base.GET("", func(ctx *gin.Context) {
			ctx.JSON(200, "Go gin starter kit by nerodev")
		})

		base.GET("/health", func(ctx *gin.Context) {
			dbStat := "connected"
			statCode := 200
			success := true

			if cfg.DB == nil || cfg.DB.Database.Conn == nil {
				dbStat = "not-configured"
				statCode = 500
				success = false
			} else {
				pingCtx, cancel := context.WithTimeout(ctx.Request.Context(), 2*time.Second)
				defer cancel()

				if err := cfg.DB.Database.Conn.Ping(pingCtx); err != nil {
					dbStat = "disconnected"
					statCode = 500
					success = false
				}
			}

			msg := "Service is healthy"
			if !success {
				msg = "Service is unhealthy"
			}

			ctx.JSON(statCode, gin.H{
				"success": success,
				"message": msg,
				"data": gin.H{
					"status":   "ok",
					"database": dbStat,
					"version":  "1.0.0",
				},
			})
		})

		v1 := base.Group("/api/v1")
		{
			if cfg.AuthModule != nil {
				cfg.AuthModule.RegisterRoutes(v1)
			}
			if cfg.UserModule != nil {
				cfg.UserModule.RegisterRoutes(v1)
			}
		}

		r.Static("/public", "./storage")

		setupPrometheus(cfg.Env, r)

		if cfg.Env.App.Mode == "dev" {
			setupSwagger(base, cfg.Env)
			pprof.Register(r)
		}
	}

	return r
}
