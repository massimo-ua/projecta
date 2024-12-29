package main

import (
	"fmt"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"os"
	"strconv"
	"time"
)

const (
	dbUri                            = "DB_URI"
	httpUri                          = "HTTP_URI"
	googleClientID                   = "GOOGLE_CLIENT_ID"
	googleClientSecret               = "GOOGLE_CLIENT_SECRET"
	jwtSecret                        = "JWT_SECRET"
	tokenTTL                         = "TOKEN_TTL"
	googleCertCacheSecondsTTL        = "GOOGLE_CERT_CACHE_SECONDS_TTL"
	defaultTokenTTL                  = 300
	defaultGoogleCertCacheSecondsTTL = 24 * 60 * 60
	defaultHttpReadTimeout           = 30 * time.Second
	defaultHttpWriteTimeout          = 45 * time.Second
	shutdownTimeout                  = 10 * time.Second
)

func loadConfig() (*core.AppConfig, error) {
	config := &core.AppConfig{
		DbUri:              os.Getenv(dbUri),
		HttpUri:            os.Getenv(httpUri),
		GoogleClientID:     os.Getenv(googleClientID),
		GoogleClientSecret: os.Getenv(googleClientSecret),
		JwtSecret:          os.Getenv(jwtSecret),
		HttpReadTimeout:    defaultHttpReadTimeout,
		HttpWriteTimeout:   defaultHttpWriteTimeout,
		ShutdownTimeout:    shutdownTimeout,
	}

	var missingConfigs []string

	if config.DbUri == "" {
		missingConfigs = append(missingConfigs, dbUri)
	}

	if config.HttpUri == "" {
		missingConfigs = append(missingConfigs, httpUri)
	}

	if config.GoogleClientID == "" {
		missingConfigs = append(missingConfigs, googleClientID)
	}

	if config.GoogleClientSecret == "" {
		missingConfigs = append(missingConfigs, googleClientSecret)
	}

	if config.JwtSecret == "" {
		missingConfigs = append(missingConfigs, jwtSecret)
	}

	tokenTTLFromEnv := os.Getenv(tokenTTL)

	if tokenTTLFromEnv == "" {
		config.TokenTTL = defaultTokenTTL
	} else {
		ttl, err := strconv.Atoi(tokenTTLFromEnv)

		if err != nil {
			missingConfigs = append(missingConfigs, tokenTTL)
		} else {
			config.TokenTTL = ttl
		}
	}

	googleCertCacheSecondsTTLFromEnv := os.Getenv(googleCertCacheSecondsTTL)

	if googleCertCacheSecondsTTLFromEnv == "" {
		config.GoogleCertTTL = defaultGoogleCertCacheSecondsTTL
	} else {
		ttl, err := strconv.Atoi(googleCertCacheSecondsTTLFromEnv)

		if err != nil {
			missingConfigs = append(missingConfigs, googleCertCacheSecondsTTL)
		} else {
			config.GoogleCertTTL = ttl
		}
	}

	if len(missingConfigs) > 0 {
		return nil, exceptions.NewInternalException("failed to read configuration", fmt.Errorf("missing required configurations: %v", missingConfigs))
	}

	return config, nil
}
