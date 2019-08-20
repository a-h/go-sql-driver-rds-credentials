package store

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/a-h/go-sql-driver-rds-credentials/store/certs"

	"github.com/go-sql-driver/mysql"
)

type secretGetter interface {
	Get(force bool) (secret string, err error)
	CallsMade() int
}

// RDS store, backed by AWS Secrets Manager.
type RDS struct {
	child    secretGetter
	config   *mysql.Config
	previous string
	m        *sync.Mutex
	dsn      string
}

// NewRDS creates a new RDS store, passing the name of the secret, and a template DSN.
// user:password@tcp(host:port)/dbname?parseTime=true&multiStatements=true&collation=utf8mb4_unicode_ci
func NewRDS(name, dbName string, params map[string]string) (rds *RDS, err error) {
	conf := mysql.NewConfig()
	conf.DBName = dbName
	conf.Params = params

	// Load the TLS certificates.
	var pem []byte
	pem, err = certs.Load()
	if err != nil {
		err = fmt.Errorf("store: could not load certificates: %v", err)
		return
	}
	rcp := x509.NewCertPool()
	if ok := rcp.AppendCertsFromPEM(pem); !ok {
		err = errors.New("store: could not append certificates from PEM")
		return
	}
	mysql.RegisterTLSConfig("rds", &tls.Config{
		RootCAs: rcp,
	})
	conf.Params["tls"] = "rds"

	rds = &RDS{
		child:  New(name),
		config: conf,
		m:      &sync.Mutex{},
	}
	return
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

// CallsMade to the underlying secret API.
func (s *RDS) CallsMade() int {
	return s.child.CallsMade()
}

type rdsSecret struct {
	Username            string `json:"username"`
	Password            string `json:"password"`
	Engine              string `json:"engine"`
	Host                string `json:"host"`
	Port                int    `json:"port"`
	DbClusterIdentifier string `json:"dbClusterIdentifier"`
}
