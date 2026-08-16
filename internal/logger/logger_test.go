package logger

import (
	"strings"
	"testing"
)

func TestParseLevel(t *testing.T) {
	cases := []struct {
		input string
		want  Level
	}{
		{"debug", LevelDebug}, {"INFO", LevelInfo}, {"warn", LevelWarn}, {"warning", LevelWarn}, {"error", LevelError}, {"bogus", LevelInfo},
	}
	for _, c := range cases {
		if got := ParseLevel(c.input); got != c.want {
			t.Errorf("ParseLevel(%q) = %v, want %v", c.input, got, c.want)
		}
	}
}

func TestLoggerFilters(t *testing.T) {
	var buf strings.Builder
	log := New(&buf, LevelWarn)
	log.Debug("debug")
	log.Info("info")
	log.Warn("warn")
	log.Error("error")
	out := buf.String()
	if strings.Contains(out, "debug") || strings.Contains(out, "info") {
		t.Error("debug/info should be filtered")
	}
	if !strings.Contains(out, "warn") || !strings.Contains(out, "error") {
		t.Error("warn/error should be emitted")
	}
}

func TestLoggerWithFields(t *testing.T) {
	var buf strings.Builder
	log := New(&buf, LevelInfo).With("service", "notekit")
	log.Info("started", "port", "8080")
	out := buf.String()
	if !strings.Contains(out, "service=notekit") {
		t.Errorf("missing parent field: %q", out)
	}
	if !strings.Contains(out, "port=8080") {
		t.Errorf("missing call field: %q", out)
	}
}

func TestLevelString(t *testing.T) {
	if LevelInfo.String() != "INFO" {
		t.Errorf("LevelInfo.String() = %q", LevelInfo.String())
	}
}
