package migrations

import (
	"context"
	"embed"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

const (
	dialect = "postgres"
	dir     = "sql"
)

//go:embed "sql/*.sql"
var embedMigrations embed.FS

func NewMigrations(pool *pgxpool.Pool) error {
	db := stdlib.OpenDBFromPool(pool)
	goose.SetDialect(dialect)
	goose.SetBaseFS(embedMigrations)

	if err := goose.Up(db, dir); err != nil {
		return fmt.Errorf("goose.Up: %w", err)
	}

	return nil
}

func RollbackLast(ctx context.Context, pool *pgxpool.Pool) (err error) {
	db := stdlib.OpenDBFromPool(pool)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to close migration database: %w", closeErr))
		}
	}()
	if err := goose.DownContext(ctx, db, "sql"); err != nil {
		return fmt.Errorf("goose.DownContext: %w", err)
	}
	return nil
}
