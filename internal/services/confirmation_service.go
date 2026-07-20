package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"marriage/internal/repositories"
)

type ConfirmationRepository interface {
	Create(ctx context.Context, confirmations []repositories.Confirmation) error
}

type ConfirmationService struct {
	repository ConfirmationRepository
}

type ConfirmationRequest struct {
	Name    string `json:"name"`
	Confirm *bool  `json:"confirm"`
	Email   string `json:"email"`
}

type ConfirmationResult struct {
	Count int `json:"count"`
}

type ValidationError struct {
	Details []string
}

func (err ValidationError) Error() string {
	return "validation failed"
}

func NewConfirmationService(repository ConfirmationRepository) ConfirmationService {
	return ConfirmationService{
		repository: repository,
	}
}

func (service ConfirmationService) Create(ctx context.Context, confirmations []ConfirmationRequest) (ConfirmationResult, error) {
	if len(confirmations) == 0 {
		return ConfirmationResult{}, ValidationError{
			Details: []string{"body must include at least one confirmation"},
		}
	}

	var details []string
	for index, confirmation := range confirmations {
		if strings.TrimSpace(confirmation.Name) == "" {
			details = append(details, fmt.Sprintf("item %d: name must be a non-empty string", index))
		}

		if confirmation.Confirm == nil {
			details = append(details, fmt.Sprintf("item %d: confirm must be a boolean", index))
		}

		if confirmation.Email != "" && !isValidEmail(confirmation.Email) {
			details = append(details, fmt.Sprintf("item %d: email must be empty or a valid email address", index))
		}
	}

	if len(details) > 0 {
		return ConfirmationResult{}, ValidationError{Details: details}
	}

	records := make([]repositories.Confirmation, 0, len(confirmations))
	now := time.Now().UTC()
	for _, confirmation := range confirmations {
		id, err := newUUID()
		if err != nil {
			return ConfirmationResult{}, fmt.Errorf("generate confirmation id: %w", err)
		}

		var email *string
		if confirmation.Email != "" {
			trimmedEmail := strings.TrimSpace(confirmation.Email)
			email = &trimmedEmail
		}

		records = append(records, repositories.Confirmation{
			ID:           id,
			Name:         strings.TrimSpace(confirmation.Name),
			Confirmation: *confirmation.Confirm,
			Email:        email,
			Timestamp:    now,
		})
	}

	if service.repository == nil {
		return ConfirmationResult{}, fmt.Errorf("confirmation repository is not configured")
	}

	if err := service.repository.Create(ctx, records); err != nil {
		return ConfirmationResult{}, fmt.Errorf("save confirmations: %w", err)
	}

	return ConfirmationResult{
		Count: len(confirmations),
	}, nil
}

func isValidEmail(email string) bool {
	address, err := mail.ParseAddress(email)
	return err == nil && address.Address == email
}

func newUUID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80

	return fmt.Sprintf(
		"%x-%x-%x-%x-%x",
		bytes[0:4],
		bytes[4:6],
		bytes[6:8],
		bytes[8:10],
		bytes[10:16],
	), nil
}
