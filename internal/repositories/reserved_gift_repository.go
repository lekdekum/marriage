package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type ReservedGift struct {
	ID        string
	Name      string
	Timestamp time.Time
}

type PostgresReservedGiftRepository struct {
	db *sql.DB
}

func NewPostgresReservedGiftRepository(db *sql.DB) PostgresReservedGiftRepository {
	return PostgresReservedGiftRepository{
		db: db,
	}
}

func (repository PostgresReservedGiftRepository) Migrate(ctx context.Context) error {
	_, err := repository.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS reserved_gifts (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			"timestamp" TIMESTAMPTZ NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("create reserved_gifts table: %w", err)
	}

	return nil
}

func (repository PostgresReservedGiftRepository) Create(ctx context.Context, reservedGift ReservedGift) (bool, error) {
	result, err := repository.db.ExecContext(ctx, `
		INSERT INTO reserved_gifts (id, name, "timestamp")
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO NOTHING
	`, reservedGift.ID, reservedGift.Name, reservedGift.Timestamp)
	if err != nil {
		return false, fmt.Errorf("insert reserved gift: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read reserved gift insert result: %w", err)
	}

	return rowsAffected > 0, nil
}

func (repository PostgresReservedGiftRepository) List(ctx context.Context) ([]ReservedGift, error) {
	rows, err := repository.db.QueryContext(ctx, `
		SELECT id, name, "timestamp"
		FROM reserved_gifts
		ORDER BY "timestamp" DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list reserved gifts: %w", err)
	}
	defer rows.Close()

	var reservedGifts []ReservedGift
	for rows.Next() {
		var reservedGift ReservedGift
		if err := rows.Scan(&reservedGift.ID, &reservedGift.Name, &reservedGift.Timestamp); err != nil {
			return nil, fmt.Errorf("scan reserved gift: %w", err)
		}

		reservedGifts = append(reservedGifts, reservedGift)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reserved gifts: %w", err)
	}

	return reservedGifts, nil
}
