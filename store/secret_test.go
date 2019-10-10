package store

import (
	"errors"
	"testing"
	"time"
)

func TestSecretRetrievalErrors(t *testing.T) {
	sm := New("secret_ARN")
	retrievalError := errors.New("retrieval error")
	sm.retrieve = func(arn string) (secret string, err error) {
		if arn != "secret_ARN" {
			t.Errorf("unexpected ARN: %v", arn)
		}
		return "expected_secret", retrievalError
	}
	_, err := sm.Refresh(time.Second)
	if err != retrievalError {
		t.Errorf("expected err: %v, got: %v", retrievalError, err)
	}
}

func TestSecret(t *testing.T) {
	tests := []struct {
		name           string
		shouldRefresh  []bool
		expectedCalls  int
		expectedSecret string
	}{
		{
			name:           "a single call to secret manager works fine",
			shouldRefresh:  []bool{false},
			expectedCalls:  1,
			expectedSecret: "expected_secret",
		},
		{
			name:           "calls to secret manager are cached",
			shouldRefresh:  []bool{false, false},
			expectedCalls:  1,
			expectedSecret: "expected_secret",
		},
		{
			name:           "calls to secret manager are cached unless force is used",
			shouldRefresh:  []bool{false, true},
			expectedCalls:  2,
			expectedSecret: "expected_secret",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var secretManagerCalls int
			sm := New("secret_ARN")
			sm.retrieve = func(arn string) (secret string, err error) {
				secretManagerCalls++
				if arn != "secret_ARN" {
					t.Errorf("unexpected ARN: %v", arn)
				}
				return "expected_secret", nil
			}
			var secret string
			var err error
			for _, shouldRefresh := range test.shouldRefresh {
				if shouldRefresh {
					secret, err = sm.Refresh(time.Duration(0))
				} else {
					secret, err = sm.Get()
				}
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if secret != test.expectedSecret {
				t.Errorf("expected secret '%v', got '%v'", test.expectedSecret, secret)
			}
			if secretManagerCalls != test.expectedCalls {
				t.Errorf("expected %d calls to secrets manager, got %v", test.expectedCalls, secretManagerCalls)
			}
			if sm.CallsMade() != secretManagerCalls {
				t.Errorf("reporting of secret manager calls, expected: %d, got %d", secretManagerCalls, sm.CallsMade())
			}
		})
	}
}
