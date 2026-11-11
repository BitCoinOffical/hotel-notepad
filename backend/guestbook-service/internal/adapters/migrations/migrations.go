package migrations

import (
	"context"
	"database/sql"
	"embed"
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

func RoolbackLast(ctx context.Context, db *sql.DB) error {
	if err := goose.DownContext(ctx, db, dir); err != nil {
		return fmt.Errorf("goose.DownContext: %w", err)
	}
	return nil
}
