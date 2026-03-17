package logger

import (
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
)

var logger = otelslog.NewLogger("profile-service", otelslog.WithSource(true))

func GetLogger() *slog.Logger {
	return logger
}
