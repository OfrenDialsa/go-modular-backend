package config

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const AppModeDev = "dev"
const AppModePreview = "pre"
const AppModeProduction = "prod"

func InitLogger(env *EnvironmentVariable) {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	switch env.App.Mode {
	case AppModePreview:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case AppModeProduction:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	}
}
