package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/lk153/import-gsheet/lib/configs"
	"github.com/lk153/import-gsheet/lib/env"
)

func Open(c configs.Config) *sql.DB {
	dbEnv := configs.DbEnv{}
	if err := env.Init(c, &dbEnv); err != nil {
		log.Panic().Err(err).Msgf("Error while initializing db env")
	}

	return open(dbEnv.DbUser, dbEnv.DbPassword, dbEnv.DbHost, dbEnv.DbName, dbEnv.DbMaxConnections)
}

func open(user, pass, host, name string, maxConn int) *sql.DB {
	driverName := "mysql"
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4,utf8&parseTime=True&loc=UTC", user, pass, host, name)

	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open database connection")
	}

	db.SetMaxOpenConns(maxConn)
	db.SetMaxIdleConns(maxConn)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		if err := db.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close the db connection after a failed ping")
		}
		log.Fatal().Err(err).Msg("Failed to ping the database")
	}

	return db
}

func Close(db *sql.DB) {
	log.Info().Array("log_tags", zerolog.Arr().Str("app").Str("shutdown")).Msg("Closing database connection")
	if err := db.Close(); err != nil {
		log.Error().Err(err).Msg("Error while closing database connection")
	}
}
