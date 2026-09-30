package database

import (
	"github/OfrenDialsa/go-gin-starter/config"
	"github/OfrenDialsa/go-gin-starter/database/postgres"
)

type WrapDB struct {
	Database *postgres.WrapDB
}

func InitDB(env *config.EnvironmentVariable) *WrapDB {
	database := postgres.InitDatabase(env)

	return &WrapDB{
		Database: database,
	}
}
