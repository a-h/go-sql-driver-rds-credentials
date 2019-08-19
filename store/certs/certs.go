package certs

import (
	"io/ioutil"

	"github.com/rakyll/statik/fs"

	// Used to import data with statik.
	_ "github.com/a-h/go-sql-driver-rds-credentials/store/certs/statik"
)

// Load the certificates.
func Load() (certs []byte, err error) {
	sfs, err := fs.New()
	if err != nil {
		return
	}
	r, err := sfs.Open("/rds-ca.pem")
	if err != nil {
		return
	}
	defer r.Close()
	return ioutil.ReadAll(r)
}
