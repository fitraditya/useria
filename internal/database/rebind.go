package database

import (
	"strconv"
	"strings"
)

// Rebind converts '?' placeholders to the numbered form the given driver
// expects. Only Postgres needs rewriting ('$1', '$2', ...); SQLite accepts
// '?' directly, so its query is returned unchanged.
func Rebind(driver, query string) string {
	switch driver {
	case "postgres", "postgresql", "pgx":
		return rebindDollar(query)
	default:
		return query
	}
}

// rebindDollar rewrites each '?' that is not inside a single-quoted string
// literal to an incrementing numbered placeholder ($1, $2, ...).
func rebindDollar(query string) string {
	if !strings.ContainsRune(query, '?') {
		return query
	}

	var b strings.Builder
	b.Grow(len(query) + 8)
	n := 0
	inQuote := false
	for i := 0; i < len(query); i++ {
		c := query[i]
		switch {
		case c == '\'':
			inQuote = !inQuote
			b.WriteByte(c)
		case c == '?' && !inQuote:
			n++
			b.WriteByte('$')
			b.WriteString(strconv.Itoa(n))
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}
