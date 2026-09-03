package repositories

import (
	"context"
	"sync"
)

type InMemoryReservedGiftRepository struct {
	mu            sync.Mutex
	reservedGifts []ReservedGift
}

func NewInMemoryReservedGiftRepository() *InMemoryReservedGiftRepository {
	return &InMemoryReservedGiftRepository{}
}

func (repository *InMemoryReservedGiftRepository) Create(ctx context.Context, reservedGift ReservedGift) (bool, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()

	for _, existingReservedGift := range repository.reservedGifts {
		if existingReservedGift.ID == reservedGift.ID {
			return false, nil
		}
	}

	repository.reservedGifts = append(repository.reservedGifts, reservedGift)
	return true, nil
}

func (repository *InMemoryReservedGiftRepository) List(ctx context.Context) ([]ReservedGift, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()

	reservedGifts := make([]ReservedGift, len(repository.reservedGifts))
	copy(reservedGifts, repository.reservedGifts)
	return reservedGifts, nil
}
