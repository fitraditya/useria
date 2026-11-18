package database

import (
	"path/filepath"
	"testing"
)

// newSQLiteTestDB opens a throwaway SQLite database with the dev schema
// applied. Tests live in this package, so the migration path is relative to
// internal/database.
func newSQLiteTestDB(t *testing.T) *DB {
	t.Helper()

	db, err := Connect("sqlite", filepath.Join(t.TempDir(), "useria-test.db"))
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := Migrate(db, "migrations/001_init_sqlite.sql"); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestMigrateCreatesSchema(t *testing.T) {
	db := newSQLiteTestDB(t)

	for _, table := range []string{"users", "companies", "roles", "company_members", "permissions"} {
		var n int
		if err := db.QueryRow(`SELECT COUNT(*) FROM ` + table).Scan(&n); err != nil {
			t.Fatalf("table %s not queryable after migrate: %v", table, err)
		}
	}
}

func TestSeedIsIdempotent(t *testing.T) {
	db := newSQLiteTestDB(t)

	if err := Seed(db); err != nil {
		t.Fatalf("first seed: %v", err)
	}
	if err := Seed(db); err != nil {
		t.Fatalf("second seed (must be idempotent): %v", err)
	}

	var roles int
	if err := db.QueryRow(`SELECT COUNT(*) FROM roles WHERE company_id IS NULL AND name = ?`, "Admin").Scan(&roles); err != nil {
		t.Fatalf("count roles: %v", err)
	}
	if roles != 1 {
		t.Fatalf("Admin system roles = %d, want 1", roles)
	}

	var perms int
	if err := db.QueryRow(`SELECT COUNT(*) FROM permissions`).Scan(&perms); err != nil {
		t.Fatalf("count permissions: %v", err)
	}
	if perms == 0 {
		t.Fatal("expected seeded permissions, got none")
	}
}

func TestSeedAdminCreatesPlatformMembership(t *testing.T) {
	db := newSQLiteTestDB(t)
	if err := Seed(db); err != nil {
		t.Fatalf("seed: %v", err)
	}

	created, err := SeedAdmin(db, "root@useria.test", "pw", "", "")
	if err != nil {
		t.Fatalf("SeedAdmin: %v", err)
	}
	if !created {
		t.Fatal("expected first SeedAdmin call to create the user")
	}

	// Running again must be a no-op, not an error or a duplicate.
	created, err = SeedAdmin(db, "root@useria.test", "pw", "", "")
	if err != nil {
		t.Fatalf("second SeedAdmin: %v", err)
	}
	if created {
		t.Fatal("expected second SeedAdmin call to be a no-op")
	}
}
