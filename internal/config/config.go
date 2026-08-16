package config

import (
	"os"
	"time"
)

// Config holds all runtime configuration for the notekit server.
type Config struct {
	Addr          string
	AuthToken     string
	DataFile      string
	SaveInterval  time.Duration
	RateLimit     int
	MaxBody       int64
	TitleMaxBytes int
}

// Default returns a config with safe default values.
func Default() Config {
	return Config{
		Addr:          getEnv("NOTEKIT_ADDR", ":8080"),
		AuthToken:     getEnv("NOTEKIT_AUTH_TOKEN", ""),
		DataFile:      getEnv("NOTEKIT_DATA_FILE", ""),
		SaveInterval:  durationEnv("NOTEKIT_SAVE_INTERVAL", 30*time.Second),
		RateLimit:     intEnv("NOTEKIT_RATE_LIMIT", 100),
		MaxBody:       int64Env("NOTEKIT_MAX_BODY", 1<<20),
		TitleMaxBytes: intEnv("NOTEKIT_TITLE_MAX_BYTES", 500),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func intEnv(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		var i int
		for _, c := range v {
			if c >= '0' && c <= '9' {
				i = i*10 + int(c-'0')
			} else {
				return fallback
			}
		}
		return i
	}
	return fallback
}

func int64Env(key string, fallback int64) int64 {
	if v := os.Getenv(key); v != "" {
		var i int64
		for _, c := range v {
			if c >= '0' && c <= '9' {
				i = i*10 + int64(c-'0')
			} else {
				return fallback
			}
		}
		return i
	}
	return fallback
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, err := time.ParseDuration(v)
		if err == nil {
			return d
		}
	}
	return fallback
}
