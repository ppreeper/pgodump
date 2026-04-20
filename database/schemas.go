package database

import (
	"context"
	"fmt"
)

// Schema struct to hold schemas
type Schema struct {
	Name string `db:"schema_name"`
}

// GetSchemas returns schema list
func (db *Database) GetSchemas(ctx context.Context) ([]Schema, error) {
	ctx, cancel := db.withTimeout(ctx)
	defer cancel()

	// Use pg_namespace to get user schemas.
	// Filter out internal postgres schemas.
	q := `
		SELECT nspname AS schema_name
		FROM pg_namespace
		WHERE nspname NOT LIKE 'pg_%'
			AND nspname != 'information_schema'`

	if len(db.IncludeSchemas) > 0 {
		q += " AND nspname = ANY($1)"
	}
	q += " ORDER BY nspname;"

	ss := []Schema{}
	var err error
	if len(db.IncludeSchemas) > 0 {
		err = db.SelectContext(ctx, &ss, q, db.IncludeSchemas)
	} else {
		err = db.SelectContext(ctx, &ss, q)
	}
	if err != nil {
		return nil, fmt.Errorf("GetSchemas: %w", err)
	}
	return ss, nil
}
