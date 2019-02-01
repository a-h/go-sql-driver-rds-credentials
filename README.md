# Example of database credential rotation

# Problem

Changing database credentials while applications are running isn't possible when using `db.Open` because the DSN (connection string) is set as a parameter, and there's no way to change it.

Using `db.OpenDB` allows this, but requires the `driver.Connector` interface to be implemented within the driver. The Go MySQL driver doesn't have it yet (https://github.com/go-sql-driver/mysql/issues/671), but the issue points out that it may be possible without changes to the library (by making an external implementation).

This repo tests whether that's possible and provides an example `driver.Connector`.

# Testing steps

* Run `docker-compose up` to start a local MySQL database.
* It will create a user called `gotest` initialised to password `first_pwd`.
* Run `main.go`
* The connection will be made and data inserted as the first test. This will be successful.
* The code then pauses and waits for enter to be pressed. At this point, you can change the user's password using MySQL workbench or CLI.
  * If the user's password is not changed, then the subsequent code fails to logon.
  * If the user's password is then changed (`SET PASSWORD FOR 'gotest' = PASSWORD('new_pwd');`), the subsequent code works perfectly.

# Next steps

* Check with developers of the MySQL driver that this is a solid implementation outside of the library.
* Update the cached credentials when connection failures occur (for when a credential has been updated, but the library doesn't know yet).