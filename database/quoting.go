package database

import (
	"fmt"
	"regexp"
	"strings"
)

func QuoteIdentifier(name string) string {
	// Already quoted — return as-is.
	if strings.HasPrefix(name, "\"") && strings.HasSuffix(name, "\"") {
		return name
	}
	// Only split on dot for unquoted schema-qualified names.
	if strings.Contains(name, ".") {
		parts := strings.SplitN(name, ".", 2)
		return "\"" + strings.ReplaceAll(parts[0], "\"", "\"\"") + "\".\"" + strings.ReplaceAll(parts[1], "\"", "\"\"") + "\""
	}
	// Per SQL standard, embedded " must be escaped as "".
	return "\"" + strings.ReplaceAll(name, "\"", "\"\"") + "\""
}

var (
	fkRegex = regexp.MustCompile(`(?i)FOREIGN KEY \((.*?)\) REFERENCES (.*?)(\((.*?)\))`)
	// idxRegex matches quoted or unquoted identifiers (schema optional) in CREATE INDEX statements.
	idxRegex = regexp.MustCompile(`(?i)^CREATE (UNIQUE )?INDEX ("(?:[^"]|"")*"|\S+) ON ((?:"(?:[^"]|"")*"|\w+)\.)?("(?:[^"]|"")*"|\S+)`)
)

func QuoteConstraint(def string, contype string) string {
	if contype != "f" {
		return def
	}
	return fkRegex.ReplaceAllStringFunc(def, func(m string) string {
		parts := fkRegex.FindStringSubmatch(m)
		cols := strings.Split(parts[1], ",")
		for i, c := range cols {
			cols[i] = QuoteIdentifier(strings.TrimSpace(c))
		}
		refTable := QuoteIdentifier(parts[2])
		refCols := strings.Split(parts[4], ",")
		for i, c := range refCols {
			refCols[i] = QuoteIdentifier(strings.TrimSpace(c))
		}
		return fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s(%s)",
			strings.Join(cols, ", "), refTable, strings.Join(refCols, ", "))
	})
}

func QuoteIndex(def string) string {
	match := idxRegex.FindStringSubmatch(def)
	if match == nil {
		return def
	}

	unique := match[1]
	idxName := match[2]
	schemaPrefix := match[3]
	tblName := match[4]

	fullTable := tblName
	if schemaPrefix != "" {
		fullTable = schemaPrefix + tblName
	}

	m := match[0]
	rest := def[len(m):]

	return fmt.Sprintf("CREATE %sINDEX %s ON %s%s",
		unique, QuoteIdentifier(idxName), QuoteIdentifier(fullTable), rest)
}
