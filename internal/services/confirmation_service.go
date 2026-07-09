package services

import (
	"context"
	"fmt"
	"io"
	"net/mail"
	"os"
	"strings"
)

type ConfirmationService struct {
	output io.Writer
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

func NewConfirmationService(output ...io.Writer) ConfirmationService {
	writer := io.Writer(os.Stdout)
	if len(output) > 0 && output[0] != nil {
		writer = output[0]
	}

	return ConfirmationService{
		output: writer,
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

	for _, confirmation := range confirmations {
		fmt.Fprintf(
			service.output,
			"name=%q confirm=%t email=%q\n",
			strings.TrimSpace(confirmation.Name),
			*confirmation.Confirm,
			confirmation.Email,
		)
	}

	return ConfirmationResult{
		Count: len(confirmations),
	}, nil
}

func isValidEmail(email string) bool {
	address, err := mail.ParseAddress(email)
	return err == nil && address.Address == email
}
