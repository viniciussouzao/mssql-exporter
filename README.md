# mssql-exporter

Prometheus exporter for Microsoft SQL Server. It collects a handful of instance-level metrics and exposes them on an HTTP endpoint for Prometheus to scrape, following a similar structure and naming style to other official database exporters (e.g. `mysqld_exporter`).

- Endpoint: `:9187/metrics`
- Driver: [`microsoft/go-mssqldb`](https://github.com/microsoft/go-mssqldb)
- Validated against SQL Server 2019 (CU32), 2022 (RTM) and 2025 (RTM) on Windows,
  including rolling upgrades of an Availability Group, database mirroring and
  log shipping across those versions

## Features

The exporter currently includes these collectors:

- Server info: basic instance information and memory/CPU facts.
- Sessions: number of sessions grouped by state and user sessions per login.
- IO file stats: per-database file IO counters and stalls.
- Wait stats: wait time and counts by wait type.
- Memory clerks: memory usage by clerk type.
- Backups: last backup timestamp (UTC, correct for servers in any time zone), duration and size per database, counting only backups taken by the instance itself (restores recorded in `backupset`, e.g. on a log shipping secondary, are ignored).
- TempDB: space usage and temp table activity.
- Performance counters: batch requests, compilations, page life expectancy, buffer cache hit ratio, deadlocks, lock waits, user connections and more.
- Databases: state, recovery model, data/log sizes and log usage per database.
- Blocking: blocked sessions, longest current block and requests by status.
- Active queries: long-running user requests and open user transactions, aggregated (no per-query labels).
- Availability Groups: replica role, synchronization health and send/redo queues (emits nothing when Always On is disabled).
- SQL Agent jobs: enabled state, last run outcome, timestamp and duration.
- Volumes: total/available disk space per volume hosting database files, as seen by SQL Server (works for remote monitoring).
- Schedulers: runnable tasks (CPU pressure signal), active workers and work queue.
- Database mirroring: role, state, safety level, witness state and send/redo queues per mirrored database (emits nothing when no database is mirrored).
- Log shipping (opt-in): last backup/copy/restore timestamps, restore latency and alert thresholds from the log shipping monitor tables.
- Top queries (opt-in): top N queries by CPU from the plan cache, per `query_hash`, with optional statement text (literals obfuscated).
- Index fragmentation (opt-in): logical fragmentation and size per index, collected in the background on its own interval (default 6h).

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
--collector.<name>             Enable/disable a SQL Server collector (all default to true
                               except log_shipping, top_queries and index_fragmentation).
                               Names: wait_stats, io_file_stats, sessions, server_info,
                               memory_clerks, backups, tempdb, perf_counters, databases,
                               blocking, active_queries, availability_groups, agent_jobs,
                               volumes, schedulers, database_mirroring, log_shipping,
                               top_queries, index_fragmentation.
                               Example: --collector.backups=false, --collector.log_shipping
--collector.top_queries.limit int
                               Queries (by total CPU) exposed by top_queries (default 50, max 1000).
--collector.top_queries.include_text
                               Expose the statement text in sqlserver_top_query_info (default false).
--collector.top_queries.text_length int
                               Maximum statement text length (default 120, 16-4000).
--collector.index_fragmentation.interval duration
                               Interval between background collections (default 6h, min 1m).
--collector.index_fragmentation.min_page_count int
                               Skip indexes with fewer 8 KB pages (default 1000).
--collector.index_fragmentation.databases strings
                               Databases to read, comma-separated (default: all accessible user databases).
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

### Encrypted connections (`encrypt=strict`)

By default the driver encrypts only the login packet. Add `encrypt=true` for a
fully encrypted TDS 7.x connection, or `encrypt=strict` for TDS 8.0 (SQL Server
2022+), where TLS is negotiated before any TDS traffic:

```
sqlserver://user:pass@NODE1:1433?database=master&encrypt=strict&certificate=C:/mssql-exporter/node1.pem
```

- SQL Server only accepts TDS 8.0 with a certificate configured for the
  instance (SQL Server Configuration Manager → Protocols → Certificate); with
  the self-generated certificate it closes the connection during the handshake.
- The driver always validates the server certificate in strict mode:
  `trustservercertificate=true` is ignored. Point it at the certificate
  instead (PEM): `certificate=<file>` validates the chain and the host name,
  `servercertificate=<file>` only compares the certificate bytes (no host name
  check). When the DSN connects by IP or alias, set
  `hostnameincertificate=<name in the certificate>`.
- Confirm on the server with `sys.dm_exec_connections`: `encrypt_option` TRUE
  and `protocol_version` 134217728 (0x08000000, TDS 8.0).
- The certificate file must end in `.pem` or `.der`. Any other extension
  (e.g. `.crt`) makes the DSN invalid and the exporter exits at startup.
- **Amazon RDS:** download the AWS CA bundle
  (`https://truststore.pki.rds.amazonaws.com/global/global-bundle.pem`) and
  use it with the instance endpoint as the host, with `encrypt=true` or
  `encrypt=strict`. Both are validated on RDS for SQL Server 2025, and no
  server-side change is needed:
  `sqlserver://user:pass@mydb.xxxx.us-east-1.rds.amazonaws.com:1433?database=master&encrypt=strict&certificate=/etc/mssql-exporter/global-bundle.pem`

### Running as a Windows service

The exporter runs natively under the Windows Service Control Manager; no
wrapper (NSSM, WinSW) is needed. From an elevated PowerShell:

```powershell
.\mssql-exporter.exe service install --config.file C:\ProgramData\mssql-exporter\config.yml
.\mssql-exporter.exe service start
.\mssql-exporter.exe service status
```

- `install` validates the config file, then registers the service with
  automatic (delayed) start and restart-on-failure recovery actions (5s, 30s,
  60s). The executable and config paths are stored as absolute paths, so keep
  the binary where it was installed from.
- The service runs as `LocalSystem`. Restrict the config file, which holds the
  DSN, to SYSTEM and Administrators:
  `icacls config.yml /inheritance:r /grant:r "SYSTEM:R" "Administrators:F"`.
- Logs go to the Windows Event Log (`Application`, source `mssql-exporter`).
- Relative paths in the config file (e.g. `web.config_file`) are resolved from
  the config file's directory.
- `--name` (default `mssql-exporter`) lets several exporters run on one host,
  e.g. one per SQL Server instance.
- Other actions: `stop`, `uninstall`. Run `mssql-exporter service --help` for
  details. Allow the listen port in Windows Firewall separately.

### Running as a Linux service (systemd)

The Linux release archives include a hardened unit file,
`contrib/systemd/mssql-exporter.service`. It runs the exporter as a dedicated
system user with no login shell and no home directory, and reads
`/etc/mssql-exporter/config.yml`. As root, or with `sudo` on each command:

```bash
VERSION=0.9.1   # release to install (archives from v0.9.1 include the unit)
ARCH=amd64      # arm64 on Graviton
URL=https://github.com/viniciussouzao/mssql-exporter/releases/download/v${VERSION}
curl -fsSLO ${URL}/mssql-exporter_${VERSION}_linux_${ARCH}.tar.gz
curl -fsSLO ${URL}/checksums.txt
sha256sum --check --ignore-missing checksums.txt
mkdir -p mssql-exporter && tar xzf mssql-exporter_${VERSION}_linux_${ARCH}.tar.gz -C mssql-exporter
cd mssql-exporter

install -m 0755 mssql-exporter /usr/local/bin/
useradd --system --no-create-home --shell /usr/sbin/nologin mssql-exporter
install -d -m 0750 -o root -g mssql-exporter /etc/mssql-exporter
install -m 0640 -o root -g mssql-exporter config.example.yml /etc/mssql-exporter/config.yml
install -m 0644 contrib/systemd/mssql-exporter.service /etc/systemd/system/
```

Edit `/etc/mssql-exporter/config.yml` (DSN and collectors), then:

```bash
systemctl daemon-reload
systemctl enable --now mssql-exporter
systemctl status mssql-exporter
curl -s localhost:9187/metrics | grep sqlserver_up
```

- The config file holds the DSN, so it is readable only by root and the
  service group (`0640 root:mssql-exporter`). Keep certificates referenced by
  the DSN (`certificate=...`) in `/etc/mssql-exporter` with the same mode.
- Logs go to the journal: `journalctl -u mssql-exporter -f`.
- Relative paths in the config file are resolved from `/etc/mssql-exporter`,
  as with the Windows service.
- The unit restarts the exporter 5s after a failure, and a stop sends
  `SIGTERM` for a clean shutdown.
- It is sandboxed: read-only system, no capabilities, no new privileges
  (`systemd-analyze security mssql-exporter` rates it 3.1, "OK").
- If Prometheus reaches the exporter through an SSH tunnel (e.g. from a
  bastion host), bind it to loopback with `web.listen_address:
  "127.0.0.1:9187"` and forward the port:
  `ssh -N -L 19187:127.0.0.1:9187 user@bastion`.
- When the exporter runs on a different host than SQL Server (a bastion, or
  any managed instance such as RDS), leave `os_collectors` disabled: they
  would describe the exporter's host, not the database server.
- To upgrade, replace `/usr/local/bin/mssql-exporter` and run
  `systemctl restart mssql-exporter`.

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

The `log_shipping` collector is **disabled by default**. To enable it
(`--collector.log_shipping` or `collectors: {log_shipping: true}`), also grant
read access to the monitor tables on the primary and secondary instances:

```sql
USE msdb;
GRANT SELECT ON dbo.log_shipping_monitor_primary TO [your_login];
GRANT SELECT ON dbo.log_shipping_monitor_secondary TO [your_login];
```

The `index_fragmentation` collector is **disabled by default**. It reads
every online user database the login can access (`HAS_DBACCESS`), and a
database where the login has no user is silently skipped; to cover all of
them without creating a user in each:

```sql
GRANT CONNECT ANY DATABASE TO [your_login];
```

`CONNECT ANY DATABASE` exists since SQL Server 2014; on older versions create
a user for the login in each database. A skipped database is not an error:
without the grant the collection succeeds with
`sqlserver_index_fragmentation_last_collection_databases 0`, so compare that
metric with the number of databases you expect (or alert on `== 0`).

On **Amazon RDS** the master user cannot grant `CONNECT ANY DATABASE`
(`Grantor does not have GRANT permission`). Create a user for the login in
each database to cover instead (it grants access to the database, not to its
data):

```sql
USE [your_database];
CREATE USER [your_login] FOR LOGIN [your_login];
```

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
- `sys.dm_os_waiting_tasks`
- `sys.databases`
- `msdb.dbo.backupset`
- `sys.database_mirroring`
- `msdb.dbo.log_shipping_monitor_primary` / `log_shipping_monitor_secondary` (log_shipping only)
- `sys.dm_exec_query_stats`, `sys.dm_exec_plan_attributes`, `sys.dm_exec_sql_text` (top_queries only)
- `sys.dm_db_index_physical_stats` and each database's `sys.indexes`/`sys.objects`/`sys.schemas` (index_fragmentation only)
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
  - `sqlserver_sessions_by_login{login_name="app_user"} 40` (user sessions
    only, `is_user_process = 1`; one series per login with open sessions,
    so expect many series where each person connects with their own
    Windows login)

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
  - `sqlserver_database_log_used_percent{db_name="master"} 71`: the `Percent
    Log Used` performance counter, reported only for databases listed in
    `sys.databases` (the counter also has instances for hidden databases such
    as `mssqlsystemresource`). On Amazon RDS the counter has no instance for
    `master`, `model`, `msdb` and `rdsadmin`, so they have no series there;
    user databases are reported normally

- Blocking
  - `sqlserver_blocked_sessions 0` (from `sys.dm_os_waiting_tasks`; a parallel
    request waiting on its own session is not counted)
  - `sqlserver_blocked_max_wait_ms 0`
  - `sqlserver_active_requests{status="running"} 3` (on an AG replica,
    `sys.dm_exec_requests` intermittently fails while an availability
    database is momentarily not accessible: error 976 on a secondary, 983
    while the replica is still resolving its role right after SQL Server
    starts; the collector retries once before reporting the scrape as failed)

- Active queries (user sessions only; the exporter's own request is excluded)
  - `sqlserver_requests_long_running{over_seconds="60"} 2`: user requests
    running for at least 10, 60, 300 and 1800 seconds (always emitted, 0 when
    idle)
  - `sqlserver_request_max_elapsed_seconds{command="SELECT"} 75`: longest
    running user request per command, only for commands currently running.
    Exclude commands that wait by design when alerting, e.g.
    `max(sqlserver_request_max_elapsed_seconds{command!~"WAITFOR|BACKUP.*|RESTORE.*|DBCC.*"}) > 600`
    (a Service Broker `WAITFOR (RECEIVE ...)` or a polling loop can run for days)
  - `sqlserver_open_transactions 4`
  - `sqlserver_oldest_open_transaction_seconds 620` (0 when there is none):
    long open transactions hold locks, keep row versions in tempdb and
    prevent log truncation
  - `sqlserver_sleeping_sessions_with_open_transaction 1`: idle sessions
    holding a transaction, typically a `BEGIN TRAN` never committed
  - Elapsed times are computed in server local time (`GETDATE()`), so a
    daylight saving change can skew them by one hour for requests spanning it
    (negative values are reported as 0)

- Availability Groups (only when Always On is enabled)
  - `sqlserver_hadr_replica_role{ag_name="ag1",replica_server="NODE1"} 1`
  - `sqlserver_hadr_replica_synchronization_health{ag_name="ag1",replica_server="NODE1"} 2`
  - `sqlserver_hadr_database_synchronization_state{ag_name="ag1",replica_server="NODE1",db_name="app"} 2`
  - `sqlserver_hadr_database_log_send_queue_kb{...} 0`
  - `sqlserver_hadr_database_redo_queue_kb{...} 0`
  - `sqlserver_hadr_database_is_suspended{...} 0` (data movement suspended; the
    log send queue is absent while suspended, so alert on this instead)
  - `sqlserver_hadr_database_suspend_reason{...} 0` (only while suspended; 0 = user,
    1 = partner/forced failover, 2 = redo, 3 = capture, 4 = apply, 5 = restart,
    6 = undo, 7 = revalidation, 8 = synchronization point error)
  - `sqlserver_hadr_database_secondary_lag_seconds{...} 0` (SQL Server 2016+;
    stays 0 while data movement is suspended, so don't alert on it alone)
  - Suspending on the primary marks the *primary's* database replica as
    suspended: alert on `max by (ag_name, db_name) (sqlserver_hadr_database_is_suspended) == 1`.

- SQL Agent jobs
  - `sqlserver_agent_job_enabled{job_name="nightly-backup"} 1`
  - `sqlserver_agent_job_last_run_status{job_name="nightly-backup"} 1`
  - `sqlserver_agent_job_last_run_timestamp_seconds{job_name="nightly-backup"} 1789344000`
    (UTC Unix time, converted from the server's local time with its current offset)
  - `sqlserver_agent_job_last_run_duration_seconds{job_name="nightly-backup"} 42`

  There is intentionally no aggregate "failed jobs" metric — derive it in
  PromQL, which also tells you *which* jobs failed:

  ```promql
  count(sqlserver_agent_job_last_run_status == 0)   # how many jobs failed last run
  sqlserver_agent_job_last_run_status == 0          # which jobs failed last run
  ```

- Database mirroring (only for mirrored databases)
  - `sqlserver_mirroring_info{db_name="app",partner_instance="NODE2"} 1`
  - `sqlserver_mirroring_role{db_name="app"} 1` (1 = principal, 2 = mirror)
  - `sqlserver_mirroring_state{db_name="app"} 4` (0 = suspended, 1 = disconnected,
    2 = synchronizing, 3 = pending failover, 4 = synchronized,
    5 = not synchronized, 6 = synchronized/failover possible)
  - `sqlserver_mirroring_safety_level{db_name="app"} 2` (1 = off/async, 2 = full/sync)
  - `sqlserver_mirroring_witness_state{db_name="app"} 1` (only with a witness;
    1 = connected, 2 = disconnected)
  - `sqlserver_mirroring_log_send_queue_kb{db_name="app"} 0`
  - `sqlserver_mirroring_redo_queue_kb{db_name="app"} 0`
    (both partners report both queues, each from its own view; the values
    can differ slightly between them)
  - `sqlserver_mirroring_log_sent_bytes_total`, `_log_received_bytes_total`,
    `_redo_bytes_total`, `_transaction_delay_ms_total`, `_write_transactions_total`
  - Role, state, safety level, witness state and `_info` come from the
    **principal** only: on the mirror, `sys.database_mirroring` returns NULL
    unless the login is sysadmin (a `GRANT SELECT` on the view doesn't
    change that). The state is a property of the session, so the principal's
    view is enough to alert on; after a failover the new principal reports
    it. The mirror still exposes the queues and throughput counters.
    Alert with `sqlserver_mirroring_state != 4` and on the principal's
    `sqlserver_up`/`up` for a principal that is down.
  - While the mirrored database is not accessible (e.g. stuck `RECOVERING`
    on the principal), `sys.database_mirroring` returns NULL even for
    sysadmin, so the state series disappear instead of changing value. The
    performance counters are still reported by both partners, so also alert
    on mirrored databases without a state:
    `count by (db_name) (sqlserver_mirroring_log_send_queue_kb) unless on (db_name) count by (db_name) (sqlserver_mirroring_state)`
    (`sqlserver_database_state{state="RECOVERING"}` covers the same case).
  - Average commit delay in high-safety mode:
    `rate(sqlserver_mirroring_transaction_delay_ms_total[5m]) / rate(sqlserver_mirroring_write_transactions_total[5m])`

- Log shipping (opt-in, `--collector.log_shipping`; timestamps are UTC)
  - `sqlserver_log_shipping_primary_last_backup_timestamp_seconds{primary_server="NODE1",primary_database="app"} 1791892800`
  - `sqlserver_log_shipping_primary_backup_threshold_seconds{...} 3600`
  - `sqlserver_log_shipping_secondary_last_copied_timestamp_seconds{secondary_server="NODE2",secondary_database="app",primary_server="NODE1",primary_database="app"} 1791892860`
  - `sqlserver_log_shipping_secondary_last_restored_timestamp_seconds{...} 1791892920`
  - `sqlserver_log_shipping_secondary_last_restored_latency_seconds{...} 120` (minute precision)
  - `sqlserver_log_shipping_secondary_restore_threshold_seconds{...} 2700`
  - Restore falling behind, using the threshold configured in SQL Server:

    ```promql
    time() - sqlserver_log_shipping_secondary_last_restored_timestamp_seconds
      > sqlserver_log_shipping_secondary_restore_threshold_seconds
    ```

  - On SQL Server 2019 RTM and CU1 the log shipping jobs run but never
    update the monitor tables ([KB4537869](https://support.microsoft.com/help/4537869),
    job history shows "Failed to convert parameter value from a SqlGuid to a
    String"), so these timestamps stay frozen — the same values
    `sp_help_log_shipping_monitor` reports. Fixed in CU2.

- Top queries (opt-in, `--collector.top_queries`; no extra grants beyond
  `VIEW SERVER STATE`)
  - The top N queries (`--collector.top_queries.limit`, default 50) by total
    CPU time in the plan cache (`sys.dm_exec_query_stats`), aggregated by
    `query_hash`; `db_name` is the database the statement ran in
  - `sqlserver_top_query_executions_total{db_name="app",query_hash="0x1A2B3C4D5E6F7081"} 72600`
  - `sqlserver_top_query_cpu_seconds_total{...}`,
    `sqlserver_top_query_elapsed_seconds_total{...}`,
    `sqlserver_top_query_logical_reads_total{...}`,
    `sqlserver_top_query_physical_reads_total{...}`,
    `sqlserver_top_query_logical_writes_total{...}`
  - Average duration:
    `rate(sqlserver_top_query_elapsed_seconds_total[5m]) / rate(sqlserver_top_query_executions_total[5m])`
  - `sqlserver_top_query_info{db_name="app",query_hash="0x1A2B...",query_text="SELECT p.parcela FROM dbo.t p WHERE p.id = ?"} 1`
    only with `--collector.top_queries.include_text`: the statement text with
    string, numeric and hex literals replaced by `?`, comments removed and
    cut to `--collector.top_queries.text_length` characters. Ad hoc
    statements carry real values (customer data, IDs) in their text, so it
    is off by default; join it in queries with
    `* on (query_hash) group_left(query_text) sqlserver_top_query_info`
  - Caveats:
    - Series come and go as queries enter and leave the top N, so `rate()`
      has gaps; and the counters only cover what is still in the plan cache:
      when a plan is evicted (memory pressure, recompile, `DBCC
      FREEPROCCACHE`) a query's counters drop, which Prometheus treats as a
      counter reset
    - The plan cache is scanned on every scrape; check
      `sqlserver_scrape_collector_duration_seconds{collector="top_queries"}`
      on instances with very large caches
    - On an idle instance the top is the exporter's own DMV queries

- Index fragmentation (opt-in, `--collector.index_fragmentation`)
  - Unlike every other collector it **does not query SQL Server on
    scrape**: `sys.dm_db_index_physical_stats` (`LIMITED` mode) reads index
    pages, which takes minutes and real I/O on large databases, so it runs in
    the background every `--collector.index_fragmentation.interval` (default
    6h, first run on the first scrape, one database at a time, each run
    bounded by the interval) and scrapes serve the last result
  - `sqlserver_index_fragmentation_percent{db_name="app",schema_name="dbo",table_name="Orders",index_name="PK_Orders"} 37.5`:
    logical fragmentation of the leaf level, weighted by page count across
    partitions; `index_name=""` for heaps
  - `sqlserver_index_page_count{...} 2048` (8 KB pages, in-row data)
  - Indexes below `--collector.index_fragmentation.min_page_count` (default
    1000 pages, ~8 MB) are skipped: fragmentation of small indexes doesn't
    matter and they would dominate the series count
  - `sqlserver_index_fragmentation_last_collection_timestamp_seconds`,
    `_last_collection_duration_seconds`, `_last_collection_databases` and
    `_last_collection_success` (0 when any database failed; the error is
    logged once per collection). The scrape itself always succeeds, so
    alert on staleness and success, e.g.
    `time() - sqlserver_index_fragmentation_last_collection_timestamp_seconds > 2 * 6 * 3600`
    and `sqlserver_index_fragmentation_last_collection_databases == 0`
    (no accessible database, usually a missing `CONNECT ANY DATABASE`)
  - Example alert (large and fragmented):
    `sqlserver_index_fragmentation_percent > 30 and on (db_name, schema_name, table_name, index_name) sqlserver_index_page_count > 10000`
  - On SSD/EBS storage, logical fragmentation matters much less than page
    density (`avg_page_space_used_in_percent`), which needs the more
    expensive `SAMPLED` mode and is not collected

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

Set the job's `scrape_timeout` above the exporter's 5s query timeout (the
example uses 10s). With a global `scrape_interval` below 10s, Prometheus
defaults the timeout to that interval: a scrape that waits on SQL Server for
the full 5s (e.g. while the instance is starting up) then fails as a whole
(`up 0`, no samples) instead of returning the metrics that were collected.

A ready-to-use dashboard is included in [`dashboards/`](dashboards/) — see
[`dashboards/README.md`](dashboards/README.md) for import instructions.

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
- For encrypted connections to SQL Server (including TDS 8.0 strict
  encryption), see [Encrypted connections](#encrypted-connections-encryptstrict).
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

## Releases

Merging a PR into `main` with one of these labels tags and publishes a new
release (binaries via GoReleaser):

| Label | Bump | Example |
|-------|------|---------|
| `release:patch` (or `release`) | patch | `v0.1.0` → `v0.1.1` |
| `release:minor` | minor | `v0.1.1` → `v0.2.0` |
| `release:major` | major | `v0.2.0` → `v1.0.0` |

With several labels the largest bump wins. When no `vX.Y.Z` tag exists yet,
the first release is `v0.1.0` regardless of the label.

## License

Apache License 2.0 — see [LICENSE](LICENSE).

---

This project is a work-in-progress exporter tailored for SQL Server, aiming for parity in usability and quality with other database exporters in the Prometheus ecosystem.
