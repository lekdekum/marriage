package routes

import (
	"net/http"

	"marriage/internal/controllers"
	"marriage/internal/services"
)

func NewRouter(confirmationService services.ConfirmationService) http.Handler {
	mux := http.NewServeMux()

	healthController := controllers.NewHealthController()
	mux.HandleFunc("GET /health", healthController.Show)

	confirmationController := controllers.NewConfirmationController(confirmationService)
	mux.HandleFunc("POST /confirmation", confirmationController.Create)

	return withCORS(mux)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if request.Method == http.MethodOptions {
			writer.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(writer, request)
	})
}
