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

- [ ] **`sqlserver_up` metric + connection resilience**
  - Do not exit when SQL Server is unreachable at startup; keep serving
    `/metrics` with `sqlserver_up 0` and let Prometheus alert
    (same failure model as `mysqld_exporter`'s `mysql_up`)
  - Avoids crash-loops during database maintenance windows
- [ ] **Per-collector enable/disable flags** (`--collector.<name>`,
  mysqld_exporter pattern), wired through `internal/config` (flag/env/YAML)
- [ ] **Exporter self-metrics**: per-collector scrape duration and success
  (`sqlserver_scrape_collector_duration_seconds`,
  `sqlserver_scrape_collector_success{collector}`)

## Phase 3 — Operational collectors

- [ ] **Blocking/locks collector** (`sys.dm_exec_requests`)
  - Blocked sessions count, max blocking wait time, requests by status
- [ ] **Availability Groups collector** (`sys.dm_hadr_availability_replica_states`,
  `sys.dm_hadr_database_replica_states`)
  - Replica role (primary/secondary), synchronization health,
    redo queue size, send queue size
  - Must degrade gracefully when HADR is disabled (emit nothing, no errors)
- [ ] **SQL Agent jobs collector** (`msdb.dbo.sysjobs` + `sysjobhistory`)
  - Last run status/outcome, last run duration, failed jobs count
  - Requires read access to `msdb` — document the extra permission

## Phase 4 — Environment & hardening

- [ ] **Volume space collector** (`sys.dm_os_volume_stats`)
  - Free/total space per volume hosting database files, via SQL —
    covers remote-monitoring setups where OS collectors don't apply
- [ ] **Scheduler/CPU pressure collector** (`sys.dm_os_schedulers`)
  - Runnable tasks count (most direct CPU starvation signal), workers
- [ ] **TLS and basic auth** for the HTTP endpoint
  (consider `prometheus/exporter-toolkit/web` + `--web.config.file`)
- [ ] **Health check endpoint** (`/healthz`)

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
