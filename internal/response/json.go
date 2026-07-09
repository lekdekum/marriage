package response

import (
	"encoding/json"
	"net/http"
)

type Error struct {
	Error string `json:"error"`
}

type ValidationError struct {
	Error   string   `json:"error"`
	Details []string `json:"details"`
}

func JSON(writer http.ResponseWriter, statusCode int, payload any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)

	if err := json.NewEncoder(writer).Encode(payload); err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
