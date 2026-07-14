# mssql-exporter

Prometheus exporter for Microsoft SQL Server. It collects a handful of instance-level metrics and exposes them on an HTTP endpoint for Prometheus to scrape, following a similar structure and naming style to other official database exporters (e.g. `mysqld_exporter`).

- Endpoint: `:9187/metrics`
- Driver: [`denisenkom/go-mssqldb`](https://github.com/denisenkom/go-mssqldb)

## Features

The exporter currently includes these collectors:

- Server info: basic instance information and memory/CPU facts.
- Sessions: number of sessions grouped by state.
- IO file stats: per-database file IO counters and stalls.
- Wait stats: wait time and counts by wait type.
- Memory clerks: memory usage by clerk type.
- Backups: last backup timestamp, duration and size per database.
- TempDB: space usage and temp table activity.
- Performance counters: batch requests, compilations, page life expectancy, buffer cache hit ratio, deadlocks, lock waits, user connections and more.
- Databases: state, recovery model, data/log sizes and log usage per database.
- Blocking: blocked sessions, longest current block and requests by status.
- Availability Groups: replica role, synchronization health and send/redo queues (emits nothing when Always On is disabled).
- SQL Agent jobs: enabled state, last run outcome, timestamp and duration.
- Volumes: total/available disk space per volume hosting database files, as seen by SQL Server (works for remote monitoring).
- Schedulers: runnable tasks (CPU pressure signal), active workers and work queue.

## Quick Start

1) Ensure you have access to a SQL Server instance and a login with permissions listed in "Permissions" below.

2) Build and run the exporter:

```bash
# Build
go build -o mssql-exporter ./

# Run (binary listens on :9187)
./mssql-exporter
```

3) Scrape with Prometheus (example snippet):

```yaml
scrape_configs:
  - job_name: 'sqlserver_exporter'
    static_configs:
      - targets: ['localhost:9187']
```

A ready-to-use `prometheus.yml` and a `docker-compose.yml` for Prometheus and Grafana are included. Run them with:

```bash
docker compose up -d
```

By default, that setup scrapes `host.docker.internal:9187`, assuming the exporter runs on your host.

## Configuration

Configuration comes from three sources, merged with the following precedence
(highest first): **CLI flag > environment variable > config file > default**.

Flags (mirroring the ergonomics of `mysqld_exporter`/`postgres_exporter`):

```
--web.listen-address string    Address to listen on for the web interface and metrics. (default ":9187")
--web.telemetry-path string    Path under which to expose metrics. (default "/metrics")
--collector.<name>             Enable/disable a SQL Server collector (all default to true).
                               Names: wait_stats, io_file_stats, sessions, server_info,
                               memory_clerks, backups, tempdb, perf_counters, databases,
                               blocking, availability_groups, agent_jobs, volumes, schedulers.
                               Example: --collector.backups=false
--collector.os-collectors      Enable OS-level collectors (CPU, memory, disk).
--config.file string           Path to an optional YAML config file.
--web.config.file string       Path to an exporter-toolkit web config file enabling TLS and/or basic auth.
--version                      Print version information and exit.
```

The database DSN has **no CLI flag on purpose**: process arguments are visible
via `ps`/task manager, which would leak credentials. Set it via the
`SQL_SERVER_DSN` environment variable (a `.env` file is loaded automatically)
or via the config file. Any DSN format supported by the `go-mssqldb` driver
works, e.g.:

```
SQL_SERVER_DSN=sqlserver://sa:MyStrongP%40ss@localhost:1433?database=master&app+name=mssql-exporter
```

Alternatively, use a YAML config file (handy for Windows services, where
editing environment variables is more cumbersome than pointing at a file next
to the executable). See `config.example.yml`:

```yaml
dsn: "sqlserver://sa:MyStrongP%40ss@localhost:1433?database=master"
web:
  listen_address: ":9187"
  telemetry_path: "/metrics"
collectors:
  os_collectors: false
  backups: false        # any SQL Server collector can be toggled by name
```

```bash
./mssql-exporter --config.file config.yml
```

Connection pool/timeouts: adjust defaults in `internal/db/options.go` (`DefaultOptions()`).

## Permissions

For **all default collectors** to succeed, the monitoring login needs the
following complete setup:

```sql
-- Server-level: required by the DMV-based collectors.
-- VIEW ANY DEFINITION is needed by io_file_stats, databases (sizes) and
-- volumes: without it sys.master_files silently returns zero rows.
GRANT VIEW SERVER STATE TO [your_login];
GRANT VIEW ANY DEFINITION TO [your_login];

-- msdb: required by the agent_jobs collector (enabled by default).
USE msdb;
CREATE USER [your_login] FOR LOGIN [your_login];
GRANT SELECT ON dbo.sysjobs TO [your_login];
GRANT SELECT ON dbo.sysjobhistory TO [your_login];
```

Without the msdb grants, `agent_jobs` reports
`sqlserver_scrape_collector_success{collector="agent_jobs"} 0` on every
scrape (the other collectors are unaffected); either apply the grants above
or disable it with `--collector.agent_jobs=false` if you don't monitor Agent
jobs (e.g. environments where SQL Agent is disabled).

Notes:

- The `backups` collector reads `msdb.dbo.backupset`, which is readable by
  `public` on default installations — no extra grant is usually needed.
- On SQL Server 2022+, `VIEW SERVER STATE` implies the more granular
  `VIEW SERVER PERFORMANCE STATE` / `VIEW SERVER SECURITY STATE` permissions
  that Microsoft now documents as the minimum for individual DMVs. You can go
  even more restrictive and grant only `VIEW SERVER PERFORMANCE STATE` if you
  disable the collectors that need more (e.g. `server_info` reads
  security-related server properties).

The exporter reads from (among others):

- `sys.dm_io_virtual_file_stats`
- `sys.master_files`
- `sys.dm_exec_sessions`
- `sys.dm_os_wait_stats`
- `sys.dm_os_sys_info`
- `sys.dm_os_sys_memory`
- `sys.dm_os_memory_clerks`
- `sys.dm_os_performance_counters`
- `sys.databases`
- `msdb.dbo.backupset`
- `SERVERPROPERTY(...)`

## Availability & Self-Metrics

The exporter starts even if SQL Server is unreachable and retries on every
scrape — alert on `sqlserver_up` instead of watching the process:

- `sqlserver_up 0|1` — whether the instance is reachable (checked per scrape).
- `sqlserver_scrape_collector_duration_seconds{collector="wait_stats"}` — duration of each collector.
- `sqlserver_scrape_collector_success{collector="wait_stats"} 0|1` — whether each collector succeeded.

When `sqlserver_up` is 0, individual collectors are skipped for that scrape.

**Known limitation**: every metric carries a `sqlserver_instance` label
resolved once at startup (`SELECT @@SERVERNAME`). If SQL Server is unreachable
at that moment the label stays empty until the exporter restarts — the metrics
themselves recover automatically, only this label does not. For a stable
identity prefer the `sqlserver_servername` metric or Prometheus target labels.

## Exposed Metrics

Metric names are stable and follow Prometheus best practices. Selected examples:

- Server info
  - `sqlserver_hostname{hostname="MACHINE"} 1`
  - `sqlserver_servername{servername="HOST\\INSTANCE"} 1`
  - `sqlserver_edition{edition="Developer Edition (64-bit)"} 1`
  - `sqlserver_product_version{product_version="16.0.4095.4"} 1`
  - `sqlserver_engine_edition 3` (numeric code)
  - `sqlserver_is_hadr_enabled 0|1`
  - `sqlserver_visible_cpus 8`
  - `sqlserver_total_physical_memory_mb 65536`
  - `sqlserver_available_physical_memory_mb 42000`

- Sessions
  - `sqlserver_sessions_total{session_state="running"} 12`
  - `sqlserver_sessions_total{session_state="sleeping"} 84`

- IO file stats (per file)
  - `sqlserver_io_file_reads_total{db_name="master",file_name="master",file_type="ROWS",physical_name="/var/opt/mssql/data/master.mdf"} 12345`
  - `sqlserver_io_file_writes_total{...} ...`
  - `sqlserver_io_file_read_bytes_total{...} ...`
  - `sqlserver_io_file_write_bytes_total{...} ...`
  - `sqlserver_io_file_stall_read_ms_total{...} ...`
  - `sqlserver_io_file_stall_write_ms_total{...} ...`

- Wait stats
  - `sqlserver_wait_time_ms{wait_type="PAGEIOLATCH_SH"} 123456`
  - `sqlserver_waiting_tasks{wait_type="PAGEIOLATCH_SH"} 789`
  - `sqlserver_max_wait_time_ms{wait_type="PAGEIOLATCH_SH"} 1024`

- Performance counters
  - `sqlserver_batch_requests_total 656`
  - `sqlserver_sql_compilations_total 628`
  - `sqlserver_sql_recompilations_total 5`
  - `sqlserver_page_life_expectancy_seconds 10064`
  - `sqlserver_buffer_cache_hit_ratio 1`
  - `sqlserver_lazy_writes_total 0`
  - `sqlserver_deadlocks_total 0`
  - `sqlserver_lock_waits_total 5`
  - `sqlserver_user_connections 6`
  - `sqlserver_page_splits_total 60`

- Databases
  - `sqlserver_database_state{db_name="master",state="ONLINE"} 1`
  - `sqlserver_database_recovery_model{db_name="master",recovery_model="SIMPLE"} 1`
  - `sqlserver_database_data_size_bytes{db_name="master"} 4915200`
  - `sqlserver_database_log_size_bytes{db_name="master"} 2097152`
  - `sqlserver_database_log_used_percent{db_name="master"} 71`

- Blocking
  - `sqlserver_blocked_sessions 0`
  - `sqlserver_blocked_max_wait_ms 0`
  - `sqlserver_active_requests{status="running"} 3`

- Availability Groups (only when Always On is enabled)
  - `sqlserver_hadr_replica_role{ag_name="ag1",replica_server="NODE1"} 1`
  - `sqlserver_hadr_replica_synchronization_health{ag_name="ag1",replica_server="NODE1"} 2`
  - `sqlserver_hadr_database_synchronization_state{ag_name="ag1",replica_server="NODE1",db_name="app"} 2`
  - `sqlserver_hadr_database_log_send_queue_kb{...} 0`
  - `sqlserver_hadr_database_redo_queue_kb{...} 0`

- SQL Agent jobs
  - `sqlserver_agent_job_enabled{job_name="nightly-backup"} 1`
  - `sqlserver_agent_job_last_run_status{job_name="nightly-backup"} 1`
  - `sqlserver_agent_job_last_run_timestamp_seconds{job_name="nightly-backup"} 1789344000`
  - `sqlserver_agent_job_last_run_duration_seconds{job_name="nightly-backup"} 42`

  There is intentionally no aggregate "failed jobs" metric — derive it in
  PromQL, which also tells you *which* jobs failed:

  ```promql
  count(sqlserver_agent_job_last_run_status == 0)   # how many jobs failed last run
  sqlserver_agent_job_last_run_status == 0          # which jobs failed last run
  ```

- Volumes
  - `sqlserver_volume_total_bytes{volume_mount_point="C:\\"} 1962694664192`
  - `sqlserver_volume_available_bytes{volume_mount_point="C:\\"} 1846100799488`

- Schedulers
  - `sqlserver_schedulers_runnable_tasks 1`
  - `sqlserver_schedulers_current_tasks 53`
  - `sqlserver_schedulers_active_workers 49`
  - `sqlserver_schedulers_work_queue 0`
  - `sqlserver_schedulers_online 8`

## Running With Prometheus and Grafana

This repository includes:

- `prometheus.yml`: simple scrape job for the exporter.
- `docker-compose.yml`: Prometheus and Grafana services.

Steps:

1) Start the exporter locally on `:9187`.
2) `docker compose up -d` to bring up Prometheus and Grafana.
3) Prometheus UI: http://localhost:9090
4) Grafana UI: http://localhost:3000 (admin/admin by default; password set in `docker-compose.yml`).

Add Prometheus (http://prometheus:9090 inside Compose network or http://localhost:9090 from your browser) as a Grafana data source and build dashboards from the metric names above.

## Security

- TLS and basic authentication are supported via
  [exporter-toolkit web configuration](https://github.com/prometheus/exporter-toolkit/blob/master/docs/web-configuration.md):
  pass `--web.config.file web-config.yml` (or the `web.config_file` key in the
  YAML config). Example enabling TLS:

  ```yaml
  tls_server_config:
    cert_file: server.crt
    key_file: server.key
  ```

  Without a web config file the exporter serves plain HTTP.
- A `/healthz` endpoint returns 200 while the process is alive (liveness);
  use `sqlserver_up` for database reachability.
- Avoid embedding secrets in source control. Keep the DSN in environment variables or a config file excluded from git.

## Troubleshooting

- Cannot connect / ping fails: verify host, port, authentication mode, and that the SQL Server accepts TCP connections.
- Permission errors or empty metrics: ensure the login has `VIEW SERVER STATE` and access to the queried DMVs.
- Panic on scrape: fixed in current code by aligning metric labels and descriptors. If you see a panic, check the logs and validate label counts vs. descriptors.
- No metrics in Prometheus: confirm Prometheus can reach `:9187/metrics` and that the scrape job is correctly targeted.

## Roadmap

See [ROADMAP.md](ROADMAP.md) for the full phased plan — phases 1-4 (core
metrics, exporter robustness, operational collectors, TLS/healthz) are done.
Next ideas include query performance metrics, index health and integration
tests with SQL Server in Docker.

---

This project is a work-in-progress exporter tailored for SQL Server, aiming for parity in usability and quality with other database exporters in the Prometheus ecosystem.
