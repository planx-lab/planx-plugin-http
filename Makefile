.PHONY: all build clean test lint package

PLUGIN_NAME := plugin-http
VERSION := 0.0.0
GIT_VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo $(VERSION))
COMMIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

LDFLAGS := -X 'main.Version=$(GIT_VERSION)' \
           -X 'main.Commit=$(COMMIT_HASH)' \
           -X 'main.BuildTime=$(BUILD_TIME)'

DIST_DIR := dist

all: build

build:
	go build -ldflags "$(LDFLAGS)" -o bin/$(PLUGIN_NAME) ./cmd/$(PLUGIN_NAME)

run: build
	./bin/$(PLUGIN_NAME) --address :50051

clean:
	rm -rf bin/ $(DIST_DIR)/

test:
	go test -race ./...

lint:
	golangci-lint run

package: clean
	mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(PLUGIN_NAME)-linux-amd64 ./cmd/$(PLUGIN_NAME)
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(PLUGIN_NAME)-linux-arm64 ./cmd/$(PLUGIN_NAME)
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(PLUGIN_NAME)-darwin-amd64 ./cmd/$(PLUGIN_NAME)
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(PLUGIN_NAME)-darwin-arm64 ./cmd/$(PLUGIN_NAME)
	@echo "Packages created in $(DIST_DIR)/"
