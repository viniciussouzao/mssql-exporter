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

Connection is configured via a single DSN, set with the `SQL_SERVER_DSN` environment variable (a `.env` file is loaded automatically). Any format supported by the `go-mssqldb` driver works:

- URL-style:

```
SQL_SERVER_DSN=sqlserver://sa:MyStrongP%40ss@localhost:1433?database=master&app+name=mssql-exporter
```

- ODBC-style:

```
SQL_SERVER_DSN=server=localhost;user id=sa;password=MyStrongP@ss;database=master;app name=mssql-exporter
```

Connection pool/timeouts: adjust defaults in `internal/db/options.go` (`DefaultOptions()`).

## Permissions

Some DMVs require `VIEW SERVER STATE`. Grant minimally:

```sql
GRANT VIEW SERVER STATE TO [your_login];
```

The exporter reads from (among others):

- `sys.dm_io_virtual_file_stats`
- `sys.master_files`
- `sys.dm_exec_sessions`
- `sys.dm_os_wait_stats`
- `sys.dm_os_sys_info`
- `sys.dm_os_sys_memory`
- `SERVERPROPERTY(...)`

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

- The exporter serves plain HTTP and does not implement TLS or authentication. Put it behind a reverse proxy or limit network access (e.g., bind to localhost and scrape through a sidecar).
- Avoid embedding secrets in source control. Move connection strings to environment variables/flags in production setups.

## Troubleshooting

- Cannot connect / ping fails: verify host, port, authentication mode, and that the SQL Server accepts TCP connections.
- Permission errors or empty metrics: ensure the login has `VIEW SERVER STATE` and access to the queried DMVs.
- Panic on scrape: fixed in current code by aligning metric labels and descriptors. If you see a panic, check the logs and validate label counts vs. descriptors.
- No metrics in Prometheus: confirm Prometheus can reach `:9187/metrics` and that the scrape job is correctly targeted.

## Roadmap

- CLI flags mirroring `SQL_SERVER_DSN` and other environment variables (listen address, timeouts, collector toggles), following the ergonomics of `mysqld_exporter`.
- Optional TLS and basic auth for the HTTP endpoint.
- More collectors (locks, memory clerks, tempdb, AGs, backups, etc.).

---

This project is a work-in-progress exporter tailored for SQL Server, aiming for parity in usability and quality with other database exporters in the Prometheus ecosystem.
