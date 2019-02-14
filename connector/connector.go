package connector

import (
	"context"
	"database/sql/driver"
	"strings"
	"sync"

	"github.com/go-sql-driver/mysql"
)

// CredentialStore is how credentials can be retrieved.
type CredentialStore interface {
	Get(force bool) (credential string, err error)
}

// New connector.
func New(store CredentialStore) *Connector {
	return &Connector{
		store: store,
		d:     defaultDriver,
		m:     &sync.Mutex{},
	}
}

func defaultDriver() driver.Driver {
	return mysql.MySQLDriver{}
}

// Connector to MySQL.
type Connector struct {
	store CredentialStore
	d     func() driver.Driver
	m     *sync.Mutex
}

// Connect implements driver.Connector interface.
// Connect returns a connection to the database.
func (c *Connector) Connect(ctx context.Context) (conn driver.Conn, err error) {
	c.m.Lock()
	defer c.m.Unlock()
	creds, err := c.store.Get(false)
	if err != nil {
		return
	}
	conn, err = c.Driver().Open(creds)
	if err != nil && strings.Contains(err.Error(), "Error 1045") {
		creds, err = c.store.Get(true)
		if err != nil {
			return
		}
		conn, err = c.Driver().Open(creds)
	}
	return
}

// Driver implements driver.Connector interface.
// Driver returns &MySQLDriver{}.
func (c *Connector) Driver() driver.Driver {
	return c.d()
}
