package core

import "time"

type AppConfig struct {
	DbUri            string
	HttpUri          string
	GoogleClientID   string
	JwtSecret        string
	TokenTTL         int
	GoogleCertTTL    int
	HttpReadTimeout  time.Duration
	HttpWriteTimeout time.Duration
	ShutdownTimeout  time.Duration
}
