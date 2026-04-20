package database

import (
	"context"
	"fmt"
	"strings"
)

type TableConstraint struct {
	Name       string `db:"constraint_name"`
	Definition string `db:"constraint_def"`
	Type       string `db:"constraint_type"`
}

func (db *Database) GetTableConstraints(ctx context.Context, schema, table string) (string, error) {
	ctx, cancel := db.withTimeout(ctx)
	defer cancel()

	q := `
		SELECT
			conname AS constraint_name,
			pg_get_constraintdef(oid) AS constraint_def,
			contype AS constraint_type
		FROM pg_constraint
		WHERE conrelid = ('"' || $1 || '"."' || $2 || '"')::regclass
			AND conislocal = true
		ORDER BY
			CASE contype WHEN 'p' THEN 0 WHEN 'u' THEN 1 WHEN 'f' THEN 2 ELSE 3 END,
			conname;`

	var constraints []TableConstraint
	if err := db.SelectContext(ctx, &constraints, q, schema, table); err != nil {
		return "", fmt.Errorf("get constraints for %s.%s: %w", schema, table, err)
	}

	var b strings.Builder
	for _, con := range constraints {
		b.WriteString(header(fmt.Sprintf("Name: %s; Type: CONSTRAINT; Schema: %s; Owner: -", con.Name, schema)))
		fmt.Fprintf(&b, "ALTER TABLE ONLY %s.%s\n", QuoteIdentifier(schema), QuoteIdentifier(table))
		fmt.Fprintf(&b, "    ADD CONSTRAINT %s %s;\n\n", QuoteIdentifier(con.Name), QuoteConstraint(con.Definition, con.Type))
	}
	return b.String(), nil
}
