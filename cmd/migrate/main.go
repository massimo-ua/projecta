package main

import (
    "fmt"
    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
    "gitlab.com/massimo-ua/projecta/pkg/env"
    "log"
)

const (
    MigrationPath = "MIGRATION_PATH"
    DbUri         = "DB_URI"
)

func main() {
    migrationPath := env.GetEnv(MigrationPath, "migrations")

    dbUri := env.GetEnv(DbUri, "postgres://projecta:projecta@localhost:5432/projecta?sslmode=disable")

    m, err := migrate.New(fmt.Sprintf("file://%s", migrationPath), dbUri)

    if err != nil {
        log.Fatalf("failed to create runner %v", err)
    }

    err = m.Up()

    if err != nil {
        log.Fatalf("failed to run migrations %v", err)
    }
}
