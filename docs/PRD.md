# PRD: pgodump v1.0 (Postgres 18 Fix Edition)

## Goal
Implement a 1:1 `pg_dump` CLI replacement in Go that correctly handles PostgreSQL inheritance, specifically targeting issues observed in PG 18.

## Requirements

### 1. CLI Interface (Parity)
- Use `spf13/cobra` and `spf13/pflag`.
- Support standard `pg_dump` flags:
  - `-h`, `--host`
  - `-p`, `--port`
  - `-U`, `--username`
  - `-d`, `--dbname`
  - `-t`, `--table`
  - `-n`, `--schema`
  - `-s`, `--schema-only`
  - `-a`, `--data-only`
  - `-f`, `--file`

### 2. Database Logic (Postgres 18 Inheritance)
- Use `pg_catalog` queries instead of `information_schema`.
- Correctly identify inherited columns using `pg_attribute.attislocal`.
- Support `CREATE TABLE ... INHERITS (...)` syntax.
- Ensure constraints (`pg_constraint`) are correctly attributed to local or inherited tables.

### 3. Output Format
- Plain SQL text (default).
- Sanitized output (strip volatile timestamps, OIDs, and search_path noise).

### 4. Testing Strategy (TDD)
- **Integration Tests:** Use `testcontainers-go` for PostgreSQL 17 and 18.
- **Golden File Comparison:** 
  - Compare `pgodump` output against a ground-truth "Golden" SQL file.
  - Automate normalization of volatile metadata before comparison.
- **Red-Green Loop:** Each feature/bug-fix starts with a failing integration test.

## Technical Stack
- Language: Go 1.22+
- CLI: Cobra
- SQL: `sqlx`, `pgx`
- Testing: `testcontainers-go`, `testify`
