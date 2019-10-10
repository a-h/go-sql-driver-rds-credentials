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
				// Make sure it's not too quick.
				time.Sleep(time.Nanosecond)
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

func TestRefreshCaching(t *testing.T) {
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
	secret, err = sm.Get()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// If the refresh period hasn't expired, it shouldn't access the secret again.
	secret, err = sm.Refresh(time.Duration(time.Hour))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if secret != "expected_secret" {
		t.Errorf("expected secret '%v', got '%v'", "expected_secret", secret)
	}
	if secretManagerCalls != 1 {
		t.Errorf("expected 1 call to secrets manager, got %v", secretManagerCalls)
	}

	// If the refresh period has expired, it should access the secret again.
	time.Sleep(time.Millisecond * 2)
	secret, err = sm.Refresh(time.Millisecond)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if secret != "expected_secret" {
		t.Errorf("expected secret '%v', got '%v'", "expected_secret", secret)
	}
	if secretManagerCalls != 2 {
		t.Errorf("expected 2 calls to secrets manager, got %v", secretManagerCalls)
	}
}
