package repository

import (
	"context"
	"database/sql"
)

// DBTX is the subset of SQL execution the repositories depend on.
// Implemented by *sql.DB and by database.DB, whose methods rewrite '?'
// placeholders into the connected backend's native form before executing.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}
