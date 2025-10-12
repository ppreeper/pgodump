package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	ec "github.com/ppreeper/pgodump/errcheck"
)

// Table list of tables
type Table struct {
	Name string `db:"table_name"`
}

func (db *Database) GetTables(schema, table_type string) []Table {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(db.Timeout)*time.Second)
	defer cancel()
	q := `select table_name
		from information_schema.tables
		where table_schema = $1 and table_type = $2
		order by table_name`
	tt := []Table{}
	if err := db.SelectContext(ctx, &tt, q, schema, table_type); err != nil {
		return tt
	}
	return tt
}

type TableComment struct {
	Comment string `db:"table_comment"`
}

func (db *Database) GetTableComment(schema, table string) string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(db.Timeout)*time.Second)
	defer cancel()
	q := `select obj_description(($1||'.'||$2)::regclass,'pg_class') as table_comment`
	tt := []TableComment{}
	if err := db.SelectContext(ctx, &tt, q, schema, table); err != nil {
		return ""
	}
	return tt[0].Comment
}

// ########
// Table Columns
// ########
type Column struct {
	ColumnName    string `db:"column_name"`
	IsNullable    string `db:"is_nullable"`
	ColumnDefault string `db:"column_default"`
	DataType      string `db:"data_type"`
	Comment       string `db:"column_comment"`
}

func (db *Database) GetColumnDetail(schema, table string) ([]Column, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(db.Timeout)*time.Second)
	defer cancel()
	q := `select c.column_name
		,case when c.is_nullable = 'NO' then 'NOT NULL' else '' end AS is_nullable
		,COALESCE (C.COLUMN_DEFAULT,'') AS column_default
		,CASE UPPER(DATA_TYPE)
			WHEN 'CHAR' THEN 'CHAR' || CASE WHEN C.CHARACTER_MAXIMUM_LENGTH::character varying IS NULL THEN '' ELSE '(' || C.CHARACTER_MAXIMUM_LENGTH::character varying || ')' END
			WHEN 'NCHAR' THEN 'CHAR' || CASE WHEN C.CHARACTER_MAXIMUM_LENGTH::character varying IS NULL THEN '' ELSE '(' || C.CHARACTER_MAXIMUM_LENGTH::character varying || ')' END
			WHEN 'VARCHAR' THEN 'VARCHAR' || CASE WHEN C.CHARACTER_MAXIMUM_LENGTH::character varying IS NULL THEN '' ELSE '(' || C.CHARACTER_MAXIMUM_LENGTH::character varying || ')' END
			WHEN 'NVARCHAR' THEN 'VARCHAR' || CASE WHEN C.CHARACTER_MAXIMUM_LENGTH::character varying IS NULL THEN '' ELSE '(' || C.CHARACTER_MAXIMUM_LENGTH::character varying || ')' END
			WHEN 'CHARACTER' THEN 'CHARACTER' || CASE WHEN C.CHARACTER_MAXIMUM_LENGTH::character varying IS NULL THEN '' ELSE '(' || C.CHARACTER_MAXIMUM_LENGTH::character varying || ')' END
			WHEN 'CHARACTER VARYING' THEN 'CHARACTER VARYING' || CASE WHEN C.CHARACTER_MAXIMUM_LENGTH::character varying IS NULL THEN '' ELSE '(' || C.CHARACTER_MAXIMUM_LENGTH::character varying || ')' END
			WHEN 'TINYINT' THEN 'TINYINT'
			WHEN 'SMALLINT' THEN 'SMALLINT'
			WHEN 'INT' THEN 'INT'
			WHEN 'DECIMAL' THEN 'DECIMAL' || case when C.NUMERIC_PRECISION::character varying IS NULL THEN '' ELSE '(' || C.NUMERIC_PRECISION::character varying || ',' || C.NUMERIC_SCALE::character varying || ')' END
			WHEN 'NUMERIC' THEN 'NUMERIC' || case when C.NUMERIC_PRECISION::character varying IS NULL THEN '' ELSE '(' || C.NUMERIC_PRECISION::character varying || ',' || C.NUMERIC_SCALE::character varying || ')' END
			WHEN 'FLOAT' THEN 'FLOAT' || CASE WHEN C.NUMERIC_PRECISION < 53 THEN '(' || C.NUMERIC_PRECISION::character varying || ')' ELSE '' END
			WHEN 'VARBINARY' THEN 'VARBINARY'
			WHEN 'DATETIME' THEN 'DATETIME'
			ELSE UPPER(DATA_TYPE) END AS data_type
		,coalesce(col_description((c.table_schema||'.'||c.table_name)::regclass,c.ordinal_position),'') AS column_comment
		FROM INFORMATION_SCHEMA.COLUMNS C
		WHERE C.TABLE_CATALOG = $1 AND C.TABLE_SCHEMA = $2 AND C.TABLE_NAME = $3
		ORDER BY C.TABLE_CATALOG,C.TABLE_SCHEMA,C.TABLE_NAME,ORDINAL_POSITION;`

	columnnames := []Column{}
	if err := db.SelectContext(ctx, &columnnames, q, db.Database, schema, table); err != nil {
		return nil, fmt.Errorf("select: %v", err)
	}
	return columnnames, nil
}

// GenTable generate table creation
func (db *Database) GenTable(schema, table string, cols []Column) (sqlc string) {
	if db.IfExists {
		sqlc += fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s (\n", schema, table)
	} else {
		sqlc += fmt.Sprintf("CREATE TABLE %s.%s (\n", schema, table)
	}

	clen := len(cols)
	for k, c := range cols {
		cdefault := ""
		if c.ColumnDefault != "" {
			cdefault += "DEFAULT " + strings.ReplaceAll(c.ColumnDefault, "getdate()", "CURRENT_TIMESTAMP")
		}
		sqlc += "    " + strings.TrimSpace(strings.Join([]string{c.ColumnName, c.DataType, c.IsNullable, cdefault}, " "))
		if k == clen-1 {
			sqlc += "\n"
		} else {
			sqlc += ",\n"
		}
	}
	sqlc += ");\n"

	return
}

func (db *Database) GenTableDrop(schema, table string, cols []Column) (sqld string) {
	if db.IfExists {
		sqld += fmt.Sprintf("\n-- DROP TABLE IF EXISTS %s.%s CASCADE;", schema, table)
	} else {
		sqld += fmt.Sprintf("\n-- DROP TABLE %s.%s CASCADE;", schema, table)
	}
	return
}

// GetTableSchema gets table definition
func (db *Database) GetTableSchema(schema, table string) (sqlc string) {
	scols, err := db.GetColumnDetail(schema, table)
	ec.CheckErr(err)
	// pcols, err := db.GetPKey(schema, table)
	// ec.CheckErr(err)
	sqlc = db.GenTable(schema, table, scols)
	tblc := db.GetTableComment(schema, table)
	if tblc != "" {
		sqlc += header(fmt.Sprintf("Name: TABLE %s; Type: COMMENT; Schema: %s; Owner: -", table, schema))
		sqlc += fmt.Sprintf("COMMENT ON TABLE %s.%s IS '%s';\n", schema, table, strings.ReplaceAll(tblc, "'", "''"))
	}
	for _, c := range scols {
		if c.Comment != "" {
			sqlc += header(fmt.Sprintf("Name: %s.%s; Type: COMMENT; Schema: %s; Owner: -", table, c.ColumnName, schema))
			sqlc += fmt.Sprintf("COMMENT ON COLUMN %s.%s.%s IS '%s';\n", schema, table, c.ColumnName, strings.ReplaceAll(c.Comment, "'", "''"))
		}
	}
	return
}
