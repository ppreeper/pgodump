package database

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

// Database holds connection config and a live *sqlx.DB.
type Database struct {
	Hostname       string
	Port           int
	Database       string
	Username       string
	Password       string
	URI            string
	SSLMode        string // default "disable"
	Timeout        int
	DataOnly       bool
	SchemaOnly     bool
	IfExists       bool
	File           string
	IncludeSchemas []string
	IncludeTables  []string
	NoOwner        bool
	NoPrivileges   bool
	*sqlx.DB
}

func NewDatabase() *Database {
	return &Database{
		Timeout: 10,
		SSLMode: "disable",
	}
}

func (db *Database) OpenDatabase() error {
	db.getURI()
	conn, err := sqlx.Open("pgx", db.URI)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	if err = conn.Ping(); err != nil {
		_ = conn.Close()
		return fmt.Errorf("ping: %w", err)
	}
	db.DB = conn
	return nil
}

// getURI builds the DSN, percent-encoding credentials to handle special characters.
func (db *Database) getURI() {
	if db.URI != "" {
		return
	}
	port := 5432
	if db.Port != 0 {
		port = db.Port
	}
	sslMode := db.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	u := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(db.Username, db.Password),
		Host:     fmt.Sprintf("%s:%d", db.Hostname, port),
		Path:     db.Database,
		RawQuery: "sslmode=" + sslMode,
	}
	db.URI = u.String()
}

// withTimeout derives a child context with the configured query timeout.
func (db *Database) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, time.Duration(db.Timeout)*time.Second)
}

func header(title string) string {
	// Keep generated SQL metadata comments single-line so object names cannot
	// inject additional SQL/comment lines in dump output.
	title = strings.ReplaceAll(title, "\n", " ")
	title = strings.ReplaceAll(title, "\r", " ")
	// Prevent embedded "--" from starting a new SQL comment within the line.
	title = strings.ReplaceAll(title, "--", "- -")
	return fmt.Sprintf("\n-- %s\n", title)
}
