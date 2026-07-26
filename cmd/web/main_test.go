package main

import (
	"testing"
	"time"

	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/pkg/dal"
)

func TestSetupAppHandler(t *testing.T) {
	cfg := &core.AppConfig{
		DbUri:              "postgres://user:pass@localhost:5432/db",
		HttpUri:            ":8080",
		GoogleClientID:     "client_id",
		GoogleClientSecret: "client_secret",
		JwtSecret:          "secret",
		TokenTTL:           300,
		GoogleCertTTL:      3600,
		HttpReadTimeout:    30 * time.Second,
		HttpWriteTimeout:   45 * time.Second,
		ShutdownTimeout:    10 * time.Second,
	}

	db, _ := dal.NewPgDbConnection("postgres://user:pass@localhost:5432/db")

	handler, err := setupAppHandler(cfg, db)
	if err != nil || handler == nil {
		t.Fatalf("unexpected error setting up app handler: %v", err)
	}
}
