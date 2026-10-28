# CLAUDE.md - AI Assistant Guide for mssql-exporter

This document provides a comprehensive guide for AI assistants working with the mssql-exporter codebase.

## Project Overview

**mssql-exporter** is a Prometheus exporter for Microsoft SQL Server written in Go. It collects instance-level metrics from SQL Server and exposes them via an HTTP endpoint (`:9187/metrics`) for Prometheus to scrape.

### Key Technologies
- **Language**: Go 1.25.0 (see `go.mod`)
- **Database Driver**: `microsoft/go-mssqldb` v1.11.2 (replaced the archived `denisenkom/go-mssqldb` in v0.5.0)
- **Metrics Library**: `prometheus/client_golang` v1.23.2
- **Configuration**: CLI flags (`spf13/pflag`), environment variables (`.env` via `joho/godotenv`), and an optional YAML config file (`gopkg.in/yaml.v3`) — see `internal/config`
- **Logging**: Standard library `log/slog`
- **HTTP server / TLS / basic auth**: `prometheus/exporter-toolkit/web` (`--web.config.file`)
- **Tests**: standard `testing` + `DATA-DOG/go-sqlmock` for collectors
- **Release**: GoReleaser (`.goreleaser.yaml`) driven by the GitHub Actions release workflow

### Purpose
This exporter follows the pattern of other official Prometheus database exporters (e.g., `mysqld_exporter`), providing SQL Server monitoring capabilities with Prometheus best practices.

## Repository Structure

```
mssql-exporter/
├── main.go                          # Entry point: DB connection, collector registration, HTTP server (/metrics, /healthz)
├── instance_label.go                # sqlserver_instance label registry, retried @@SERVERNAME
├── service*.go, eventlog_windows.go # Windows service install/run and Event Log handler
├── go.mod                           # Go module dependencies
├── go.sum                           # Dependency checksums
├── Makefile                         # Build, test, docker and release-dry-run targets
├── README.md                        # User-facing documentation
├── ROADMAP.md                       # Phased feature roadmap (authoritative)
├── BACKLOG-HOMOLOG.md               # Homologation findings (HML-NNN): EC2 Windows, RDS, Cloud SQL, Azure, home lab
├── LICENSE                          # Apache-2.0
├── Dockerfile                       # Multi-stage build, non-root, /healthz HEALTHCHECK
├── .goreleaser.yaml                 # Cross-platform binaries, checksums, changelog
├── .gitignore                       # Git ignore patterns
├── .env                             # Environment variables (not committed, create locally)
├── docker-compose.yml               # Prometheus + Grafana stack for local testing
├── prometheus.yml                   # Prometheus scrape configuration
├── config.example.yml               # Example YAML config file (--config.file)
│
├── .github/workflows/
│   ├── ci.yml                       # PR checks: vet, golangci-lint, test, build
│   └── release.yml                  # Tag + GoReleaser on merged PRs labeled release:*
│
├── dashboards/
│   ├── mssql-exporter-grafana-dashboard.json        # Main Grafana dashboard
│   ├── mssql-exporter-rds-cloudwatch-dashboard.json # Amazon RDS: exporter + CloudWatch
│   ├── mssql-exporter-cloudsql-dashboard.json       # Google Cloud SQL: exporter + Cloud Monitoring
│   ├── mssql-exporter-azure-mi-dashboard.json       # Azure SQL Managed Instance: exporter + Azure Monitor
│   ├── mssql-exporter-homolog-coverage.json         # Metric coverage view used in homologation
│   └── README.md                    # Dashboard import guide
│
├── contrib/systemd/
│   └── mssql-exporter.service       # Hardened Linux unit, shipped in the release archives
│
├── hack/loadgen/                    # Load generator (main.go, setup.sql, cleanup.sql) for homologation
│
└── internal/
    ├── config/
    │   └── config.go                # Flag/env/YAML config loading and precedence
    │
    ├── version/
    │   └── version.go               # Version/Commit/Date injected via -ldflags
    │
    ├── oscollector/                 # Optional OS-level collectors (Linux and Windows)
    │   └── *.go, *_windows.go       # CPU, memory, disk I/O/usage, network, process, boot time
    │
    ├── db/
    │   ├── connection.go            # Database connection management
    │   └── options.go               # Connection pool configuration
    │
    └── collector/
        ├── exporter.go              # Exporter (prometheus.Collector), Scraper interface, registry
        ├── options.go               # Options of the collectors that need more than on/off
        ├── databases_list.go        # Accessible user databases (shared by per-database collectors)
        ├── engine_edition.go        # SERVERPROPERTY('EngineEdition') helper for platform-specific queries
        ├── server_info.go           # Server properties and system info metrics
        ├── sessions.go              # Session state metrics and sessions by login
        ├── wait_stats.go            # Wait statistics metrics
        ├── io_file_stats.go         # File I/O statistics metrics
        ├── memory_clerks.go         # Memory clerk metrics
        ├── backups.go               # Last backup timestamp/duration/size per database
        ├── tempdb.go                # TempDB space usage and temp table activity
        ├── perf_counters.go         # Instance-level performance counters
        ├── databases.go             # Database state, recovery model, sizes, log usage and log reuse wait
        ├── transaction_log.go       # VLFs, active log, log since last log backup/checkpoint (dm_db_log_stats)
        ├── blocking.go              # Blocked sessions and requests by status
        ├── active_queries.go        # Long-running requests and open transactions, aggregated
        ├── availability_groups.go   # Always On replica/database health, queues, suspension and lag
        ├── agent_jobs.go            # SQL Agent job outcomes (needs msdb grants)
        ├── volumes.go               # Disk space per volume hosting database files (instance storage on Azure MI)
        ├── schedulers.go            # Scheduler/CPU pressure signals
        ├── database_mirroring.go    # Mirroring role, state, safety, witness and queues
        ├── log_shipping.go          # Log shipping monitor (opt-in, needs msdb grants)
        ├── top_queries.go           # Top N queries by CPU from the plan cache (opt-in)
        ├── sqlobfuscate.go          # Literal obfuscation for top_queries statement text
        ├── index_fragmentation.go   # Index fragmentation, collected in the background (opt-in)
        └── query_store.go           # Query Store health per database, in the background (opt-in)
```

Tests live beside the code as `*_test.go` in `internal/collector` (all collectors and the
Exporter, using `go-sqlmock`; shared helpers in `testutil_test.go`), `internal/config` and
`internal/db`, plus the `main` package (`service_test.go`, `instance_label_test.go`).
`internal/oscollector` and `internal/version` have no tests (the OS collectors are
validated on the home lab: Windows notebook and the Raspberry Pi for Linux).

## Architecture & Design Patterns

### 1. Database Connection Management (`internal/db/`)

**Pattern**: Singleton database connection with configurable options.

- `connection.go`: Provides a package-level `*sql.DB` instance via `Connect()` and `DB()` functions
- `options.go`: Defines connection pool settings (max connections, timeouts, lifetimes)
- Default settings in `DefaultOptions()`: 5 max connections, 5s query timeout, 5min connection lifetime;
  `main.go` overrides max open/idle connections with `--sqlserver.max-connections`

**Key Functions**:
```go
Connect(connString string, opts *Options) error  // Initialize DB connection
DB() *sql.DB                                     // Get DB instance
CurrentOptions() *Options                        // Get current connection options
```

### 2. Collectors (`internal/collector/`)

**Pattern**: Each collector implements the `Scraper` interface (defined in
`internal/collector/exporter.go`). A single top-level `Exporter`
(prometheus.Collector) pings SQL Server per scrape (`sqlserver_up`), runs
every enabled scraper and emits per-scraper self-metrics
(`sqlserver_scrape_collector_duration_seconds` / `_success`). When the ping
fails, scrapers are skipped and only `sqlserver_up 0` is exposed — the
process never exits because SQL Server is down. `Collect` waits for the ping
and each scraper at most until the scrape deadline plus `deadlineGrace`
(500 ms) and then answers with what it has: the driver can block past a
cancelled context during a failover, and without that bound the whole
scrape outlived Prometheus' `scrape_timeout` and was lost (HML-037, v0.12.1).
Simultaneous scrapes share one collection (singleflight in `Collect`): a
scrape that arrives while another one is collecting waits for it and
exposes the same const metrics, so several Prometheus servers do not
compete for the connection pool (HML-034, v0.14.0). There is no cache:
scrapes that do not overlap collect again.

```go
type Scraper interface {
    Name() string                                  // snake_case, used in --collector.<name>
    Describe(ch chan<- *prometheus.Desc)
    // ctx carries the shared scrape deadline; scrapers run concurrently
    Scrape(ctx context.Context, ch chan<- prometheus.Metric) error
}
```

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

func (c *CollectorName) Name() string { return "collector_name" }

func (c *CollectorName) Describe(ch chan<- *prometheus.Desc) {
    ch <- c.metricName
}

func (c *CollectorName) Scrape(ctx context.Context, ch chan<- prometheus.Metric) error {
    // Query SQL Server DMVs (ctx carries the shared scrape deadline)
    rows, err := c.db.QueryContext(ctx, "SELECT ...")
    if err != nil {
        return fmt.Errorf("failed to query ...: %w", err)
    }
    defer func() { _ = rows.Close() }()

    // Scan and emit metrics
    var scanErrs []error
    for rows.Next() {
        // scan errors: accumulate and `continue` — later rows still emit
        // metrics, but the scrape is reported as failed (success 0)
        ch <- prometheus.MustNewConstMetric(c.metricName, prometheus.GaugeValue, value, labels...)
    }

    return errors.Join(append(scanErrs, rows.Err())...)
}
```

**Important Conventions**:
- Never create your own context: use the `ctx` passed to `Scrape` (it carries
  the shared scrape deadline; scrapers run concurrently under one timeout)
- Always `defer func() { _ = rows.Close() }()` after querying
- Query failures: return the error (the Exporter logs it and reports
  `scrape_collector_success 0`)
- Row-scan failures: accumulate in a `scanErrs` slice and `continue`, then
  `return errors.Join(append(scanErrs, rows.Err())...)` — partial metrics are
  still emitted, but any failed row marks the scrape as unsuccessful
- Multi-query scrapers combine errors with `errors.Join` (see tempdb, databases)
- Metric names follow pattern: `sqlserver_<category>_<metric>`
- Cumulative DMV counters use `prometheus.CounterValue` with `_total` suffix;
  point-in-time values use `prometheus.GaugeValue`

### 3. Main Application Flow (`main.go`)

1. Load `.env` file (using `godotenv.Load()`)
2. Resolve configuration with `config.Load(os.Args[1:])` (flags/env/YAML, see below); logs the build version
3. Initialize connection pool with `db.Connect(cfg.DSN, opts)` (validates the DSN only — does not dial)
4. Ping database (best effort: logs a warning and continues if unreachable; `sqlserver_up` reports 0 until it recovers)
5. Build `collector.NewExporterWithOptions(db, cfg.EnabledCollectors, cfg.CollectorOptions)` and
   register it, plus `oscollector.Register(reg)` if `cfg.EnableOSCollectors`, in an
   `instanceRegistry` (`instance_label.go`) that adds the `sqlserver_instance` label
6. Register `mssql_exporter_build_info` (version, commit, date, goversion labels) in the
   default registry; the handler serves `prometheus.Gatherers{default, instanceRegistry}`
7. Serve `cfg.WebTelemetryPath` (metrics) and `/healthz` (liveness only — use
   `sqlserver_up` for database reachability) via `web.ListenAndServe` from
   exporter-toolkit, on `cfg.WebListenAddress`, with TLS/basic auth from
   `cfg.WebConfigFile`

`--version` prints `version.String()` and exits (handled in `config.Load`).

Note: the `sqlserver_instance` label is `instance_label` when configured,
otherwise `@@SERVERNAME` read at startup; if SQL Server is down at that
moment the label starts empty and `resolveInstanceName` retries every 15 s,
then swaps in a new registry with the name (HML-029, v0.15.0; the name is
not read again afterwards). On Amazon RDS
the server name is a generic host name that changes on stop/start and
failover, so set `instance_label` there. On Google Cloud SQL and Azure SQL
Managed Instance `@@SERVERNAME` survives failover (a generated host name on
Cloud SQL, the FQDN on MI), but `instance_label` is still recommended.

## Configuration Sources

Resolved in `internal/config.Load()`. Precedence per field, highest first:
**CLI flag > environment variable > YAML config file (`--config.file`) > default**.

```bash
--web.listen-address string    Address to listen on for the web interface and metrics. (default ":9187")
--web.telemetry-path string    Path under which to expose metrics. (default "/metrics")
--web.config.file string       exporter-toolkit web config for TLS and basic auth (YAML: web.config_file).
--collector.<name>             Enable/disable a SQL Server collector (default true for all
                               except log_shipping, top_queries, index_fragmentation and
                               query_store, see scrapersDisabledByDefault).
                               Names come from collector.ScraperNames(); YAML equivalent is
                               the `collectors:` map keyed by the same names.
--collector.wait_stats.include_idle
--collector.top_queries.limit|include_text|text_length
--collector.index_fragmentation.interval|min_page_count|databases
--collector.query_store.interval|databases
                               Collector options (YAML: `wait_stats:`, `top_queries:`,
                               `index_fragmentation:`, `query_store:` sections); they do not
                               enable the collector. wait_stats skips idle/background wait
                               types (`idleWaitTypes`) and never-seen ones unless include_idle.
--sqlserver.instance-label string
                               Fixed sqlserver_instance label (env SQL_SERVER_INSTANCE_LABEL,
                               YAML instance_label; default @@SERVERNAME).
--sqlserver.max-connections int
                               Connection pool size, 1-100 (env SQL_SERVER_MAX_CONNECTIONS,
                               YAML max_connections; default 5).
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
- Go 1.25.0 or later
- Access to a SQL Server instance
- SQL Server login with `VIEW SERVER STATE` permission

### Common Make Targets

```bash
make help           # Show all available targets
make build          # Build binary to bin/mssql-exporter
make run            # Build and run the exporter
make test           # Run unit tests
make tidy           # go mod tidy
make fmt            # Format code with gofmt
make vet            # Run go vet
make check          # Run fmt, vet, and tests
make install        # go install to GOBIN
make clean          # Remove build artifacts (and Go build/test cache)
make compose-up     # Start Prometheus + Grafana
make compose-down   # Stop Prometheus + Grafana
make docker-build   # Build the Docker image
make release-dry-run  # GoReleaser snapshot build (no publish, no tag needed)
```

### Build Configuration
- `CGO_ENABLED=0` (static binary, no C dependencies)
- Build flags: `-trimpath`; `-ldflags` inject `internal/version` (Makefile, Dockerfile and GoReleaser)
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
    "errors"
    "fmt"

    "github.com/prometheus/client_golang/prometheus"
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

func (c *YourCollector) Name() string { return "your_collector" }

func (c *YourCollector) Describe(ch chan<- *prometheus.Desc) {
    ch <- c.metricDesc
}

func (c *YourCollector) Scrape(ctx context.Context, ch chan<- prometheus.Metric) error {
    rows, err := c.db.QueryContext(ctx, `
        SELECT column1, column2
        FROM sys.dm_your_dmv
    `)

    if err != nil {
        return fmt.Errorf("failed to query your_dmv: %w", err)
    }
    defer func() { _ = rows.Close() }()

    var scanErrs []error
    for rows.Next() {
        var label string
        var value float64
        if err := rows.Scan(&label, &value); err != nil {
            scanErrs = append(scanErrs, fmt.Errorf("failed to scan your_dmv row: %w", err))
            continue
        }

        ch <- prometheus.MustNewConstMetric(
            c.metricDesc,
            prometheus.GaugeValue,
            value,
            label,
        )
    }

    return errors.Join(append(scanErrs, rows.Err())...)
}
```

### 2. Register Collector

Edit `internal/collector/exporter.go` and add the collector to `scraperNames`
(this defines the registration order AND the `--collector.<name>` flag /
`collectors:` YAML key) and to `scraperFactories`:

```go
var scraperNames = []string{
    // ...
    "your_collector",
}

var scraperFactories = map[string]func(*sql.DB, Options) Scraper{
    // ...
    "your_collector": func(db *sql.DB, _ Options) Scraper { return NewYourCollector(db) },
}
```

No change in `main.go` is needed — the Exporter picks it up automatically,
including the enable/disable toggle.

- **Opt-in collectors** (extra grants, expensive queries or high cardinality)
  also go in `scrapersDisabledByDefault`.
- **Options** beyond on/off: add a struct to `Options` in `options.go` with
  its defaults in `DefaultOptions()`, and the flag/env/YAML plumbing and
  validation in `internal/config/config.go` (see `top_queries`).
- **Expensive or per-database collectors** run in the background on their
  own interval and `Scrape` serves the last result plus
  `_last_collection_{timestamp_seconds,duration_seconds,success,databases}`
  (see `index_fragmentation.go` and `query_store.go`). Per-database queries
  use `listAccessibleDatabases` (`HAS_DBACCESS`) and three-part names quoted
  with `quoteIdentifier`.

### 3. Test

1. Add `internal/collector/your_collector_test.go` using `newSQLMock` from
   `testutil_test.go` (covers happy path, query error, scan error, empty result)
2. Verify SQL query works against your SQL Server instance
3. Run `make build && make run`
4. Check metrics: `curl http://localhost:9187/metrics | grep sqlserver_your_metric`
5. Check self-metrics: `curl http://localhost:9187/metrics | grep 'scrape_collector.*your_collector'`
6. Run `make check` to ensure code quality
7. Document the metrics in README.md and, if useful, the Grafana dashboard

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
| `sys.dm_os_memory_clerks` | VIEW SERVER STATE | memory_clerks | Memory usage by clerk |
| `msdb.dbo.backupset` | msdb read access | backups | Backup history |
| `tempdb.sys.dm_db_file_space_usage` | VIEW SERVER STATE | tempdb | TempDB space usage |
| `sys.dm_os_performance_counters` | VIEW SERVER STATE | perf_counters, databases, tempdb | Instance/database counters |
| `sys.databases` | - | databases | Database state, recovery model and log reuse wait |
| `sys.dm_db_log_stats` + `sys.dm_hadr_database_replica_states` | VIEW SERVER STATE | transaction_log | VLFs, active log, log since last log backup/checkpoint (2016 SP2+; AG secondaries skipped) |
| `sys.dm_exec_requests` | VIEW SERVER STATE | blocking, active_queries | Blocked sessions, request status, long-running requests |
| `sys.dm_tran_active_transactions` + `dm_tran_session_transactions` | VIEW SERVER STATE | active_queries | Open transactions and their age |
| `sys.dm_hadr_availability_replica_states` | VIEW SERVER STATE | availability_groups | Replica role and sync health |
| `sys.dm_hadr_database_replica_states` | VIEW SERVER STATE | availability_groups | Send/redo queues per database |
| `msdb.dbo.sysjobs` + `sysjobhistory` | SELECT grants in msdb | agent_jobs | Agent job outcomes |
| `sys.dm_os_volume_stats` | VIEW SERVER STATE + VIEW ANY DEFINITION | volumes | Disk space per volume |
| `sys.dm_os_schedulers` | VIEW SERVER STATE | schedulers | CPU pressure signals |
| `sys.database_mirroring` | VIEW SERVER STATE | database_mirroring | Mirroring role, state and safety |
| `msdb.dbo.log_shipping_monitor_primary` / `_secondary` | SELECT grants in msdb | log_shipping | Backup/copy/restore times |
| `sys.dm_exec_query_stats`, `dm_exec_plan_attributes`, `dm_exec_sql_text` | VIEW SERVER STATE | top_queries | Plan cache statistics and statement text |
| `sys.dm_db_index_physical_stats` + per-database `sys.indexes`/`objects`/`schemas` | VIEW SERVER STATE + database access | index_fragmentation | Index fragmentation and size |
| Per-database `sys.database_query_store_options` + `sys.query_store_plan` | VIEW SERVER STATE + database access | query_store | Query Store state, storage and forced plans |
| `master.sys.server_resource_stats` | VIEW SERVER STATE | volumes (Azure MI only) | Reserved and used instance storage |
| `SERVERPROPERTY()` | - | server_info, engine_edition.go | Server properties; `EngineEdition` selects platform-specific queries |

**Note**: `sys.master_files` silently returns zero rows without
`VIEW ANY DEFINITION` — the io_file_stats, databases (sizes) and volumes
collectors report success but emit nothing. Grant both `VIEW SERVER STATE`
and `VIEW ANY DEFINITION` to the monitoring login.

**Required Permission**: `GRANT VIEW SERVER STATE TO [your_login];`

**Per-database collectors** (index_fragmentation, query_store) skip databases
without a user for the login: grant `CONNECT ANY DATABASE` (not grantable by
the master user on Amazon RDS or on Google Cloud SQL — create a user per
database there).

### Managed platforms

`SERVERPROPERTY('EngineEdition')`: 2/3/4 = boxed SQL Server (also RDS and
Cloud SQL), 8 = Azure SQL Managed Instance, 5 = Azure SQL Database.
Platform-specific behaviour is gated on the engine edition (helper
`queryEngineEdition` in `engine_edition.go`, or an inline
`CAST(SERVERPROPERTY('EngineEdition') AS int) = 8` in the SQL) so every other
platform keeps the exact same query results. Current MI-only branches:

| Collector | On Azure SQL MI (EngineEdition 8) |
|---|---|
| volumes | `dm_os_volume_stats` reports file sizes there, so it emits `sqlserver_instance_storage_{reserved,used}_bytes` from `master.sys.server_resource_stats` instead (v0.13.0) |
| databases | Log counters of user databases use `physical_database_name` (a GUID) as instance name; a `CROSS APPLY` prefers `d.name` because master/model also have `replicatedmaster`/`managed_model` instances (v0.13.1) |
| backups | `backupset.server_name` is NULL for MI's automatic backups; accepted only on MI (v0.13.2) |
| active_queries | The always-running `TdService` telemetry session (`NT AUTHORITY\SYSTEM`) is excluded (v0.13.2) |
| server_info | `MachineName` and other properties may be NULL (handled on every platform, v0.12.2) |

Other platform facts from the homologation (details in `BACKLOG-HOMOLOG.md`
and README):
- **Cloud SQL** is SQL Server on **Linux**; grants only work as
  `GRANT ... TO [x] AS [CustomerDbRootRole]`, msdb is closed (agent_jobs
  fails on permission), `encrypt=strict` is refused (`encrypt=true` with the
  instance CA works). Like every SQL Server on Linux (also the Docker
  images), `dm_os_volume_stats` returns NULL mount point and volume name, so
  `volumes` labels the series with the directory of the database files.
- **Azure MI** behaves like boxed SQL Server for grants (including
  `CONNECT ANY DATABASE` and msdb users); `encrypt=strict` works with the
  public Microsoft CA; the exporter must run inside the VNet (the gateway
  redirects to the node).
- **Azure SQL Database** is not supported yet: no `sys.master_files`, no
  msdb (error 40515), no `dm_os_sys_memory`; only 10 of 19 collectors work
  with a contained user (HML-053). The "Azure SQL Database mode" in
  `ROADMAP.md` (Later) is the sized plan.
- **Windows without a pagefile**: gopsutil `SwapMemory` fails; the Windows
  memory collector emits zeros when `SwapDevices()` is empty (v0.13.3).

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
- Unit tests for `internal/config` (flag/env/YAML precedence, validation), `internal/db`,
  the Exporter and every collector (`go-sqlmock`) — run with `make test`
  (185 test functions, 227 runs counting subtests)
- CI (`.github/workflows/ci.yml`) runs vet, golangci-lint, tests and build on every PR
- No integration tests against a real SQL Server yet
- Manual homologation recorded in `BACKLOG-HOMOLOG.md` (HML-001..053), all
  cloud resources destroyed after each phase:
  - AWS: EC2 Windows (2019 → 2022 → 2025, AG, mirroring, log shipping) and
    RDS for SQL Server 2025 SE (Single/Multi-AZ)
  - Google Cloud SQL for SQL Server 2022 SE (Linux, zonal → regional HA)
  - Azure SQL Managed Instance (GP, free offer) and an Azure SQL Database
    reconnaissance
  - GCP and Azure infrastructure is Terraform in `hml_resources_gcp/` and
    `hml_resources_azure/` (local only, never committed; the maintainer runs
    plan/apply/destroy)
- Permanent home lab: Windows 11 pt-BR notebook with SQL Server 2025
  Developer (exporter as a Windows service, OS collectors on), scraped by
  Prometheus/Grafana on a Raspberry Pi 5
- Manual testing via local SQL Server containers (Docker: 2025 with and
  without Agent, Express)
- Integration testing with Prometheus scraping

### Testing Checklist (for new changes)
1. Verify connection to SQL Server
2. Check metrics endpoint: `curl localhost:9187/metrics`
3. Validate metric format (Prometheus text exposition format)
4. Test with Prometheus scraping (use docker-compose setup)
5. Verify no panics or errors in logs
6. Check metric cardinality is reasonable

### Future Testing Improvements
- Integration tests with SQL Server in Docker (including an Availability Group scenario)
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
4. Query and emit metric in `Scrape()` method
5. Update README.md with new metric documentation

### Task: Update Dependencies
1. Edit `go.mod` if needed
2. Run `make tidy`
3. Test with `make build && make run`
4. Verify metrics still work correctly

### Task: Improve Error Handling
1. Query failures must be returned from `Scrape()` (not just logged)
2. Ensure `defer func() { _ = rows.Close() }()` is present
3. Accumulate row-scan errors and join them with `rows.Err()` at the end
4. Verify queries use the `ctx` passed to `Scrape()` (no local timeouts)

## Known Issues & Future Work

### Current Limitations
- Query timeout (5s) and connection lifetimes are not configurable; only the
  pool size is (`--sqlserver.max-connections`)
- No integration tests against a real SQL Server (manual homologation is tracked in `BACKLOG-HOMOLOG.md`)
- Azure SQL Database is not supported (see "Managed platforms")
- availability_groups was validated against a real 2-replica Always On AG on EC2 (SQL Server 2019, 2026-10-12).
  A secondary's DMVs only show its own replica, so the same replica series is reported by every exporter
  in the AG (told apart by `sqlserver_instance`); aggregate with `max by (ag_name, replica_server, db_name)`
- No `--config.file` auto-discovery (must be passed explicitly)

### Roadmap
See `ROADMAP.md` for the authoritative phased plan. Phases 1–7 are done (core
metrics, robustness, operational collectors, hardening, legacy HA/DR,
workload and index insights, transaction log and Query Store health); the
"Later" section holds the next candidates. Update it as features land: move
items to the "Done" section and keep phases current.

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

### Releases
Merging a PR into `main` with a label publishes a release (`release.yml` tags the
next version and runs GoReleaser): `release:patch` (or legacy `release`),
`release:minor`, `release:major`. The largest label wins; with no `vX.Y.Z` tag
yet the first release is `v0.1.0`. v0.1.0 was published manually with
`goreleaser release --clean`. Never add a release label unless a release is intended.

### Pull Requests
- Ensure `make check` passes
- Update README.md if adding/changing metrics
- Test locally with real SQL Server instance
- Include example metric output in PR description

## Dependencies & Module Management

### Core Dependencies
```
github.com/microsoft/go-mssqldb v1.11.2         # SQL Server driver
github.com/prometheus/client_golang v1.23.2     # Prometheus metrics
github.com/prometheus/exporter-toolkit v0.17.1  # HTTP server, TLS, basic auth
github.com/spf13/pflag v1.0.10                  # CLI flags
gopkg.in/yaml.v3 v3.0.1                         # YAML config file
github.com/joho/godotenv v1.5.1                 # .env file support
github.com/DATA-DOG/go-sqlmock v1.5.2           # Collector tests
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
- Default: 5 max connections, 5 idle; `--sqlserver.max-connections` sets both
- Every enabled collector of a scrape runs at the same time on this pool
- Monitor connection usage with `db.Stats()`

## Security Best Practices

### Credential Management
- Never commit `.env` file (listed in `.gitignore`)
- Use least-privilege SQL Server login (VIEW SERVER STATE only)
- Consider SQL Server authentication over Windows auth for containers

### Network Security
- Plain HTTP by default; enable TLS and/or basic auth with `--web.config.file` (exporter-toolkit format)
- Otherwise bind to localhost or use a reverse proxy with TLS
- Prometheus scraping should be on trusted network

### SQL Injection Prevention
- All queries use static SQL (no user input)
- No dynamic query building currently
- If adding user input, use parameterized queries

## Additional Resources

- [Prometheus Naming Best Practices](https://prometheus.io/docs/practices/naming/)
- [SQL Server DMV Reference](https://docs.microsoft.com/en-us/sql/relational-databases/system-dynamic-management-views/)
- [go-mssqldb Documentation](https://github.com/microsoft/go-mssqldb)
- [Prometheus Client Golang](https://github.com/prometheus/client_golang)

---

**Last Updated**: 2026-10-27
**Maintainer**: viniciussouzao
**AI Assistant**: This file is specifically designed to help AI assistants understand and work with this codebase effectively.
