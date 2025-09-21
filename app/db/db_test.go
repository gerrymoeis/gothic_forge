package db

import (
	"context"
	"os"
	"testing"
)

func TestConnect_NoDatabaseURL(t *testing.T) {
	old := os.Getenv("DATABASE_URL")
	_ = os.Unsetenv("DATABASE_URL")
	defer func() {
		if old != "" {
			os.Setenv("DATABASE_URL", old)
		}
	}()

	p, err := Connect(context.Background())
	if err != nil {
		t.Fatalf("Connect returned error: %v", err)
	}
	if p != nil {
		t.Fatalf("expected nil pool when DATABASE_URL is empty")
	}
}
