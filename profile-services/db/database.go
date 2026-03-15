package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/lib/pq"
)

const (
	host   = "hostname"
	port   = "5431"
	user   = "postgres"
	dbName = "profile-service"
)

var (
	db  *sql.DB
	err error
)

func InitDB() (*sql.DB, error) {
	psConfig := fmt.Sprintf("host=%s port=%s user=%s "+"dbname=%s sslmode=disable", host, port, user, dbName)

	db, err = sql.Open("postgres", psConfig)
	if err != nil {
		slog.Error("Unable to validate database configutation: %v", err)
		return nil, err
	}

	defer db.Close()

	if err = db.Ping(); err != nil {
		slog.Error("unable to ping db: %v", err)
		return nil, err
	}

	_, err = db.ExecContext(
		context.Background(),
		`
		CREATE TABLE IF NOT EXISTS Users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) NOT NULL,
			bio VARCHAR(250) NOT NULL,
			location VARCHAR(250),
		);
		`,
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}
