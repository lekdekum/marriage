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

	if allowOrigin := response.Header.Get("Access-Control-Allow-Origin"); allowOrigin != "*" {
		t.Fatalf("expected CORS allow origin *, got %q", allowOrigin)
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

func TestConfirmationRouteAcceptsCorsPreflight(t *testing.T) {
	server := httptest.NewServer(NewRouter())
	defer server.Close()

	request, err := http.NewRequest(http.MethodOptions, server.URL+"/confirmation", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	request.Header.Set("Origin", "http://localhost:5173")
	request.Header.Set("Access-Control-Request-Method", http.MethodPost)
	request.Header.Set("Access-Control-Request-Headers", "content-type")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("OPTIONS /confirmation failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, response.StatusCode)
	}

	if allowOrigin := response.Header.Get("Access-Control-Allow-Origin"); allowOrigin != "*" {
		t.Fatalf("expected CORS allow origin *, got %q", allowOrigin)
	}

	if allowMethods := response.Header.Get("Access-Control-Allow-Methods"); allowMethods != "GET, POST, OPTIONS" {
		t.Fatalf("expected allowed methods, got %q", allowMethods)
	}

	if allowHeaders := response.Header.Get("Access-Control-Allow-Headers"); allowHeaders != "Content-Type" {
		t.Fatalf("expected allowed headers, got %q", allowHeaders)
	}
}
