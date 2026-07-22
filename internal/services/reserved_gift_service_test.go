package services

import (
	"context"
	"testing"
	"time"

	"marriage/internal/repositories"
)

type fakeReservedGiftRepository struct {
	reservedGifts []repositories.ReservedGift
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

func TestReservedGiftServiceCreatesReservedGift(t *testing.T) {
	repository := &fakeReservedGiftRepository{}
	service := NewReservedGiftService(repository)

	reservedGift, err := service.Create(context.Background(), ReservedGiftRequest{
		ID:   "stand-mixer",
		Name: "Maria Silva",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if reservedGift.ID != "stand-mixer" {
		t.Fatalf("expected reserved gift id, got %q", reservedGift.ID)
	}

	if reservedGift.Name != "Maria Silva" {
		t.Fatalf("expected reserved gift name, got %q", reservedGift.Name)
	}

	if len(repository.reservedGifts) != 1 {
		t.Fatalf("expected 1 saved reserved gift, got %d", len(repository.reservedGifts))
	}

	if repository.reservedGifts[0].Timestamp.IsZero() {
		t.Fatal("expected saved reserved gift to include timestamp")
	}
}

func TestReservedGiftServiceRejectsDuplicateGiftID(t *testing.T) {
	repository := &fakeReservedGiftRepository{}
	service := NewReservedGiftService(repository)

	_, err := service.Create(context.Background(), ReservedGiftRequest{
		ID:   "stand-mixer",
		Name: "Maria Silva",
	})
	if err != nil {
		t.Fatalf("expected first reservation to succeed, got %v", err)
	}

	_, err = service.Create(context.Background(), ReservedGiftRequest{
		ID:   "stand-mixer",
		Name: "Joao Silva",
	})
	if err == nil {
		t.Fatal("expected duplicate reservation to fail")
	}

	if _, ok := err.(ConflictError); !ok {
		t.Fatalf("expected ConflictError, got %T", err)
	}

	if len(repository.reservedGifts) != 1 {
		t.Fatalf("expected duplicate gift id not to create another reservation, got %d", len(repository.reservedGifts))
	}

	if repository.reservedGifts[0].Name != "Maria Silva" {
		t.Fatalf("expected original buyer to remain unchanged, got %q", repository.reservedGifts[0].Name)
	}
}

func TestReservedGiftServiceListsReservedGiftsWithoutTimestamp(t *testing.T) {
	repository := &fakeReservedGiftRepository{
		reservedGifts: []repositories.ReservedGift{
			{ID: "stand-mixer", Name: "Maria Silva", Timestamp: time.Now().UTC()},
		},
	}
	service := NewReservedGiftService(repository)

	reservedGifts, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(reservedGifts) != 1 {
		t.Fatalf("expected 1 reserved gift, got %d", len(reservedGifts))
	}

	if reservedGifts[0].ID != "stand-mixer" {
		t.Fatalf("expected reserved gift id, got %q", reservedGifts[0].ID)
	}

	if reservedGifts[0].Name != "Maria Silva" {
		t.Fatalf("expected reserved gift name, got %q", reservedGifts[0].Name)
	}
}
