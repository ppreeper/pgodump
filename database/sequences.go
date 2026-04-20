package database

import (
	"context"
	"fmt"
	"log"
	"strings"
)

type Sequence struct {
	Name string `db:"seq_name"`
}

func (db *Database) GetSequences(ctx context.Context, schema string) ([]Sequence, error) {
	ctx, cancel := db.withTimeout(ctx)
	defer cancel()

	q := `
		SELECT c.relname AS seq_name
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = $1 AND c.relkind = 'S'`

	if len(db.IncludeTables) > 0 {
		q += " AND c.relname = ANY($2)"
	}
	q += " ORDER BY c.relname;"

	seqs := []Sequence{}
	var err error
	if len(db.IncludeTables) > 0 {
		err = db.SelectContext(ctx, &seqs, q, schema, db.IncludeTables)
	} else {
		err = db.SelectContext(ctx, &seqs, q, schema)
	}

	if err != nil {
		return nil, fmt.Errorf("get sequences for schema %s: %w", schema, err)
	}
	return seqs, nil
}

func (db *Database) GetSequenceDetails(ctx context.Context, schema, name string) (string, error) {
	ctx, cancel := db.withTimeout(ctx)
	defer cancel()

	// 1. Get sequence definition
	qDef := `
		SELECT
			format_type(seqtypid, 0) AS data_type,
			seqstart AS start_value,
			seqmin AS min_value,
			seqmax AS max_value,
			seqincrement AS increment_by,
			seqcycle AS is_cycled,
			seqcache AS cache_size
		FROM pg_sequence
		WHERE seqrelid = ('"' || $1 || '"."' || $2 || '"')::regclass;`

	type SeqDef struct {
		DataType    string `db:"data_type"`
		StartValue  int64  `db:"start_value"`
		MinValue    int64  `db:"min_value"`
		MaxValue    int64  `db:"max_value"`
		IncrementBy int64  `db:"increment_by"`
		IsCycled    bool   `db:"is_cycled"`
		CacheSize   int64  `db:"cache_size"`
	}

	var sd SeqDef
	if err := db.GetContext(ctx, &sd, qDef, schema, name); err != nil {
		return "", fmt.Errorf("get sequence definition for %s.%s: %w", schema, name, err)
	}

	var b strings.Builder
	b.WriteString(header(fmt.Sprintf("Name: %s; Type: SEQUENCE; Schema: %s; Owner: -", name, schema)))
	if db.IfExists {
		fmt.Fprintf(&b, "CREATE SEQUENCE IF NOT EXISTS %s.%s\n", QuoteIdentifier(schema), QuoteIdentifier(name))
	} else {
		fmt.Fprintf(&b, "CREATE SEQUENCE %s.%s\n", QuoteIdentifier(schema), QuoteIdentifier(name))
	}
	fmt.Fprintf(&b, "    AS %s\n", sd.DataType)
	fmt.Fprintf(&b, "    START WITH %d\n", sd.StartValue)
	fmt.Fprintf(&b, "    INCREMENT BY %d\n", sd.IncrementBy)
	fmt.Fprintf(&b, "    MINVALUE %d\n", sd.MinValue)
	fmt.Fprintf(&b, "    MAXVALUE %d\n", sd.MaxValue)
	fmt.Fprintf(&b, "    CACHE %d", sd.CacheSize)
	if sd.IsCycled {
		b.WriteString("\n    CYCLE")
	}
	b.WriteString(";\n\n")
	aclSQL, err := db.GetOwnershipAndPrivileges(ctx, schema, name, "SEQUENCE")
	if err != nil {
		return "", fmt.Errorf("get ACL for sequence %s.%s: %w", schema, name, err)
	}
	b.WriteString(aclSQL)

	// 2. Get sequence state (last_value)
	qVal := fmt.Sprintf("SELECT last_value, is_called FROM %s.%s", QuoteIdentifier(schema), QuoteIdentifier(name))
	type SeqVal struct {
		LastValue int64 `db:"last_value"`
		IsCalled  bool  `db:"is_called"`
	}
	var sv SeqVal
	if err := db.GetContext(ctx, &sv, qVal); err != nil {
		log.Printf("WARNING: could not read last_value for sequence %s.%s: %v", schema, name, err)
	} else {
		qualifiedName := QuoteIdentifier(schema) + "." + QuoteIdentifier(name)
		escapedName := strings.ReplaceAll(qualifiedName, "'", "''")
		b.WriteString(header(fmt.Sprintf("Name: %s; Type: SEQUENCE SET; Schema: %s; Owner: -", name, schema)))
		fmt.Fprintf(&b, "SELECT pg_catalog.setval('%s', %d, %t);\n\n",
			escapedName, sv.LastValue, sv.IsCalled)
	}

	return b.String(), nil
}
