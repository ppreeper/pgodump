package database

import (
	"context"
	"testing"

	"github.com/ppreeper/pgodump/internal/testutils"
	"github.com/stretchr/testify/require"
)

func TestTableInheritance(t *testing.T) {
	ctx := context.Background()

	// Test against Postgres 17 (known good)
	pg17, err := testutils.SetupPostgres(ctx, "postgres:17-alpine")
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := pg17.Terminate(ctx); err != nil {
			t.Errorf("terminate pg17: %v", err)
		}
	})

	db17 := NewDatabase()
	db17.URI = pg17.URI
	require.NoError(t, db17.OpenDatabase())
	t.Cleanup(func() {
		if err := db17.Close(); err != nil {
			t.Errorf("close db17: %v", err)
		}
	})

	// Setup inheritance schema
	setupSQL := `
		CREATE TABLE parent (id int PRIMARY KEY, name text);
		CREATE TABLE child (age int) INHERITS (parent);
		ALTER TABLE child ADD CONSTRAINT age_check CHECK (age > 0);
		INSERT INTO parent VALUES (1, 'parent');
		INSERT INTO child VALUES (2, 'child', 25);
	`
	_, err = db17.Exec(setupSQL)
	require.NoError(t, err)

	// Test pgodump output for child table
	output, err := db17.GetTableSchema(ctx, "public", "child")
	require.NoError(t, err)

	// Child should only have 'age' column and 'age_check' constraint.
	// Primary key 'id' is inherited, so it shouldn't appear in CREATE TABLE or ALTER TABLE for child.
	require.Contains(t, output, "CREATE TABLE \"public\".\"child\" (\n    \"age\" integer")
	require.NotContains(t, output, "    \"id\" integer")
	require.Contains(t, output, "INHERITS (\"public\".\"parent\")")
	require.Contains(t, output, "ADD CONSTRAINT \"age_check\" CHECK ((age > 0))")
	require.NotContains(t, output, "ADD CONSTRAINT parent_pkey")

	// Test Index handling
	_, err = db17.Exec("CREATE INDEX age_idx ON child (age)")
	require.NoError(t, err)
	output, err = db17.GetTableSchema(ctx, "public", "child")
	require.NoError(t, err)

	normalized := testutils.NormalizeSQL(output)
	require.Contains(t, normalized, "CREATE INDEX \"age_idx\" ON \"public\".\"child\" USING btree (age)")

	// Test View handling
	_, err = db17.Exec("CREATE VIEW child_view AS SELECT * FROM child")
	require.NoError(t, err)
	viewOutput, err := db17.GetViewDetails(ctx, "public", "child_view")
	require.NoError(t, err)
	normalizedView := testutils.NormalizeSQL(viewOutput)
	require.Contains(t, normalizedView, "CREATE VIEW \"public\".\"child_view\" AS")
	require.Contains(t, viewOutput, "SELECT id,")
	require.Contains(t, viewOutput, "name,")
	require.Contains(t, viewOutput, "age")
	require.Contains(t, viewOutput, "FROM child")

	// Test Sequence handling
	_, err = db17.Exec("CREATE SEQUENCE test_seq START 100")
	require.NoError(t, err)
	seqOutput, err := db17.GetSequenceDetails(ctx, "public", "test_seq")
	require.NoError(t, err)
	require.Contains(t, seqOutput, "CREATE SEQUENCE \"public\".\"test_seq\"")
	require.Contains(t, seqOutput, "START WITH 100")
	require.Contains(t, seqOutput, "SELECT pg_catalog.setval('\"public\".\"test_seq\"', 100, false)")

	// Test Data dumping
	dataOutput, err := db17.GetTableData(ctx, "public", "child")
	require.NoError(t, err)
	require.Contains(t, dataOutput, "COPY \"public\".\"child\" (\"id\", \"name\", \"age\") FROM stdin;")
	require.Contains(t, dataOutput, "2\tchild\t25")
	require.Contains(t, dataOutput, "\\.")

	// Test Ownership
	require.Contains(t, output, "OWNER TO \"testuser\"")

	// Test against Postgres 18 (if available, otherwise skip)
	t.Run("Postgres 18", func(t *testing.T) {
		pg18, err := testutils.SetupPostgres(ctx, "postgres:18-alpine")
		if err != nil {
			t.Skip("Postgres 18 not available")
			return
		}
		t.Cleanup(func() {
			if err := pg18.Terminate(ctx); err != nil {
				t.Errorf("terminate pg18: %v", err)
			}
		})

		db18 := NewDatabase()
		db18.URI = pg18.URI
		require.NoError(t, db18.OpenDatabase())
		t.Cleanup(func() {
			if err := db18.Close(); err != nil {
				t.Errorf("close db18: %v", err)
			}
		})

		_, err = db18.Exec(setupSQL)
		require.NoError(t, err)

		output18, err := db18.GetTableSchema(ctx, "public", "child")
		require.NoError(t, err)
		require.Contains(t, output18, "CREATE TABLE \"public\".\"child\" (\n    \"age\" integer")
		require.Contains(t, output18, "INHERITS (\"public\".\"parent\")")
	})
}
