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
	EmailSentAt  *time.Time
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
			"timestamp" TIMESTAMPTZ NOT NULL,
			email_sent_at TIMESTAMPTZ
		)
	`)
	if err != nil {
		return fmt.Errorf("create confirmations table: %w", err)
	}

	_, err = repository.db.ExecContext(ctx, `
		ALTER TABLE confirmations
		ADD COLUMN IF NOT EXISTS email_sent_at TIMESTAMPTZ
	`)
	if err != nil {
		return fmt.Errorf("add email_sent_at column: %w", err)
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

func (repository PostgresConfirmationRepository) List(ctx context.Context) ([]Confirmation, error) {
	rows, err := repository.db.QueryContext(ctx, `
		SELECT id, name, confirmation, email, "timestamp", email_sent_at
		FROM confirmations
		ORDER BY "timestamp" DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list confirmations: %w", err)
	}
	defer rows.Close()

	var confirmations []Confirmation
	for rows.Next() {
		var confirmation Confirmation
		var email sql.NullString
		var emailSentAt sql.NullTime

		if err := rows.Scan(
			&confirmation.ID,
			&confirmation.Name,
			&confirmation.Confirmation,
			&email,
			&confirmation.Timestamp,
			&emailSentAt,
		); err != nil {
			return nil, fmt.Errorf("scan confirmation: %w", err)
		}

		if email.Valid {
			emailValue := email.String
			confirmation.Email = &emailValue
		}

		if emailSentAt.Valid {
			emailSentAtValue := emailSentAt.Time
			confirmation.EmailSentAt = &emailSentAtValue
		}

		confirmations = append(confirmations, confirmation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate confirmations: %w", err)
	}

	return confirmations, nil
}

func (repository PostgresConfirmationRepository) Update(ctx context.Context, confirmation Confirmation) (bool, error) {
	var email sql.NullString
	if confirmation.Email != nil {
		email = sql.NullString{
			String: *confirmation.Email,
			Valid:  true,
		}
	}

	result, err := repository.db.ExecContext(ctx, `
		UPDATE confirmations
		SET name = $2, confirmation = $3, email = $4
		WHERE id = $1
	`, confirmation.ID, confirmation.Name, confirmation.Confirmation, email)
	if err != nil {
		return false, fmt.Errorf("update confirmation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read update result: %w", err)
	}

	return rowsAffected > 0, nil
}

func (repository PostgresConfirmationRepository) HasEmailSent(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := repository.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM confirmations
			WHERE lower(email) = lower($1)
				AND email_sent_at IS NOT NULL
		)
	`, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check email sent: %w", err)
	}

	return exists, nil
}

func (repository PostgresConfirmationRepository) MarkEmailSent(ctx context.Context, id string, sentAt time.Time) error {
	_, err := repository.db.ExecContext(ctx, `
		UPDATE confirmations
		SET email_sent_at = $2
		WHERE id = $1
	`, id, sentAt)
	if err != nil {
		return fmt.Errorf("mark email sent: %w", err)
	}

	return nil
}

func (repository PostgresConfirmationRepository) Delete(ctx context.Context, id string) (bool, error) {
	result, err := repository.db.ExecContext(ctx, `
		DELETE FROM confirmations
		WHERE id = $1
	`, id)
	if err != nil {
		return false, fmt.Errorf("delete confirmation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read delete result: %w", err)
	}

	return rowsAffected > 0, nil
}
