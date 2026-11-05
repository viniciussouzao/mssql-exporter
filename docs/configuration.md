# Configuration

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
`SQL_SERVER_DSN` environment variable (a `.env` file in the current directory
is loaded automatically when running in the foreground, not as a Windows
service) or via the config file. Any DSN format supported by the `go-mssqldb` driver
works, e.g.:

```
SQL_SERVER_DSN=sqlserver://mssql_exporter:MyStrongP%40ss@localhost:1433?database=master&app+name=mssql-exporter
```

Alternatively, use a YAML config file (handy for Windows services, where
editing environment variables is more cumbersome than pointing at a file next
to the executable). Unknown keys are an error, so a misspelled key (for
example `web.config_file`) stops the exporter instead of being ignored. See
`config.example.yml`:

```yaml
dsn: "sqlserver://mssql_exporter:MyStrongP%40ss@localhost:1433?database=master"
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
  encryption), see [Encrypted connections](tls.md).
- A `/healthz` endpoint returns 200 while the process is alive (liveness);
  use `sqlserver_up` for database reachability. It is served by the same
  listener as the metrics, so TLS and basic auth apply to it too. The Docker
  image's `HEALTHCHECK` probes `http://localhost:10060/healthz`: with TLS,
  basic auth or another listen address, override it (`healthcheck:` in
  Compose, `--health-cmd` in `docker run`) or disable it
  (`--no-healthcheck`).
- Avoid embedding secrets in source control. Keep the DSN in environment variables or a config file excluded from git.
- Release binaries and the Docker image are built with the latest patch of a
  supported Go release, and CI runs `govulncheck` on every pull request.
  To report a vulnerability, see [SECURITY.md](../SECURITY.md).
