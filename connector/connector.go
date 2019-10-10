package connector

import (
	"context"
	"database/sql/driver"
	"strings"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
)

// CredentialStore is how credentials can be retrieved.
type CredentialStore interface {
	Get() (credential string, err error)
	Refresh(ifOlderThan time.Duration) (credential string, err error)
}

// New connector.
func New(store CredentialStore) *Connector {
	return &Connector{
		store:            store,
		d:                defaultDriver,
		m:                &sync.Mutex{},
		MaxRatePerSecond: 2,
	}
}

func defaultDriver() driver.Driver {
	return mysql.MySQLDriver{}
}

// Connector to MySQL.
type Connector struct {
	store            CredentialStore
	d                func() driver.Driver
	m                *sync.Mutex
	MaxRatePerSecond int
}

// Connect implements driver.Connector interface.
// Connect returns a connection to the database.
func (c *Connector) Connect(ctx context.Context) (conn driver.Conn, err error) {
	c.m.Lock()
	defer c.m.Unlock()
	creds, err := c.store.Get()
	if err != nil {
		return
	}
	conn, err = c.Driver().Open(creds)
	if err != nil && strings.Contains(err.Error(), "Error 1045") {
		wait := time.Duration(float64(time.Second) / float64(c.MaxRatePerSecond))
		creds, err = c.store.Refresh(wait)
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
