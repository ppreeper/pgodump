package database

import (
	"context"
	"fmt"
	"strings"
)

// RelKind represents a pg_class.relkind value.
type RelKind string

// relkind constants per pg_class.relkind
const (
	RelkindTable            RelKind = "r" // ordinary table
	RelkindView             RelKind = "v" // view
	RelkindMatView          RelKind = "m" // materialized view
	RelkindPartitionedTable RelKind = "p" // partitioned table  — TODO: dump PARTITION BY clauses
	RelkindForeignTable     RelKind = "f" // foreign table      — TODO: dump CREATE FOREIGN TABLE ... SERVER
)

// Table list of tables
type Table struct {
	Name string `db:"table_name"`
}

func (db *Database) GetTables(ctx context.Context, schema string, kind RelKind) ([]Table, error) {
	ctx, cancel := db.withTimeout(ctx)
	defer cancel()

	q := `
		SELECT c.relname AS table_name
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		LEFT JOIN pg_inherits inh ON inh.inhrelid = c.oid
		WHERE n.nspname = $1 AND c.relkind = $2`

	if len(db.IncludeTables) > 0 {
		q += " AND c.relname = ANY($3)"
	}
	q += `
		GROUP BY c.relname
		ORDER BY count(inh.inhparent), c.relname;`

	tt := []Table{}
	var err error
	if len(db.IncludeTables) > 0 {
		err = db.SelectContext(ctx, &tt, q, schema, string(kind), db.IncludeTables)
	} else {
		err = db.SelectContext(ctx, &tt, q, schema, string(kind))
	}

	if err != nil {
		return nil, fmt.Errorf("get tables for schema %s (kind %s): %w", schema, kind, err)
	}
	return tt, nil
}

func (db *Database) GetTableComment(ctx context.Context, schema, table string) (string, error) {
	ctx, cancel := db.withTimeout(ctx)
	defer cancel()
	q := `SELECT obj_description(('"' || $1 || '"."' || $2 || '"')::regclass,'pg_class') AS table_comment`
	var comment *string
	if err := db.GetContext(ctx, &comment, q, schema, table); err != nil {
		return "", fmt.Errorf("get table comment for %s.%s: %w", schema, table, err)
	}
	if comment == nil {
		return "", nil
	}
	return *comment, nil
}

// Column details using pg_catalog
type Column struct {
	ColumnName    string  `db:"column_name"`
	IsNullable    string  `db:"is_nullable"`
	ColumnDefault *string `db:"column_default"`
	DataType      string  `db:"data_type"`
	Comment       string  `db:"column_comment"`
	IsLocal       bool    `db:"is_local"`
}

func (db *Database) GetColumnDetail(ctx context.Context, schema, table string) ([]Column, error) {
	ctx, cancel := db.withTimeout(ctx)
	defer cancel()
	q := `
		SELECT
			a.attname AS column_name,
			CASE WHEN a.attnotnull THEN 'NOT NULL' ELSE '' END AS is_nullable,
			pg_get_expr(d.adbin, d.adrelid) AS column_default,
			format_type(a.atttypid, a.atttypmod) AS data_type,
			COALESCE(col_description(a.attrelid, a.attnum), '') AS column_comment,
			a.attislocal AS is_local
		FROM pg_attribute a
		JOIN pg_class c ON a.attrelid = c.oid
		JOIN pg_namespace n ON c.relnamespace = n.oid
		LEFT JOIN pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
		WHERE n.nspname = $1
			AND c.relname = $2
			AND a.attnum > 0
			AND NOT a.attisdropped
		ORDER BY a.attnum;`

	cols := []Column{}
	if err := db.SelectContext(ctx, &cols, q, schema, table); err != nil {
		return nil, fmt.Errorf("get columns for %s.%s: %w", schema, table, err)
	}
	return cols, nil
}

type Inheritance struct {
	ParentSchema string `db:"parent_schema"`
	ParentTable  string `db:"parent_table"`
}

func (db *Database) GetInheritance(ctx context.Context, schema, table string) ([]Inheritance, error) {
	ctx, cancel := db.withTimeout(ctx)
	defer cancel()
	q := `
		SELECT
			pn.nspname AS parent_schema,
			pc.relname AS parent_table
		FROM pg_inherits i
		JOIN pg_class c ON i.inhrelid = c.oid
		JOIN pg_namespace n ON c.relnamespace = n.oid
		JOIN pg_class pc ON i.inhparent = pc.oid
		JOIN pg_namespace pn ON pc.relnamespace = pn.oid
		WHERE n.nspname = $1 AND c.relname = $2
		ORDER BY i.inhseqno;`

	parents := []Inheritance{}
	if err := db.SelectContext(ctx, &parents, q, schema, table); err != nil {
		return nil, fmt.Errorf("get inheritance for %s.%s: %w", schema, table, err)
	}
	return parents, nil
}

// GenTable generates a CREATE TABLE DDL statement.
func (db *Database) GenTable(ctx context.Context, schema, table string, cols []Column) (string, error) {
	parents, err := db.GetInheritance(ctx, schema, table)
	if err != nil {
		return "", fmt.Errorf("GetInheritance %s.%s: %w", schema, table, err)
	}

	var b strings.Builder
	if db.IfExists {
		fmt.Fprintf(&b, "CREATE TABLE IF NOT EXISTS %s.%s (\n", QuoteIdentifier(schema), QuoteIdentifier(table))
	} else {
		fmt.Fprintf(&b, "CREATE TABLE %s.%s (\n", QuoteIdentifier(schema), QuoteIdentifier(table))
	}

	localCols := []string{}
	for _, c := range cols {
		if !c.IsLocal {
			continue
		}
		parts := []string{QuoteIdentifier(c.ColumnName), c.DataType}
		if c.IsNullable != "" {
			parts = append(parts, c.IsNullable)
		}
		if c.ColumnDefault != nil {
			parts = append(parts, "DEFAULT "+*c.ColumnDefault)
		}
		localCols = append(localCols, "    "+strings.Join(parts, " "))
	}
	if len(localCols) > 0 {
		b.WriteString(strings.Join(localCols, ",\n") + "\n")
	}
	b.WriteString(")")

	if len(parents) > 0 {
		pnames := make([]string, 0, len(parents))
		for _, p := range parents {
			pnames = append(pnames, fmt.Sprintf("%s.%s", QuoteIdentifier(p.ParentSchema), QuoteIdentifier(p.ParentTable)))
		}
		fmt.Fprintf(&b, " INHERITS (%s)", strings.Join(pnames, ", "))
	}
	b.WriteString(";\n")

	return b.String(), nil
}

// GetTableSchema gets the full DDL for a table.
func (db *Database) GetTableSchema(ctx context.Context, schema, table string) (string, error) {
	scols, err := db.GetColumnDetail(ctx, schema, table)
	if err != nil {
		return "", fmt.Errorf("GetColumnDetail %s.%s: %w", schema, table, err)
	}
	tblSQL, err := db.GenTable(ctx, schema, table, scols)
	if err != nil {
		return "", fmt.Errorf("GenTable %s.%s: %w", schema, table, err)
	}

	var b strings.Builder
	b.WriteString(header(fmt.Sprintf("Name: %s; Type: TABLE; Schema: %s; Owner: -", table, schema)))
	b.WriteString(tblSQL)

	constraintSQL, err := db.GetTableConstraints(ctx, schema, table)
	if err != nil {
		return "", fmt.Errorf("GetTableConstraints %s.%s: %w", schema, table, err)
	}
	b.WriteString(constraintSQL)

	indexSQL, err := db.GetTableIndexes(ctx, schema, table)
	if err != nil {
		return "", fmt.Errorf("GetTableIndexes %s.%s: %w", schema, table, err)
	}
	b.WriteString(indexSQL)

	aclSQL, err := db.GetOwnershipAndPrivileges(ctx, schema, table, "TABLE")
	if err != nil {
		return "", fmt.Errorf("GetOwnershipAndPrivileges %s.%s: %w", schema, table, err)
	}
	b.WriteString(aclSQL)

	tblc, err := db.GetTableComment(ctx, schema, table)
	if err != nil {
		return "", fmt.Errorf("GetTableComment %s.%s: %w", schema, table, err)
	}
	if tblc != "" {
		b.WriteString(header(fmt.Sprintf("Name: TABLE %s; Type: COMMENT; Schema: %s; Owner: -", table, schema)))
		fmt.Fprintf(&b, "COMMENT ON TABLE %s.%s IS '%s';\n", QuoteIdentifier(schema), QuoteIdentifier(table), strings.ReplaceAll(tblc, "'", "''"))
	}
	for _, c := range scols {
		if c.Comment != "" {
			b.WriteString(header(fmt.Sprintf("Name: %s.%s; Type: COMMENT; Schema: %s; Owner: -", table, c.ColumnName, schema)))
			fmt.Fprintf(&b, "COMMENT ON COLUMN %s.%s.%s IS '%s';\n", QuoteIdentifier(schema), QuoteIdentifier(table), QuoteIdentifier(c.ColumnName), strings.ReplaceAll(c.Comment, "'", "''"))
		}
	}
	return b.String(), nil
}
