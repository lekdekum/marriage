package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"marriage/internal/repositories"
)

type fakeConfirmationRepository struct {
	confirmations []repositories.Confirmation
}

func (repository *fakeConfirmationRepository) Create(ctx context.Context, confirmations []repositories.Confirmation) error {
	repository.confirmations = append(repository.confirmations, confirmations...)
	return nil
}

func (repository *fakeConfirmationRepository) List(ctx context.Context) ([]repositories.Confirmation, error) {
	return repository.confirmations, nil
}

func (repository *fakeConfirmationRepository) Update(ctx context.Context, confirmation repositories.Confirmation) (bool, error) {
	for index := range repository.confirmations {
		if repository.confirmations[index].ID == confirmation.ID {
			repository.confirmations[index].Name = confirmation.Name
			repository.confirmations[index].Confirmation = confirmation.Confirmation
			repository.confirmations[index].Email = confirmation.Email
			return true, nil
		}
	}

	return false, nil
}

func (repository *fakeConfirmationRepository) Delete(ctx context.Context, id string) (bool, error) {
	for index := range repository.confirmations {
		if repository.confirmations[index].ID == id {
			repository.confirmations = append(repository.confirmations[:index], repository.confirmations[index+1:]...)
			return true, nil
		}
	}

	return false, nil
}

func (repository *fakeConfirmationRepository) HasEmailSent(ctx context.Context, email string) (bool, error) {
	for _, confirmation := range repository.confirmations {
		if confirmation.Email != nil && *confirmation.Email == email && confirmation.EmailSentAt != nil {
			return true, nil
		}
	}

	return false, nil
}

func (repository *fakeConfirmationRepository) MarkEmailSent(ctx context.Context, id string, sentAt time.Time) error {
	for index := range repository.confirmations {
		if repository.confirmations[index].ID == id {
			repository.confirmations[index].EmailSentAt = &sentAt
			return nil
		}
	}

	return nil
}

type fakeConfirmationEmailSender struct {
	emails []ConfirmationEmail
	err    error
}

func (sender *fakeConfirmationEmailSender) SendConfirmation(ctx context.Context, email ConfirmationEmail) error {
	sender.emails = append(sender.emails, email)
	return sender.err
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

func TestConfirmationServiceDoesNotSendEmailWhenFireEmailIsOmitted(t *testing.T) {
	repository := &fakeConfirmationRepository{}
	sender := &fakeConfirmationEmailSender{}
	service := NewConfirmationService(repository, sender)
	yes := true

	result, err := service.Create(context.Background(), []ConfirmationRequest{
		{Name: "Maria Silva", Confirm: &yes, Email: "maria@example.com"},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Count != 1 {
		t.Fatalf("expected count 1, got %d", result.Count)
	}

	if len(sender.emails) != 0 {
		t.Fatalf("expected no emails to be sent, got %d", len(sender.emails))
	}
}

func TestConfirmationServiceDoesNotSendEmailWhenFireEmailIsFalse(t *testing.T) {
	repository := &fakeConfirmationRepository{}
	sender := &fakeConfirmationEmailSender{}
	service := NewConfirmationService(repository, sender)
	yes := true

	result, err := service.Create(context.Background(), []ConfirmationRequest{
		{Name: "Maria Silva", Confirm: &yes, Email: "maria@example.com", FireEmail: false},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Count != 1 {
		t.Fatalf("expected count 1, got %d", result.Count)
	}

	if len(sender.emails) != 0 {
		t.Fatalf("expected no emails to be sent, got %d", len(sender.emails))
	}
}

func TestConfirmationServiceDoesNotSendEmailWhenEmailIsEmpty(t *testing.T) {
	repository := &fakeConfirmationRepository{}
	sender := &fakeConfirmationEmailSender{}
	service := NewConfirmationService(repository, sender)
	yes := true

	result, err := service.Create(context.Background(), []ConfirmationRequest{
		{Name: "Maria Silva", Confirm: &yes, Email: "", FireEmail: true},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Count != 1 {
		t.Fatalf("expected count 1, got %d", result.Count)
	}

	if len(sender.emails) != 0 {
		t.Fatalf("expected no emails to be sent, got %d", len(sender.emails))
	}
}

func TestConfirmationServiceSendsEmailOnceForNewEmail(t *testing.T) {
	repository := &fakeConfirmationRepository{}
	sender := &fakeConfirmationEmailSender{}
	service := NewConfirmationService(repository, sender)
	yes := true

	result, err := service.Create(context.Background(), []ConfirmationRequest{
		{Name: "Maria Silva", Confirm: &yes, Email: "maria@example.com", FireEmail: true},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Count != 1 {
		t.Fatalf("expected count 1, got %d", result.Count)
	}

	if len(sender.emails) != 1 {
		t.Fatalf("expected 1 email to be sent, got %d", len(sender.emails))
	}

	if sender.emails[0].To != "maria@example.com" {
		t.Fatalf("expected email recipient, got %q", sender.emails[0].To)
	}

	if repository.confirmations[0].EmailSentAt == nil {
		t.Fatal("expected confirmation to be marked as emailed")
	}

	result, err = service.Create(context.Background(), []ConfirmationRequest{
		{Name: "Maria Silva", Confirm: &yes, Email: "maria@example.com", FireEmail: true},
	})
	if err != nil {
		t.Fatalf("expected no error on duplicate email, got %v", err)
	}

	if result.Count != 1 {
		t.Fatalf("expected count 1, got %d", result.Count)
	}

	if len(sender.emails) != 1 {
		t.Fatalf("expected no duplicate email, got %d emails", len(sender.emails))
	}
}

func TestConfirmationServiceContinuesWhenEmailSendFails(t *testing.T) {
	repository := &fakeConfirmationRepository{}
	sender := &fakeConfirmationEmailSender{err: errors.New("email provider failed")}
	service := NewConfirmationService(repository, sender)
	yes := true

	result, err := service.Create(context.Background(), []ConfirmationRequest{
		{Name: "Maria Silva", Confirm: &yes, Email: "maria@example.com", FireEmail: true},
	})
	if err != nil {
		t.Fatalf("expected confirmation flow to continue, got %v", err)
	}

	if result.Count != 1 {
		t.Fatalf("expected count 1, got %d", result.Count)
	}

	if len(repository.confirmations) != 1 {
		t.Fatalf("expected confirmation to be saved, got %d", len(repository.confirmations))
	}

	if len(sender.emails) != 1 {
		t.Fatalf("expected exactly 1 email attempt, got %d", len(sender.emails))
	}

	if repository.confirmations[0].EmailSentAt != nil {
		t.Fatal("expected failed email not to be marked as sent")
	}
}

func TestConfirmationServiceContinuesWhenEmailSenderIsNotConfigured(t *testing.T) {
	repository := &fakeConfirmationRepository{}
	service := NewConfirmationService(repository)
	yes := true

	result, err := service.Create(context.Background(), []ConfirmationRequest{
		{Name: "Maria Silva", Confirm: &yes, Email: "maria@example.com", FireEmail: true},
	})
	if err != nil {
		t.Fatalf("expected confirmation flow to continue, got %v", err)
	}

	if result.Count != 1 {
		t.Fatalf("expected count 1, got %d", result.Count)
	}

	if len(repository.confirmations) != 1 {
		t.Fatalf("expected confirmation to be saved, got %d", len(repository.confirmations))
	}

	if repository.confirmations[0].EmailSentAt != nil {
		t.Fatal("expected missing email sender not to be marked as sent")
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

func TestConfirmationServiceListsConfirmations(t *testing.T) {
	email := "maria@example.com"
	repository := &fakeConfirmationRepository{
		confirmations: []repositories.Confirmation{
			{ID: "72f22b5a-1489-4c38-a74e-f6611a9c7042", Name: "Maria Silva", Confirmation: true, Email: &email},
		},
	}
	service := NewConfirmationService(repository)

	confirmations, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(confirmations) != 1 {
		t.Fatalf("expected 1 confirmation, got %d", len(confirmations))
	}

	if confirmations[0].ID != "72f22b5a-1489-4c38-a74e-f6611a9c7042" {
		t.Fatalf("expected confirmation id, got %q", confirmations[0].ID)
	}
}

func TestConfirmationServiceUpdatesConfirmation(t *testing.T) {
	repository := &fakeConfirmationRepository{
		confirmations: []repositories.Confirmation{
			{ID: "72f22b5a-1489-4c38-a74e-f6611a9c7042", Name: "Maria Silva", Confirmation: true},
		},
	}
	service := NewConfirmationService(repository)
	no := false

	confirmation, err := service.Update(context.Background(), UpdateConfirmationRequest{
		ID:           "72f22b5a-1489-4c38-a74e-f6611a9c7042",
		Name:         "Maria Santos",
		Confirmation: &no,
		Email:        "",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if confirmation.Name != "Maria Santos" {
		t.Fatalf("expected updated name, got %q", confirmation.Name)
	}

	if confirmation.Confirmation {
		t.Fatal("expected updated confirmation to be false")
	}
}

func TestConfirmationServiceRejectsInvalidUpdate(t *testing.T) {
	repository := &fakeConfirmationRepository{}
	service := NewConfirmationService(repository)

	_, err := service.Update(context.Background(), UpdateConfirmationRequest{
		ID:    "",
		Name:  "",
		Email: "not-valid",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationError, ok := err.(ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if len(validationError.Details) != 4 {
		t.Fatalf("expected 4 validation details, got %d", len(validationError.Details))
	}
}

func TestConfirmationServiceDeletesConfirmation(t *testing.T) {
	repository := &fakeConfirmationRepository{
		confirmations: []repositories.Confirmation{
			{ID: "72f22b5a-1489-4c38-a74e-f6611a9c7042", Name: "Maria Silva", Confirmation: true},
		},
	}
	service := NewConfirmationService(repository)

	err := service.Delete(context.Background(), DeleteConfirmationRequest{
		ID: "72f22b5a-1489-4c38-a74e-f6611a9c7042",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(repository.confirmations) != 0 {
		t.Fatalf("expected confirmation to be deleted, got %d records", len(repository.confirmations))
	}
}
