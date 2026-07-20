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
	List(ctx context.Context) ([]repositories.Confirmation, error)
	Update(ctx context.Context, confirmation repositories.Confirmation) (bool, error)
	Delete(ctx context.Context, id string) (bool, error)
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

type ConfirmationResponse struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Confirmation bool       `json:"confirmation"`
	Email        *string    `json:"email"`
	Timestamp    *time.Time `json:"timestamp,omitempty"`
}

type UpdateConfirmationRequest struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Confirmation *bool  `json:"confirmation"`
	Email        string `json:"email"`
}

type DeleteConfirmationRequest struct {
	ID string `json:"id"`
}

type ValidationError struct {
	Details []string
}

type NotFoundError struct{}

func (err NotFoundError) Error() string {
	return "confirmation not found"
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

func (service ConfirmationService) List(ctx context.Context) ([]ConfirmationResponse, error) {
	if service.repository == nil {
		return nil, fmt.Errorf("confirmation repository is not configured")
	}

	confirmations, err := service.repository.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list confirmations: %w", err)
	}

	responses := make([]ConfirmationResponse, 0, len(confirmations))
	for _, confirmation := range confirmations {
		timestamp := confirmation.Timestamp
		responses = append(responses, ConfirmationResponse{
			ID:           confirmation.ID,
			Name:         confirmation.Name,
			Confirmation: confirmation.Confirmation,
			Email:        confirmation.Email,
			Timestamp:    &timestamp,
		})
	}

	return responses, nil
}

func (service ConfirmationService) Update(ctx context.Context, request UpdateConfirmationRequest) (ConfirmationResponse, error) {
	var details []string
	if strings.TrimSpace(request.ID) == "" {
		details = append(details, "id must be a non-empty string")
	}

	if strings.TrimSpace(request.Name) == "" {
		details = append(details, "name must be a non-empty string")
	}

	if request.Confirmation == nil {
		details = append(details, "confirmation must be a boolean")
	}

	if request.Email != "" && !isValidEmail(request.Email) {
		details = append(details, "email must be empty or a valid email address")
	}

	if len(details) > 0 {
		return ConfirmationResponse{}, ValidationError{Details: details}
	}

	if service.repository == nil {
		return ConfirmationResponse{}, fmt.Errorf("confirmation repository is not configured")
	}

	var email *string
	if request.Email != "" {
		trimmedEmail := strings.TrimSpace(request.Email)
		email = &trimmedEmail
	}

	confirmation := repositories.Confirmation{
		ID:           strings.TrimSpace(request.ID),
		Name:         strings.TrimSpace(request.Name),
		Confirmation: *request.Confirmation,
		Email:        email,
	}

	updated, err := service.repository.Update(ctx, confirmation)
	if err != nil {
		return ConfirmationResponse{}, fmt.Errorf("update confirmation: %w", err)
	}

	if !updated {
		return ConfirmationResponse{}, NotFoundError{}
	}

	return ConfirmationResponse{
		ID:           confirmation.ID,
		Name:         confirmation.Name,
		Confirmation: confirmation.Confirmation,
		Email:        confirmation.Email,
	}, nil
}

func (service ConfirmationService) Delete(ctx context.Context, request DeleteConfirmationRequest) error {
	if strings.TrimSpace(request.ID) == "" {
		return ValidationError{Details: []string{"id must be a non-empty string"}}
	}

	if service.repository == nil {
		return fmt.Errorf("confirmation repository is not configured")
	}

	deleted, err := service.repository.Delete(ctx, strings.TrimSpace(request.ID))
	if err != nil {
		return fmt.Errorf("delete confirmation: %w", err)
	}

	if !deleted {
		return NotFoundError{}
	}

	return nil
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
