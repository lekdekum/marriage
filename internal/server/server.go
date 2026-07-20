package server

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"

	"marriage/internal/repositories"
	"marriage/internal/routes"
	"marriage/internal/services"
)

func Run() {
	addr := ":" + envOrDefault("PORT", "8080")
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		slog.Error("DATABASE_URL is required")
		os.Exit(1)
	}
	adminToken := os.Getenv("ADMIN_TOKEN")

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	confirmationRepository := repositories.NewPostgresConfirmationRepository(db)
	if err := confirmationRepository.Migrate(ctx); err != nil {
		slog.Error("failed to migrate database", "error", err)
		os.Exit(1)
	}

	server := &http.Server{
		Addr:         addr,
		Handler:      routes.NewRouter(services.NewConfirmationService(confirmationRepository), adminToken),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	slog.Info("starting api server", "addr", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("api server stopped unexpectedly", "error", err)
		os.Exit(1)
	}
}

func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
