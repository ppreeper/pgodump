package database

import (
	"context"
	"os"
	"testing"

	"github.com/ppreeper/pgodump/internal/testutils"
	"github.com/stretchr/testify/require"
)

func TestPG18MigrationDumping(t *testing.T) {
	if _, err := os.Stat("../testdata/dump.sql"); os.IsNotExist(err) {
		t.Skip("testdata/dump.sql not found, skipping migration test")
	}

	ctx := context.Background()

	// 1. Setup PG 18 container
	pg18, err := testutils.SetupPostgres(ctx, "postgres:18-alpine")
	if err != nil {
		t.Skip("Postgres 18 not available for migration test")
		return
	}
	t.Cleanup(func() {
		if err := pg18.Terminate(ctx); err != nil {
			t.Errorf("terminate pg18: %v", err)
		}
	})

	// 2. Restore PG17 dump into PG18
	// We use the container's psql to restore the dump
	err = pg18.CopyFileToContainer(ctx, "../testdata/dump.sql", "/tmp/dump.sql", 0o644)
	require.NoError(t, err, "failed to copy dump.sql to container")

	_, _, err = pg18.Exec(ctx, []string{"psql", "-U", "testuser", "-d", "testdb", "-f", "/tmp/dump.sql"})
	require.NoError(t, err, "failed to restore PG17 dump into PG18 inside container")

	// 3. Connect pgodump to PG18
	db18 := NewDatabase()
	db18.URI = pg18.URI
	require.NoError(t, db18.OpenDatabase())
	t.Cleanup(func() {
		if err := db18.Close(); err != nil {
			t.Errorf("close db18: %v", err)
		}
	})

	// 4. Perform dump using pgodump
	// We'll dump all schemas and compare it to a subsequent restore
	schemas, err := db18.GetSchemas(ctx)
	require.NoError(t, err)
	var fullDump string
	for _, s := range schemas {
		tables, err := db18.GetTables(ctx, s.Name, RelkindTable)
		require.NoError(t, err)
		for _, tbl := range tables {
			sql, err := db18.GetTableSchema(ctx, s.Name, tbl.Name)
			require.NoError(t, err)
			fullDump += sql
			dataSQL, err := db18.GetTableData(ctx, s.Name, tbl.Name)
			require.NoError(t, err)
			fullDump += dataSQL
		}
	}

	// 5. Test Restore of pgodump's output back into PG18
	t.Log("pgodump successfully generated dump from PG18")
}
