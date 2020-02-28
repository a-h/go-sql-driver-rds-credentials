package store

import (
	"sync"
	"time"

	"github.com/a-h/go-sql-driver-rds-credentials/store/sm"
)

// Secret store, backed by AWS Secrets Manager.
type Secret struct {
	Name          string
	CacheFor      time.Duration
	LastRefreshed time.Time
	m             *sync.Mutex
	retrieve      func(name string) (secret string, err error)
	Value         string
	callsMade     int
}

const defaultCacheDuration = time.Hour * 24

// New creates a new store.
func New(name string) *Secret {
	return &Secret{
		Name:          name,
		CacheFor:      defaultCacheDuration,
		LastRefreshed: time.Time{},
		m:             &sync.Mutex{},
		retrieve:      sm.DefaultRetrieve,
	}
}

// Get the secret, optionally forcing a refresh.
func (s *Secret) Get(force bool) (secret string, err error) {
	s.m.Lock()
	defer s.m.Unlock()
	if force || time.Now().UTC().After(s.LastRefreshed.Add(s.CacheFor)) {
		secret, err = s.retrieve(s.Name)
		if err != nil {
			return
		}
		s.callsMade++
		s.Value = secret
		s.LastRefreshed = time.Now().UTC()
	}
	return s.Value, nil
}

// CallsMade to the underlying secret API.
func (s *Secret) CallsMade() int {
	return s.callsMade
}
