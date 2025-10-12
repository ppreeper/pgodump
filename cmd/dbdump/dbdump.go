package main

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/ppreeper/pgodump/database"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	// Automatically read environment variables
	viper.AutomaticEnv()

	db := database.NewDatabase()

	cmd := &cobra.Command{
		Use:          "dbdump",
		Short:        "A tool to export PostgreSQL databases.",
		SilenceUsage: false,
		Run: func(cmd *cobra.Command, args []string) {
			// Assign the value retrieved by viper to your App struct's field
			db.Password = viper.GetString("password")

			if db.SchemaOnly && !db.DataOnly {
				fmt.Println("schema-only")
			} else if db.DataOnly && !db.SchemaOnly {
				fmt.Println("data-only")
			} else if db.DataOnly && db.SchemaOnly {
				fmt.Println("cant be both data-only and schema-only")
				os.Exit(1)
			} else if !db.DataOnly && !db.SchemaOnly {
				fmt.Println("full")
			}

			db.OpenDatabase()
			defer db.Close()

			fmt.Println(header("PostgreSQL database dump"))
			// unaccent_schema

			// extensions

			// schemas := db.GetSchemas()

			// for _, schema := range schemas {
			// 	fmt.Println(header("Schema: " + schema.Name))
			// 	tables := db.GetTables(schema.Name, "BASE TABLE")
			// 	for _, table := range tables {
			// 		// table header
			// 		fmt.Println(header("Name: " + table.Name + "; Type: TABLE; Schema: " + schema.Name + "; Owner: -"))
			// 		// table
			// 		sqlc := db.GetTableSchema(schema.Name, table.Name)
			// 		fmt.Println(sqlc)

			// 		pkey := db.GetPKeySQL(schema.Name, table.Name)
			// 		fmt.Println(pkey)

			// 		// table comment
			// 		// table column comment
			// 	}
			// }

			// table header
			// fmt.Println(header("Name: " + "account_account_tag" + "; Type: TABLE; Schema: " + "public" + "; Owner: -"))
			// // table
			// fmt.Println(db.GetTableSchema("public", "account_account_tag"))

			// fmt.Println(db.GetPrimaryKeyConstraint("public", "account_account_tag"))
			// fmt.Println(db.GetUniqueConstraints("public", "account_account_tag"))
			// fmt.Println(db.GetCheckConstraints("public", "account_account_tag"))
			// fmt.Println(db.GetForeignKeyConstraints("public", "account_account_tag"))

			fmt.Println(header("Name: " + "discuss_channel" + "; Type: TABLE; Schema: " + "public" + "; Owner: -"))
			// table
			fmt.Println(db.GetTableSchema("public", "discuss_channel"))

			fmt.Println(db.GetPrimaryKeyConstraint("public", "discuss_channel"))
			fmt.Println(db.GetUniqueConstraints("public", "discuss_channel"))
			fmt.Println(db.GetCheckConstraints("public", "discuss_channel"))
			fmt.Println(db.GetForeignKeyConstraints("public", "discuss_channel"))

			// sequences
			// sequence alter

			// =======
			// get schemas
			// =======

			// s := database.Schema{}
			// sSchemas = []database.Schema{s}

			// for _, s := range sSchemas {

			// 	data := database.Conn{
			// 		Source:  sdb,
			// 		Dest:    ddb,
			// 		SSchema: s.Name,
			// 		DSchema: DSchema,
			// 	}

			// 	getTables(&config, &data)

			// 	getViews(&config, &data)

			// 	getRoutines(&config, &data)

			// 	getIndexes(&config, &data)
			// }
		},
	}

	cmd.Flags().BoolP("help", "", false, "show this help, then exit")

	// General options:
	cmd.Flags().StringVarP(&db.File, "file", "f", "", "output file or directory name")
	cmd.Flags().BoolVarP(&db.Verbose, "verbose", "v", false, "verbose mode")

	// Options controlling the output content:
	cmd.Flags().BoolVarP(&db.DataOnly, "data-only", "a", false, "dump only the data, not the schema or statistics")
	cmd.Flags().BoolVarP(&db.SchemaOnly, "schema-only", "s", false, "dump only the schema, no data or statistics")
	cmd.Flags().BoolVarP(&db.IfExists, "if-exists", "", false, "use IF EXISTS when dropping objects")
	cmd.Flags().BoolVarP(&db.Inserts, "inserts", "", false, "dump data as INSERT commands, rather than COPY")
	cmd.Flags().BoolVarP(&db.NoComments, "no-comments", "", false, "do not dump comment commands")
	cmd.Flags().BoolVarP(&db.OnConflictDoNothing, "on-conflict-do-nothing", "", false, "add ON CONFLICT DO NOTHING to INSERT commands")

	// Connection options:
	cmd.Flags().StringVarP(&db.Database, "dbname", "d", "odoo19_local", "database to dump")
	cmd.Flags().StringVarP(&db.Hostname, "host", "h", "db.local", "database server host or socket directory")
	cmd.Flags().IntVarP(&db.Port, "port", "p", 5432, "database server port number")
	cmd.Flags().StringVarP(&db.Username, "username", "U", "odoo", "connect as specified database user")

	cmd.Flags().StringVar(&db.Password, "password", "odooodoo", "password from PGPASSWORD or env var")
	viper.BindPFlag("password", cmd.Flags().Lookup("password"))
	viper.BindEnv("password", "PGPASSWORD")

	if err := fang.Execute(context.Background(), cmd); err != nil {
		fmt.Println("Error executing command:", err)
		os.Exit(1)
	}
}

func header(title string) string {
	return fmt.Sprintf("\n--\n-- %s\n--\n\n", title)
}

func comment(text string) string {
	return fmt.Sprintf("-- %s\n", text)
}
