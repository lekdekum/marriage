package services

import (
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func TestAuthServiceLoginReturnsValidJWT(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	service := NewAuthService(string(passwordHash), "test-jwt-secret", time.Hour)

	result, err := service.Login(LoginRequest{Password: "correct-password"})
	if err != nil {
		t.Fatalf("expected login to succeed, got %v", err)
	}

	if result.Token == "" {
		t.Fatal("expected JWT token")
	}

	if err := service.ValidateToken(result.Token); err != nil {
		t.Fatalf("expected JWT token to validate, got %v", err)
	}
}

func TestAuthServiceLoginRejectsWrongPassword(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	service := NewAuthService(string(passwordHash), "test-jwt-secret", time.Hour)

	_, err = service.Login(LoginRequest{Password: "wrong-password"})
	if err == nil {
		t.Fatal("expected login to fail")
	}
}

func TestAuthServiceLoginRequiresConfiguration(t *testing.T) {
	service := NewAuthService("", "", time.Hour)

	_, err := service.Login(LoginRequest{Password: "correct-password"})
	if err == nil {
		t.Fatal("expected login to fail without configuration")
	}
}

func TestAuthServiceRejectsInvalidJWT(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	service := NewAuthService(string(passwordHash), "test-jwt-secret", time.Hour)

	if err := service.ValidateToken("not-a-jwt"); err == nil {
		t.Fatal("expected invalid JWT to fail")
	}
}
