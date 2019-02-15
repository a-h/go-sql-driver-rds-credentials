# go-sql-driver-rds-credentials

Use AWS Secrets Manager with Go for automated database credential rotation without downtime.

Changing database credentials while applications are running isn't possible when using `db.Open` because the DSN (connection string) is set as a parameter, and there's no way to change it.

Using `db.OpenDB` allows this, but requires the `driver.Connector` interface to be implemented within the driver. The Go MySQL driver doesn't have it yet (https://github.com/go-sql-driver/mysql/issues/671), but the issue points out that it may be possible without changes to the library (by making an external implementation).

This repo implements that connector, and demonstrates combining that with AWS Secrets Manager to allow automatic rotation of AWS RDS connection details from Go programs without having to shut down your programs.

# Usage

```go
import (
  "github.com/a-h/go-sql-driver-rds-credentials/store"
  "github.com/a-h/go-sql-driver-rds-credentials/connector"
)

func main() {
	s := store.NewRDS("secret_ARN", "databaseName", map[string]string{
		"parseTime":       "true",
		"multiStatements": "true",
		"collation":       "utf8mb4_unicode_ci",
	})
	c := connector.New(s)
  db := sql.OpenDB(c)
  err := db.Ping()
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	fmt.Println("OK")
}
```

# Structure

* /connector
  * See `/test/main.go` for an example which uses the connector instead of passing a DSN directly to `db.Open`.
* /store
  * Uses the AWS SDK to load secrets and to cache them locally as per the Java example provided by AWS. It also unmarshals the RDS secrets stored in AWS Secrets Manager back into a DSN for use with the Go MySQL driver.
  * The contents of the `cmd` directory contain an example of retrieving secrets from AWS.
* /test
  * Contains an example of connecting to MySQL using the connector, but with a file-based implementation of the credential store (instead of using the AWS SDK).

# Manual testing

* `cd test`
* Run `docker-compose up` to start a local MySQL database.
* It will create a user called `gotest` initialised to password `first_pwd`.
* Run `main.go`
* The connection will be made and data inserted as the first test. This will be successful.
* The code then pauses and waits for enter to be pressed. At this point, you can change the user's password using MySQL workbench or CLI.
  * If the user's password is not changed, then the subsequent code fails to logon.
  * If the user's password is then changed (`SET PASSWORD FOR 'gotest' = PASSWORD('new_pwd');`), the subsequent code works fine. Hit Ctrl-C or press enter to exit the program.
