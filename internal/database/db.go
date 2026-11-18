package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

// sqlDriver maps a logical DB_DRIVER value to the database/sql driver name
// its blank import registers (Postgres is served by pgx's database/sql
// adapter, registered as "pgx").
func sqlDriver(driver string) string {
	switch driver {
	case "postgres", "postgresql", "pgx":
		return "pgx"
	case "sqlite", "sqlite3":
		return "sqlite"
	default:
		return driver
	}
}

// DB wraps *sql.DB and rewrites '?' placeholders to the backend's native
// form before each query, so repositories can keep writing one
// dialect-independent query string. SQLite uses '?' as-is; Postgres needs
// the numbered '$1', '$2', ... form.
type DB struct {
	*sql.DB
	driver string
}

// Connect opens a connection pool for the given logical driver and DSN.
func Connect(driver, dsn string) (*DB, error) {
	db, err := sql.Open(sqlDriver(driver), dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if sqlDriver(driver) == "sqlite" {
		// SQLite allows one writer at a time; a single pooled connection
		// avoids "database is locked" errors under concurrent access.
		db.SetMaxOpenConns(1)
	} else {
		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(25)
	}
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return &DB{DB: db, driver: driver}, nil
}

// Wrap adapts an existing *sql.DB (e.g. one opened directly by a test) to
// the placeholder-rewriting DB.
func Wrap(db *sql.DB, driver string) *DB { return &DB{DB: db, driver: driver} }

func (d *DB) rebind(query string) string { return Rebind(d.driver, query) }

func (d *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return d.DB.ExecContext(ctx, d.rebind(query), args...)
}

func (d *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return d.DB.QueryContext(ctx, d.rebind(query), args...)
}

func (d *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return d.DB.QueryRowContext(ctx, d.rebind(query), args...)
}

func (d *DB) Begin() (*Tx, error) {
	tx, err := d.DB.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{Tx: tx, driver: d.driver}, nil
}

func (d *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := d.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{Tx: tx, driver: d.driver}, nil
}

// Tx is the *sql.Tx counterpart to DB, rewriting placeholders identically.
type Tx struct {
	*sql.Tx
	driver string
}

func (t *Tx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.Tx.ExecContext(ctx, Rebind(t.driver, query), args...)
}

func (t *Tx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return t.Tx.QueryContext(ctx, Rebind(t.driver, query), args...)
}

func (t *Tx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return t.Tx.QueryRowContext(ctx, Rebind(t.driver, query), args...)
}

// Migrate applies a .sql file statement-by-statement: it strips full-line
// "--" comments and splits on ';'. That is sufficient for this project's
// schema files, which contain no dollar-quoted function bodies.
func Migrate(db *DB, path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", path, err)
	}

	var cleaned []string
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "--") {
			continue
		}
		cleaned = append(cleaned, line)
	}

	for _, stmt := range strings.Split(strings.Join(cleaned, "\n"), ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("exec migration statement: %w\n%s", err, stmt)
		}
	}
	return nil
}
