package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Path-Software-Engineer/rl-simulation-control-platform/services/control-api/internal/application"
	"github.com/Path-Software-Engineer/rl-simulation-control-platform/services/control-api/internal/infrastructure/postgres"
	"github.com/Path-Software-Engineer/rl-simulation-control-platform/services/control-api/internal/infrastructure/redisstream"
	"github.com/Path-Software-Engineer/rl-simulation-control-platform/services/control-api/internal/interfaces/httpapi"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	database, err := postgres.Open(ctx, required("DATABASE_URL"))
	if err != nil {
		slog.Error("database initialization failed", "error", err)
		os.Exit(1)
	}
	defer database.Close()
	stream, err := redisstream.Open(required("REDIS_URL"))
	if err != nil {
		slog.Error("redis initialization failed", "error", err)
		os.Exit(1)
	}
	defer stream.Close()

	hub := httpapi.NewHub()
	service := application.NewService(database, stream)
	operatorToken := required("OPERATOR_TOKEN")
	if len(operatorToken) < 16 || len(operatorToken) > 128 || strings.ContainsAny(operatorToken, " \t\r\n,") {
		slog.Error("OPERATOR_TOKEN must be 16-128 protocol-safe characters")
		os.Exit(1)
	}
	dispatcher := redisstream.NewOutboxDispatcher(database, stream)
	projector := redisstream.NewEventProjector(
		database,
		stream,
		hub,
		getenv("EVENT_STREAM", "rl.events.v1"),
		getenv("PROJECTOR_CONSUMER_GROUP", "control-api-projector-v1"),
		getenv("HOSTNAME", "control-api-local"),
		getenv("DEAD_LETTER_STREAM", "rl.dead-letter.v1"),
	)
	go dispatcher.Run(ctx)
	go projector.Run(ctx)

	router := httpapi.NewRouter(
		service,
		hub,
		operatorToken,
		strings.Split(getenv("ALLOWED_ORIGINS", "http://127.0.0.1:3000,http://localhost:3000"), ","),
		getenv("OPENAPI_PATH", "/workspace/contracts/http/openapi.json"),
	)
	server := &http.Server{
		Addr:              getenv("HTTP_ADDRESS", ":8080"),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    32 << 10,
	}
	go func() {
		slog.Info("control API listening", "address", server.Addr)
		if listenErr := server.ListenAndServe(); listenErr != nil && !errors.Is(listenErr, http.ErrServerClosed) {
			slog.Error("control API stopped", "error", listenErr)
			stop()
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
	}
}

func required(name string) string {
	value := os.Getenv(name)
	if value == "" {
		slog.Error("required configuration is missing", "name", name)
		os.Exit(1)
	}
	return value
}

func getenv(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
