package database

import (
	"context"
	"fmt"
	"strings"
)

// Routine holds the DDL for a single function or procedure overload.
type Routine struct {
	Name       string `db:"routine_name"`
	Type       string `db:"routine_type"`
	Definition string `db:"routine_definition"`
}

// GetRoutines returns the distinct routine names in a schema.
func (db *Database) GetRoutines(ctx context.Context, schema string) ([]string, error) {
	ctx, cancel := db.withTimeout(ctx)
	defer cancel()
	q := `
		SELECT DISTINCT p.proname AS routine_name
		FROM pg_proc p
		JOIN pg_namespace n ON n.oid = p.pronamespace
		WHERE n.nspname = $1
			AND p.prokind IN ('f', 'p')
		ORDER BY p.proname`
	var names []string
	if err := db.SelectContext(ctx, &names, q, schema); err != nil {
		return nil, fmt.Errorf("get routines for schema %s: %w", schema, err)
	}
	return names, nil
}

// GetRoutineSchema returns all overloads of a routine (function or procedure).
// pg_get_functiondef returns a complete CREATE OR REPLACE statement for both
// functions and procedures in PG11+, so we use it for both kinds.
func (db *Database) GetRoutineSchema(ctx context.Context, schema, routine string) ([]Routine, error) {
	ctx, cancel := db.withTimeout(ctx)
	defer cancel()
	q := `
		SELECT
			p.proname AS routine_name,
			CASE p.prokind WHEN 'p' THEN 'PROCEDURE' ELSE 'FUNCTION' END AS routine_type,
			pg_get_functiondef(p.oid) AS routine_definition
		FROM pg_proc p
		JOIN pg_namespace n ON n.oid = p.pronamespace
		WHERE n.nspname = $1 AND p.proname = $2
		ORDER BY p.oid`
	rr := []Routine{}
	if err := db.SelectContext(ctx, &rr, q, schema, routine); err != nil {
		return nil, fmt.Errorf("get routine schema for %s.%s: %w", schema, routine, err)
	}
	return rr, nil
}

// GetRoutine returns the SQL DDL for a single routine overload.
// pg_get_functiondef already returns a complete CREATE OR REPLACE statement.
func (db *Database) GetRoutine(schema string, r Routine) string {
	var b strings.Builder
	b.WriteString(header(fmt.Sprintf("Name: %s; Type: %s; Schema: %s; Owner: -", r.Name, r.Type, schema)))
	b.WriteString(strings.TrimSpace(r.Definition))
	b.WriteString("\n\n")
	return b.String()
}
