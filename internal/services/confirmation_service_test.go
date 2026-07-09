package services

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestConfirmationServiceAcceptsValidConfirmationsAndPrintsThem(t *testing.T) {
	var output bytes.Buffer
	service := NewConfirmationService(&output)
	yes := true
	no := false

	result, err := service.Create(context.Background(), []ConfirmationRequest{
		{Name: "Maria Silva", Confirm: &yes, Email: "maria@example.com"},
		{Name: "Joao Silva", Confirm: &no, Email: ""},
	})
	if err != nil {
		t.Fatalf("expected no validation error, got %v", err)
	}

	if result.Count != 2 {
		t.Fatalf("expected count 2, got %d", result.Count)
	}

	printed := output.String()
	for _, expected := range []string{"Maria Silva", "true", "maria@example.com", "Joao Silva", "false"} {
		if !strings.Contains(printed, expected) {
			t.Fatalf("expected printed output to contain %q, got %q", expected, printed)
		}
	}
}

func TestConfirmationServiceRejectsInvalidConfirmations(t *testing.T) {
	var output bytes.Buffer
	service := NewConfirmationService(&output)
	yes := true

	_, err := service.Create(context.Background(), []ConfirmationRequest{
		{Name: "   ", Confirm: &yes, Email: "maria@example.com"},
		{Name: "Joao Silva", Email: "not-valid"},
	})
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationError, ok := err.(ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if len(validationError.Details) != 3 {
		t.Fatalf("expected 3 validation details, got %d", len(validationError.Details))
	}

	if output.Len() != 0 {
		t.Fatalf("expected no output for invalid confirmations, got %q", output.String())
	}
}
