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
  - Decision (2026-07-14): the originally planned aggregate "failed jobs
    count" metric was dropped in favor of PromQL derivation
    (`count(sqlserver_agent_job_last_run_status == 0)`), documented in the
    README — per-job status is strictly more informative
  - Decision (2026-07-14): stays enabled by default (SQL Agent is rarely
    disabled outside dev/QA); the required msdb grants are documented as
    part of the standard permissions setup

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

## Phase 5 — Legacy HA/DR

- [x] **Database mirroring collector** (`sys.database_mirroring` +
  `Database Mirroring` performance counters) —
  `internal/collector/database_mirroring.go`: role, state, safety level,
  witness state, send/redo queues, log throughput and transaction delay;
  emits nothing when no database is mirrored
- [x] **Log shipping collector** (`msdb.dbo.log_shipping_monitor_primary` /
  `_secondary`) — `internal/collector/log_shipping.go`: last backup, copy and
  restore timestamps (already UTC), restore latency and thresholds
  - Decision (2026-10-13): disabled by default (`--collector.log_shipping`
    to enable) — niche setup that needs SELECT grants on the two monitor
    tables; enabling it by default would report scrape failures on most
    instances after an upgrade

## Phase 6 — Workload & index insights

Inspired by a Datadog DBM proof of concept (top queries with text, execution
count, duration and currently running load). Decided 2026-10-15; implement in
this order.

- [x] **Sessions by login** (`sys.dm_exec_sessions`, user processes only) —
  `sqlserver_sessions_by_login{login_name}` in the `sessions` collector.
  Cardinality follows the number of logins with open sessions (bounded in
  practice; document it for Windows-auth shops with one login per person)
- [x] **Active queries, aggregated** (`sys.dm_exec_requests`,
  `sys.dm_tran_active_transactions`) — no per-query labels: running user
  requests above a duration threshold, longest running request duration,
  age of the oldest open transaction. Cheap, alert-friendly ("a query has
  been stuck for 20 minutes"); candidate for default on. Reuse the AG
  976/983 retry from `blocking`. Done as the `active_queries` collector
  (default on): `sqlserver_requests_long_running{over_seconds}`,
  `sqlserver_request_max_elapsed_seconds{command}` (per command so alerts can
  skip `WAITFOR`/`BACKUP`), open transactions, oldest open transaction age
  and sleeping sessions holding a transaction
- [x] **Top queries** (`sys.dm_exec_query_stats` aggregated by `query_hash`) —
  **disabled by default**. Follows `postgres_exporter`'s `stat_statements`
  and `mysqld_exporter`'s `perf_schema.eventsstatements`:
  - top N by CPU (`--collector.top_queries.limit`, default 50–100), labels
    `db_name` + `query_hash`; counters for executions, CPU, elapsed, logical
    and physical reads, writes (average duration = rate(elapsed)/rate(executions))
  - query text **off by default**; when enabled, a truncated label (e.g. 120
    chars) with **literals obfuscated** (`'...'`/numbers → `?`, as Datadog
    does): ad hoc statements carry literal values (customer data, IDs) that
    must not land in Prometheus
  - document: series churn when queries enter/leave the top N (gaps in
    `rate()`), counter resets on plan cache eviction, DMV scan cost on large
    plan caches (may need its own collection interval, see below)
  - Done as the `top_queries` collector: one plan cache scan per scrape
    (window functions + TOP before the per-row APPLYs), options
    `--collector.top_queries.limit|include_text|text_length` (also
    `top_queries:` in the YAML config), statement text in a separate
    `sqlserver_top_query_info` series joined on `query_hash`
- [x] **Index fragmentation** (`sys.dm_db_index_physical_stats`, `LIMITED`
  mode) — **disabled by default**, heavy I/O on large objects:
  - collected **in the background on its own interval** (e.g.
    `--collector.index_fragmentation.interval=6h`), one database at a time
    with its own timeout; the scrape serves the cached result plus
    `_last_collection_timestamp_seconds` / `_last_collection_duration_seconds`.
    First collector outside the per-scrape model — build the async
    infrastructure so other expensive collectors can reuse it
  - labels db/schema/table/index; filters: minimum page count (e.g. 1000),
    database include/exclude list, skip inaccessible databases (mirror,
    restoring, AG non-readable secondary)
  - metrics: `avg_fragmentation_in_percent`, `page_count`. Page density
    (`avg_page_space_used_in_percent`) needs `SAMPLED` mode — out of scope
    for now; document that on SSD/EBS storage density matters more than
    logical fragmentation
  - Done as the `index_fragmentation` collector: background loop started on
    the first scrape, bounded by the interval, scrapes serve the cached
    result; `HAS_DBACCESS` database filter (needs `CONNECT ANY DATABASE` to
    cover databases without a user for the login), three-part catalog names
    with bracket-quoted database names, heaps with an empty `index_name`
- Validated on EC2 with SQL Server 2025 (HML-021 to HML-024); repeat in the
  RDS phase

## Later / ideas (unprioritized)

- Missing indexes (`sys.dm_db_missing_index_*`)
- `volumes`: skip databases changing state (`sys.databases.state` filter) to
  narrow the window where `sys.dm_os_volume_stats` waits and the query times
  out (HML-019; low impact, the next scrape succeeds)
- Windows OS collector: `os_disk_read_queue_length` / `os_disk_write_queue_length`
  are exposed as gauges over raw cumulative WMI counters; fixing needs a new
  name/type aligned with Linux (pending from HML-003)
- Replication status
- Unit tests for collectors (mock `*sql.DB`), integration tests with
  SQL Server in Docker

## Done

- [x] Lint fixes (`errcheck` on `rows.Close`), error check on listen (PR #7)
- [x] Single-DSN connection (`SQL_SERVER_DSN`, any go-mssqldb format) (PR #7)
- [x] CLI flags (`pflag`) + optional YAML config file with
  flag > env > file > default precedence (PR #7)
- [x] First unit tests (`internal/config`) (PR #7)
- [x] Linux service: hardened systemd unit (`contrib/systemd`) shipped in the
  release archives, install steps in the README (HML-025)
