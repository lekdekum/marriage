package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	passwordHash string
	jwtSecret    string
	tokenTTL     time.Duration
}

type LoginRequest struct {
	Password string `json:"password"`
}

type LoginResult struct {
	Token string `json:"token"`
}

type authClaims struct {
	Subject  string `json:"sub"`
	IssuedAt int64  `json:"iat"`
	Expires  int64  `json:"exp"`
}

func NewAuthService(passwordHash string, jwtSecret string, tokenTTL time.Duration) AuthService {
	return AuthService{
		passwordHash: strings.TrimSpace(passwordHash),
		jwtSecret:    strings.TrimSpace(jwtSecret),
		tokenTTL:     tokenTTL,
	}
}

func (service AuthService) Configured() bool {
	return service.passwordHash != "" && service.jwtSecret != ""
}

func (service AuthService) Login(request LoginRequest) (LoginResult, error) {
	if !service.Configured() {
		return LoginResult{}, fmt.Errorf("auth service is not configured")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(service.passwordHash), []byte(request.Password)); err != nil {
		return LoginResult{}, ErrInvalidCredentials
	}

	token, err := service.newToken(time.Now().UTC())
	if err != nil {
		return LoginResult{}, err
	}

	return LoginResult{Token: token}, nil
}

func (service AuthService) ValidateToken(token string) error {
	if service.jwtSecret == "" {
		return fmt.Errorf("auth service is not configured")
	}

	claims, err := service.parseToken(token)
	if err != nil {
		return err
	}

	if claims.Subject != "admin" {
		return ErrInvalidToken
	}

	if time.Now().UTC().Unix() >= claims.Expires {
		return ErrInvalidToken
	}

	return nil
}

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrInvalidToken = errors.New("invalid token")

func (service AuthService) newToken(now time.Time) (string, error) {
	ttl := service.tokenTTL
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}

	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}
	claims := authClaims{
		Subject:  "admin",
		IssuedAt: now.Unix(),
		Expires:  now.Add(ttl).Unix(),
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("encode jwt header: %w", err)
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("encode jwt claims: %w", err)
	}

	encodedHeader := base64.RawURLEncoding.EncodeToString(headerJSON)
	encodedClaims := base64.RawURLEncoding.EncodeToString(claimsJSON)
	unsignedToken := encodedHeader + "." + encodedClaims

	signature := service.sign(unsignedToken)
	return unsignedToken + "." + signature, nil
}

func (service AuthService) parseToken(token string) (authClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return authClaims{}, ErrInvalidToken
	}

	unsignedToken := parts[0] + "." + parts[1]
	expectedSignature := service.sign(unsignedToken)
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSignature)) {
		return authClaims{}, ErrInvalidToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return authClaims{}, ErrInvalidToken
	}

	var claims authClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return authClaims{}, ErrInvalidToken
	}

	return claims, nil
}

func (service AuthService) sign(unsignedToken string) string {
	mac := hmac.New(sha256.New, []byte(service.jwtSecret))
	mac.Write([]byte(unsignedToken))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
