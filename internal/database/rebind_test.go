package database

import (
	"strconv"
	"strings"
	"testing"
)

func TestRebind(t *testing.T) {
	cases := []struct {
		name   string
		driver string
		in     string
		want   string
	}{
		{
			name:   "postgres single placeholder",
			driver: "postgres",
			in:     "SELECT id FROM users WHERE email = ?",
			want:   "SELECT id FROM users WHERE email = $1",
		},
		{
			name:   "postgres increments across the statement",
			driver: "postgres",
			in:     "INSERT INTO company_members (id, company_id, user_id, status) VALUES (?, ?, ?, ?)",
			want:   "INSERT INTO company_members (id, company_id, user_id, status) VALUES ($1, $2, $3, $4)",
		},
		{
			name:   "pgx alias is treated as postgres",
			driver: "pgx",
			in:     "UPDATE users SET first_name = ?, last_name = ?, updated_at = ? WHERE id = ?",
			want:   "UPDATE users SET first_name = $1, last_name = $2, updated_at = $3 WHERE id = $4",
		},
		{
			name:   "question mark inside a string literal is left alone",
			driver: "postgres",
			in:     "SELECT '?' AS q, a FROM t WHERE b = ?",
			want:   "SELECT '?' AS q, a FROM t WHERE b = $1",
		},
		{
			name:   "query without placeholders is unchanged",
			driver: "postgres",
			in:     "SELECT COUNT(*) FROM companies WHERE deleted_at IS NULL",
			want:   "SELECT COUNT(*) FROM companies WHERE deleted_at IS NULL",
		},
		{
			name:   "sqlite keeps '?'",
			driver: "sqlite",
			in:     "SELECT id FROM users WHERE email = ?",
			want:   "SELECT id FROM users WHERE email = ?",
		},
		{
			name:   "unknown driver keeps '?'",
			driver: "",
			in:     "SELECT ?",
			want:   "SELECT ?",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Rebind(c.driver, c.in); got != c.want {
				t.Errorf("Rebind(%q, %q)\n got: %q\nwant: %q", c.driver, c.in, got, c.want)
			}
		})
	}
}

func TestRebindManyPlaceholders(t *testing.T) {
	in := "(" + "?" + strings.Repeat(", ?", 20) + ")"
	got := Rebind("postgres", in)

	for n := 1; n <= 21; n++ {
		if want := "$" + strconv.Itoa(n); !strings.Contains(got, want) {
			t.Fatalf("rebound query missing %s:\n%s", want, got)
		}
	}
	if strings.Contains(got, "?") {
		t.Fatalf("unrebound '?' left in query:\n%s", got)
	}
}
