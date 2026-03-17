package db

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"phlcode.club/profile-service/logger"

	_ "github.com/lib/pq"
)

var (
	db  *sql.DB
	err error
)

func InitDB() (*sql.DB, error) {
	l := logger.GetLogger()
	psConfig := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		"5432",
		"phlcodeclub",
		"super-secret-password",
		"otel",
	)

	db, err = otelsql.Open("postgres", psConfig,
		otelsql.WithAttributes(semconv.DBSystemPostgreSQL),
		otelsql.WithDBName("otel"))
	if err != nil {
		l.Error("Unable to validate database configutation: %v", "error", err)
		return nil, err
	}

	if err = db.Ping(); err != nil {
		l.Error("unable to ping db: %v", "error", err)
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	return db, nil
}
