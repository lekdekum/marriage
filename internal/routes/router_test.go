package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthRoute(t *testing.T) {
	server := httptest.NewServer(NewRouter())
	defer server.Close()

	response, err := http.Get(server.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.StatusCode)
	}

	if contentType := response.Header.Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected content type application/json, got %q", contentType)
	}
}

func TestConfirmationRouteAcceptsValidPayload(t *testing.T) {
	server := httptest.NewServer(NewRouter())
	defer server.Close()

	body := []byte(`[
		{"name":"Maria Silva","confirm":true,"email":"maria@example.com"},
		{"name":"Joao Silva","confirm":false,"email":""}
	]`)

	response, err := http.Post(server.URL+"/confirmation", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /confirmation failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, response.StatusCode)
	}

	var responseBody struct {
		Count int `json:"count"`
	}
	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if responseBody.Count != 2 {
		t.Fatalf("expected count 2, got %d", responseBody.Count)
	}
}

func TestConfirmationRouteRejectsInvalidPayload(t *testing.T) {
	server := httptest.NewServer(NewRouter())
	defer server.Close()

	body := []byte(`[
		{"name":"","confirm":true,"email":"maria@example.com"},
		{"name":"Joao Silva","email":"not-valid"}
	]`)

	response, err := http.Post(server.URL+"/confirmation", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /confirmation failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.StatusCode)
	}

	var responseBody struct {
		Error   string   `json:"error"`
		Details []string `json:"details"`
	}
	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if responseBody.Error == "" {
		t.Fatal("expected validation error message")
	}

	if len(responseBody.Details) != 3 {
		t.Fatalf("expected 3 validation details, got %d", len(responseBody.Details))
	}
}
