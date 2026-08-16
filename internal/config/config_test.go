package config

import (
	"testing"
	"time"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.Addr == "" {
		t.Error("Addr should not be empty")
	}
	if cfg.SaveInterval <= 0 {
		t.Error("SaveInterval should be positive")
	}
	if cfg.MaxBody <= 0 {
		t.Error("MaxBody should be positive")
	}
}

func TestIntEnv(t *testing.T) {
	t.Setenv("TEST_INT", "42")
	if got := intEnv("TEST_INT", 0); got != 42 {
		t.Errorf("intEnv = %d, want 42", got)
	}
	if got := intEnv("MISSING", 99); got != 99 {
		t.Errorf("intEnv fallback = %d, want 99", got)
	}
}

func TestDurationEnv(t *testing.T) {
	t.Setenv("TEST_DUR", "5s")
	if got := durationEnv("TEST_DUR", 0); got != 5*time.Second {
		t.Errorf("durationEnv = %v, want 5s", got)
	}
	if got := durationEnv("MISSING", 10*time.Second); got != 10*time.Second {
		t.Errorf("durationEnv fallback = %v, want 10s", got)
	}
}
