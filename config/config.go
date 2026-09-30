package config

import (
	"fmt"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

const EnvFolder = "env"
const EnvSecretFilename = ".secret.env"
const EnvFilename = ".env"

func LoadEnv() (env *EnvironmentVariable, err error) {
	envFile := fmt.Sprintf("%s/%s", EnvFolder, EnvFilename)

	v := viper.New()
	if _, err := os.Stat(envFile); err == nil {
		v.SetConfigFile(envFile)
		if err := v.ReadInConfig(); err != nil {
			log.Printf("Error reading .env file: %v", err)
			panic(err)
		}
		log.Info().Msg(".env file loaded successfully")
	} else {
		v.AutomaticEnv()
		log.Info().Msg(".env file not found, skipping loading")
	}

	err = v.Unmarshal(&env)
	if err != nil {
		log.Error().Err(err).Msg("viper error unmarshal config")
	}

	err = env.validateRequiredValue()
	if err != nil {
		log.Error().Err(err).Msg("Some required configuration are missing")
		return nil, err
	}

	log.Info().Msg("Env Loaded")
	return
}

type EnvironmentVariable struct {
	App struct {
		Host string `mapstructure:"HOST" validate:"required"`
		Mode string `mapstructure:"MODE" validate:"required"`
	} `mapstructure:"APP"`
	Database struct {
		Timeout  time.Duration `mapstructure:"TIMEOUT" validate:"required"`
		Postgres struct {
			Host     string `mapstructure:"HOST" validate:"required"`
			Port     string `mapstructure:"PORT" validate:"required"`
			Name     string `mapstructure:"NAME" validate:"required"`
			Username string `mapstructure:"USERNAME" validate:"required"`
			Password string `mapstructure:"PASSWORD" validate:"required"`
		} `mapstructure:"POSTGRES"`
		Redis struct {
			IsEnabled bool   `mapstructure:"IS_ENABLED"`
			Host      string `mapstructure:"HOST"`
			Port      int    `mapstructure:"PORT"`
			Password  string `mapstructure:"PASSWORD"`
		} `mapstructure:"REDIS"`
	} `mapstructure:"DATABASE"`
	Swagger struct {
		BasePath    string `mapstructure:"BASE_PATH"`
		Host        string `mapstructure:"HOST"`
		Title       string `mapstructure:"TITLE"`
		Description string `mapstructure:"DESCRIPTION"`
		Version     string `mapstructure:"VERSION"`
	} `mapstructure:"SWAGGER"`
	JWT struct {
		SecretKey struct {
			Access  string `mapstructure:"ACCESS"`
			Refresh string `mapstructure:"REFRESH"`
		} `mapstructure:"SECRET_KEY"`
		Token struct {
			AccessLifeTime time.Duration `mapstructure:"ACCESS_LIFE_TIME"`
		} `mapstructure:"TOKEN"`
	} `mapstructure:"JWT"`
	External struct {
		ResetPasswordURL string `mapstructure:"RESET_PASSWORD_URL"`
		VerifyEmailURL   string `mapstructure:"VERIFY_EMAIL_URL"`
		FrontendURL      string `mapstructure:"FRONTEND_URL"`
	} `mapstructure:"EXTERNAL"`
	Mail struct {
		From     string `mapstructure:"FROM"`
		FromName string `mapstructure:"FROM_NAME"`
		SMTP     struct {
			Host     string `mapstructure:"HOST"`
			Port     int    `mapstructure:"PORT"`
			User     string `mapstructure:"USER"`
			Password string `mapstructure:"PASSWORD"`
		} `mapstructure:"SMTP"`
	} `mapstructure:"MAIL"`
	Storage struct {
		Type  string `mapstructure:"TYPE" validate:"required"`
		Local struct {
			BasePath  string `mapstructure:"BASE_PATH"`
			PublicUrl string `mapstructure:"PUBLIC_URL"`
		} `mapstructure:"LOCAL"`
		S3 struct {
			Endpoint     string `mapstructure:"ENDPOINT"`
			Region       string `mapstructure:"REGION"`
			Bucket       string `mapstructure:"BUCKET"`
			PublicUrl    string `mapstructure:"PUBLIC_URL"`
			AccessKey    string `mapstructure:"ACCESS_KEY"`
			SecretKey    string `mapstructure:"SECRET_KEY"`
			SSL          bool   `mapstructure:"SSL"`
			UsePathStyle bool   `mapstructure:"USE_PATH_STYLE"`
		} `mapstructure:"S3"`
	} `mapstructure:"STORAGE"`
	Prometheus struct {
		Username string `mapstructure:"USERNAME"`
		Password string `mapstructure:"PASSWORD"`
	} `mapstructure:"PROMETHEUS"`
}

func (e *EnvironmentVariable) validateRequiredValue() error {
	validate := validator.New()

	if err := validate.Struct(e); err != nil {
		return err
	}

	return nil
}
