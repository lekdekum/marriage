package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"marriage/internal/repositories"
)

type ReservedGiftRepository interface {
	Create(ctx context.Context, reservedGift repositories.ReservedGift) (bool, error)
	List(ctx context.Context) ([]repositories.ReservedGift, error)
}

type ReservedGiftService struct {
	repository ReservedGiftRepository
}

type ReservedGiftRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ReservedGiftResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ConflictError struct{}

func (err ConflictError) Error() string {
	return "resource already exists"
}

func NewReservedGiftService(repository ReservedGiftRepository) ReservedGiftService {
	return ReservedGiftService{
		repository: repository,
	}
}

func (service ReservedGiftService) Create(ctx context.Context, request ReservedGiftRequest) (ReservedGiftResponse, error) {
	var details []string
	if strings.TrimSpace(request.ID) == "" {
		details = append(details, "id must be a non-empty string")
	}

	if strings.TrimSpace(request.Name) == "" {
		details = append(details, "name must be a non-empty string")
	}

	if len(details) > 0 {
		return ReservedGiftResponse{}, ValidationError{Details: details}
	}

	if service.repository == nil {
		return ReservedGiftResponse{}, fmt.Errorf("reserved gift repository is not configured")
	}

	reservedGift := repositories.ReservedGift{
		ID:        strings.TrimSpace(request.ID),
		Name:      strings.TrimSpace(request.Name),
		Timestamp: time.Now().UTC(),
	}

	created, err := service.repository.Create(ctx, reservedGift)
	if err != nil {
		return ReservedGiftResponse{}, fmt.Errorf("save reserved gift: %w", err)
	}

	if !created {
		return ReservedGiftResponse{}, ConflictError{}
	}

	return ReservedGiftResponse{
		ID:   reservedGift.ID,
		Name: reservedGift.Name,
	}, nil
}

func (service ReservedGiftService) List(ctx context.Context) ([]ReservedGiftResponse, error) {
	if service.repository == nil {
		return nil, fmt.Errorf("reserved gift repository is not configured")
	}

	reservedGifts, err := service.repository.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list reserved gifts: %w", err)
	}

	responses := make([]ReservedGiftResponse, 0, len(reservedGifts))
	for _, reservedGift := range reservedGifts {
		responses = append(responses, ReservedGiftResponse{
			ID:   reservedGift.ID,
			Name: reservedGift.Name,
		})
	}

	return responses, nil
}
