package repositories

import (
	"context"
	"sync"
	"time"
)

type InMemoryConfirmationRepository struct {
	mu            sync.Mutex
	confirmations []Confirmation
}

func NewInMemoryConfirmationRepository() *InMemoryConfirmationRepository {
	return &InMemoryConfirmationRepository{}
}

func (repository *InMemoryConfirmationRepository) Create(ctx context.Context, confirmations []Confirmation) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()

	repository.confirmations = append(repository.confirmations, confirmations...)
	return nil
}

func (repository *InMemoryConfirmationRepository) List(ctx context.Context) ([]Confirmation, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()

	confirmations := make([]Confirmation, len(repository.confirmations))
	copy(confirmations, repository.confirmations)
	return confirmations, nil
}

func (repository *InMemoryConfirmationRepository) Update(ctx context.Context, confirmation Confirmation) (bool, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()

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

func (repository *InMemoryConfirmationRepository) Delete(ctx context.Context, id string) (bool, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()

	for index := range repository.confirmations {
		if repository.confirmations[index].ID == id {
			repository.confirmations = append(repository.confirmations[:index], repository.confirmations[index+1:]...)
			return true, nil
		}
	}

	return false, nil
}

func (repository *InMemoryConfirmationRepository) HasEmailSent(ctx context.Context, email string) (bool, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()

	for _, confirmation := range repository.confirmations {
		if confirmation.Email != nil && *confirmation.Email == email && confirmation.EmailSentAt != nil {
			return true, nil
		}
	}

	return false, nil
}

func (repository *InMemoryConfirmationRepository) MarkEmailSent(ctx context.Context, id string, sentAt time.Time) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()

	for index := range repository.confirmations {
		if repository.confirmations[index].ID == id {
			repository.confirmations[index].EmailSentAt = &sentAt
			return nil
		}
	}

	return nil
}
