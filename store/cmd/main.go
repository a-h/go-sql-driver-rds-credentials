package main

import (
	"flag"
	"fmt"

	"github.com/a-h/go-sql-driver-rds-credentials/store"
)

var secretNameFlag = flag.String("secret", "", "The name of the secret.")

func main() {
	flag.Parse()

	fmt.Println("Getting plain secret twice (to validate caching):")
	s := store.New(*secretNameFlag)
	secret, err := s.Get()
	if err != nil {
		panic(err)
	}
	secret, err = s.Get()
	if err != nil {
		panic(err)
	}
	fmt.Println(secret)
	fmt.Println("Calls Made:", s.CallsMade())
	fmt.Println()

	fmt.Println("Getting database secret twice (to validate caching):")
	rds, err := store.NewRDS(*secretNameFlag, "databaseName", map[string]string{
		"parseTime":       "true",
		"multiStatements": "true",
		"collation":       "utf8mb4_unicode_ci",
	})
	if err != nil {
		panic(err)
	}
	secret, err = rds.Get()
	if err != nil {
		panic(err)
	}
	secret, err = rds.Get()
	if err != nil {
		panic(err)
	}
	fmt.Println("Calls Made:", rds.CallsMade())
	fmt.Println(secret)
}
