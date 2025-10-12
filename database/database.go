package database

import (
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	ec "github.com/ppreeper/pgodump/errcheck"
)

// Database struct contains sql pointer
type Database struct {
	Name                string
	Hostname            string
	Port                int
	Database            string
	Username            string
	Password            string
	URI                 string
	Timeout             int `default:"10"`
	Verbose             bool
	DataOnly            bool
	SchemaOnly          bool
	IfExists            bool
	Inserts             bool
	NoComments          bool
	OnConflictDoNothing bool
	File                string
	*sqlx.DB
}

func NewDatabase() *Database {
	return &Database{
		Timeout: 10,
	}
}

func (db *Database) OpenDatabase() {
	var err error
	db.getURI()
	db.DB, err = sqlx.Open("pgx", db.URI)
	ec.FatalErr(err, "cannot open database")
	if err = db.Ping(); err != nil {
		ec.FatalErr(err, "cannot ping database")
	}
}

// GenURI generate db uri string
func (db *Database) getURI() {
	port := 5432
	if db.Port != 0 {
		port = db.Port
	}
	db.URI = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", db.Username, db.Password, db.Hostname, port, db.Database)
}

func header(title string) string {
	return fmt.Sprintf("\n--\n-- %s\n--\n\n", title)
}

func comment(text string) string {
	return fmt.Sprintf("-- %s\n", text)
}
