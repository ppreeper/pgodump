# pgodump

A high-fidelity PostgreSQL dump utility written in Go, designed as a replacement for `pg_dump` with specific enhancements for PostgreSQL 18+ migrations, Odoo-style inheritance, and reserved keyword handling.

## Features

- **PG18 Ready**: Correctly handles inheritance structures that cause issues in standard `pg_dump` on PostgreSQL 18.
- **Hardened Quoting**: Automatically double-quotes all identifiers to prevent collisions with reserved keywords (e.g., `user`, `limit`, `order`).
- **Inheritance Aware**: Intelligent table ordering ensures parent tables are created before children during restoration.
- **ACL to GRANT**: Converts internal PostgreSQL ACL strings into standard SQL `GRANT` statements.
- **Library & CLI**: Can be used as a standalone binary or imported as a Go module.

---

## CLI Installation & Usage

### Installation

```bash
# Clone the repository
git clone https://github.com/ppreeper/pgodump.git
cd pgodump

# Build the binary
make build

# Install to $GOPATH/bin
make install
```

### Basic Usage

`pgodump` follows the standard `pg_dump` CLI flag conventions.

```bash
# Dump a schema-only version of a database
./pgodump my_database -h localhost -U my_user -W my_password --schema-only > schema.sql

# Dump specific tables and exclude ownership metadata
./pgodump my_database -t res_users -t res_company --no-owner > odoo_metadata.sql

# Full data and schema dump
./pgodump my_database > backup.sql
```

### Available Flags

| Flag | Description |
| :--- | :--- |
| `-h`, `--host` | Database server host |
| `-p`, `--port` | Database server port |
| `-U`, `--username` | Database user |
| `-W`, `--password` | Database password |
| `-t`, `--table` | Dump only named table(s) |
| `-n`, `--schema` | Dump only named schema(s) |
| `-s`, `--schema-only` | Dump only the schema, no data |
| `-a`, `--data-only` | Dump only the data, no schema |
| `-x`, `--no-privileges` | Do not dump privileges (grant/revoke) |
| `-O`, `--no-owner` | Do not dump ownership (ALTER OWNER) |

---

## Testing

Integration tests use [testcontainers-go](https://golang.testcontainers.org/) to spin up real PostgreSQL instances. Docker (or a compatible runtime) must be available on the host.

### Run all tests

```bash
make test
```

### Run unit tests only (no Docker required)

```bash
make test-unit
```

### Run integration tests only

Integration tests cover inheritance DDL correctness, circular dump/restore fidelity, and PG17 → PG18 migration behaviour.

```bash
make test-integration
```

### Run tests manually with `go test`

```bash
# All tests with race detector
go test -v -race -timeout 120s ./...

# A specific test by name
go test -v -run TestTableInheritance ./database/...
```

### What the tests cover

| Test | Description |
| :--- | :--- |
| `TestTableInheritance` | Validates inheritance DDL: parent-first ordering, local-only columns, `INHERITS (...)` clause, constraint and index generation |
| `TestPG17toPG18Migration` | Loads a PG17 dump into PG18 via pgodump and verifies the output is restorable |

---

## Library Integration

`pgodump` is designed to be easily integrated into other Go projects for automated migrations, backups, or database introspection.

### Import the Package

```go
import "github.com/ppreeper/pgodump/database"
```

### Integration Example

The following example demonstrates how to connect to a database and generate DDL for all public tables.

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ppreeper/pgodump/database"
)

func main() {
	ctx := context.Background()

	// 1. Initialize and configure
	db := database.NewDatabase()
	db.Hostname = "localhost"
	db.Username = "postgres"
	db.Password = "password"
	db.Database = "target_db"

	// 2. Set dump options
	db.SchemaOnly = true
	db.NoOwner = true

	// 3. Open connection
	if err := db.OpenDatabase(); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 4. Fetch and generate DDL
	tables, err := db.GetTables(ctx, "public", "BASE TABLE")
	if err != nil {
		log.Fatal(err)
	}

	for _, t := range tables {
		// GetTableSchema returns full DDL: CREATE TABLE, Constraints, Indexes, ACLs
		ddl, err := db.GetTableSchema(ctx, "public", t.Name)
		if err != nil {
			log.Printf("skipping %s: %v", t.Name, err)
			continue
		}
		fmt.Print(ddl)
	}
}
```

### Core API

All methods accept `context.Context` as the first argument, enabling timeout and cancellation control from the caller.

| Function | Returns | Description |
| :--- | :--- | :--- |
| `GetSchemas(ctx)` | `([]Schema, error)` | All user schemas |
| `GetTables(ctx, schema, kind)` | `([]Table, error)` | Tables or views in a schema |
| `GetTableSchema(ctx, schema, table)` | `(string, error)` | Full DDL for a table |
| `GetTableData(ctx, schema, table)` | `string` | Data in PostgreSQL `COPY` format |
| `GetSequences(ctx, schema)` | `([]Sequence, error)` | Sequences in a schema |
| `GetSequenceDetails(ctx, schema, name)` | `string` | DDL and `setval` for a sequence |
| `GetViewDetails(ctx, schema, view)` | `string` | DDL for a view |
| `QuoteIdentifier(name)` | `string` | Safely double-quotes a SQL identifier |

## License

MIT
