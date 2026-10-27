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

## Phase 7 — Transaction log & Query Store health

Decided 2026-10-19 after a production incident: a lagging AG secondary kept
the primary's log from truncating (`AVAILABILITY_REPLICA`); the same happens
with CDC/replication (`REPLICATION`). Alert-friendly, low-cardinality signals
first; implement in this order.

- [x] **Log reuse wait** (`sys.databases.log_reuse_wait_desc`) in the
  `databases` collector: `sqlserver_database_log_reuse_wait{db_name,reason}`
  (1 for the current reason). Works on every version and on RDS with no new
  permission. Document that `LOG_BACKUP` is normal for FULL recovery between
  log backups and that the value is refreshed on truncation attempts
  (checkpoint/log backup), so alert with `for:` and together with
  `sqlserver_database_log_used_percent`
- [x] **Query Store health** (new `query_store` collector, per database like
  `index_fragmentation`, `HAS_DBACCESS` filter):
  - `sys.database_query_store_options`: `actual_state` vs `desired_state`
    (the silent switch to `READ_ONLY` when the store fills up),
    `readonly_reason`, `current_storage_size_mb` / `max_storage_size_mb`
  - forced plans that fail (`sys.query_store_plan.force_failure_count`,
    `last_force_failure_reason_desc`)
  - no per-query series here (see Later)
  - Done as the `query_store` collector (opt-in): background loop like
    index_fragmentation (default every 5m, options
    `--collector.query_store.interval|databases`), database listing shared
    with index_fragmentation; one-hot `actual_state`/`desired_state`,
    `readonly_reason`, storage used/max bytes, forced plans and failing
    forced plans. Validated on SQL Server 2025: `READ_ONLY` with reason
    65536 after lowering `MAX_STORAGE_SIZE_MB`, forced plan failing with
    `NO_INDEX` after dropping its index, `OFF`; needs only database access
    on top of `VIEW SERVER STATE`

## Later / ideas (unprioritized)

- Missing indexes (`sys.dm_db_missing_index_*`)
- `wait_stats`: filter idle/background waits in the exporter (today only
  `SLEEP%`, `XE%` and `BROKER_%` are skipped), e.g. `LAZYWRITER_SLEEP`,
  `CHECKPOINT_QUEUE`, `HADR_WORK_QUEUE`, `DIRTY_PAGE_POLL`, keeping waits
  that slow queries such as `HADR_SYNC_COMMIT`; the RDS dashboard applies
  the same list in PromQL meanwhile (HML-031). On Azure SQL Managed Instance
  the top waits are all background ones: `SOS_WORK_DISPATCHER`,
  `HADR_WORK_QUEUE`, `DISPATCHER_QUEUE_SEMAPHORE`, `LOGMGR_QUEUE` (HML-039). Changes the series exposed:
  make it a documented default, possibly with an option to keep everything
- Automatic `sqlserver_instance` label resolved on the first successful
  connection instead of only at startup (today it stays empty until a
  restart when SQL Server is down at startup; `instance_label` avoids it)
  (HML-029)
- Transaction log health from `sys.dm_db_log_stats` (one row per database,
  works on RDS where the `Percent Log Used` counter misses the system
  databases): VLF count and log generated since the last log backup
  (HML-026). The truncation reason is already covered by
  `sqlserver_database_log_reuse_wait` (Phase 7)
- Per-query statistics from Query Store (`sys.query_store_runtime_stats`):
  history that survives restarts and plan cache eviction, plan changes, waits
  per query (2017+). Deferred: data is aggregated in intervals (60 min by
  default) and cleanup purges old intervals, so totals can decrease and break
  `rate()`; per-database scans are expensive and series churn is the same as
  `top_queries`. The Query Store UI remains the better tool for analysis
- `volumes`: skip databases changing state (`sys.databases.state IN (0, 1)`,
  keeping `RESTORING` secondaries) to narrow the window where
  `sys.dm_os_volume_stats` waits and the query times out (HML-019; low
  impact, the next scrape succeeds)
- Concurrent scrapes: with 5 or more simultaneous `/metrics` requests the
  fixed pool of 5 connections saturates and collectors fail with `context
  deadline exceeded` behind an HTTP 200 (HML-034). Let simultaneous scrapes
  wait for the collection in flight and share its result (singleflight), and
  make the pool size configurable
- `volumes`: SQL Server on Linux (Cloud SQL and the Docker images) returns
  NULL mount point, name and file system from `sys.dm_os_volume_stats`, so
  the series has an empty `volume_mount_point`; fall back to a label derived
  from the file path (HML-032, HML-038)
- Windows OS collector: `os_disk_read_queue_length` / `os_disk_write_queue_length`
  are exposed as gauges over raw cumulative WMI counters; fixing needs a new
  name/type aligned with Linux (pending from HML-003)
- Replication status
- Unit tests for collectors (mock `*sql.DB`), integration tests with
  SQL Server in Docker
- Microsoft Entra ID authentication (Azure SQL Managed Instance and SQL
  Database): `go-mssqldb/azuread` and the `azuresql` driver when the DSN has
  `fedauth=` (managed identity, service principal, default credential)
- Azure SQL Database mode: one connection per database (the unit is the
  database, no cross-database queries), `sys.dm_db_resource_stats`, elastic
  pools and geo-replication instead of AGs; detect it with
  `SERVERPROPERTY('EngineEdition') = 5`. Size it after the reconnaissance
  pass of the Azure homologation phase
- Helm chart and a homologation of the exporter on GKE, after the first
  public release

## Done

- [x] Lint fixes (`errcheck` on `rows.Close`), error check on listen (PR #7)
- [x] Single-DSN connection (`SQL_SERVER_DSN`, any go-mssqldb format) (PR #7)
- [x] CLI flags (`pflag`) + optional YAML config file with
  flag > env > file > default precedence (PR #7)
- [x] First unit tests (`internal/config`) (PR #7)
- [x] Linux service: hardened systemd unit (`contrib/systemd`) shipped in the
  release archives, install steps in the README (HML-025)
- [x] Azure SQL Managed Instance storage: `volumes` reports
  `sqlserver_instance_storage_{reserved,used}_bytes` from
  `sys.server_resource_stats` instead of `dm_os_volume_stats`, whose
  `total_bytes` is the size of the files there, not the volume capacity
  (Azure homologation phase)
