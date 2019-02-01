package store

import (
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

// Secret store, backed by AWS Secrets Manager.
type Secret struct {
	Name          string
	CacheFor      time.Duration
	LastRefreshed time.Time
	m             *sync.Mutex
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
	}
}

// Get the secret, optionally forcing a refresh.
func (s *Secret) Get(force bool) (secret string, err error) {
	s.m.Lock()
	defer s.m.Unlock()
	if force || time.Now().UTC().After(s.LastRefreshed.Add(s.CacheFor)) {
		svc := secretsmanager.New(session.New())
		input := &secretsmanager.GetSecretValueInput{
			SecretId:     aws.String(s.Name),
			VersionStage: aws.String("AWSCURRENT"),
		}
		var result *secretsmanager.GetSecretValueOutput
		result, err = svc.GetSecretValue(input)
		if err != nil {
			return
		}
		s.callsMade++
		s.Value = *result.SecretString
		s.LastRefreshed = time.Now().UTC()
	}
	return s.Value, nil
}

// CallsMade to the underlying secret API.
func (s *Secret) CallsMade() int {
	return s.callsMade
}
