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
// Views
//########

// View list of views
type View struct {
	Name       string `db:"TABLE_NAME"`
	Definition string `db:"VIEW_DEFINITION"`
}

type ViewList struct {
	Name string `db:"TABLE_NAME"`
}

// GetViews returns list of views and definitions
func (db *Database) GetViews(schema string, timeout int) ([]ViewList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	q := ""
	q += `SELECT TABLE_NAME AS "TABLE_NAME"
		FROM INFORMATION_SCHEMA.VIEWS
		WHERE TABLE_SCHEMA = $1
		ORDER BY TABLE_NAME`

	vv := []ViewList{}
	if err := db.SelectContext(ctx, &vv, q, schema); err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}
	return vv, nil
}

// GetViewSchema returns views and definition
func (db *Database) GetViewSchema(schema, view string) (View, error) {
	q := ""
	q += `SELECT TABLE_NAME AS "TABLE_NAME", VIEW_DEFINITION AS "VIEW_DEFINITION"
		FROM INFORMATION_SCHEMA.VIEWS
		WHERE TABLE_SCHEMA = $1 AND TABLE_NAME = $2
		ORDER BY TABLE_NAME`
	vv := View{}
	if err := db.Get(&vv, q, schema, view); err != nil {
		return View{}, fmt.Errorf("select: %w", err)
	}
	return vv, nil
}

// GetView gets view definition
func (db *Database) GetView(d Database, schema string, view View, dbg bool) {
	q := ""
	q += "DROP VIEW " + schema + "." + view.Name + ";\n"
	q += "CREATE VIEW " + schema + "." + view.Name + " AS \n"
	q += view.Definition

	t := strings.Replace(view.Name, "/", "_", -1)
	fname := fmt.Sprintf("%s.%s.%s.VIEW.sql", d.Database, schema, t)
	f, err := os.Create(fname)
	ec.CheckErr(err)
	defer f.Close()
	f.Write([]byte(q))
}
