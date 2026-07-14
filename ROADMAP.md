# Roadmap

Feature roadmap for mssql-exporter, ordered by implementation phases. Each
phase is meant to be delivered as one or more focused PRs. Prioritization
rationale: highest dashboard/alerting value per effort first, then exporter
robustness before adding heavier collectors.

Progress legend: [ ] not started · [~] in progress · [x] done

## Phase 1 — Core metrics gap (highest value per effort)

- [x] **Performance counters collector** (`sys.dm_os_performance_counters`) —
  `internal/collector/perf_counters.go`
  - Batch Requests, SQL Compilations/Re-Compilations, Page Life Expectancy,
    Buffer Cache Hit Ratio (ratio/base pair divided in Go), Lazy Writes,
    Deadlocks, Lock Waits, User Connections, Page Splits
- [x] **Databases collector** (`sys.databases`, `sys.master_files`,
  `Percent Log Used` perf counter) — `internal/collector/databases.go`
  - State and recovery model (value 1, labeled), data/log sizes in bytes,
    log used percent per database

## Phase 2 — Exporter robustness (before adding heavier collectors)

- [x] **`sqlserver_up` metric + connection resilience** — the exporter starts
  even if SQL Server is down, pings per scrape and exposes `sqlserver_up 0/1`;
  scrapers are skipped while down (`internal/collector/exporter.go`)
- [x] **Per-collector enable/disable flags** (`--collector.<name>`) and YAML
  `collectors:` map, wired through `internal/config`
- [x] **Exporter self-metrics**: `sqlserver_scrape_collector_duration_seconds`
  and `sqlserver_scrape_collector_success{collector}`
- Note: collectors now implement the `Scraper` interface (`Scrape() error`)
  instead of raw `prometheus.Collector` — see CLAUDE.md "Adding a New Collector"

## Phase 3 — Operational collectors

- [x] **Blocking collector** (`sys.dm_exec_requests`) —
  `internal/collector/blocking.go`: blocked sessions, longest current block,
  requests by status
- [x] **Availability Groups collector** (`sys.dm_hadr_availability_replica_states`,
  `sys.dm_hadr_database_replica_states`) —
  `internal/collector/availability_groups.go`: replica role, synchronization
  health, send/redo queues; emits nothing when HADR is disabled
- [x] **SQL Agent jobs collector** (`msdb.dbo.sysjobs` + `sysjobhistory`) —
  `internal/collector/agent_jobs.go`: enabled state, last run outcome,
  timestamp and duration; requires SELECT grants on the two msdb tables
  (documented in README)

## Phase 4 — Environment & hardening

- [x] **Volume space collector** (`sys.dm_os_volume_stats`) —
  `internal/collector/volumes.go`: total/available bytes per volume, via SQL
  (works for remote monitoring); requires `VIEW ANY DEFINITION` in addition
  to `VIEW SERVER STATE` (sys.master_files is silently empty without it)
- [x] **Scheduler/CPU pressure collector** (`sys.dm_os_schedulers`) —
  `internal/collector/schedulers.go`: runnable tasks, current tasks, active
  workers, work queue, online schedulers
- [x] **TLS and basic auth** via `prometheus/exporter-toolkit/web` and
  `--web.config.file` (also `web.config_file` in the YAML config)
- [x] **Health check endpoint** (`/healthz`) — liveness only; use
  `sqlserver_up` for database reachability

## Later / ideas (unprioritized)

- Query performance metrics (`sys.dm_exec_query_stats`) — high cardinality,
  needs careful label design and possibly a longer scrape interval
- Index health (fragmentation, missing indexes) — expensive queries, opt-in only
- Log shipping / replication status
- Unit tests for collectors (mock `*sql.DB`), integration tests with
  SQL Server in Docker

## Done

- [x] Lint fixes (`errcheck` on `rows.Close`), error check on listen (PR #7)
- [x] Single-DSN connection (`SQL_SERVER_DSN`, any go-mssqldb format) (PR #7)
- [x] CLI flags (`pflag`) + optional YAML config file with
  flag > env > file > default precedence (PR #7)
- [x] First unit tests (`internal/config`) (PR #7)
