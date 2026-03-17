package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"phlcode.club/profile-service/db"
	"phlcode.club/profile-service/handlers"
	"phlcode.club/profile-service/logger"
	"phlcode.club/profile-service/telemetry"

	_ "github.com/lib/pq"
)

func main() {
	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	l := logger.GetLogger()
	otelShutdown, err := telemetry.SetupOTelSDK(ctx)
	if err != nil {
		l.Error("unable to bootstrap otel sdk", "error", err)
		os.Exit(66)
	}
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()
	db, err := db.InitDB()
	defer db.Close()
	if err != nil {
		l.Error("database connection refused", "error", err)
		os.Exit(66)
	}

	h := handlers.NewSession(db)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handlers.ValidateUser(h.GetProfile))
	mux.HandleFunc("PUT /", handlers.ValidateUser(h.UpdateProfile))
	handler := otelhttp.NewHandler(mux, "/")
	server := &http.Server{
		Addr:         ":8080",
		BaseContext:  func(net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      handler,
	}
	l.Info("Server is running on port 8080")
	srvErr := make(chan error, 1)
	go func() {
		srvErr <- server.ListenAndServe()
	}()
	// Wait for interruption.
	select {
	case err = <-srvErr:
		// Error when starting HTTP server.
		l.Error("Error starting server", "error", err)
	case <-ctx.Done():
		// Wait for first CTRL+C.
		// Stop receiving signal notifications as soon as possible.
		stop()
	}

	// When Shutdown is called, ListenAndServe immediately returns ErrServerClosed.
	err = server.Shutdown(context.Background())
}
