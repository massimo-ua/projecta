# Variables
BINARY_NAME=projecta-web
VERSION?=1.0.0
BUILD_DIR=builds
GOLANG_CROSS_VERSION?=1.26

.PHONY: clean build build-linux dev-up dev-down dev-logs dev-build test-unit test-integration

clean:
	rm -rf $(BUILD_DIR)

# Optional: for multiple architectures
build-linux:
	cd cmd/web && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-w -s -X main.Version=$(VERSION)" \
		-o ../../$(BUILD_DIR)/$(BINARY_NAME) .

# Build for all platforms
build: clean build-linux

# Unit tests (no external dependencies required)
test-unit:
	go test -race -count=1 ./internal/... ./pkg/... ./cmd/...

# Integration & HTTP component tests against an isolated Postgres container
TEST_DB_URI ?= postgres://projecta_test:projecta_test@localhost:5434/projecta_test?sslmode=disable
test-integration:
	@set -e; \
	trap 'docker compose -f docker-compose.test.yml down -v --remove-orphans' EXIT; \
	docker compose -f docker-compose.test.yml up -d --wait; \
	TEST_DB_URI="$(TEST_DB_URI)" go test -v -race -count=1 ./pkg/dal/... ./cmd/web/...

# Docker local development commands
dev-up:
	docker compose up -d --build

dev-down:
	docker compose down

dev-logs:
	docker compose logs -f

dev-build:
	docker compose build