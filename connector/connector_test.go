package connector

import (
	"context"
	"database/sql/driver"
	"errors"
	"reflect"
	"testing"

	"github.com/go-sql-driver/mysql"
)

func TestDefaultDriver(t *testing.T) {
	d := defaultDriver()
	_, ok := d.(mysql.MySQLDriver)
	if !ok {
		t.Errorf("expected MySQL driver, got: %t", d)
	}
}

func Test(t *testing.T) {
	tests := []struct {
		name                      string
		store                     *mockStore
		driver                    *mockDriver
		expectedErr               error
		expectedStoreGets         int
		expectedStoreGetForced    int
		expectedOpenAttempts      int
		expectedConnectionStrings []string
	}{
		{
			name: "store errors are immediately returned",
			store: &mockStore{
				GetResults: []StoreGetResult{
					StoreGetResult{
						Credential: "",
						Err:        errors.New("failure"),
					},
				},
			},
			driver: &mockDriver{
				GetResults: []DriverGetResult{
					{
						Err: nil,
					},
				},
			},
			expectedErr:               errors.New("failure"),
			expectedStoreGets:         1,
			expectedOpenAttempts:      0,
			expectedConnectionStrings: nil,
		},
		{
			name: "the connector attempts to open the connection using the credential store's connection string",
			store: &mockStore{
				GetResults: []StoreGetResult{
					StoreGetResult{
						Credential: "pharmacy:test@tcp(nowhere.example.com:3306)/testdb?parseTime=true&multiStatements=true&collation=utf8mb4_unicode_ci",
						Err:        nil,
					},
				},
			},
			driver: &mockDriver{
				GetResults: []DriverGetResult{
					{
						Err: nil,
					},
				},
			},
			expectedStoreGets:         1,
			expectedOpenAttempts:      1,
			expectedConnectionStrings: []string{"pharmacy:test@tcp(nowhere.example.com:3306)/testdb?parseTime=true&multiStatements=true&collation=utf8mb4_unicode_ci"},
		},
		{
			name: "errors containing the text 'Error 1045' result in a retry",
			store: &mockStore{
				GetResults: []StoreGetResult{
					{
						Credential: "pharmacy:test@tcp(nowhere.example.com:3306)/testdb?parseTime=true&multiStatements=true&collation=utf8mb4_unicode_ci",
						Err:        nil,
					},
					{
						Credential: "pharmacy:test@tcp(nowhere.example.com:3306)/testdb?parseTime=true&multiStatements=true&collation=utf8mb4_unicode_ci",
						Err:        nil,
					},
				},
			},
			driver: &mockDriver{
				GetResults: []DriverGetResult{
					{
						Err: errors.New("Error 1045"),
					},
					{
						Err: nil,
					},
				},
			},
			expectedStoreGets:    2,
			expectedOpenAttempts: 2,
			expectedConnectionStrings: []string{
				"pharmacy:test@tcp(nowhere.example.com:3306)/testdb?parseTime=true&multiStatements=true&collation=utf8mb4_unicode_ci",
				"pharmacy:test@tcp(nowhere.example.com:3306)/testdb?parseTime=true&multiStatements=true&collation=utf8mb4_unicode_ci"},
		},
		{
			name: "errors containing the text 'Error 1045' result in a retry where the credential is forced to reload. An error retrieving the credential would be returned",
			store: &mockStore{
				GetResults: []StoreGetResult{
					{
						Credential: "pharmacy:test@tcp(nowhere.example.com:3306)/testdb?parseTime=true&multiStatements=true&collation=utf8mb4_unicode_ci",
						Err:        nil,
					},
					{
						Credential: "pharmacy:test@tcp(nowhere.example.com:3306)/testdb?parseTime=true&multiStatements=true&collation=utf8mb4_unicode_ci",
						Err:        errors.New("error getting secret"),
					},
				},
			},
			driver: &mockDriver{
				GetResults: []DriverGetResult{
					{
						Err: errors.New("Error 1045"),
					},
					{
						Err: errors.New("Some other error"),
					},
				},
			},
			expectedStoreGets:      2,
			expectedStoreGetForced: 1,
			expectedOpenAttempts:   1,
			expectedConnectionStrings: []string{
				"pharmacy:test@tcp(nowhere.example.com:3306)/testdb?parseTime=true&multiStatements=true&collation=utf8mb4_unicode_ci",
			},
			expectedErr: errors.New("error getting secret"),
		},
		{
			name: "errors containing the text 'Error 1045' result in a retry, but the second error would be returned",
			store: &mockStore{
				GetResults: []StoreGetResult{
					{
						Credential: "pharmacy:test@tcp(nowhere.example.com:3306)/testdb?parseTime=true&multiStatements=true&collation=utf8mb4_unicode_ci",
						Err:        nil,
					},
					{
						Credential: "pharmacy:test@tcp(nowhere.example.com:3306)/testdb?parseTime=true&multiStatements=true&collation=utf8mb4_unicode_ci",
						Err:        nil,
					},
				},
			},
			driver: &mockDriver{
				GetResults: []DriverGetResult{
					{
						Err: errors.New("Error 1045"),
					},
					{
						Err: errors.New("Some other error"),
					},
				},
			},
			expectedStoreGets:    2,
			expectedOpenAttempts: 2,
			expectedConnectionStrings: []string{
				"pharmacy:test@tcp(nowhere.example.com:3306)/testdb?parseTime=true&multiStatements=true&collation=utf8mb4_unicode_ci",
				"pharmacy:test@tcp(nowhere.example.com:3306)/testdb?parseTime=true&multiStatements=true&collation=utf8mb4_unicode_ci",
			},
			expectedErr: errors.New("Some other error"),
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			c := New(test.store)
			c.d = func() driver.Driver { return test.driver }
			_, err := c.Connect(context.Background())
			if !errorsEqual(test.expectedErr, err) {
				t.Errorf("expected error %v, got: %v", test.expectedErr, err)
			}
			if test.driver.OpenCalls != test.expectedOpenAttempts {
				t.Errorf("expected to connect to the database %d times, got %d", test.expectedOpenAttempts, test.driver.OpenCalls)
			}
			if !reflect.DeepEqual(test.expectedConnectionStrings, test.driver.ConnectionStrings) {
				t.Errorf("expected connection strings %v, got %v", test.expectedConnectionStrings, test.driver.ConnectionStrings)
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

type mockStore struct {
	GetCalls       int
	GetCallsForced int
	GetResults     []StoreGetResult
}

type StoreGetResult struct {
	Credential string
	Err        error
}

func (ms *mockStore) Get(force bool) (credential string, err error) {
	results := ms.GetResults[ms.GetCalls]
	credential, err = results.Credential, results.Err
	ms.GetCalls++
	if force {
		ms.GetCallsForced++
	}
	return
}

type mockDriver struct {
	OpenCalls         int
	GetResults        []DriverGetResult
	ConnectionStrings []string
}

type DriverGetResult struct {
	Conn driver.Conn
	Err  error
}

func (md *mockDriver) Open(dsn string) (conn driver.Conn, err error) {
	md.ConnectionStrings = append(md.ConnectionStrings, dsn)
	results := md.GetResults[md.OpenCalls]
	conn, err = results.Conn, results.Err
	md.OpenCalls++
	return
}
