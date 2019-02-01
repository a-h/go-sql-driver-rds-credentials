package connector

import (
	"context"
	"database/sql/driver"
	"sync"

	"github.com/go-sql-driver/mysql"
)

// New connector.
func New(dsn string) *Connector {
	return &Connector{
		dsn: dsn,
		d:   &mysql.MySQLDriver{},
		m:   &sync.Mutex{},
	}
}

// Connector to MySQL.
type Connector struct {
	dsn string
	d   *mysql.MySQLDriver
	m   *sync.Mutex
}

// UpdateDSN updates the DSN.
func (c *Connector) UpdateDSN(new string) {
	c.m.Lock()
	defer c.m.Unlock()
	c.dsn = new
}

// Connect implements driver.Connector interface.
// Connect returns a connection to the database.
func (c *Connector) Connect(ctx context.Context) (driver.Conn, error) {
	c.m.Lock()
	defer c.m.Unlock()
	return c.Driver().Open(c.dsn)
}

// Driver implements driver.Connector interface.
// Driver returns &MySQLDriver{}.
func (c *Connector) Driver() driver.Driver {
	return c.d
}
