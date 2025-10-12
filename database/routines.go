package database

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	ec "github.com/ppreeper/pgodump/errcheck"
)

//########
// Routines
//########

// Routine list of routines (procedures, functions)
type Routine struct {
	Name             string `db:"ROUTINE_NAME"`
	Type             string `db:"ROUTINE_TYPE"`
	Definition       string `db:"ROUTINE_DEFINITION"`
	DataType         string `db:"DATA_TYPE"`
	ExternalLanguage string `db:"EXTERNAL_LANGUAGE"`
}

type RoutineList struct {
	Name string `db:"ROUTINE_NAME"`
}

// GetRoutines returns list of routines and definitions
func (db *Database) GetRoutines(schema string, timeout int) ([]RoutineList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	q := ""
	q += `SELECT ROUTINE_NAME "ROUTINE_NAME"
		FROM INFORMATION_SCHEMA.ROUTINES
		WHERE ROUTINE_SCHEMA = $1
		AND ROUTINE_DEFINITION IS NOT NULL
		ORDER BY ROUTINE_NAME`
	rr := []RoutineList{}
	if err := db.SelectContext(ctx, &rr, q, schema); err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}
	return rr, nil
}

// GetRoutineSchema returns routine and definition
func (db *Database) GetRoutineSchema(schema, routine string) (Routine, error) {
	q := ""
	q += `SELECT ROUTINE_NAME "ROUTINE_NAME"
		ROUTINE_TYPE "ROUTINE_TYPE"
		ROUTINE_DEFINITION "ROUTINE_DEFINITION"
		CASE WHEN DATA_TYPE IS NULL THEN '' ELSE DATA_TYPE END "DATA_TYPE"
		CASE WHEN EXTERNAL_LANGUAGE IS NULL THEN '' ELSE EXTERNAL_LANGUAGE END "EXTERNAL_LANGUAGE"
		FROM INFORMATION_SCHEMA.ROUTINES
		WHERE ROUTINE_SCHEMA = $1 AND ROUTINE_NAME = $2
		AND ROUTINE_DEFINITION IS NOT NULL
		ORDER BY ROUTINE_NAME`
	rr := Routine{}
	if err := db.Get(&rr, q, schema, routine); err != nil {
		return Routine{}, fmt.Errorf("select: %w", err)
	}
	return rr, nil
}

// GetRoutine gets procedure definition
func (db *Database) GetRoutine(schema string, r Routine, dbg bool) {
	fmt.Printf("\n-- ROUTINE: %s.%s", schema, r.Name)
	q := ""
	if r.Type == "PROCEDURE" {
		q += "DROP " + r.Type + " IF EXISTS " + schema + "." + r.Name + "();\n"
		q += "CREATE OR REPLACE " + r.Type + " " + schema + "." + r.Name + "()\n"
		q += "LANGUAGE sql\n"
		q += "AS $procedure$"
		q += r.Definition
		q += "$procedure$\n;"
	} else if r.Type == "FUNCTION" {
		q += r.Definition
	}

	t := strings.Replace(r.Name, "/", "_", -1)
	fname := fmt.Sprintf("%s.%s.%s.%s.sql", db.Name, schema, t, r.Type)
	f, err := os.Create(fname)
	ec.CheckErr(err)
	defer f.Close()
	f.Write([]byte(q))
}
