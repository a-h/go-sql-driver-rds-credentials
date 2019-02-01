package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/a-h/rotationtest/connector"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// CREATE USER 'gotest' IDENTIFIED BY 'first_pwd';
	// GRANT ALL ON test.* TO 'gotest';

	// SET PASSWORD FOR 'gotest' = PASSWORD('first_pwd');
	// SET PASSWORD FOR 'gotest' = PASSWORD('new_pwd');
	c := connector.New("gotest:first_pwd@tcp(localhost:3306)/test?parseTime=true&multiStatements=true")
	db := sql.OpenDB(c)
	err := db.Ping()
	if err != nil {
		panic("can't ping DB: " + err.Error())
	}
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)

	// The ping initiated a connection, let's check it again.
	if err := run(db); err != nil {
		fmt.Println("ERROR running first test:", err)
	}

	fmt.Println("SET PASSWORD FOR 'gotest' = PASSWORD('new_pwd');")

	// Wait to press enter.
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')

	c.UpdateDSN("gotest:new_pwd@tcp(localhost:3306)/test?parseTime=true&multiStatements=true")

	// Run loads more tests.
	for i := 0; i < 5; i++ {
		go func() {
			for {
				if err := run(db); err != nil {
					fmt.Println("ERROR running test:", err)
				}
				time.Sleep(time.Millisecond * 50)
			}
		}()
	}

	reader.ReadString('\n')
}

func run(db *sql.DB) error {
	_, err := db.Exec(`INSERT INTO table1 (s) VALUES ('a');`)
	if err != nil {
		return fmt.Errorf("insert err %v", err)
	}

	_, err = db.Exec(`DELETE FROM table1 WHERE s = 'a';`)
	if err != nil {
		return fmt.Errorf("delete err %v", err)
	}

	return nil
}
