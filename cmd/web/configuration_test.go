package main

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Run("loadConfig missing env vars error", func(t *testing.T) {
		os.Unsetenv("DB_URI")
		os.Unsetenv("HTTP_URI")
		os.Unsetenv("GOOGLE_CLIENT_ID")
		os.Unsetenv("GOOGLE_CLIENT_SECRET")
		os.Unsetenv("JWT_SECRET")

		_, err := loadConfig()
		if err == nil {
			t.Errorf("expected error when required env vars are missing")
		}
	})

	t.Run("loadConfig success and custom TTLs", func(t *testing.T) {
		t.Setenv("DB_URI", "postgres://user:pass@localhost:5432/db")
		t.Setenv("HTTP_URI", ":8080")
		t.Setenv("GOOGLE_CLIENT_ID", "client_id")
		t.Setenv("GOOGLE_CLIENT_SECRET", "client_secret")
		t.Setenv("JWT_SECRET", "secret")
		t.Setenv("TOKEN_TTL", "600")
		t.Setenv("GOOGLE_CERT_CACHE_SECONDS_TTL", "3600")

		cfg, err := loadConfig()
		if err != nil || cfg == nil {
			t.Fatalf("unexpected loadConfig error: %v", err)
		}

		if cfg.TokenTTL != 600 || cfg.GoogleCertTTL != 3600 {
			t.Errorf("TTL values mismatch: tokenTTL=%d, certTTL=%d", cfg.TokenTTL, cfg.GoogleCertTTL)
		}
	})

	t.Run("loadConfig invalid TTL env vars", func(t *testing.T) {
		t.Setenv("DB_URI", "postgres://user:pass@localhost:5432/db")
		t.Setenv("HTTP_URI", ":8080")
		t.Setenv("GOOGLE_CLIENT_ID", "client_id")
		t.Setenv("GOOGLE_CLIENT_SECRET", "client_secret")
		t.Setenv("JWT_SECRET", "secret")
		t.Setenv("TOKEN_TTL", "invalid")
		t.Setenv("GOOGLE_CERT_CACHE_SECONDS_TTL", "invalid")

		_, err := loadConfig()
		if err == nil {
			t.Errorf("expected error for invalid TTL format")
		}
	})
}
