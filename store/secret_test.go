package store

import (
	"errors"
	"testing"
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
	_, err := sm.Get(true)
	if err != retrievalError {
		t.Errorf("expected err: %v, got: %v", retrievalError, err)
	}
}

func TestSecret(t *testing.T) {
	tests := []struct {
		name           string
		getParameters  []bool
		expectedCalls  int
		expectedSecret string
	}{
		{
			name:           "a single call to secret manager works fine",
			getParameters:  []bool{false},
			expectedCalls:  1,
			expectedSecret: "expected_secret",
		},
		{
			name:           "calls to secret manager are cached",
			getParameters:  []bool{false, false},
			expectedCalls:  1,
			expectedSecret: "expected_secret",
		},
		{
			name:           "calls to secret manager are cached unless force is used",
			getParameters:  []bool{false, true},
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
			for _, force := range test.getParameters {
				secret, err = sm.Get(force)
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
