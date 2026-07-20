package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
)

const resendEmailEndpoint = "https://api.resend.com/emails"

type ResendConfirmationEmailSender struct {
	apiKey     string
	from       string
	templateID string
	endpoint   string
	client     *http.Client
}

func NewResendConfirmationEmailSender(apiKey string, from string) ResendConfirmationEmailSender {
	return ResendConfirmationEmailSender{
		apiKey:   strings.TrimSpace(apiKey),
		from:     strings.TrimSpace(from),
		endpoint: resendEmailEndpoint,
		client:   http.DefaultClient,
	}
}

func NewResendConfirmationEmailSenderWithTemplate(apiKey string, from string, templateID string) ResendConfirmationEmailSender {
	return ResendConfirmationEmailSender{
		apiKey:     strings.TrimSpace(apiKey),
		from:       strings.TrimSpace(from),
		templateID: strings.TrimSpace(templateID),
		endpoint:   resendEmailEndpoint,
		client:     http.DefaultClient,
	}
}

func newTestResendConfirmationEmailSender(apiKey string, from string, templateID string, endpoint string, client *http.Client) ResendConfirmationEmailSender {
	return ResendConfirmationEmailSender{
		apiKey:     strings.TrimSpace(apiKey),
		from:       strings.TrimSpace(from),
		templateID: strings.TrimSpace(templateID),
		endpoint:   endpoint,
		client:     client,
	}
}

func (sender ResendConfirmationEmailSender) SendConfirmation(ctx context.Context, email ConfirmationEmail) error {
	if sender.apiKey == "" {
		return fmt.Errorf("resend api key is required")
	}

	if sender.from == "" {
		return fmt.Errorf("email from address is required")
	}

	payload := sender.emailRequest(email)

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode resend email request: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, sender.endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create resend email request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+sender.apiKey)
	request.Header.Set("Content-Type", "application/json")

	client := sender.client
	if client == nil {
		client = http.DefaultClient
	}

	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("send resend email request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		responseBody, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return fmt.Errorf("resend email request failed with status %d: %s", response.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	return nil
}

func (sender ResendConfirmationEmailSender) emailRequest(email ConfirmationEmail) resendEmailRequest {
	request := resendEmailRequest{
		From:    sender.from,
		To:      []string{email.To},
		Subject: "Wedding confirmation received",
	}

	if sender.templateID != "" {
		request.Template = &resendEmailTemplate{
			ID: sender.templateID,
			Variables: map[string]string{
				"NAME": email.Name,
			},
		}
		return request
	}

	request.Text = confirmationEmailText(email)
	request.HTML = confirmationEmailHTML(email)
	return request
}

type resendEmailRequest struct {
	From     string               `json:"from"`
	To       []string             `json:"to"`
	Subject  string               `json:"subject"`
	Text     string               `json:"text,omitempty"`
	HTML     string               `json:"html,omitempty"`
	Template *resendEmailTemplate `json:"template,omitempty"`
}

type resendEmailTemplate struct {
	ID        string            `json:"id"`
	Variables map[string]string `json:"variables"`
}

func confirmationEmailText(email ConfirmationEmail) string {
	status := "confirmed"
	if !email.Confirmation {
		status = "not confirmed"
	}

	return fmt.Sprintf("Hi %s,\n\nWe received your wedding RSVP as %s.\n\nThank you.", email.Name, status)
}

func confirmationEmailHTML(email ConfirmationEmail) string {
	status := "confirmed"
	if !email.Confirmation {
		status = "not confirmed"
	}

	return fmt.Sprintf(
		"<p>Hi %s,</p><p>We received your wedding RSVP as <strong>%s</strong>.</p><p>Thank you.</p>",
		html.EscapeString(email.Name),
		html.EscapeString(status),
	)
}
