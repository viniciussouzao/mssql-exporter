# mssql-exporter

Prometheus exporter for Microsoft SQL Server. It collects a handful of instance-level metrics and exposes them on an HTTP endpoint for Prometheus to scrape, following a similar structure and naming style to other official database exporters (e.g. `mysqld_exporter`).

- Endpoint: `:10060/metrics`. Up to v0.17 the default port was 9187, which
  the [Prometheus port registry](https://github.com/prometheus/prometheus/wiki/Default-port-allocations)
  assigns to `postgres_exporter`. When upgrading, either point the scrape
  targets (and firewall rules) at 10060 or keep the old port with
  `--web.listen-address=:9187`
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
- Backups: last backup timestamp (UTC, correct for servers in any time zone), duration and size per database, counting only backups taken by the instance itself (restores recorded in `backupset`, e.g. on a log shipping secondary, are ignored) and skipping backups written to the `NUL` device, which are discarded (Amazon RDS takes such backups at instance creation). On Amazon RDS, automated backups show up as daily snapshot FULL backups (duration ~1s, size = database size) and log backups every 5 minutes. On Azure SQL Managed Instance, the automatic backups of user databases are reported (recorded with a NULL `server_name`); system databases have no backup history there.
- TempDB: space usage and temp table activity.
- Performance counters: batch requests, compilations, page life expectancy, buffer cache hit ratio, deadlocks, lock waits, user connections and more.
- Databases: state, recovery model, data/log sizes and log usage per database.
- Transaction log: VLF count, active log and log generated since the last log backup and checkpoint, per database (`sys.dm_db_log_stats`, SQL Server 2016 SP2 and later).
- Blocking: blocked sessions, longest current block and requests by status.
- Active queries: long-running user requests and open user transactions, aggregated (no per-query labels).
- Availability Groups: replica role, synchronization health and send/redo queues (emits nothing when Always On is disabled).
- SQL Agent jobs: enabled state, last run outcome, timestamp and duration.
- Volumes: total/available disk space per volume hosting database files, as seen by SQL Server (works for remote monitoring); reserved/used instance storage on Azure SQL Managed Instance.
- Schedulers: runnable tasks (CPU pressure signal), active workers and work queue.
- Database mirroring: role, state, safety level, witness state and send/redo queues per mirrored database (emits nothing when no database is mirrored).
- Log shipping (opt-in): last backup/copy/restore timestamps, restore latency and alert thresholds from the log shipping monitor tables.
- Top queries (opt-in): top N queries by CPU from the plan cache, per `query_hash`, with optional statement text (literals obfuscated).
- Index fragmentation (opt-in): logical fragmentation and size per index, collected in the background on its own interval (default 6h).
- Missing indexes (opt-in): the top N indexes the optimizer reported as missing (`sys.dm_db_missing_index_*`), by estimated improvement, with their columns.
- Replication (opt-in, on the Distributor): status, last activity and delivery latency of the transactional replication Log Reader and Distribution agents, and the commands still waiting in the distribution database.
- Query Store health (opt-in): actual vs desired state (catches the silent switch to `READ_ONLY` when the store fills up), storage used and forced plans that fail, per database, collected in the background (default 5m).

## Quick Start

1) Ensure you have access to a SQL Server instance and a login with permissions listed in "Permissions" below.

2) Build and run the exporter:

```bash
# Build
go build -o mssql-exporter ./

# Run (binary listens on :10060)
./mssql-exporter
```

3) Scrape with Prometheus (example snippet):

```yaml
scrape_configs:
  - job_name: 'sqlserver_exporter'
    static_configs:
      - targets: ['localhost:10060']
```

A `docker-compose.yml` for Prometheus and Grafana is included; it mounts a
local `prometheus.yml` that you create first (see
[Running With Prometheus and Grafana](#running-with-prometheus-and-grafana)).

## Configuration

Configuration comes from three sources, merged with the following precedence
(highest first): **CLI flag > environment variable > config file > default**.

Flags (mirroring the ergonomics of `mysqld_exporter`/`postgres_exporter`):

```
--web.listen-address string    Address to listen on for the web interface and metrics. (default ":10060")
--web.telemetry-path string    Path under which to expose metrics. (default "/metrics")
--collector.<name>             Enable/disable a SQL Server collector (all default to true
                               except log_shipping, top_queries, index_fragmentation,
                               query_store, missing_indexes and replication).
                               Names: wait_stats, io_file_stats, sessions, server_info,
                               memory_clerks, backups, tempdb, perf_counters, databases,
                               transaction_log, blocking, active_queries, availability_groups, agent_jobs,
                               volumes, schedulers, database_mirroring, log_shipping,
                               top_queries, index_fragmentation, query_store,
                               missing_indexes, replication.
                               Example: --collector.backups=false, --collector.log_shipping
--collector.wait_stats.include_idle
                               Also expose idle/background wait types and wait types that never
                               happened (default false).
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
--collector.query_store.interval duration
                               Interval between background Query Store collections (default 5m, min 1m).
--collector.query_store.databases strings
                               Databases to read, comma-separated (default: all accessible user databases).
--collector.missing_indexes.limit int
                               Missing index suggestions exposed (default 20, 1-600).
--collector.os-collectors      Enable OS-level collectors (CPU, memory, disk).
--config.file string           Path to an optional YAML config file.
--sqlserver.instance-label string
                               Fixed value of the sqlserver_instance label (default: @@SERVERNAME).
                               Env: SQL_SERVER_INSTANCE_LABEL. YAML: instance_label.
--sqlserver.max-connections int
                               Connections to SQL Server shared by the collectors of a scrape
                               (default 5, 1-100). Env: SQL_SERVER_MAX_CONNECTIONS. YAML: max_connections.
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
  listen_address: ":10060"
  telemetry_path: "/metrics"
collectors:
  os_collectors: false
  backups: false        # any SQL Server collector can be toggled by name
```

```bash
./mssql-exporter --config.file config.yml
```

Connection pool: every enabled collector runs at the same time on each
scrape, sharing `--sqlserver.max-connections` connections (default 5). Raise
it when many collectors are enabled and the scrape is slow while SQL Server
is not busy; lower it to limit the sessions the exporter opens. The query
timeout (5s) is fixed in `internal/db/options.go` (`DefaultOptions()`).

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
- **Google Cloud SQL:** use `encrypt=true` with the instance's server CA
  (`gcloud sql instances describe <instance> --format="value(serverCaCert.cert)"`,
  saved as `.pem`). The server certificate lists the private IP and the
  instance DNS name (`<id>.<region>.sql.goog`), so connecting by IP needs no
  `hostnameincertificate`. Cloud SQL does **not** accept `encrypt=strict`: the
  handshake fails with `EOF`, even with `trustservercertificate=true`
  (validated on Cloud SQL for SQL Server 2022):
  `sqlserver://user:pass@10.0.0.3:1433?database=master&encrypt=true&certificate=/etc/mssql-exporter/cloudsql-server-ca.pem`
- **Azure SQL Managed Instance:** the certificate is issued by a public
  Microsoft CA for `*.<dns-zone>.database.windows.net`, so `encrypt=true` and
  `encrypt=strict` work with the system CAs and the instance FQDN as the host,
  with no certificate file and no server-side change; the instance refuses
  unencrypted connections (validated on a General Purpose instance):
  `sqlserver://user:pass@<instance>.<dns-zone>.database.windows.net:1433?database=master&encrypt=strict`.
  Run the exporter inside the instance's virtual network (or a peered one):
  the gateway redirects the connection to the node's private IP, so a port
  forward to the gateway alone is not enough.

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
  `icacls config.yml /inheritance:r /grant:r "*S-1-5-18:R" "*S-1-5-32-544:F"`
  (the well-known SIDs of SYSTEM and Administrators: the group names are
  translated on non-English Windows, e.g. `Administradores`).
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
curl -s localhost:10060/metrics | grep sqlserver_up
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
  "127.0.0.1:10060"` and forward the port:
  `ssh -N -L 10060:127.0.0.1:10060 user@bastion`.
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

The `index_fragmentation` and `query_store` collectors are **disabled by
default**. They read every online user database the login can access
(`HAS_DBACCESS`), and a database where the login has no user is silently
skipped; to cover all of them without creating a user in each:

```sql
GRANT CONNECT ANY DATABASE TO [your_login];
```

`CONNECT ANY DATABASE` exists since SQL Server 2014; on older versions create
a user for the login in each database. A skipped database is not an error:
without the grant the collection succeeds with
`sqlserver_index_fragmentation_last_collection_databases 0` (or
`sqlserver_query_store_last_collection_databases 0`), so compare that
metric with the number of databases you expect (or alert on `== 0`). Besides
database access, `query_store` needs no grant beyond `VIEW SERVER STATE`
(validated on SQL Server 2025).

The `missing_indexes` collector is **disabled by default**. It needs only
`VIEW SERVER STATE`, but the schema and table names come from each
database's catalog: without access to a database (`CONNECT ANY DATABASE` or
a user there) its suggestions are still exposed, with empty `schema` and
`table` labels.

The `replication` collector is **disabled by default** and only emits on a
**Distributor** (an instance with a distribution database; elsewhere it
succeeds with no series). It reads the agent tables of each distribution
database the login can access; the `replmonitor` role is not enough (it
only grants the monitoring procedures), so grant SELECT on them:

```sql
USE distribution;
CREATE USER [your_login] FOR LOGIN [your_login];
GRANT SELECT ON dbo.MSdistribution_agents TO [your_login];
GRANT SELECT ON dbo.MSdistribution_history TO [your_login];
GRANT SELECT ON dbo.MSdistribution_status TO [your_login];
GRANT SELECT ON dbo.MSlogreader_agents TO [your_login];
GRANT SELECT ON dbo.MSlogreader_history TO [your_login];
GRANT SELECT ON dbo.MSreplservers TO [your_login];
```

On **Amazon RDS** the master user cannot grant `CONNECT ANY DATABASE`
(`Grantor does not have GRANT permission`). Create a user for the login in
each database to cover instead (it grants access to the database, not to its
data):

```sql
USE [your_database];
CREATE USER [your_login] FOR LOGIN [your_login];
```

On **Google Cloud SQL** there is no `sysadmin`. The built-in `sqlserver` user
grants server permissions on behalf of its role, `CustomerDbRootRole`, which
must be named as the grantor (without `AS` the grant fails with
`Grantor does not have GRANT permission`); run it from `master`:

```sql
USE master;
GRANT VIEW SERVER STATE TO [your_login] AS [CustomerDbRootRole];
GRANT VIEW ANY DEFINITION TO [your_login] AS [CustomerDbRootRole];
```

- `CONNECT ANY DATABASE` cannot be granted and the `##MS_...##` server roles
  do not accept new members: create a user per database, as on RDS.
- `msdb` is protected: no user can be created there for your login, so
  `msdb.dbo.sysjobs` is not readable and `agent_jobs` fails. Disable it
  (`--collector.agent_jobs=false`). Logins created through Cloud SQL (console
  or API) get `CustomerDbRootRole` and can read the jobs, but that role can
  also manage logins and create databases. Cloud SQL publishes job outcomes
  itself in Cloud Monitoring (`sqlserver/sql_agent/jobs`).
- `backups` works with no grant (`public` can read `backupset`).

On **Azure SQL Managed Instance** the SQL admin of the instance is
`sysadmin`, so every grant above works as on a self-managed instance,
including `CONNECT ANY DATABASE` and the `msdb` user for `agent_jobs`
(validated on a General Purpose instance).

Notes:

- The `backups` collector reads `msdb.dbo.backupset` and
  `msdb.dbo.backupmediafamily`, which are readable by `public` on default
  installations (and on Amazon RDS) — no extra grant is usually needed.
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
- `sys.dm_db_log_stats` (transaction_log), `sys.dm_hadr_database_replica_states`
- `msdb.dbo.backupset`, `msdb.dbo.backupmediafamily`
- `sys.database_mirroring`
- `msdb.dbo.log_shipping_monitor_primary` / `log_shipping_monitor_secondary` (log_shipping only)
- `sys.dm_exec_query_stats`, `sys.dm_exec_plan_attributes`, `sys.dm_exec_sql_text` (top_queries only)
- `sys.dm_db_index_physical_stats` and each database's `sys.indexes`/`sys.objects`/`sys.schemas` (index_fragmentation only)
- Each database's `sys.database_query_store_options` and `sys.query_store_plan` (query_store only)
- `sys.dm_db_missing_index_details`, `_groups`, `_group_stats` (missing_indexes only)
- Each distribution database's `MSdistribution_agents`/`_history`/`_status`, `MSlogreader_agents`/`_history` and `MSreplservers` (replication only)
- `SERVERPROPERTY(...)`

## Availability & Self-Metrics

The exporter starts even if SQL Server is unreachable and retries on every
scrape — alert on `sqlserver_up` instead of watching the process:

- `sqlserver_up 0|1` — whether the instance is reachable (checked per scrape).
- `sqlserver_scrape_collector_duration_seconds{collector="wait_stats"}` — duration of each collector.
- `sqlserver_scrape_collector_success{collector="wait_stats"} 0|1` — whether each collector succeeded.

When `sqlserver_up` is 0, individual collectors are skipped for that scrape.

Every metric carries a `sqlserver_instance` label. By default it is
`@@SERVERNAME`, read at startup. If SQL Server is unreachable at that moment
the label is empty at first and the exporter retries every 15 seconds; once
SQL Server answers, the label is set without a restart (the series with the
empty label stop and new ones start). The name is not read again after
that.

Set `instance_label` (or `--sqlserver.instance-label`, or
`SQL_SERVER_INSTANCE_LABEL`) to fix the label to a name of your choice. The
exporter then uses it as is and does not read `@@SERVERNAME`, so the label is
also set when SQL Server is down at startup. Recommended on **Amazon RDS**:
there `@@SERVERNAME` is a generic host name (`EC2AMAZ-...`) that says nothing
about which instance it is, and it changes when the instance is stopped and
started or fails over. Use the DB instance identifier, e.g.
`instance_label: "orders-prod"`. On **Google Cloud SQL** `@@SERVERNAME` is a
generated host name as well; use the instance name. The
`sqlserver_servername` metric still reports the current `@@SERVERNAME`.

## Exposed Metrics

Metric names are stable and follow Prometheus best practices. Selected examples:

- Server info
  - `sqlserver_hostname{hostname="MACHINE"} 1`
  - `sqlserver_servername{servername="HOST\\INSTANCE"} 1`
  - `sqlserver_edition{edition="Developer Edition (64-bit)"} 1`
  - `sqlserver_product_version{product_version="16.0.4095.4"} 1`
  - `sqlserver_engine_edition 3` (numeric code; 8 on Azure SQL Managed Instance)
  - `sqlserver_is_hadr_enabled 0|1`
  - `sqlserver_visible_cpus 8`
  - `sqlserver_total_physical_memory_kb 67108864`
  - `sqlserver_available_physical_memory_kb 43008000`
  - A property that `SERVERPROPERTY` returns as NULL omits only its own
    metric: Azure SQL Managed Instance has no `MachineName`, so it exposes no
    `sqlserver_hostname`.

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
  - Idle and background waits (lazy writer, checkpoint, Always On and
    mirroring queues, Query Store and trace flushes, `SOS_WORK_DISPATCHER`
    and others that only accumulate while SQL Server is idle) and wait types
    that never happened are skipped: about 75 wait types instead of ~1,500.
    Waits that slow queries, such as `HADR_SYNC_COMMIT`, are kept.
    `--collector.wait_stats.include_idle` exposes everything again (only
    `SLEEP%`, `XE%` and `BROKER_%` excluded, as before v0.15.0).

- OS collectors (`--collector.os-collectors`, Linux and Windows), disk I/O
  per device
  - `os_disk_reads_total`, `os_disk_writes_total`, `os_disk_read_bytes_total`,
    `os_disk_write_bytes_total`, `os_disk_io_time_ms_total`
  - `os_disk_read_time_seconds_total{device="0 C:"}` and
    `os_disk_write_time_seconds_total`: time spent on reads/writes summed
    over concurrent requests; `rate()` is the average read/write queue
    length, e.g. `rate(os_disk_write_time_seconds_total[5m])`. Same
    semantics as node_exporter's `node_disk_*_time_seconds_total`.
  - `os_disk_queue_length`: requests in progress at the time of the scrape.
  - Before v0.15.0 Windows exposed `os_disk_read_queue_length` and
    `os_disk_write_queue_length` as gauges over cumulative counters, which
    only ever grew; they are replaced by the two counters above.

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
  - `sqlserver_database_log_reuse_wait{db_name="AppDB",reason="AVAILABILITY_REPLICA"} 1`:
    why the transaction log cannot be truncated (`log_reuse_wait_desc` in
    `sys.databases`; `NOTHING` when it can). `LOG_BACKUP` is the normal value
    of a FULL recovery database between log backups. SQL Server refreshes it
    when it tries to truncate the log (checkpoint or log backup), so it can
    lag a few minutes; alert with `for:` and together with the log usage, e.g.
    `sqlserver_database_log_reuse_wait{reason!~"NOTHING|LOG_BACKUP|CHECKPOINT"} == 1
    and on (instance, db_name) sqlserver_database_log_used_percent > 60`.
    `AVAILABILITY_REPLICA` points to a lagging AG secondary (see the
    `sqlserver_hadr_*` queues), `REPLICATION` to CDC or transactional
    replication, `ACTIVE_TRANSACTION` to a long-open transaction
  - `sqlserver_database_data_size_bytes{db_name="master"} 4915200`
  - `sqlserver_database_log_size_bytes{db_name="master"} 2097152`
  - `sqlserver_database_log_used_percent{db_name="master"} 71`: the `Percent
    Log Used` performance counter, reported only for databases listed in
    `sys.databases` (the counter also has instances for hidden databases such
    as `mssqlsystemresource`). On Amazon RDS the counter has no instance for
    `master`, `model`, `msdb` and `rdsadmin`, so they have no series there;
    user databases are reported normally. On Azure SQL Managed Instance the
    counter instance of a user database is its `physical_database_name` (a
    GUID); the series still carries the database name

- Transaction log (`sys.dm_db_log_stats`, SQL Server 2016 SP2 and later; on
  older versions disable it with `--collector.transaction_log=false`). Every
  `ONLINE` database the login can access; local secondary replicas of an
  Availability Group are skipped (their log follows the primary's)
  - `sqlserver_database_log_vlf_count{db_name="AppDB"} 12` and
    `sqlserver_database_log_active_vlf_count{db_name="AppDB"} 8`: thousands
    of VLFs (many small autogrowths) slow down recovery, restores and redo
  - `sqlserver_database_log_active_bytes{db_name="AppDB"} 91561984`
  - `sqlserver_database_log_since_last_log_backup_bytes{db_name="AppDB"} 91533312`:
    log generated since the last log backup; absent for databases without a
    log backup chain (SIMPLE recovery, or FULL before the first full backup).
    A value that only grows means log backups stopped:
    `deriv(sqlserver_database_log_since_last_log_backup_bytes[1h]) > 0 and sqlserver_database_log_since_last_log_backup_bytes > 10e9`
  - `sqlserver_database_log_since_last_checkpoint_bytes{db_name="AppDB"} 3293184`
  - Works on Amazon RDS for the system databases too, where the `Percent Log
    Used` counter has no instance for them

- Blocking
  - `sqlserver_blocked_sessions 0` (from `sys.dm_os_waiting_tasks`; a parallel
    request waiting on its own session is not counted)
  - `sqlserver_blocked_max_wait_ms 0`
  - `sqlserver_active_requests{status="running"} 3` (on an AG replica,
    `sys.dm_exec_requests` intermittently fails while an availability
    database is momentarily not accessible: error 976 on a secondary, 983
    while the replica is still resolving its role right after SQL Server
    starts; the collector retries once before reporting the scrape as failed)

- Active queries (user sessions only; the exporter's own request is excluded,
  and so is the `TdService` telemetry session of Azure SQL Managed Instance,
  which keeps a request open from instance start)
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

- Query Store health (opt-in, `--collector.query_store`)
  - Query Store is per database and its catalog views get slow on large
    stores, so like index_fragmentation it **collects in the background**
    every `--collector.query_store.interval` (default 5m, first run on the
    first scrape) and scrapes serve the last result. Requires SQL Server
    2016+ (the views don't exist before; every database fails)
  - `sqlserver_query_store_actual_state{db_name="AppDB",state="READ_ONLY"} 1`
    and `sqlserver_query_store_desired_state{db_name="AppDB",state="READ_WRITE"} 1`
    (`OFF`, `READ_ONLY`, `READ_WRITE`, `ERROR`). Query Store switches itself
    to `READ_ONLY` when it fills up and silently stops capturing; alert when
    the actual state differs from the configured one:
    `sqlserver_query_store_desired_state == 1 unless on (instance, db_name, state) sqlserver_query_store_actual_state == 1`
  - `sqlserver_query_store_readonly_reason{db_name="AppDB"} 65536`: bitmask
    of why it is `READ_ONLY` (0 when it isn't). 65536 = max storage size
    reached, 131072 = statement limit, 262144 = in-memory limit, 524288 =
    database disk size limit; 1/2/4 = database read-only, single user,
    emergency; 8 = secondary replica, which is expected on a readable AG
    secondary (exclude it from the state alert there). To alert only on
    capture stopped by a limit: `sqlserver_query_store_readonly_reason >= 65536`
  - `sqlserver_query_store_storage_used_bytes` and
    `sqlserver_query_store_storage_max_bytes` (`MAX_STORAGE_SIZE_MB`);
    alert before it fills up, e.g. `used / max > 0.9`
  - `sqlserver_query_store_forced_plans{db_name="AppDB"} 3` and
    `sqlserver_query_store_forced_plans_failing{db_name="AppDB"} 1`: forced
    plans whose forcing has failed (`force_failure_count > 0`, e.g. the
    forced plan uses an index that was dropped), so the query runs with
    another plan while everyone believes it is pinned. See
    `last_force_failure_reason_desc` in `sys.query_store_plan`
  - `sqlserver_query_store_last_collection_timestamp_seconds`,
    `_last_collection_duration_seconds`, `_last_collection_databases` and
    `_last_collection_success`, as in index_fragmentation
  - No per-query series: the Query Store UI is the tool for analysis (see
    `ROADMAP.md`)

- Missing indexes (opt-in, `--collector.missing_indexes`): the top
  `--collector.missing_indexes.limit` (default 20) suggestions of the query
  optimizer by estimated improvement, labeled by `db_name`, `schema`,
  `table` and `index_handle`
  - `sqlserver_missing_index_improvement_measure{db_name="AppDB",schema="dbo",table="Orders",index_handle="3"} 93.4`:
    `avg_total_user_cost * avg_user_impact / 100 * (user_seeks + user_scans)`,
    the usual ranking; compare suggestions with each other, not with a fixed
    threshold
  - `sqlserver_missing_index_user_seeks_total`, `_user_scans_total`,
    `sqlserver_missing_index_avg_user_impact_percent` (estimated cost
    reduction) and `sqlserver_missing_index_avg_total_user_cost`
  - `sqlserver_missing_index_info{...,equality_columns="[CustomerID]",inequality_columns="",included_columns=""} 1`:
    the suggested key and included columns
  - The suggestions are reset when SQL Server restarts, dropped for a table
    when an index is created on it or its schema changes, and capped at 600
    by SQL Server, so series come and go. Review them before creating
    indexes: they overlap and ignore the cost of maintaining the index

- Replication (opt-in, `--collector.replication`; run it against the
  **Distributor**; transactional replication; merge and snapshot agents
  are not covered). Labels: `distribution_db`, `agent_type` (`logreader` or
  `distribution`), `agent_name`, `publisher`, `publisher_db`, `publication`,
  `subscriber`, `subscriber_db` (the last three are empty for the Log
  Reader). Virtual subscriptions of `immediate_sync` publications are skipped
  - `sqlserver_replication_agent_run_status{agent_type="distribution",...} 3`:
    status of the last history entry: 1 started, 2 succeeded, 3 in
    progress, 4 idle, 5 retrying, 6 failed. A stopped agent keeps its last
    status (2 after a clean stop), so also alert on the two metrics below
  - `sqlserver_replication_agent_last_history_timestamp_seconds{...}` (UTC)
    and `sqlserver_replication_agent_delivery_latency_seconds{...}` from the
    same entry
  - `sqlserver_replication_pending_commands{agent_type="distribution",...} 300`:
    commands in the distribution database not yet delivered to the
    subscription (`MSdistribution_status`). Alert when they keep growing:
    `sqlserver_replication_pending_commands > 10000 and deriv(sqlserver_replication_pending_commands[15m]) > 0`.
    On busy distributors that view can take seconds; keep an eye on
    `sqlserver_scrape_collector_duration_seconds{collector="replication"}`

- Volumes
  - `sqlserver_volume_total_bytes{volume_mount_point="C:\\"} 1962694664192`
  - `sqlserver_volume_available_bytes{volume_mount_point="C:\\"} 1846100799488`
  - SQL Server on Linux (Google Cloud SQL, the Docker images) reports no
    mount point, so the label is the directory of the database files, e.g.
    `volume_mount_point="/var/opt/mssql/data/"`. Directories on the same
    file system show up as separate series with the same values.
  - Only databases that are `ONLINE` or `RESTORING` (AG and log shipping
    secondaries) count: a database changing state could hold the query
    until the scrape timeout.
  - Azure SQL Managed Instance (`sqlserver_engine_edition 8`): no volume
    series. There `dm_os_volume_stats` reports in `total_bytes` the size of
    the files on the volume, not its capacity (smaller than
    `available_bytes`), so the collector reports the instance storage from
    `sys.server_resource_stats` instead (latest 15-second sample, published a
    few minutes late):
    - `sqlserver_instance_storage_reserved_bytes 68719476736` (the storage
      size of the instance)
    - `sqlserver_instance_storage_used_bytes 352321536`

- Schedulers
  - `sqlserver_schedulers_runnable_tasks 1`
  - `sqlserver_schedulers_current_tasks 53`
  - `sqlserver_schedulers_active_workers 49`
  - `sqlserver_schedulers_work_queue 0`
  - `sqlserver_schedulers_online 8`

## Running With Prometheus and Grafana

This repository includes `docker-compose.yml` with Prometheus and Grafana
services. It mounts `./prometheus.yml`, which is not versioned (it is in
`.gitignore`, so each environment keeps its own targets). Create it next to
`docker-compose.yml`:

```yaml
global:
  scrape_interval: 5s

scrape_configs:
  - job_name: 'sqlserver_exporter'
    scrape_interval: 15s
    # Must exceed the exporter's 5s query timeout, or a slow scrape (e.g. SQL
    # Server starting up) fails as a whole instead of returning partial data.
    scrape_timeout: 10s
    static_configs:
      # host.docker.internal reaches the exporter running on the Docker host;
      # switch to 'mssql-exporter:10060' if you enable the exporter service in docker-compose.yml
      - targets: ['host.docker.internal:10060']
```

Steps:

1) Start the exporter locally on `:10060`.
2) `docker compose up -d` (or `make compose-up`) to bring up Prometheus and Grafana.
3) Prometheus UI: http://localhost:9090
4) Grafana UI: http://localhost:3000 (admin/admin by default; password set in `docker-compose.yml`).

Add Prometheus (http://prometheus:9090 inside Compose network or http://localhost:9090 from your browser) as a Grafana data source and build dashboards from the metric names above.

Set the job's `scrape_timeout` above the exporter's 5s query timeout (the
example uses 10s). With a global `scrape_interval` below 10s, Prometheus
defaults the timeout to that interval: a scrape that waits on SQL Server for
the full 5s (e.g. while the instance is starting up) then fails as a whole
(`up 0`, no samples) instead of returning the metrics that were collected.
The exporter itself always answers within the query timeout plus 0.5s, even
when the driver hangs on a connection the server stopped answering (seen
during a Cloud SQL failover): a ping that does not return reports
`sqlserver_up 0`, and a collector that does not return reports
`sqlserver_scrape_collector_success 0`.

Simultaneous scrapes (two Prometheus servers in HA, a dashboard or `curl`
next to Prometheus) share one collection: a scrape that arrives while
another one is collecting waits for it and gets the same metrics, so SQL
Server is queried once and the scrapes do not compete for the connection
pool. Before v0.14.0 each scrape ran every collector again, and from about
5 simultaneous scrapes collectors failed on the query timeout while waiting
for a connection.

Ready-to-use dashboards are included in [`dashboards/`](dashboards/): a
general one, one for **Amazon RDS** that shows the exporter's engine
metrics next to the host metrics from CloudWatch, one for **Google Cloud
SQL** that does the same with Cloud Monitoring, and one for **Azure SQL
Managed Instance** with Azure Monitor. See
[`dashboards/README.md`](dashboards/README.md) for import instructions.

### Alerting rules

[`contrib/prometheus/alerts.yml`](contrib/prometheus/alerts.yml) (also in
the release archives) has example alerting rules built from the
expressions of [Exposed Metrics](#exposed-metrics): exporter and SQL Server
reachability, failing collectors, databases in `SUSPECT`/`RECOVERY_PENDING`,
a log that cannot be truncated or whose log backups stopped, stale full and
log backups, many VLFs, full volumes, blocking, long requests and
transactions, CPU pressure, Availability Groups, mirroring, log shipping,
failed Agent jobs, replication agents (failed, stale history, growing
queue, latency) and the Query Store and background collections. Load it
with `rule_files:` and tune thresholds and `for:` windows to your workload;
rules over opt-in collectors never fire while they are disabled. The
`MSSQLExporterDown` rule selects `job="sqlserver_exporter"`, the job name of
the example above.

```yaml
rule_files:
  - alerts.yml
```

`make test-alerts` validates the file and runs its unit tests
([`alerts_test.yml`](contrib/prometheus/alerts_test.yml)) with `promtool`
from the Prometheus image; CI runs it on every pull request.

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
- Release binaries and the Docker image are built with the latest patch of a
  supported Go release, and CI runs `govulncheck` on every pull request.
  To report a vulnerability, see [SECURITY.md](SECURITY.md).

## Troubleshooting

- Cannot connect / ping fails: verify host, port, authentication mode, and that the SQL Server accepts TCP connections.
- Permission errors or empty metrics: ensure the login has `VIEW SERVER STATE` and access to the queried DMVs.
- Panic on scrape: fixed in current code by aligning metric labels and descriptors. If you see a panic, check the logs and validate label counts vs. descriptors.
- No metrics in Prometheus: confirm Prometheus can reach `:10060/metrics` and that the scrape job is correctly targeted.

## Roadmap

See [ROADMAP.md](ROADMAP.md) for the phased plan and the next candidates.

## Development

```bash
make check             # gofmt, go vet and unit tests (go-sqlmock, no SQL Server needed)
make test-integration  # integration tests against a disposable SQL Server container
make test-alerts       # promtool check + unit tests of contrib/prometheus/alerts.yml (Docker)
make scan-secrets      # gitleaks over the whole git history (Docker)
make vulncheck         # govulncheck: known vulnerabilities reachable from the code
```

`make test-integration` starts `mcr.microsoft.com/mssql/server:2025-latest`
(`IT_IMAGE=...:2022-latest` for another version) with a random `sa`
password on `127.0.0.1:14339`, runs `go test -tags integration ./integration/...`
and stops the container (it is removed on stop). The tests create a database,
a backup and a monitoring login with exactly the grants of
[Permissions](#permissions), enable **every** collector as that login and
require each one to succeed, so a missing grant in the documentation fails
the run. To use your own disposable instance, set
`MSSQL_EXPORTER_TEST_DSN` to a sysadmin `sqlserver://` DSN and run the
`go test` command above. CI runs them on every pull request against SQL
Server 2022 and 2025.

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
