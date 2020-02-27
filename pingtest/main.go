package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	"github.com/marrickmedical/go-sql-driver-rds-credentials/connector"
	"github.com/marrickmedical/go-sql-driver-rds-credentials/store"
)

var secretARNFlag = flag.String("secret", "", "The name of the secret.")
var databaseNameFlag = flag.String("dbName", "", "The name of the database to connect to.")

func main() {
	flag.Parse()
	fmt.Println("Getting secret.")
	err := pingDB(*secretARNFlag, *databaseNameFlag)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	fmt.Println("OK")
}

func pingDB(secretARN, databaseName string) error {
	s, err := store.NewRDS(secretARN, databaseName, map[string]string{
		"parseTime":       "true",
		"multiStatements": "true",
		"collation":       "utf8mb4_unicode_ci",
	})
	if err != nil {
		return fmt.Errorf("pingDB: failed to create store: %v", err)
	}
	c := connector.New(s)
	db := sql.OpenDB(c)
	return db.Ping()
}
