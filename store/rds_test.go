package store

import (
	"errors"
	"testing"
	"time"
)

func TestRDS(t *testing.T) {
	tests := []struct {
		name                string
		secret              *mockSecret
		previousSecret      string
		expectedSecret      string
		expectedSecretCalls int
		expectedErr         error
	}{
		{
			name: "credential retrieve errors result in an error",
			secret: &mockSecret{
				GetResults: []SecretGetResult{
					{
						Credential: "abc",
						Err:        errors.New("failure"),
					},
				},
			},
			expectedSecret:      "",
			expectedSecretCalls: 1,
			expectedErr:         errors.New("failure"),
		},
		{
			name: "credentials aren't unmarshalled if nothing has changed",
			secret: &mockSecret{
				GetResults: []SecretGetResult{
					{
						Credential: "abc",
						Err:        nil,
					},
				},
			},
			previousSecret:      "abc",
			expectedSecret:      "",
			expectedSecretCalls: 1,
			expectedErr:         nil,
		},
		{
			name: "unmarshalling errors are returned to the client",
			secret: &mockSecret{
				GetResults: []SecretGetResult{
					{
						Credential: "definitely not JSON",
						Err:        nil,
					},
				},
			},
			previousSecret:      "N/A",
			expectedSecret:      "",
			expectedSecretCalls: 1,
			expectedErr:         errors.New("invalid character 'd' looking for beginning of value"),
		},
		{
			name: "RDS data is unmarshalled",
			secret: &mockSecret{
				GetResults: []SecretGetResult{
					{
						Credential: `{ "username": "user", "password": "pwd", "engine": "mysql", "host": "host_name", "port": 3306, "dbClusterIdentifier": "dbcid" }`,
						Err:        nil,
					},
				},
			},
			previousSecret:      "N/A",
			expectedSecret:      "user:pwd@tcp(host_name:3306)/databaseName?collation=utf8mb4_unicode_ci&multiStatements=true&parseTime=true&tls=rds",
			expectedSecretCalls: 1,
			expectedErr:         nil,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			rds, err := NewRDS("secret_ARN", "databaseName", map[string]string{
				"parseTime":       "true",
				"multiStatements": "true",
				"collation":       "utf8mb4_unicode_ci",
			})
			if err != nil {
				t.Fatalf("unepxected error creating RDS: %v", err)
			}
			rds.previous = test.previousSecret
			rds.child = test.secret
			secret, err := rds.Get()
			if !errorsEqual(err, test.expectedErr) {
				t.Fatalf("expected error: %v, got: %v", test.expectedErr, err)
			}
			if secret != test.expectedSecret {
				t.Errorf("expected secret '%v', got '%v'", test.expectedSecret, secret)
			}
			if rds.CallsMade() != test.expectedSecretCalls {
				t.Errorf("expected %d calls, got %d", test.expectedSecretCalls, rds.CallsMade())
			}
		})
	}
}

func errorsEqual(a, b error) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil && b != nil {
		return false
	}
	if a != nil && b == nil {
		return false
	}
	return a.Error() == b.Error()
}

type mockSecret struct {
	GetCalls       int
	GetCallsForced int
	GetResults     []SecretGetResult
}

type SecretGetResult struct {
	Credential string
	Err        error
}

func (ms *mockSecret) Get() (credential string, err error) {
	results := ms.GetResults[ms.GetCalls]
	credential, err = results.Credential, results.Err
	ms.GetCalls++
	return
}

func (ms *mockSecret) Refresh(ifOlderThan time.Duration) (credential string, err error) {
	results := ms.GetResults[ms.GetCalls]
	credential, err = results.Credential, results.Err
	ms.GetCalls++
	ms.GetCallsForced++
	return
}

func (ms *mockSecret) CallsMade() int {
	return ms.GetCalls
}
