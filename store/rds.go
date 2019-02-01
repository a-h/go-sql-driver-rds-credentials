package store

import (
	"encoding/json"
	"strconv"
	"sync"

	"github.com/go-sql-driver/mysql"
)

// RDS store, backed by AWS Secrets Manager.
type RDS struct {
	child    *Secret
	config   *mysql.Config
	previous string
	m        *sync.Mutex
	dsn      string
}

// NewRDS creates a new RDS store, passing the name of the secret, and a template DSN.
// user:password@tcp(host:port)/dbname?parseTime=true&multiStatements=true&collation=utf8mb4_unicode_ci
func NewRDS(name, dbName string, params map[string]string) *RDS {
	conf := mysql.NewConfig()
	conf.DBName = dbName
	conf.Params = params
	return &RDS{
		child:  New(name),
		config: conf,
		m:      &sync.Mutex{},
	}
}

// Get the secret, optionally forcing a refresh.
func (s *RDS) Get(force bool) (secret string, err error) {
	j, err := s.child.Get(force)
	if err != nil {
		return
	}
	if j == s.previous {
		// Don't bother unmarshalling from JSON if nothing has changed.
		return s.dsn, nil
	}
	var r rdsSecret
	err = json.Unmarshal([]byte(j), &r)
	if err != nil {
		return
	}
	s.previous = j
	// It's changed, so update the cached dsn.
	s.m.Lock()
	defer s.m.Unlock()
	s.config.User = r.Username
	s.config.Passwd = r.Password
	s.config.Net = "tcp"
	s.config.Addr = r.Host + ":" + strconv.Itoa(r.Port)
	s.dsn = s.config.FormatDSN()
	return s.dsn, nil
}

type rdsSecret struct {
	Username            string `json:"username"`
	Password            string `json:"password"`
	Engine              string `json:"engine"`
	Host                string `json:"host"`
	Port                int    `json:"port"`
	DbClusterIdentifier string `json:"dbClusterIdentifier"`
}
