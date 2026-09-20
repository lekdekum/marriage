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
	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	emailSender := newConfirmationEmailSenderFromEnv()
	authService := services.NewAuthService(os.Getenv("ADMIN_PASSWORD_HASH"), os.Getenv("JWT_SECRET"), 24*time.Hour)

	handler, cleanup := newHandler(databaseURL, emailSender, authService, allowedOrigin)
	defer cleanup()

	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
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

func newHandler(
	databaseURL string,
	emailSender services.ConfirmationEmailSender,
	authService services.AuthService,
	allowedOrigin string,
) (http.Handler, func()) {
	if databaseURL == "" {
		return newInMemoryHandler("DATABASE_URL is not configured", nil, emailSender, authService, allowedOrigin)
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return newInMemoryHandler("failed to open database", err, emailSender, authService, allowedOrigin)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			slog.Error("failed to close database", "error", closeErr)
		}
		return newInMemoryHandler("failed to connect to database", err, emailSender, authService, allowedOrigin)
	}

	confirmationRepository := repositories.NewPostgresConfirmationRepository(db)
	if err := confirmationRepository.Migrate(ctx); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			slog.Error("failed to close database", "error", closeErr)
		}
		return newInMemoryHandler("failed to migrate database", err, emailSender, authService, allowedOrigin)
	}

	reservedGiftRepository := repositories.NewPostgresReservedGiftRepository(db)
	if err := reservedGiftRepository.Migrate(ctx); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			slog.Error("failed to close database", "error", closeErr)
		}
		return newInMemoryHandler("failed to migrate reserved gifts database", err, emailSender, authService, allowedOrigin)
	}

	return routes.NewRouter(
			services.NewConfirmationService(confirmationRepository, emailSender),
			services.NewReservedGiftService(reservedGiftRepository),
			authService,
			allowedOrigin,
		), func() {
			if err := db.Close(); err != nil {
				slog.Error("failed to close database", "error", err)
			}
		}
}

func newInMemoryHandler(
	reason string,
	err error,
	emailSender services.ConfirmationEmailSender,
	authService services.AuthService,
	allowedOrigin string,
) (http.Handler, func()) {
	if requireDatabase() {
		if err == nil {
			slog.Error(reason)
		} else {
			slog.Error(reason, "error", err)
		}
		os.Exit(1)
	}

	if err == nil {
		slog.Warn(reason + "; using in-memory repositories")
	} else {
		slog.Warn(reason+"; using in-memory repositories", "error", err)
	}

	return routes.NewRouter(
		services.NewConfirmationService(repositories.NewInMemoryConfirmationRepository(), emailSender),
		services.NewReservedGiftService(repositories.NewInMemoryReservedGiftRepository()),
		authService,
		allowedOrigin,
	), func() {}
}

func requireDatabase() bool {
	return os.Getenv("REQUIRE_DATABASE") == "true"
}

func newConfirmationEmailSenderFromEnv() services.ConfirmationEmailSender {
	provider := os.Getenv("EMAIL_PROVIDER")
	if provider == "" {
		slog.Info("confirmation email sender disabled")
		return nil
	}

	if provider != "resend" {
		slog.Error("unsupported email provider", "provider", provider)
		return nil
	}

	apiKey := os.Getenv("RESEND_API_KEY")
	from := os.Getenv("EMAIL_FROM")
	templateID := os.Getenv("RESEND_CONFIRMATION_TEMPLATE_ID")
	if apiKey == "" || from == "" {
		slog.Error("resend email sender is not configured")
		return nil
	}

	return services.NewResendConfirmationEmailSenderWithTemplate(apiKey, from, templateID)
}

func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
