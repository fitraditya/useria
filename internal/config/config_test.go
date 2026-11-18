package config

import "testing"

func TestDSN(t *testing.T) {
	cases := []struct {
		name string
		cfg  Config
		want string
	}{
		{
			name: "postgres builds a connection URL",
			cfg: Config{
				DBDriver: "postgres", DBHost: "db.internal", DBPort: "5432",
				DBUser: "useria", DBPassword: "secret", DBName: "useria", DBSSLMode: "disable",
			},
			want: "postgres://useria:secret@db.internal:5432/useria?sslmode=disable",
		},
		{
			name: "postgresql alias works the same way",
			cfg: Config{
				DBDriver: "postgresql", DBHost: "h", DBPort: "5432",
				DBUser: "u", DBPassword: "p", DBName: "d", DBSSLMode: "require",
			},
			want: "postgres://u:p@h:5432/d?sslmode=require",
		},
		{
			name: "sqlite returns the file path verbatim",
			cfg:  Config{DBDriver: "sqlite", DBSQLitePath: "./useria.db"},
			want: "./useria.db",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.cfg.DSN(); got != c.want {
				t.Errorf("DSN() = %q, want %q", got, c.want)
			}
		})
	}
}

func TestMigrationPath(t *testing.T) {
	cases := []struct {
		driver string
		want   string
	}{
		{"postgres", "internal/database/migrations/001_init_postgres.sql"},
		{"postgresql", "internal/database/migrations/001_init_postgres.sql"},
		{"sqlite", "internal/database/migrations/001_init_sqlite.sql"},
	}
	for _, c := range cases {
		t.Run(c.driver, func(t *testing.T) {
			if got := (&Config{DBDriver: c.driver}).MigrationPath(); got != c.want {
				t.Errorf("MigrationPath() = %q, want %q", got, c.want)
			}
		})
	}
}
