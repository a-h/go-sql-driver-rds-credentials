package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/a-h/rotationtest/connector"

	_ "github.com/go-sql-driver/mysql"
)

type fileCredentialStore struct {
	Filename string
}

func (fcs fileCredentialStore) Get(force bool) (credential string, err error) {
	bytes, err := ioutil.ReadFile(fcs.Filename)
	if err != nil {
		return
	}
	fmt.Printf("Credentials read: %v\n", string(bytes))
	return string(bytes), err
}

func main() {
	// CREATE USER 'gotest' IDENTIFIED BY 'first_pwd';
	// GRANT ALL ON test.* TO 'gotest';

	// SET PASSWORD FOR 'gotest' = PASSWORD('first_pwd');
	// SET PASSWORD FOR 'gotest' = PASSWORD('new_pwd');
	fcs := fileCredentialStore{
		Filename: "credentials.txt",
	}
	c := connector.New(fcs)
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

	fmt.Println("Update the gotest user, and update credentials.txt")
	fmt.Println("SET PASSWORD FOR 'gotest' = PASSWORD('new_pwd');")

	// Wait to press enter.
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')

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
