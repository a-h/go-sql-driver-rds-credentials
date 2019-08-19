package certs

import (
	"testing"
)

func TestLoad(t *testing.T) {
	bytes, err := Load()
	if err != nil {
		t.Fatalf("failed to load certs: %v", err)
	}
	if len(bytes) == 0 {
		t.Fatal("failed to read all cert data")
	}
}
