package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"marriage/internal/repositories"
	"marriage/internal/services"

	"golang.org/x/crypto/bcrypt"
)

type fakeConfirmationRepository struct {
	confirmations []repositories.Confirmation
}

type fakeReservedGiftRepository struct {
	reservedGifts []repositories.ReservedGift
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

func (repository *fakeReservedGiftRepository) Create(ctx context.Context, reservedGift repositories.ReservedGift) (bool, error) {
	for _, existingReservedGift := range repository.reservedGifts {
		if existingReservedGift.ID == reservedGift.ID {
			return false, nil
		}
	}

	repository.reservedGifts = append(repository.reservedGifts, reservedGift)
	return true, nil
}

func (repository *fakeReservedGiftRepository) List(ctx context.Context) ([]repositories.ReservedGift, error) {
	return repository.reservedGifts, nil
}

func newTestRouter() http.Handler {
	return newTestRouterWithRepositories(&fakeConfirmationRepository{}, &fakeReservedGiftRepository{})
}

func newTestRouterWithRepository(repository *fakeConfirmationRepository) http.Handler {
	return newTestRouterWithRepositories(repository, &fakeReservedGiftRepository{})
}

func newTestRouterWithRepositoryAndOrigin(repository *fakeConfirmationRepository, allowedOrigin string) http.Handler {
	return newTestRouterWithRepositoriesAndOrigin(repository, &fakeReservedGiftRepository{}, allowedOrigin)
}

func newTestRouterWithRepositories(confirmationRepository *fakeConfirmationRepository, reservedGiftRepository *fakeReservedGiftRepository) http.Handler {
	return newTestRouterWithRepositoriesAndOrigin(confirmationRepository, reservedGiftRepository, "*")
}

func newTestRouterWithRepositoriesAndOrigin(
	confirmationRepository *fakeConfirmationRepository,
	reservedGiftRepository *fakeReservedGiftRepository,
	allowedOrigin string,
) http.Handler {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("test-password"), bcrypt.MinCost)
	if err != nil {
		panic(err)
	}

	authService := services.NewAuthService(string(passwordHash), "test-jwt-secret", time.Hour)
	return NewRouter(
		services.NewConfirmationService(confirmationRepository),
		services.NewReservedGiftService(reservedGiftRepository),
		authService,
		allowedOrigin,
	)
}

func loginAdmin(t *testing.T, serverURL string) string {
	t.Helper()

	loginBody := []byte(`{"password":"test-password"}`)
	loginResponse, err := http.Post(serverURL+"/login", "application/json", bytes.NewReader(loginBody))
	if err != nil {
		t.Fatalf("POST /login failed: %v", err)
	}
	defer loginResponse.Body.Close()

	var loginResponseBody struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(loginResponse.Body).Decode(&loginResponseBody); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}

	if loginResponseBody.Token == "" {
		t.Fatal("expected login token")
	}

	return loginResponseBody.Token
}

func TestHealthRoute(t *testing.T) {
	server := httptest.NewServer(newTestRouter())
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
	server := httptest.NewServer(newTestRouter())
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
	server := httptest.NewServer(newTestRouter())
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

func TestReservedGiftRouteCreatesReservedGift(t *testing.T) {
	repository := &fakeReservedGiftRepository{}
	server := httptest.NewServer(newTestRouterWithRepositories(&fakeConfirmationRepository{}, repository))
	defer server.Close()

	body := []byte(`{"id":"stand-mixer","name":"Maria Silva"}`)
	response, err := http.Post(server.URL+"/reserved_gift", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /reserved_gift failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, response.StatusCode)
	}

	if len(repository.reservedGifts) != 1 {
		t.Fatalf("expected 1 saved reserved gift, got %d", len(repository.reservedGifts))
	}

	if repository.reservedGifts[0].Timestamp.IsZero() {
		t.Fatal("expected saved reserved gift to include timestamp")
	}
}

func TestReservedGiftRouteRejectsDuplicateGiftID(t *testing.T) {
	repository := &fakeReservedGiftRepository{}
	server := httptest.NewServer(newTestRouterWithRepositories(&fakeConfirmationRepository{}, repository))
	defer server.Close()

	firstBody := []byte(`{"id":"stand-mixer","name":"Maria Silva"}`)
	firstResponse, err := http.Post(server.URL+"/reserved_gift", "application/json", bytes.NewReader(firstBody))
	if err != nil {
		t.Fatalf("first POST /reserved_gift failed: %v", err)
	}
	defer firstResponse.Body.Close()

	if firstResponse.StatusCode != http.StatusCreated {
		t.Fatalf("expected first status %d, got %d", http.StatusCreated, firstResponse.StatusCode)
	}

	secondBody := []byte(`{"id":"stand-mixer","name":"Joao Silva"}`)
	secondResponse, err := http.Post(server.URL+"/reserved_gift", "application/json", bytes.NewReader(secondBody))
	if err != nil {
		t.Fatalf("second POST /reserved_gift failed: %v", err)
	}
	defer secondResponse.Body.Close()

	if secondResponse.StatusCode != http.StatusConflict {
		t.Fatalf("expected second status %d, got %d", http.StatusConflict, secondResponse.StatusCode)
	}

	if len(repository.reservedGifts) != 1 {
		t.Fatalf("expected duplicate gift id not to create another reservation, got %d", len(repository.reservedGifts))
	}

	if repository.reservedGifts[0].Name != "Maria Silva" {
		t.Fatalf("expected original buyer to remain unchanged, got %q", repository.reservedGifts[0].Name)
	}
}

func TestReservedGiftRouteListsReservedGifts(t *testing.T) {
	repository := &fakeReservedGiftRepository{
		reservedGifts: []repositories.ReservedGift{
			{ID: "stand-mixer", Name: "Maria Silva", Timestamp: time.Now().UTC()},
		},
	}
	server := httptest.NewServer(newTestRouterWithRepositories(&fakeConfirmationRepository{}, repository))
	defer server.Close()

	response, err := http.Get(server.URL + "/reserved_gifts")
	if err != nil {
		t.Fatalf("GET /reserved_gifts failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.StatusCode)
	}

	var responseBody []services.ReservedGiftResponse
	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if len(responseBody) != 1 {
		t.Fatalf("expected 1 reserved gift, got %d", len(responseBody))
	}

	if responseBody[0].ID != "stand-mixer" {
		t.Fatalf("expected reserved gift id, got %q", responseBody[0].ID)
	}

	if responseBody[0].Name != "Maria Silva" {
		t.Fatalf("expected reserved gift name, got %q", responseBody[0].Name)
	}
}

func TestConfirmationRouteAcceptsCorsPreflight(t *testing.T) {
	server := httptest.NewServer(newTestRouter())
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

	if allowMethods := response.Header.Get("Access-Control-Allow-Methods"); allowMethods != "GET, POST, PATCH, DELETE, OPTIONS" {
		t.Fatalf("expected allowed methods, got %q", allowMethods)
	}

	if allowHeaders := response.Header.Get("Access-Control-Allow-Headers"); allowHeaders != "Authorization, Content-Type" {
		t.Fatalf("expected allowed headers, got %q", allowHeaders)
	}
}

func TestConfirmationRouteUsesConfiguredCorsOrigin(t *testing.T) {
	server := httptest.NewServer(newTestRouterWithRepositoryAndOrigin(
		&fakeConfirmationRepository{},
		"https://marriage.example.com",
	))
	defer server.Close()

	request, err := http.NewRequest(http.MethodOptions, server.URL+"/confirmation", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	request.Header.Set("Origin", "https://marriage.example.com")
	request.Header.Set("Access-Control-Request-Method", http.MethodPost)
	request.Header.Set("Access-Control-Request-Headers", "content-type")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("OPTIONS /confirmation failed: %v", err)
	}
	defer response.Body.Close()

	if allowOrigin := response.Header.Get("Access-Control-Allow-Origin"); allowOrigin != "https://marriage.example.com" {
		t.Fatalf("expected configured CORS allow origin, got %q", allowOrigin)
	}
}

func TestLoginRouteReturnsJWT(t *testing.T) {
	server := httptest.NewServer(newTestRouter())
	defer server.Close()

	body := []byte(`{"password":"test-password"}`)
	response, err := http.Post(server.URL+"/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /login failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.StatusCode)
	}

	var responseBody struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if responseBody.Token == "" {
		t.Fatal("expected JWT token")
	}
}

func TestLoginRouteRejectsWrongPassword(t *testing.T) {
	server := httptest.NewServer(newTestRouter())
	defer server.Close()

	body := []byte(`{"password":"wrong-password"}`)
	response, err := http.Post(server.URL+"/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /login failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.StatusCode)
	}
}

func TestAdminConfirmationRouteRejectsMissingToken(t *testing.T) {
	server := httptest.NewServer(newTestRouter())
	defer server.Close()

	response, err := http.Get(server.URL + "/admin/confirmations")
	if err != nil {
		t.Fatalf("GET /admin/confirmations failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.StatusCode)
	}
}

func TestAdminConfirmationRouteAcceptsLoginJWT(t *testing.T) {
	email := "maria@example.com"
	repository := &fakeConfirmationRepository{
		confirmations: []repositories.Confirmation{
			{ID: "72f22b5a-1489-4c38-a74e-f6611a9c7042", Name: "Maria Silva", Confirmation: true, Email: &email},
		},
	}
	server := httptest.NewServer(newTestRouterWithRepository(repository))
	defer server.Close()

	token := loginAdmin(t, server.URL)

	request, err := http.NewRequest(http.MethodGet, server.URL+"/admin/confirmations", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	request.Header.Set("Authorization", "Bearer "+token)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("GET /admin/confirmations failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.StatusCode)
	}
}

func TestAdminConfirmationRouteListsConfirmations(t *testing.T) {
	email := "maria@example.com"
	repository := &fakeConfirmationRepository{
		confirmations: []repositories.Confirmation{
			{ID: "72f22b5a-1489-4c38-a74e-f6611a9c7042", Name: "Maria Silva", Confirmation: true, Email: &email},
		},
	}
	server := httptest.NewServer(newTestRouterWithRepository(repository))
	defer server.Close()

	request, err := http.NewRequest(http.MethodGet, server.URL+"/admin/confirmations", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	request.Header.Set("Authorization", "Bearer "+loginAdmin(t, server.URL))

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("GET /admin/confirmations failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.StatusCode)
	}

	var responseBody []services.ConfirmationResponse
	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if len(responseBody) != 1 {
		t.Fatalf("expected 1 confirmation, got %d", len(responseBody))
	}
}

func TestAdminConfirmationRouteUpdatesConfirmation(t *testing.T) {
	repository := &fakeConfirmationRepository{
		confirmations: []repositories.Confirmation{
			{ID: "72f22b5a-1489-4c38-a74e-f6611a9c7042", Name: "Maria Silva", Confirmation: true},
		},
	}
	server := httptest.NewServer(newTestRouterWithRepository(repository))
	defer server.Close()

	body := []byte(`{"id":"72f22b5a-1489-4c38-a74e-f6611a9c7042","name":"Maria Santos","confirmation":false,"email":""}`)
	request, err := http.NewRequest(http.MethodPatch, server.URL+"/admin/confirmations", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	request.Header.Set("Authorization", "Bearer "+loginAdmin(t, server.URL))
	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("PATCH /admin/confirmations failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.StatusCode)
	}

	if repository.confirmations[0].Name != "Maria Santos" {
		t.Fatalf("expected updated name, got %q", repository.confirmations[0].Name)
	}
}

func TestAdminConfirmationRouteDeletesConfirmation(t *testing.T) {
	repository := &fakeConfirmationRepository{
		confirmations: []repositories.Confirmation{
			{ID: "72f22b5a-1489-4c38-a74e-f6611a9c7042", Name: "Maria Silva", Confirmation: true},
		},
	}
	server := httptest.NewServer(newTestRouterWithRepository(repository))
	defer server.Close()

	body := []byte(`{"id":"72f22b5a-1489-4c38-a74e-f6611a9c7042"}`)
	request, err := http.NewRequest(http.MethodDelete, server.URL+"/admin/confirmations", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	request.Header.Set("Authorization", "Bearer "+loginAdmin(t, server.URL))
	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("DELETE /admin/confirmations failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, response.StatusCode)
	}

	if len(repository.confirmations) != 0 {
		t.Fatalf("expected confirmation to be deleted, got %d records", len(repository.confirmations))
	}
}
