package testutils

import (
	"regexp"
	"strings"
)

var (
	reTimestamp = regexp.MustCompile(`-- Dumped on \d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`)
	reVersion   = regexp.MustCompile(`-- Dumped by pg_dump version .*`)
)

// NormalizeSQL removes volatile metadata from pg_dump/pgodump output
func NormalizeSQL(sql string) string {
	// Remove "Dumped on" timestamps
	sql = reTimestamp.ReplaceAllString(sql, "-- Dumped on [TIMESTAMP]")

	// Remove "Dumped by" versions
	sql = reVersion.ReplaceAllString(sql, "-- Dumped by pg_dump version [VERSION]")

	// Normalize whitespace
	lines := strings.Split(sql, "\n")
	var normalizedLines []string
	for _, line := range lines {
		trimmed := strings.TrimRight(line, " \t\r")
		// Skip empty lines at end of blocks if needed, but keeping for now for strictness
		normalizedLines = append(normalizedLines, trimmed)
	}

	return strings.Join(normalizedLines, "\n")
}
