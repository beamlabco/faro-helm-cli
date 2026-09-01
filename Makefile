.PHONY: build build-staging build-prod run run-staging run-prod install clean test fmt lint build-all help

BINARY_NAME=faro
BUILD_DIR=bin

DEV_API_URL=http://localhost:3001
STAGING_API_URL=https://api-faro-helm.beamlab.dev
PROD_API_URL=https://api.helm.farohelm.com

LDFLAGS_DEV=-ldflags "-X main.defaultBaseURL=$(DEV_API_URL)"
LDFLAGS_STAGING=-ldflags "-X main.defaultBaseURL=$(STAGING_API_URL)"
LDFLAGS_PROD=-ldflags "-X main.defaultBaseURL=$(PROD_API_URL)"

build:
	@echo "Building $(BINARY_NAME) (dev)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS_DEV) -o $(BUILD_DIR)/$(BINARY_NAME) cmd/faro/main.go
	@echo "✓ $(BUILD_DIR)/$(BINARY_NAME) → $(DEV_API_URL)"

build-staging:
	@echo "Building $(BINARY_NAME) (staging)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS_STAGING) -o $(BUILD_DIR)/$(BINARY_NAME) cmd/faro/main.go
	@echo "✓ $(BUILD_DIR)/$(BINARY_NAME) → $(STAGING_API_URL)"

build-prod:
	@echo "Building $(BINARY_NAME) (production)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS_PROD) -o $(BUILD_DIR)/$(BINARY_NAME) cmd/faro/main.go
	@echo "✓ $(BUILD_DIR)/$(BINARY_NAME) → $(PROD_API_URL)"

run:
	go run $(LDFLAGS_DEV) cmd/faro/main.go

run-staging:
	go run $(LDFLAGS_STAGING) cmd/faro/main.go

run-prod:
	go run $(LDFLAGS_PROD) cmd/faro/main.go

install: build-prod
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "✓ Installed to /usr/local/bin/$(BINARY_NAME)"

clean:
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME)

test:
	go test -v ./...

fmt:
	go fmt ./...

lint:
	golangci-lint run ./...

build-all:
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS_PROD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64  cmd/faro/main.go
	GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS_PROD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64  cmd/faro/main.go
	GOOS=linux   GOARCH=amd64 go build $(LDFLAGS_PROD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64   cmd/faro/main.go
	GOOS=linux   GOARCH=arm64 go build $(LDFLAGS_PROD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64   cmd/faro/main.go
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS_PROD) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe cmd/faro/main.go
	@echo "✓ All platforms built in $(BUILD_DIR)/"

help:
	@echo "Faro Helm CLI — Makefile commands:"
	@echo "  make build         Build (dev → localhost:3001)"
	@echo "  make build-staging Build (staging → api-faro-helm.beamlab.dev)"
	@echo "  make build-prod    Build (production)"
	@echo "  make run           Run directly (dev)"
	@echo "  make run-staging   Run directly (staging)"
	@echo "  make run-prod      Run directly (production)"
	@echo "  make install       Install to /usr/local/bin"
	@echo "  make build-all     Cross-platform production builds"
	@echo "  make test          Run tests"
	@echo "  make fmt           Format code"
	@echo "  make lint          Run linters"
	@echo "  make clean         Remove build artifacts"
