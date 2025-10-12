package database

import (
	"context"
	"fmt"
	"time"
)

//########
// Schemas
//########

// Schema struct to hold schemas
type Schema struct {
	Name string `db:"schema_name"`
}

// GetSchemas returns schema list
func (db *Database) GetSchemas() []Schema {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(db.Timeout)*time.Second)
	defer cancel()
	q := `select schema_name
		from information_schema.schemata
		where schema_name not in ('pg_catalog','information_schema')
		order by schema_name`
	fmt.Println(q)
	ss := []Schema{}
	if err := db.SelectContext(ctx, &ss, q); err != nil {
		fmt.Println("Error getting schemas:", err)
		return ss
	}
	return ss
}
