# Variables
BINARY_NAME=projecta-web
VERSION?=1.0.0
BUILD_DIR=builds
GOLANG_CROSS_VERSION?=1.26

.PHONY: clean build build-linux dev-up dev-down dev-logs dev-build

clean:
	rm -rf $(BUILD_DIR)

# Optional: for multiple architectures
build-linux:
	cd cmd/web && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-w -s -X main.Version=$(VERSION)" \
		-o ../../$(BUILD_DIR)/$(BINARY_NAME) .

# Build for all platforms
build: clean build-linux

# Docker local development commands
dev-up:
	docker compose up -d --build

dev-down:
	docker compose down

dev-logs:
	docker compose logs -f

dev-build:
	docker compose build