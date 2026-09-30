package main

import (
	"context"
	"fmt"
	"github/OfrenDialsa/go-gin-starter/cmd/api"
	"github/OfrenDialsa/go-gin-starter/config"
	"github/OfrenDialsa/go-gin-starter/database"
	"github/OfrenDialsa/go-gin-starter/internal/infra/metrics"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

// @title Go Backend starter API
// @version 1.0
// @description Api starter by Nerodev
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	env, err := config.LoadEnv()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
		panic(err)
	}

	wrapDB := database.InitDB(env)

	config.InitLogger(env)
	config.InitSwagger(env)

	setup, err := api.Init(env, wrapDB)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to initialize services")
		panic(err)
	}

	defer func() {
		if setup != nil && setup.WrapDB != nil {
			if setup.WrapDB.Database != nil && setup.WrapDB.Database.Conn != nil {
				setup.WrapDB.Database.Conn.Close()
				log.Info().Msg("Database connection closed")
			}
		}
	}()

	log.Info().Msg("Initializing Prometheus Metrics...")
	metrics.Init()

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s", env.App.Host),
		Handler:      setup.Router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Msgf("Server starting on %s", env.App.Host)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exiting")
}
