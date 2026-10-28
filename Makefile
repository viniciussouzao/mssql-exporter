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

.PHONY: help build run test test-integration test-alerts scan-secrets vulncheck tidy fmt vet check clean install compose-up compose-down docker-build release-dry-run

help:
	@echo "Targets:"
	@echo "  build        Build the exporter binary into bin/"
	@echo "  run          Build and run the exporter (listens on :10060)"
	@echo "  test         Run unit tests"
	@echo "  test-integration  Run the integration tests against a disposable SQL Server container"
	@echo "  test-alerts  Check and unit test contrib/prometheus/alerts.yml with promtool (Docker)"
	@echo "  scan-secrets Scan the whole git history for secrets with gitleaks (Docker)"
	@echo "  vulncheck    Report known vulnerabilities reachable from the code (govulncheck)"
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
PROMTOOL_IMAGE ?= prom/prometheus:v3.5.0
GITLEAKS_IMAGE ?= ghcr.io/gitleaks/gitleaks:v8.30.1
GOVULNCHECK_VERSION ?= v1.9.0

test-integration:
	@pw="It-$$(openssl rand -hex 12)-Aa1"; \
	docker run -d --rm --name $(IT_CONTAINER) -e ACCEPT_EULA=Y -e MSSQL_SA_PASSWORD="$$pw" \
		-p 127.0.0.1:$(IT_PORT):1433 $(IT_IMAGE) >/dev/null || exit 1; \
	MSSQL_EXPORTER_TEST_DSN="sqlserver://sa:$$pw@127.0.0.1:$(IT_PORT)?encrypt=disable" \
		$(GO) test -tags integration -count=1 -v ./integration/...; \
	rc=$$?; docker stop -t 10 $(IT_CONTAINER) >/dev/null; exit $$rc

# promtool ships in the Prometheus image; no local install needed.
test-alerts:
	docker run --rm -v "$(CURDIR)/contrib/prometheus:/rules:ro" -w /rules --entrypoint promtool \
		$(PROMTOOL_IMAGE) check rules alerts.yml
	docker run --rm -v "$(CURDIR)/contrib/prometheus:/rules:ro" -w /rules --entrypoint promtool \
		$(PROMTOOL_IMAGE) test rules alerts_test.yml

# Every commit of every local ref, with the findings' values redacted. Run it
# from the main checkout: a worktree's .git points outside the mounted directory.
scan-secrets:
	docker run --rm -v "$(CURDIR):/repo:ro" $(GITLEAKS_IMAGE) \
		git /repo --redact --verbose --no-banner --log-opts="--all"

# Fails only on vulnerabilities the code can reach, in the standard library of
# the Go in use and in the modules. Run it with the Go that builds the releases.
vulncheck:
	$(GO) run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) ./...

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

