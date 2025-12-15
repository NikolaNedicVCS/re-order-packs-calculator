package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/app/db"
	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/app/models"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

var defaultPackSizes = []int{250, 500, 1000, 2000, 5000}

func DefaultPackSizes() []int {
	out := make([]int, len(defaultPackSizes))
	copy(out, defaultPackSizes)
	return out
}

type PackSizesRepository interface {
	ResetToDefault(ctx context.Context) ([]int, error)
	List(ctx context.Context) ([]models.PackSize, error)
	Create(ctx context.Context, size int) (*models.PackSize, error)
	Update(ctx context.Context, id int64, size int) (*models.PackSize, error)
	Delete(ctx context.Context, id int64) error
}

type sqlitePackSizesRepository struct{}

var packSizesRepo PackSizesRepository = &sqlitePackSizesRepository{}

func PackSizes() PackSizesRepository {
	return packSizesRepo
}

// SetPackSizesRepository swaps the repository implementation (primarily for tests).
func SetPackSizesRepository(repo PackSizesRepository) {
	if repo == nil {
		panic("PackSizesRepository must not be nil")
	}
	packSizesRepo = repo
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

	if _, err := tx.ExecContext(ctx, `DELETE FROM pack_sizes`); err != nil {
		return nil, fmt.Errorf("error deleting pack sizes: %w", err)
	}

	// Best-effort reset of AUTOINCREMENT counter so IDs start from 1 again.
	// (sqlite_sequence may not exist in some configurations; ignore errors.)
	_, _ = tx.ExecContext(ctx, `DELETE FROM sqlite_sequence WHERE name = 'pack_sizes'`)

	defaults := DefaultPackSizes()
	for _, s := range defaults {
		if _, err := tx.ExecContext(ctx, `INSERT INTO pack_sizes(size) VALUES(?)`, s); err != nil {
			return nil, fmt.Errorf("error inserting pack size %d: %w", s, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction at ResetToDefault: %w", err)
	}

	return defaults, nil
}

func (r *sqlitePackSizesRepository) List(ctx context.Context) ([]models.PackSize, error) {
	if err := r.ensureTable(ctx); err != nil {
		return nil, fmt.Errorf("error ensuring pack_sizes table: %w", err)
	}

	conn, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("error getting database connection: %w", err)
	}

	rows, err := conn.QueryContext(ctx, `SELECT id, size FROM pack_sizes ORDER BY size ASC`)
	if err != nil {
		return nil, fmt.Errorf("list pack sizes: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.PackSize
	for rows.Next() {
		var p models.PackSize
		if err := rows.Scan(&p.ID, &p.Size); err != nil {
			return nil, fmt.Errorf("scan pack size: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pack sizes: %w", err)
	}
	return out, nil
}

func (r *sqlitePackSizesRepository) Create(ctx context.Context, size int) (*models.PackSize, error) {
	if err := r.ensureTable(ctx); err != nil {
		return nil, fmt.Errorf("error ensuring pack_sizes table: %w", err)
	}

	conn, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("error getting database connection: %w", err)
	}

	res, err := conn.ExecContext(ctx, `INSERT INTO pack_sizes(size) VALUES(?)`, size)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("%w", ErrConflict)
		}
		return nil, fmt.Errorf("insert pack size: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("read insert id: %w", err)
	}
	return &models.PackSize{ID: id, Size: size}, nil
}

func (r *sqlitePackSizesRepository) Update(ctx context.Context, id int64, size int) (*models.PackSize, error) {
	if err := r.ensureTable(ctx); err != nil {
		return nil, fmt.Errorf("error ensuring pack_sizes table: %w", err)
	}

	conn, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("error getting database connection: %w", err)
	}

	res, err := conn.ExecContext(ctx, `UPDATE pack_sizes SET size = ? WHERE id = ?`, size, id)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("%w", ErrConflict)
		}
		return nil, fmt.Errorf("update pack size: %w", err)
	}
	ra, err := res.RowsAffected()
	if err == nil && ra == 0 {
		return nil, fmt.Errorf("%w", ErrNotFound)
	}
	return &models.PackSize{ID: id, Size: size}, nil
}

func (r *sqlitePackSizesRepository) Delete(ctx context.Context, id int64) error {
	if err := r.ensureTable(ctx); err != nil {
		return fmt.Errorf("error ensuring pack_sizes table: %w", err)
	}

	conn, err := db.DB()
	if err != nil {
		return fmt.Errorf("error getting database connection: %w", err)
	}

	res, err := conn.ExecContext(ctx, `DELETE FROM pack_sizes WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete pack size: %w", err)
	}
	ra, err := res.RowsAffected()
	if err == nil && ra == 0 {
		return fmt.Errorf("%w", ErrNotFound)
	}
	return nil
}

func isUniqueViolation(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "unique constraint failed")
}
