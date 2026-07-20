package services

import (
	"context"
	"testing"

	"marriage/internal/repositories"
)

type fakeConfirmationRepository struct {
	confirmations []repositories.Confirmation
}

func (repository *fakeConfirmationRepository) Create(ctx context.Context, confirmations []repositories.Confirmation) error {
	repository.confirmations = append(repository.confirmations, confirmations...)
	return nil
}

func TestConfirmationServiceAcceptsValidConfirmationsAndSavesThem(t *testing.T) {
	repository := &fakeConfirmationRepository{}
	service := NewConfirmationService(repository)
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

	if len(repository.confirmations) != 2 {
		t.Fatalf("expected 2 saved confirmations, got %d", len(repository.confirmations))
	}

	if repository.confirmations[0].ID == "" {
		t.Fatal("expected saved confirmation to include UUID")
	}

	if repository.confirmations[0].Name != "Maria Silva" {
		t.Fatalf("expected trimmed name, got %q", repository.confirmations[0].Name)
	}

	if !repository.confirmations[0].Confirmation {
		t.Fatal("expected first confirmation to be true")
	}

	if repository.confirmations[0].Email == nil || *repository.confirmations[0].Email != "maria@example.com" {
		t.Fatalf("expected saved email, got %v", repository.confirmations[0].Email)
	}

	if repository.confirmations[1].Email != nil {
		t.Fatalf("expected empty email to be saved as nil, got %v", repository.confirmations[1].Email)
	}

	if repository.confirmations[0].Timestamp.IsZero() {
		t.Fatal("expected saved confirmation to include timestamp")
	}
}

func TestConfirmationServiceRejectsInvalidConfirmations(t *testing.T) {
	repository := &fakeConfirmationRepository{}
	service := NewConfirmationService(repository)
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

	if len(repository.confirmations) != 0 {
		t.Fatalf("expected no saved confirmations, got %d", len(repository.confirmations))
	}
}
