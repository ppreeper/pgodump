package database

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// WriteTableData streams COPY-format table data directly to w, avoiding
// buffering the entire table in memory. The caller is responsible for
// buffering w (e.g. via bufio.Writer) if needed.
func (db *Database) WriteTableData(ctx context.Context, w io.Writer, schema, table string) (retErr error) {
	// Defensive: callers also guard with !SchemaOnly, but we check here too
	// so WriteTableData is safe to call unconditionally.
	if db.SchemaOnly {
		return nil
	}

	ctx, cancel := db.withTimeout(ctx)
	defer cancel()

	cols, err := db.GetColumnDetail(ctx, schema, table)
	if err != nil {
		return fmt.Errorf("get columns for data dump %s.%s: %w", schema, table, err)
	}

	quotedColNames := make([]string, 0, len(cols))
	for _, c := range cols {
		quotedColNames = append(quotedColNames, QuoteIdentifier(c.ColumnName))
	}

	q := fmt.Sprintf("SELECT %s FROM %s.%s",
		strings.Join(quotedColNames, ", "),
		QuoteIdentifier(schema), QuoteIdentifier(table))

	rows, err := db.QueryxContext(ctx, q)
	if err != nil {
		return fmt.Errorf("query table data %s.%s: %w", schema, table, err)
	}
	defer func() {
		if err := rows.Close(); err != nil && retErr == nil {
			retErr = fmt.Errorf("close rows for table %s.%s: %w", schema, table, err)
		}
	}()

	// Track the first write error so we can stop early and report it.
	var writeErr error
	write := func(s string) {
		if writeErr != nil {
			return
		}
		_, writeErr = io.WriteString(w, s)
	}
	writef := func(format string, args ...any) {
		if writeErr != nil {
			return
		}
		_, writeErr = fmt.Fprintf(w, format, args...)
	}

	// Query succeeded — safe to start the COPY block.
	write(header(fmt.Sprintf("Data for Name: %s; Type: TABLE DATA; Schema: %s; Owner: -", table, schema)))
	writef("COPY %s.%s (%s) FROM stdin;\n",
		QuoteIdentifier(schema), QuoteIdentifier(table), strings.Join(quotedColNames, ", "))

	rowStrs := make([]string, 0, len(cols))
	for rows.Next() {
		if writeErr != nil {
			return fmt.Errorf("write data for table %s.%s: %w", schema, table, writeErr)
		}

		vals, err := rows.SliceScan()
		if err != nil {
			return fmt.Errorf("scan row for table %s.%s: %w", schema, table, err)
		}

		rowStrs = rowStrs[:0]
		for _, v := range vals {
			if v == nil {
				rowStrs = append(rowStrs, "\\N")
			} else {
				var s string
				switch val := v.(type) {
				case []uint8:
					s = string(val)
				default:
					s = fmt.Sprintf("%v", v)
				}
				// COPY text escape order: backslash first, then control chars.
				s = strings.ReplaceAll(s, "\\", "\\\\")
				s = strings.ReplaceAll(s, "\t", "\\t")
				s = strings.ReplaceAll(s, "\n", "\\n")
				s = strings.ReplaceAll(s, "\r", "\\r")
				rowStrs = append(rowStrs, s)
			}
		}
		write(strings.Join(rowStrs, "\t"))
		write("\n")
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read rows for table %s.%s: %w", schema, table, err)
	}
	if writeErr != nil {
		return fmt.Errorf("write data for table %s.%s: %w", schema, table, writeErr)
	}
	write("\\.\n\n")
	if writeErr != nil {
		return fmt.Errorf("write data for table %s.%s: %w", schema, table, writeErr)
	}
	return nil
}

// GetTableData returns COPY-format table data as a string.
// Prefer WriteTableData for production use to avoid buffering large tables.
func (db *Database) GetTableData(ctx context.Context, schema, table string) (string, error) {
	var b strings.Builder
	if err := db.WriteTableData(ctx, &b, schema, table); err != nil {
		return "", err
	}
	return b.String(), nil
}
