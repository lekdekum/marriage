package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Confirmation struct {
	ID           string
	Name         string
	Confirmation bool
	Email        *string
	Timestamp    time.Time
}

type PostgresConfirmationRepository struct {
	db *sql.DB
}

func NewPostgresConfirmationRepository(db *sql.DB) PostgresConfirmationRepository {
	return PostgresConfirmationRepository{
		db: db,
	}
}

func (repository PostgresConfirmationRepository) Migrate(ctx context.Context) error {
	_, err := repository.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS confirmations (
			id UUID PRIMARY KEY,
			name TEXT NOT NULL,
			confirmation BOOLEAN NOT NULL,
			email TEXT,
			"timestamp" TIMESTAMPTZ NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("create confirmations table: %w", err)
	}

	return nil
}

func (repository PostgresConfirmationRepository) Create(ctx context.Context, confirmations []Confirmation) error {
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin confirmation transaction: %w", err)
	}
	defer tx.Rollback()

	statement, err := tx.PrepareContext(ctx, `
		INSERT INTO confirmations (id, name, confirmation, email, "timestamp")
		VALUES ($1, $2, $3, $4, $5)
	`)
	if err != nil {
		return fmt.Errorf("prepare confirmation insert: %w", err)
	}
	defer statement.Close()

	for _, confirmation := range confirmations {
		var email sql.NullString
		if confirmation.Email != nil {
			email = sql.NullString{
				String: *confirmation.Email,
				Valid:  true,
			}
		}

		if _, err := statement.ExecContext(
			ctx,
			confirmation.ID,
			confirmation.Name,
			confirmation.Confirmation,
			email,
			confirmation.Timestamp,
		); err != nil {
			return fmt.Errorf("insert confirmation: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit confirmation transaction: %w", err)
	}

	return nil
}
