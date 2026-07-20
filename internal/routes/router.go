package routes

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"marriage/internal/controllers"
	"marriage/internal/response"
	"marriage/internal/services"
)

func NewRouter(confirmationService services.ConfirmationService, adminToken string) http.Handler {
	mux := http.NewServeMux()

	healthController := controllers.NewHealthController()
	mux.HandleFunc("GET /health", healthController.Show)

	confirmationController := controllers.NewConfirmationController(confirmationService)
	mux.HandleFunc("POST /confirmation", confirmationController.Create)
	mux.HandleFunc("GET /admin/confirmations", withAdminToken(adminToken, confirmationController.List))
	mux.HandleFunc("PATCH /admin/confirmations", withAdminToken(adminToken, confirmationController.Update))
	mux.HandleFunc("DELETE /admin/confirmations", withAdminToken(adminToken, confirmationController.Delete))

	return withCORS(mux)
}

func withAdminToken(adminToken string, handler http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if adminToken == "" {
			response.JSON(writer, http.StatusServiceUnavailable, response.Error{
				Error: "admin token is not configured",
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
		if token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(adminToken)) != 1 {
			response.JSON(writer, http.StatusUnauthorized, response.Error{
				Error: http.StatusText(http.StatusUnauthorized),
			})
			return
		}

		handler(writer, request)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

		if request.Method == http.MethodOptions {
			writer.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(writer, request)
	})
}
