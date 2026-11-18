package database

import (
	"os"
	"testing"
)

// TestPostgresIntegrationRebind exercises the '?' -> '$n' rewriting against a
// real Postgres server. It is skipped unless TEST_POSTGRES_DSN points at a
// disposable database — the test creates and drops only its own throwaway
// table, so it is safe to point at an empty scratch database:
//
//	TEST_POSTGRES_DSN='postgres://postgres:password@localhost:5432/useria_test?sslmode=disable' go test ./internal/database
func TestPostgresIntegrationRebind(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set TEST_POSTGRES_DSN to run the Postgres integration test")
	}

	db, err := Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS cc_rebind_probe (id TEXT PRIMARY KEY, note TEXT)`); err != nil {
		t.Fatalf("create probe table: %v", err)
	}
	t.Cleanup(func() { db.Exec(`DROP TABLE IF EXISTS cc_rebind_probe`) })

	// Every statement below uses '?' placeholders, which only work if the DB
	// wrapper rewrites them to $1/$2 for Postgres.
	if _, err := db.Exec(`DELETE FROM cc_rebind_probe WHERE id = ?`, "1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO cc_rebind_probe (id, note) VALUES (?, ?)`, "1", "hello"); err != nil {
		t.Fatalf("insert: %v", err)
	}

	var note string
	if err := db.QueryRow(`SELECT note FROM cc_rebind_probe WHERE id = ?`, "1").Scan(&note); err != nil {
		t.Fatalf("select: %v", err)
	}
	if note != "hello" {
		t.Fatalf("note = %q, want %q", note, "hello")
	}
}
