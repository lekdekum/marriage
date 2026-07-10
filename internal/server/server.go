package server

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"marriage/internal/routes"
)

func Run() {
	addr := ":" + envOrDefault("PORT", "8080")

	server := &http.Server{
		Addr:         addr,
		Handler:      routes.NewRouter(),
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
