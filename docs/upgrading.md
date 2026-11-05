# Upgrading

Changes to check when upgrading, newest first. Upgrading is replacing the
binary (or the image tag) and restarting; the configuration format is the
same across versions unless a section below says otherwise. Every release is
listed on the [releases page](https://github.com/viniciussouzao/mssql-exporter/releases).

## v0.19.0

- **Wait statistics renamed to counters.** `sqlserver_wait_time_ms` is now
  `sqlserver_wait_time_ms_total` and `sqlserver_waiting_tasks` is now
  `sqlserver_waiting_tasks_total`, both counters (they were gauges over
  values that only grow). Update queries, recording rules and dashboards
  built on the old names, and re-import the dashboards of this repository.
  `sqlserver_max_wait_time_ms` is unchanged.

## v0.18.4

- **Unknown keys in the config file are an error.** A misspelled key (for
  example `web.config_file`) used to be ignored, silently dropping TLS, basic
  auth or the listen address; now the exporter stops at startup and names
  the key. Start the new version once in the foreground with your
  `--config.file` to check it.
- `--web.telemetry-path` rejects `/healthz` and paths the HTTP router does
  not accept, instead of crashing at startup.
- `.env` is read only when running in the foreground: not by the Windows
  service nor by `service install`, which used to validate the config with
  the variables of the current directory.

## v0.18.2

Metric names are unchanged, but some values change:

- `sqlserver_memory_clerk_used_kb` adds `pages_kb` to
  `virtual_memory_committed_kb`. The page allocator holds most of the memory
  since SQL Server 2012, so values go up (on the home lab, from 107 MB to
  465 MB in total).
- `top_queries` aggregates per `query_hash` **and** `db_name`: a query that
  runs in several databases now has one series per database instead of one
  under the database of its most expensive plan.
- `sqlserver_schedulers_online` counts only `VISIBLE ONLINE` schedulers, so
  it drops where affinity or licensing keeps schedulers offline.
- The statement text of `sqlserver_top_query_info` also replaces money
  literals (`$12.50`) and `"..."` strings under `QUOTED_IDENTIFIER OFF` with
  `?`, so the `query_text` label of those queries changes.

## v0.18.0

- **Default port 9187 → 10060.** 9187 is the port the
  [Prometheus port registry](https://github.com/prometheus/prometheus/wiki/Default-port-allocations)
  assigns to `postgres_exporter`; 10060 is registered for this exporter.
  Point the scrape targets (and firewall rules) at 10060, or keep the old
  port with `--web.listen-address=:9187` (`web.listen_address` in the config
  file).

## v0.15.0

- **Windows disk queue metrics replaced.** `os_disk_read_queue_length` and
  `os_disk_write_queue_length` were gauges over cumulative counters, which
  only ever grew. They are replaced by the counters
  `os_disk_read_time_seconds_total` and `os_disk_write_time_seconds_total`,
  whose `rate()` is the average queue length (as node_exporter's
  `node_disk_*_time_seconds_total`).
- **Idle waits filtered.** `wait_stats` skips idle and background wait types
  and wait types that never happened: about 75 series instead of ~1,500.
  `--collector.wait_stats.include_idle` restores the previous behaviour
  (only `SLEEP%`, `XE%` and `BROKER_%` excluded).

## v0.14.0

- **Simultaneous scrapes share one collection.** A scrape that arrives while
  another one is collecting waits for it and gets the same metrics. No
  configuration change. Before, each scrape ran every collector again, and
  from about 5 simultaneous scrapes collectors failed on the query timeout
  while waiting for a connection.
