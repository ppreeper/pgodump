package database

import (
	"strings"

	ec "github.com/ppreeper/pgodump/errcheck"
)

//########
// Generate
//########

// GenTableIndexSQL generate table index sql
func (db *Database) GenTableIndexSQL(schema, table string) (sqld, sqlc string) {
	idxs, err := db.GetTableIndexSchema(schema, table)
	ec.CheckErr(err)
	for _, i := range idxs {
		idx := "\"" + strings.Replace(strings.Replace(i.Table+`_`+i.Columns+"_idx", "\"", "", -1), ",", "_", -1) + "\""
		exists := ""
		notexists := ""
		exists = "IF EXISTS "
		notexists = "IF NOT EXISTS "
		sqld += `DROP INDEX ` + exists + `"` + schema + `".` + idx + `;` + "\n"
		sqlc += `CREATE INDEX ` + notexists + `` + idx + ` ON "` + schema + `"."` + i.Table + `" (` + i.Columns + `);` + "\n"
	}
	return
}
