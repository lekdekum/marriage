package routes

import (
	"net/http"
	"strings"

	"marriage/internal/controllers"
	"marriage/internal/response"
	"marriage/internal/services"
)

func NewRouter(confirmationService services.ConfirmationService, authService services.AuthService, allowedOrigin string) http.Handler {
	mux := http.NewServeMux()

	healthController := controllers.NewHealthController()
	mux.HandleFunc("GET /health", healthController.Show)

	authController := controllers.NewAuthController(authService)
	mux.HandleFunc("POST /login", authController.Login)

	confirmationController := controllers.NewConfirmationController(confirmationService)
	mux.HandleFunc("POST /confirmation", confirmationController.Create)
	mux.HandleFunc("GET /admin/confirmations", withAdminAuth(authService, confirmationController.List))
	mux.HandleFunc("PATCH /admin/confirmations", withAdminAuth(authService, confirmationController.Update))
	mux.HandleFunc("DELETE /admin/confirmations", withAdminAuth(authService, confirmationController.Delete))

	return withCORS(mux, allowedOrigin)
}

func withAdminAuth(authService services.AuthService, handler http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if !authService.Configured() {
			response.JSON(writer, http.StatusServiceUnavailable, response.Error{
				Error: "admin auth is not configured",
			})
			return
		}

		authorization := request.Header.Get("Authorization")
		if !strings.HasPrefix(authorization, "Bearer ") {
			response.JSON(writer, http.StatusUnauthorized, response.Error{
				Error: http.StatusText(http.StatusUnauthorized),
			})
			return
		}

		token := strings.TrimPrefix(authorization, "Bearer ")
		if token == "" || authService.ValidateToken(token) != nil {
			response.JSON(writer, http.StatusUnauthorized, response.Error{
				Error: http.StatusText(http.StatusUnauthorized),
			})
			return
		}

		handler(writer, request)
	}
}

func withCORS(next http.Handler, allowedOrigin string) http.Handler {
	if allowedOrigin == "" {
		allowedOrigin = "*"
	}

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

		if request.Method == http.MethodOptions {
			writer.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(writer, request)
	})
}
