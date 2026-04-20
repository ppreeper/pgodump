package database

import (
	"context"
	"fmt"
	"strings"
)

type View struct {
	Name       string `db:"view_name"`
	Definition string `db:"view_def"`
}

func (db *Database) GetViewDetails(ctx context.Context, schema, view string) (string, error) {
	ctx, cancel := db.withTimeout(ctx)
	defer cancel()

	q := `
		SELECT
			c.relname AS view_name,
			pg_get_viewdef(c.oid) AS view_def
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = $1 AND c.relname = $2
		AND c.relkind = 'v';`

	var v View
	if err := db.GetContext(ctx, &v, q, schema, view); err != nil {
		return "", fmt.Errorf("get view %s.%s: %w", schema, view, err)
	}

	var b strings.Builder
	b.WriteString(header(fmt.Sprintf("Name: %s; Type: VIEW; Schema: %s; Owner: -", v.Name, schema)))
	if db.IfExists {
		fmt.Fprintf(&b, "CREATE VIEW IF NOT EXISTS %s.%s AS\n%s;\n\n",
			QuoteIdentifier(schema), QuoteIdentifier(v.Name), strings.TrimSuffix(v.Definition, ";"))
	} else {
		fmt.Fprintf(&b, "CREATE VIEW %s.%s AS\n%s;\n\n",
			QuoteIdentifier(schema), QuoteIdentifier(v.Name), strings.TrimSuffix(v.Definition, ";"))
	}
	aclSQL, err := db.GetOwnershipAndPrivileges(ctx, schema, view, "VIEW")
	if err != nil {
		return "", fmt.Errorf("get ACL for view %s.%s: %w", schema, view, err)
	}
	b.WriteString(aclSQL)
	return b.String(), nil
}

// GetMatViewDetails returns the DDL for a materialized view.
func (db *Database) GetMatViewDetails(ctx context.Context, schema, matview string) (string, error) {
	ctx, cancel := db.withTimeout(ctx)
	defer cancel()

	q := `
		SELECT
			c.relname AS view_name,
			pg_get_viewdef(c.oid) AS view_def
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = $1 AND c.relname = $2
		AND c.relkind = 'm';`

	var v View
	if err := db.GetContext(ctx, &v, q, schema, matview); err != nil {
		return "", fmt.Errorf("get materialized view %s.%s: %w", schema, matview, err)
	}

	var b strings.Builder
	b.WriteString(header(fmt.Sprintf("Name: %s; Type: MATERIALIZED VIEW; Schema: %s; Owner: -", v.Name, schema)))
	if db.IfExists {
		fmt.Fprintf(&b, "CREATE MATERIALIZED VIEW IF NOT EXISTS %s.%s AS\n%s\nWITH NO DATA;\n\n",
			QuoteIdentifier(schema), QuoteIdentifier(v.Name), strings.TrimSuffix(v.Definition, ";"))
	} else {
		fmt.Fprintf(&b, "CREATE MATERIALIZED VIEW %s.%s AS\n%s\nWITH NO DATA;\n\n",
			QuoteIdentifier(schema), QuoteIdentifier(v.Name), strings.TrimSuffix(v.Definition, ";"))
	}

	indexSQL, err := db.GetTableIndexes(ctx, schema, matview)
	if err != nil {
		return "", fmt.Errorf("get indexes for matview %s.%s: %w", schema, matview, err)
	}
	b.WriteString(indexSQL)

	aclSQL, err := db.GetOwnershipAndPrivileges(ctx, schema, matview, "MATERIALIZED VIEW")
	if err != nil {
		return "", fmt.Errorf("get ACL for matview %s.%s: %w", schema, matview, err)
	}
	b.WriteString(aclSQL)
	return b.String(), nil
}
