# CLAUDE.md - AI Assistant Guide for mssql-exporter

This document provides a comprehensive guide for AI assistants working with the mssql-exporter codebase.

## Project Overview

**mssql-exporter** is a Prometheus exporter for Microsoft SQL Server written in Go. It collects instance-level metrics from SQL Server and exposes them via an HTTP endpoint (`:9187/metrics`) for Prometheus to scrape.

### Key Technologies
- **Language**: Go 1.24.2
- **Database Driver**: `denisenkom/go-mssqldb` v0.12.3
- **Metrics Library**: `prometheus/client_golang` v1.23.2
- **Configuration**: CLI flags (`spf13/pflag`), environment variables (`.env` via `joho/godotenv`), and an optional YAML config file (`gopkg.in/yaml.v3`) — see `internal/config`
- **Logging**: Standard library `log/slog`

### Purpose
This exporter follows the pattern of other official Prometheus database exporters (e.g., `mysqld_exporter`), providing SQL Server monitoring capabilities with Prometheus best practices.

## Repository Structure

```
mssql-exporter/
├── main.go                          # Entry point: DB connection, collector registration, HTTP server
├── go.mod                           # Go module dependencies
├── go.sum                           # Dependency checksums
├── Makefile                         # Build, test, and deployment targets
├── README.md                        # User-facing documentation
├── .gitignore                       # Git ignore patterns
├── .env                             # Environment variables (not committed, create locally)
├── docker-compose.yml               # Prometheus + Grafana stack for local testing
├── prometheus.yml                   # Prometheus scrape configuration
├── config.example.yml               # Example YAML config file (--config.file)
│
└── internal/
    ├── config/
    │   └── config.go                # Flag/env/YAML config loading and precedence
    │
    ├── db/
    │   ├── connection.go            # Database connection management
    │   └── options.go               # Connection pool configuration
    │
    └── collector/
        ├── server_info.go           # Server properties and system info metrics
        ├── sessions.go              # Session state metrics
        ├── wait_stats.go            # Wait statistics metrics
        ├── io_file_stats.go         # File I/O statistics metrics
        └── memory_clerks.go         # Memory clerk metrics (collector exists but not registered)
```

## Architecture & Design Patterns

### 1. Database Connection Management (`internal/db/`)

**Pattern**: Singleton database connection with configurable options.

- `connection.go`: Provides a package-level `*sql.DB` instance via `Connect()` and `DB()` functions
- `options.go`: Defines connection pool settings (max connections, timeouts, lifetimes)
- Default settings in `DefaultOptions()`: 5 max connections, 5s query timeout, 5min connection lifetime

**Key Functions**:
```go
Connect(connString string, opts *Options) error  // Initialize DB connection
DB() *sql.DB                                     // Get DB instance
CurrentOptions() *Options                        // Get current connection options
```

### 2. Prometheus Collectors (`internal/collector/`)

**Pattern**: Each collector implements `prometheus.Collector` interface.

All collectors follow this structure:
```go
type CollectorName struct {
    db          *sql.DB
    metricName  *prometheus.Desc
    // ... more metric descriptors
}

func NewCollectorName(db *sql.DB) *CollectorName {
    return &CollectorName{
        db: db,
        metricName: prometheus.NewDesc(
            "sqlserver_metric_name",     // Metric name
            "Metric description",          // Help text
            []string{"label1", "label2"},  // Label names
            nil,                           // Const labels
        ),
    }
}

func (c *CollectorName) Describe(ch chan<- *prometheus.Desc) {
    ch <- c.metricName
}

func (c *CollectorName) Collect(ch chan<- prometheus.Metric) {
    ctx, cancel := context.WithTimeout(context.Background(), db.CurrentOptions().ConnTimeout)
    defer cancel()

    // Query SQL Server DMVs
    rows, err := c.db.QueryContext(ctx, "SELECT ...")
    if err != nil {
        slog.Error("...", slog.Any("error", err))
        return
    }
    defer rows.Close()

    // Scan and emit metrics
    for rows.Next() {
        // ... scan values
        ch <- prometheus.MustNewConstMetric(c.metricName, prometheus.GaugeValue, value, labels...)
    }
}
```

**Important Conventions**:
- Always use `context.WithTimeout()` with `db.CurrentOptions().ConnTimeout`
- Always `defer rows.Close()` after querying
- Use `slog.Error()` for logging errors (never panic on query failures)
- Metric names follow pattern: `sqlserver_<category>_<metric>`
- All metrics use `prometheus.GaugeValue` (counters are tracked via SQL Server's cumulative DMVs)

### 3. Main Application Flow (`main.go`)

1. Load `.env` file (using `godotenv.Load()`)
2. Resolve configuration with `config.Load(os.Args[1:])` (flags/env/YAML, see below)
3. Initialize database connection with `db.Connect(cfg.DSN, opts)`
4. Ping database to verify connectivity
5. Register all collectors with `prometheus.MustRegister()`
6. Start HTTP server on `cfg.WebListenAddress` with `cfg.WebTelemetryPath` endpoint

## Configuration Sources

Resolved in `internal/config.Load()`. Precedence per field, highest first:
**CLI flag > environment variable > YAML config file (`--config.file`) > default**.

```bash
--web.listen-address string    Address to listen on for the web interface and metrics. (default ":9187")
--web.telemetry-path string    Path under which to expose metrics. (default "/metrics")
--collector.os-collectors      Enable OS-level collectors (CPU, memory, disk).
--config.file string           Path to an optional YAML config file.
--version                      Print version information and exit.
```

`SQL_SERVER_DSN` has **no CLI flag on purpose** (process args are visible via `ps`/task manager, which would leak credentials). Set it via the `SQL_SERVER_DSN` env var or the `dsn` key in the YAML config file — see `config.example.yml`. `ENABLE_OS_COLLECTORS=true` remains supported as the env-var equivalent of `--collector.os-collectors`.

**DSN Format**: the full DSN is passed through as-is to `db.Connect()`. Any format the `go-mssqldb` driver accepts works, e.g.:
- URL-style: `sqlserver://{user}:{password}@{host}:{port}?database={database}&app+name={appName}`
- ODBC-style: `server={host};user id={user};password={password};port={port};database={database};app name={appName}`

## Development Workflow

### Prerequisites
- Go 1.24.2 or later
- Access to a SQL Server instance
- SQL Server login with `VIEW SERVER STATE` permission

### Common Make Targets

```bash
make help           # Show all available targets
make build          # Build binary to bin/mssql-exporter
make run            # Build and run the exporter
make test           # Run unit tests
make fmt            # Format code with gofmt
make vet            # Run go vet
make check          # Run fmt, vet, and tests
make clean          # Remove build artifacts
make compose-up     # Start Prometheus + Grafana
make compose-down   # Stop Prometheus + Grafana
```

### Build Configuration
- `CGO_ENABLED=0` (static binary, no C dependencies)
- Build flags: `-trimpath` (reproducible builds)
- Binary output: `bin/mssql-exporter`

### Local Testing Setup

1. Create `.env` file with SQL Server credentials
2. Run exporter: `make run` or `./bin/mssql-exporter`
3. Start monitoring stack: `make compose-up`
4. Access services:
   - Exporter metrics: http://localhost:9187/metrics
   - Prometheus: http://localhost:9090
   - Grafana: http://localhost:3000 (admin/admin)

## Adding a New Collector

Follow these steps to add a new collector:

### 1. Create Collector File

Create `internal/collector/your_collector.go`:

```go
package collector

import (
    "context"
    "database/sql"
    "log/slog"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/viniciussouzao/mssql-exporter/internal/db"
)

type YourCollector struct {
    db         *sql.DB
    metricDesc *prometheus.Desc
}

func NewYourCollector(db *sql.DB) *YourCollector {
    return &YourCollector{
        db: db,
        metricDesc: prometheus.NewDesc(
            "sqlserver_your_metric",
            "Description of your metric",
            []string{"label1"},  // Label names
            nil,
        ),
    }
}

func (c *YourCollector) Describe(ch chan<- *prometheus.Desc) {
    ch <- c.metricDesc
}

func (c *YourCollector) Collect(ch chan<- prometheus.Metric) {
    ctx, cancel := context.WithTimeout(context.Background(), db.CurrentOptions().ConnTimeout)
    defer cancel()

    rows, err := c.db.QueryContext(ctx, `
        SELECT column1, column2
        FROM sys.dm_your_dmv
    `)

    if err != nil {
        slog.Error("failed to query your_dmv", slog.Any("error", err))
        return
    }
    defer rows.Close()

    for rows.Next() {
        var label, value string
        if err := rows.Scan(&label, &value); err != nil {
            slog.Error("failed to scan row", slog.Any("error", err))
            continue
        }

        ch <- prometheus.MustNewConstMetric(
            c.metricDesc,
            prometheus.GaugeValue,
            parseFloat(value),
            label,
        )
    }
}
```

### 2. Register Collector

Edit `main.go` and add to `prometheus.MustRegister()`:

```go
prometheus.MustRegister(
    collector.NewWaitStatsCollector(db),
    collector.NewIOFileStatsCollector(db),
    collector.NewSessionsCollector(db),
    collector.NewServerInfoCollector(db),
    collector.NewYourCollector(db),  // Add here
)
```

### 3. Test

1. Verify SQL query works against your SQL Server instance
2. Run `make build && make run`
3. Check metrics: `curl http://localhost:9187/metrics | grep sqlserver_your_metric`
4. Run `make check` to ensure code quality

## SQL Server DMV Reference

The exporter queries these Dynamic Management Views (DMVs):

| DMV | Permission | Collector | Purpose |
|-----|-----------|-----------|---------|
| `sys.dm_os_wait_stats` | VIEW SERVER STATE | wait_stats | Wait types and durations |
| `sys.dm_exec_sessions` | VIEW SERVER STATE | sessions | Session states |
| `sys.dm_io_virtual_file_stats` | VIEW SERVER STATE | io_file_stats | File I/O counters |
| `sys.master_files` | VIEW SERVER STATE | io_file_stats | File metadata |
| `sys.dm_os_sys_info` | VIEW SERVER STATE | server_info | CPU count |
| `sys.dm_os_sys_memory` | VIEW SERVER STATE | server_info | Memory info |
| `SERVERPROPERTY()` | - | server_info | Server properties |

**Required Permission**: `GRANT VIEW SERVER STATE TO [your_login];`

## Coding Conventions

### Error Handling
- **Never panic** on database errors; log and continue
- Use `slog.Error()` with structured logging: `slog.Error("message", slog.Any("error", err))`
- Context timeouts prevent hanging queries

### Metric Naming
- Prefix: `sqlserver_`
- Pattern: `sqlserver_{component}_{metric}[_{unit}]`
- Examples: `sqlserver_sessions_total`, `sqlserver_wait_time_ms`
- Use `_total` suffix for counters
- Include unit in name if not obvious (e.g., `_ms`, `_kb`, `_bytes`)

### Label Design
- Keep cardinality reasonable (avoid unbounded labels like session IDs)
- Common labels: `wait_type`, `session_state`, `db_name`, `file_type`
- Use descriptive label names (not abbreviations)

### Query Best Practices
- Always use `QueryContext()` with timeout
- Filter out noisy wait types (SLEEP%, XE%, BROKER_%)
- Prefer aggregated queries over row-by-row processing
- Use column aliases for clarity

### Code Style
- Follow standard Go conventions (`gofmt`, `go vet`)
- No commented-out code in production
- Keep collector files focused (one responsibility)
- Document complex SQL queries

## Testing Strategy

### Current State
- Unit tests for `internal/config` (flag/env/YAML precedence, validation) — run with `make test`
- No unit tests for collectors yet
- Manual testing via local SQL Server instance
- Integration testing with Prometheus scraping

### Testing Checklist (for new changes)
1. Verify connection to SQL Server
2. Check metrics endpoint: `curl localhost:9187/metrics`
3. Validate metric format (Prometheus text exposition format)
4. Test with Prometheus scraping (use docker-compose setup)
5. Verify no panics or errors in logs
6. Check metric cardinality is reasonable

### Future Testing Improvements
- Add unit tests for collectors (mock `*sql.DB`)
- Integration tests with SQL Server in Docker
- Metric validation tests

## Common Tasks for AI Assistants

### Task: Fix a Bug in a Collector
1. Read the relevant collector file (`internal/collector/*.go`)
2. Check SQL query logic and DMV usage
3. Verify metric descriptor labels match `MustNewConstMetric()` calls
4. Test locally with `make run`
5. Verify fix doesn't break other collectors

### Task: Add a New Metric to Existing Collector
1. Add new `*prometheus.Desc` field to collector struct
2. Initialize descriptor in constructor with proper labels
3. Add to `Describe()` method
4. Query and emit metric in `Collect()` method
5. Update README.md with new metric documentation

### Task: Update Dependencies
1. Edit `go.mod` if needed
2. Run `make tidy`
3. Test with `make build && make run`
4. Verify metrics still work correctly

### Task: Improve Error Handling
1. Look for error conditions that aren't logged
2. Add `slog.Error()` calls with context
3. Ensure `defer rows.Close()` is present
4. Verify context timeouts are used

## Known Issues & Future Work

### Current Limitations
- No TLS/authentication on HTTP endpoint
- Connection pool settings not configurable at runtime
- Limited test coverage
- No `--config.file` auto-discovery (must be passed explicitly)

### Roadmap
See `ROADMAP.md` for the authoritative phased plan (performance counters,
databases collector, `sqlserver_up` + per-collector toggles, blocking/AG/Agent
jobs collectors, TLS/basic auth, health check endpoint). Update it as features
land: move items to the "Done" section and keep phases current.

## Troubleshooting Guide

### Cannot connect to SQL Server
- Check environment variables in `.env`
- Verify SQL Server accepts TCP connections
- Test connectivity: `telnet <host> <port>`
- Check firewall rules

### Permission errors
- Grant: `GRANT VIEW SERVER STATE TO [your_login];`
- Verify login can query DMVs directly in SSMS

### Panic on scrape
- Check metric descriptor labels count matches `MustNewConstMetric()` call
- Verify label slice length equals descriptor's label count
- Review recent collector changes

### Metrics not appearing in Prometheus
- Verify exporter is running: `curl localhost:9187/metrics`
- Check Prometheus targets page (http://localhost:9090/targets)
- Verify Prometheus can reach `host.docker.internal:9187`
- Check for errors in Prometheus logs

### High memory usage
- Review collector queries for unbounded result sets
- Check metric cardinality (too many unique label combinations)
- Consider filtering DMV results more aggressively

## Git Workflow

### Branch Naming
- Feature branches: `feat/<description>`
- Bug fixes: `fix/<description>`
- Claude AI branches: `claude/claude-md-<session-id>`

### Commit Messages
- Follow conventional commits style
- Examples:
  - `feat: add database locks collector`
  - `fix: correct label count in sessions collector`
  - `docs: update README with new metrics`
  - `refactor: extract connection string builder`

### Pull Requests
- Ensure `make check` passes
- Update README.md if adding/changing metrics
- Test locally with real SQL Server instance
- Include example metric output in PR description

## Dependencies & Module Management

### Core Dependencies
```
github.com/denisenkom/go-mssqldb v0.12.3        # SQL Server driver
github.com/prometheus/client_golang v1.23.2     # Prometheus metrics
github.com/joho/godotenv v1.5.1                 # .env file support
```

### Updating Dependencies
```bash
go get -u github.com/prometheus/client_golang@latest
make tidy
make test
```

### Vendoring
- Not currently using vendoring
- Dependencies managed via `go.mod`/`go.sum`

## Performance Considerations

### Query Optimization
- DMV queries are generally fast (< 100ms)
- Some DMVs (like `dm_io_virtual_file_stats`) can be slower with many databases
- Consider filtering or sampling for large deployments

### Scrape Interval
- Default: 15 seconds (see `prometheus.yml`)
- Balance freshness vs. SQL Server load
- For high-cardinality metrics, consider longer intervals

### Connection Pooling
- Default: 5 max connections, 5 idle
- Adjust in `internal/db/options.go` if needed
- Monitor connection usage with `db.Stats()`

## Security Best Practices

### Credential Management
- Never commit `.env` file (listed in `.gitignore`)
- Use least-privilege SQL Server login (VIEW SERVER STATE only)
- Consider SQL Server authentication over Windows auth for containers

### Network Security
- Exporter serves plain HTTP (no TLS)
- Recommended: bind to localhost or use reverse proxy with TLS
- Prometheus scraping should be on trusted network

### SQL Injection Prevention
- All queries use static SQL (no user input)
- No dynamic query building currently
- If adding user input, use parameterized queries

## Additional Resources

- [Prometheus Naming Best Practices](https://prometheus.io/docs/practices/naming/)
- [SQL Server DMV Reference](https://docs.microsoft.com/en-us/sql/relational-databases/system-dynamic-management-views/)
- [go-mssqldb Documentation](https://github.com/denisenkom/go-mssqldb)
- [Prometheus Client Golang](https://github.com/prometheus/client_golang)

---

**Last Updated**: 2025-11-27
**Maintainer**: viniciussouzao
**AI Assistant**: This file is specifically designed to help AI assistants understand and work with this codebase effectively.
