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

.PHONY: help build run test test-integration tidy fmt vet check clean install compose-up compose-down docker-build release-dry-run

help:
	@echo "Targets:"
	@echo "  build        Build the exporter binary into bin/"
	@echo "  run          Build and run the exporter (listens on :10060)"
	@echo "  test         Run unit tests"
	@echo "  test-integration  Run the integration tests against a disposable SQL Server container"
	@echo "  tidy         Run 'go mod tidy'"
	@echo "  fmt          Format code (gofmt -s -w .)"
	@echo "  vet          Run 'go vet'"
	@echo "  check        Run fmt, vet, and tests"
	@echo "  install      Install the exporter to GOBIN"
	@echo "  clean        Remove build artifacts"
	@echo "  compose-up   Start Prometheus and Grafana via docker compose"
	@echo "  compose-down Stop Prometheus and Grafana"
	@echo "  docker-build Build the mssql-exporter Docker image"
	@echo "  release-dry-run  Run GoReleaser in snapshot mode (no publish, no tag needed)"

build:
	@mkdir -p bin
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(OUT) ./

run: build
	./$(OUT)

test:
	$(GO) test ./...

# Integration tests: starts a throwaway SQL Server container (removed when
# it stops) with a random sa password and runs ./integration against it.
# Override IT_IMAGE to test another version, e.g. IT_IMAGE=mcr.microsoft.com/mssql/server:2022-latest
IT_IMAGE ?= mcr.microsoft.com/mssql/server:2025-latest
IT_PORT ?= 14339
IT_CONTAINER ?= mssql-exporter-it

test-integration:
	@pw="It-$$(openssl rand -hex 12)-Aa1"; \
	docker run -d --rm --name $(IT_CONTAINER) -e ACCEPT_EULA=Y -e MSSQL_SA_PASSWORD="$$pw" \
		-p 127.0.0.1:$(IT_PORT):1433 $(IT_IMAGE) >/dev/null || exit 1; \
	MSSQL_EXPORTER_TEST_DSN="sqlserver://sa:$$pw@127.0.0.1:$(IT_PORT)?encrypt=disable" \
		$(GO) test -tags integration -count=1 -v ./integration/...; \
	rc=$$?; docker stop -t 10 $(IT_CONTAINER) >/dev/null; exit $$rc

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

release-dry-run:
	goreleaser release --snapshot --clean

