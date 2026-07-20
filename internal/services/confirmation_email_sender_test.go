package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResendConfirmationEmailSenderSendsExpectedRequest(t *testing.T) {
	var requestBody resendEmailRequest
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Fatalf("expected POST request, got %s", request.Method)
		}

		if authorization := request.Header.Get("Authorization"); authorization != "Bearer test-api-key" {
			t.Fatalf("expected bearer token, got %q", authorization)
		}

		if contentType := request.Header.Get("Content-Type"); contentType != "application/json" {
			t.Fatalf("expected JSON content type, got %q", contentType)
		}

		if err := json.NewDecoder(request.Body).Decode(&requestBody); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		writer.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := newTestResendConfirmationEmailSender(
		"test-api-key",
		"Marriage <no-reply@example.com>",
		server.URL,
		server.Client(),
	)

	err := sender.SendConfirmation(context.Background(), ConfirmationEmail{
		To:           "maria@example.com",
		Name:         "Maria Silva",
		Confirmation: true,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if requestBody.From != "Marriage <no-reply@example.com>" {
		t.Fatalf("expected from address, got %q", requestBody.From)
	}

	if len(requestBody.To) != 1 || requestBody.To[0] != "maria@example.com" {
		t.Fatalf("expected recipient, got %v", requestBody.To)
	}

	if requestBody.Subject == "" {
		t.Fatal("expected subject")
	}

	if requestBody.Text == "" {
		t.Fatal("expected text body")
	}

	if requestBody.HTML == "" {
		t.Fatal("expected html body")
	}
}

func TestResendConfirmationEmailSenderReturnsErrorForProviderFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Error(writer, "invalid api key", http.StatusForbidden)
	}))
	defer server.Close()

	sender := newTestResendConfirmationEmailSender(
		"test-api-key",
		"Marriage <no-reply@example.com>",
		server.URL,
		server.Client(),
	)

	err := sender.SendConfirmation(context.Background(), ConfirmationEmail{
		To:           "maria@example.com",
		Name:         "Maria Silva",
		Confirmation: true,
	})
	if err == nil {
		t.Fatal("expected provider failure error")
	}
}

func TestResendConfirmationEmailSenderRequiresConfiguration(t *testing.T) {
	sender := NewResendConfirmationEmailSender("", "")

	err := sender.SendConfirmation(context.Background(), ConfirmationEmail{
		To:           "maria@example.com",
		Name:         "Maria Silva",
		Confirmation: true,
	})
	if err == nil {
		t.Fatal("expected missing configuration error")
	}
}
