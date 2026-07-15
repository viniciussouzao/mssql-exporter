# Simple Makefile for mssql-exporter, following common Prometheus exporter patterns

GO ?= go
BIN ?= mssql-exporter
OUT ?= bin/$(BIN)
CGO_ENABLED ?= 0

MODULE  := github.com/viniciussouzao/mssql-exporter
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w \
	-X $(MODULE)/internal/version.Version=$(VERSION) \
	-X $(MODULE)/internal/version.Commit=$(COMMIT) \
	-X $(MODULE)/internal/version.Date=$(DATE)

.PHONY: help build run test tidy fmt vet check clean install compose-up compose-down docker-build

help:
	@echo "Targets:"
	@echo "  build        Build the exporter binary into bin/"
	@echo "  run          Build and run the exporter (listens on :9187)"
	@echo "  test         Run unit tests"
	@echo "  tidy         Run 'go mod tidy'"
	@echo "  fmt          Format code (gofmt -s -w .)"
	@echo "  vet          Run 'go vet'"
	@echo "  check        Run fmt, vet, and tests"
	@echo "  install      Install the exporter to GOBIN"
	@echo "  clean        Remove build artifacts"
	@echo "  compose-up   Start Prometheus and Grafana via docker compose"
	@echo "  compose-down Stop Prometheus and Grafana"
	@echo "  docker-build Build the mssql-exporter Docker image"

build:
	@mkdir -p bin
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(OUT) ./

run: build
	./$(OUT)

test:
	$(GO) test ./...

tidy:
	$(GO) mod tidy

fmt:
	gofmt -s -w .

vet:
	$(GO) vet ./...

check: fmt vet test

install:
	$(GO) install ./

clean:
	rm -rf bin
	$(GO) clean -cache -testcache

compose-up:
	docker compose up -d

compose-down:
	docker compose down

docker-build:
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg DATE=$(DATE) \
		-t mssql-exporter:$(VERSION) .

