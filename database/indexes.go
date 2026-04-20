package database

import (
	"context"
	"fmt"
	"strings"
)

type Index struct {
	Name       string `db:"indexname"`
	Definition string `db:"indexdef"`
}

func (db *Database) GetTableIndexes(ctx context.Context, schema, table string) (string, error) {
	ctx, cancel := db.withTimeout(ctx)
	defer cancel()

	// Query pg_catalog directly rather than the pg_indexes convenience view.
	// Exclude indexes that back constraints (they are emitted by GetTableConstraints).
	q := `
		SELECT
			ic.relname AS indexname,
			pg_get_indexdef(i.indexrelid) AS indexdef
		FROM pg_index i
		JOIN pg_class tc ON tc.oid = i.indrelid
		JOIN pg_namespace n ON n.oid = tc.relnamespace
		JOIN pg_class ic ON ic.oid = i.indexrelid
		WHERE n.nspname = $1
			AND tc.relname = $2
			AND NOT EXISTS (
				SELECT 1 FROM pg_constraint c
				WHERE c.conindid = i.indexrelid
			)
		ORDER BY ic.relname;`

	var indexes []Index
	if err := db.SelectContext(ctx, &indexes, q, schema, table); err != nil {
		return "", fmt.Errorf("get indexes for %s.%s: %w", schema, table, err)
	}

	var b strings.Builder
	for _, idx := range indexes {
		b.WriteString(header(fmt.Sprintf("Name: %s; Type: INDEX; Schema: %s; Owner: -", idx.Name, schema)))
		b.WriteString(QuoteIndex(idx.Definition) + ";\n\n")
	}
	return b.String(), nil
}
