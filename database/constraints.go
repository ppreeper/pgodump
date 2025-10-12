package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	ec "github.com/ppreeper/pgodump/errcheck"
)

type TableConstraint struct {
	Name string `db:"constraint_name"`
}

func (db *Database) getConstraintName(schema, table, constraint_type string) []TableConstraint {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(db.Timeout)*time.Second)
	defer cancel()
	q := `select tc.constraint_name
		from information_schema.table_constraints tc
		join information_schema.constraint_table_usage ctu
		on tc.table_catalog = ctu.table_catalog
		and tc.table_schema = ctu.table_schema
		and tc.table_name = ctu.table_name
		and tc.constraint_catalog = ctu.constraint_catalog
		and tc.constraint_schema = ctu.constraint_schema
		and tc.constraint_name = ctu.constraint_name
		where tc.table_catalog = $1
		and tc.table_schema = $2
		and tc.table_name = $3
		AND tc.constraint_type = $4;`
	var checkConstraints []TableConstraint
	if err := db.SelectContext(ctx, &checkConstraints, q, db.Database, schema, table, constraint_type); err != nil {
		fmt.Println("Error getting constraints:", err)
		return []TableConstraint{}
	}
	return checkConstraints
}

type ConstraintColumns struct {
	Name string `db:"column_name"`
}

func (db *Database) getConstraintColumns(schema, table, constraintName string) ([]ConstraintColumns, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(db.Timeout)*time.Second)
	defer cancel()
	q := `select c.column_name
		from information_schema.constraint_column_usage c
		join information_schema.columns clm	on
			c.table_catalog = clm.table_catalog and
			c.table_schema = clm.table_schema and
			c.table_name = clm.table_name and
			c.column_name = clm.column_name
		where c.table_catalog = $1
		and c.table_schema = $2
		and c.table_name in ($3)
		and c.constraint_name in ($4)
		order by clm.ordinal_position;`

	var pkey []ConstraintColumns
	if err := db.SelectContext(ctx, &pkey, q, db.Database, schema, table, constraintName); err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}
	return pkey, nil
}

func (db *Database) GetPrimaryKeyConstraint(schema, table string) (sqlc string) {
	constraints := db.getConstraintName(schema, table, "PRIMARY KEY")
	for _, constraint := range constraints {
		columns, err := db.getConstraintColumns(schema, table, constraint.Name)
		ec.CheckErr(err)
		if len(columns) > 0 {
			sqlc += header(fmt.Sprintf("Name: %s; Type: TABLE; Schema: %s; Owner: -", table, schema))
			sqlc += fmt.Sprintf("ALTER TABLE ONLY %s.%s\n", schema, table)
			sqlc += fmt.Sprintf("    ADD CONSTRAINT %s PRIMARY KEY (", constraint.Name)
			pkeys := make([]string, 0, len(columns))
			for _, p := range columns {
				pkeys = append(pkeys, fmt.Sprintf("%s", p.Name))
			}
			sqlc += strings.Join(pkeys, ", ")
			sqlc += ");\n\n"
		}
	}
	return sqlc
}

// UNIQUE CONSTRAINTS
func (db *Database) GetUniqueConstraints(schema, table string) (sqlc string) {
	constraints := db.getConstraintName(schema, table, "UNIQUE")
	for _, constraint := range constraints {
		columns, err := db.getConstraintColumns(schema, table, constraint.Name)
		ec.CheckErr(err)
		if len(columns) > 0 {
			sqlc += header(fmt.Sprintf("Name: %s; Type: CONSTRAINT; Schema: %s; Owner: -", table, schema))
			sqlc += fmt.Sprintf("ALTER TABLE ONLY %s.%s\n", schema, table)
			sqlc += fmt.Sprintf("    ADD CONSTRAINT %s UNIQUE (", constraint.Name)
			pkeys := make([]string, 0, len(columns))
			for _, p := range columns {
				pkeys = append(pkeys, fmt.Sprintf("%s", p.Name))
			}
			sqlc += strings.Join(pkeys, ", ")
			sqlc += ");\n\n"
		}
	}
	return sqlc
}

// CHECK CONSTRAINTS
func (db *Database) getCheckConstraints(schema, table string)  {
}

func (db *Database) GetCheckConstraints(schema, table string) (sqlc string) {
	constraints := db.getConstraintName(schema, table, "CHECK")
	for _, constraint := range constraints {
		columns, err := db.getConstraintColumns(schema, table, constraint.Name)
		ec.CheckErr(err)
		if len(columns) > 0 {
			sqlc += header(fmt.Sprintf("Name: %s %s; Type: CONSTRAINT; Schema: %s; Owner: -", table, constraint.Name, schema))
			sqlc += fmt.Sprintf("ALTER TABLE ONLY %s.%s\n", schema, table)
			sqlc += fmt.Sprintf("    ADD CONSTRAINT %s CHECK (", constraint.Name)
			pkeys := make([]string, 0, len(columns))
			for _, p := range columns {
				pkeys = append(pkeys, fmt.Sprintf("%s", p.Name))
			}
			sqlc += strings.Join(pkeys, ", ")
			sqlc += ");\n\n"
		}
	}
	return sqlc
}

// FOREIGN KEY CONSTRAINTS
func (db *Database) GetForeignKeyConstraints(schema, table string) (sqlc string) {
	constraints := db.getConstraintName(schema, table, "FOREIGN KEY")
	for _, constraint := range constraints {
		columns, err := db.getConstraintColumns(schema, table, constraint.Name)
		ec.CheckErr(err)
		if len(columns) > 0 {
			sqlc += header(fmt.Sprintf("Name: %s %s; Type: FK CONSTRAINT; Schema: %s; Owner: -", table, constraint.Name, schema))
			sqlc += fmt.Sprintf("ALTER TABLE ONLY %s.%s\n", schema, table)
			sqlc += fmt.Sprintf("    ADD CONSTRAINT %s FOREIGN KEY (", constraint.Name)
			pkeys := make([]string, 0, len(columns))
			for _, p := range columns {
				pkeys = append(pkeys, fmt.Sprintf("%s", p.Name))
			}
			sqlc += strings.Join(pkeys, ", ")
			sqlc += ");\n\n"
		}
	}
	return sqlc
}

// --
// -- Name: account_account_account_journal_rel account_account_account_journal_rel_account_journal_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
// --

// ALTER TABLE ONLY public.account_account_account_journal_rel
//     ADD CONSTRAINT account_account_account_journal_rel_account_journal_id_fkey FOREIGN KEY (account_journal_id) REFERENCES public.account_journal(id) ON DELETE CASCADE;
