package repository

import (
	"context"
	"fmt"

	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/app/db"
)

var DefaultPackSizes = []int{250, 500, 1000, 2000, 5000}

type PackSizesRepository interface {
	ResetToDefault(ctx context.Context) ([]int, error)
}

type sqlitePackSizesRepository struct{}

var packSizesRepo PackSizesRepository = &sqlitePackSizesRepository{}

func PackSizes() PackSizesRepository {
	return packSizesRepo
}

func (r *sqlitePackSizesRepository) ensureTable(ctx context.Context) error {
	conn, err := db.DB()
	if err != nil {
		return fmt.Errorf("error getting database connection: %w", err)
	}

	_, err = conn.ExecContext(ctx, `
	CREATE TABLE IF NOT EXISTS pack_sizes (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	size INTEGER NOT NULL UNIQUE
	);`)
	if err != nil {
		return fmt.Errorf("ensure pack_sizes table: %w", err)
	}
	return nil
}

func (r *sqlitePackSizesRepository) ResetToDefault(ctx context.Context) ([]int, error) {
	if err := r.ensureTable(ctx); err != nil {
		return nil, fmt.Errorf("error ensuring pack_sizes table: %w", err)
	}

	conn, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("error getting database connection: %w", err)
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error beginning transaction at ResetToDefault: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var count int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(1) FROM pack_sizes`).Scan(&count); err != nil {
		return nil, fmt.Errorf("error counting pack sizes: %w", err)
	}

	if count > 0 {
		if _, err := tx.ExecContext(ctx, `DELETE FROM pack_sizes`); err != nil {
			return nil, fmt.Errorf("error deleting pack sizes: %w", err)
		}
		// Reset AUTOINCREMENT counter so IDs start from 1 again.
		if _, err := tx.ExecContext(ctx, `DELETE FROM sqlite_sequence WHERE name = 'pack_sizes'`); err != nil {
			return nil, fmt.Errorf("error resetting sqlite_sequence: %w", err)
		}
	}

	for _, s := range DefaultPackSizes {
		if _, err := tx.ExecContext(ctx, `INSERT INTO pack_sizes(size) VALUES(?)`, s); err != nil {
			return nil, fmt.Errorf("error inserting pack size %d: %w", s, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction at ResetToDefault: %w", err)
	}

	return DefaultPackSizes, nil
}
