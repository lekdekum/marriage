package routes

import (
	"net/http"

	"marriage/internal/controllers"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()

	healthController := controllers.NewHealthController()
	mux.HandleFunc("GET /health", healthController.Show)

	confirmationController := controllers.NewConfirmationController()
	mux.HandleFunc("POST /confirmation", confirmationController.Create)

	return mux
}
