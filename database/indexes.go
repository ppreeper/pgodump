package database

import (
	"context"
	"fmt"
	"time"
)

//########
// Indexes
//########

// Index list of Indexes
type Index struct {
	Schema     string `db:"schemaname"`
	Table      string `db:"tablename"`
	Name       string `db:"indexname"`
	Columns    string `db:"indexcolumns"`
	Definition string `db:"indexdef"`
}

type IndexList struct {
	Name string `db:"indexname"`
}

// GetIndexes returns list of Indexes and definitions
func (db *Database) GetIndexes(schema string, timeout int) ([]IndexList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	q := ""
	q += `select p.indexname from pg_catalog.pg_indexes p
		left join (
			SELECT CONSTRAINT_NAME`
	q += fmt.Sprintf("FROM %s.INFORMATION_SCHEMA.TABLE_CONSTRAINTS", db.Name)
	q += `where CONSTRAINT_TYPE = 'PRIMARY KEY'
		) c on p.indexname = c.constraint_name
		where p.schemaname not in ('information_schema','pg_catalog')
		and c.constraint_name is null
		and p.schemaname = $1`
	vv := []IndexList{}
	if err := db.SelectContext(ctx, &vv, q, schema); err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}
	return vv, nil
}

// GetIndexeschema returns Indexes and definition
func (db *Database) GetIndexSchema(schema, index string) (Index, error) {
	q := ""
	q += `select p.schemaname,p.tablename,p.indexname
		,'"'||replace(replace(split_part(split_part(p.indexdef,'(',2),')',1),'"',''),',','","')||'"' as indexcolumns
		,p.indexdef
		from pg_catalog.pg_indexes p
		left join (
			SELECT CONSTRAINT_NAME`
	q += fmt.Sprintf("FROM %s.INFORMATION_SCHEMA.TABLE_CONSTRAINTS", db.Name)
	q += `) c on p.indexname = c.constraint_name
		where p.schemaname not in ('information_schema','pg_catalog')
		and c.constraint_name is null
		and p.schemaname = $1 and p.indexname = $2
		order by schemaname,tablename`
	vv := Index{}
	if err := db.Get(&vv, q, schema, index); err != nil {
		return Index{}, fmt.Errorf("select: %w", err)
	}
	return vv, nil
}

// GetIndexeschema returns Indexes and definition
func (db *Database) GetTableIndexSchema(schema, table string) ([]Index, error) {
	q := ""
	q += "select p.schemaname,p.tablename,p.indexname" + "\n"
	q += `,'"'||replace(replace(split_part(split_part(p.indexdef,'(',2),')',1),'"',''),',','","')||'"' as indexcolumns` + "\n"
	q += `,p.indexdef` + "\n"
	q += `from pg_catalog.pg_indexes p` + "\n"
	q += `left join (` + "\n"
	q += `SELECT CONSTRAINT_NAME` + "\n"
	q += fmt.Sprintf("FROM %s.INFORMATION_SCHEMA.TABLE_CONSTRAINTS", db.Name) + "\n"
	q += `) c on p.indexname = c.constraint_name` + "\n"
	q += `where p.schemaname not in ('information_schema','pg_catalog')` + "\n"
	q += `and c.constraint_name is null` + "\n"
	q += `and p.schemaname = $1 and p.tablename = $2` + "\n"
	q += `order by schemaname,tablename,indexname`
	vv := []Index{}
	if err := db.Select(&vv, q, schema, table); err != nil {
		return []Index{}, fmt.Errorf("select: %w", err)
	}
	return vv, nil
}

// func (db *Database) GetTableIndexSchema(schema, table string) ([]Index, error) {
// 	q := ""
// 	q += "select p.schemaname,p.tablename,p.indexname" + "\n"
// 	q += `,'"'||replace(replace(split_part(split_part(p.indexdef,'(',2),')',1),'"',''),',','","')||'"' as indexcolumns` + "\n"
// 	q += `,p.indexdef` + "\n"
// 	q += `from pg_catalog.pg_indexes p` + "\n"
// 	q += `left join (` + "\n"
// 	q += `SELECT CONSTRAINT_NAME` + "\n"
// 	q += fmt.Sprintf("FROM %s.INFORMATION_SCHEMA.TABLE_CONSTRAINTS", db.Name) + "\n"
// 	q += `) c on p.indexname = c.constraint_name` + "\n"
// 	q += `where p.schemaname not in ('information_schema','pg_catalog')` + "\n"
// 	q += `and c.constraint_name is null` + "\n"
// 	q += `and p.schemaname = '` + schema + `'` + "\n"
// 	q += `and p.tablename = '` + table + `'` + "\n"
// 	q += `order by schemaname,tablename,indexname`

// 	// fmt.Println(q)
// 	vv := []Index{}
// 	if err := db.Select(&vv, q); err != nil {
// 		return []Index{}, fmt.Errorf("select: %w", err)
// 	}
// 	return vv, nil
// }
