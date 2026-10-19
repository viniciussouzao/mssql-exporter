# Grafana Dashboards

Two dashboards are included: the general one below, and one for Amazon RDS
that combines the exporter with CloudWatch (see
[Amazon RDS dashboard](#amazon-rds-dashboard-exporter--cloudwatch)).

`mssql-exporter-grafana-dashboard.json` is a ready-to-use Grafana dashboard
covering the metrics exposed by this exporter: instance overview, throughput
and compilation, memory, blocking/concurrency, CPU/schedulers, I/O, databases,
TempDB, storage volumes, Availability Groups, database mirroring, log shipping,
and exporter/collector self-health. The HA/DR rows show "No data" when the
feature isn't in use; the Log Shipping row also needs the opt-in
`log_shipping` collector (`--collector.log_shipping`).

## Requirements

- A Prometheus data source scraping this exporter (see the root
  [`README.md`](../README.md#running-with-prometheus-and-grafana) for the
  local Prometheus + Grafana stack via `docker-compose.yml`).
- The `databases`, `tempdb`, `volumes` and `schedulers` collectors enabled
  (default) — several panels rely on their metrics.
- At least one successful scrape with SQL Server reachable, so the
  `sqlserver_instance` label is populated (it's resolved once at exporter
  startup; see the note in the root README).

## Importing

1. In Grafana, go to **Dashboards → New → Import**.
2. Upload `mssql-exporter-grafana-dashboard.json` (or paste its contents into
   the "Import via panel json" box).
3. On the mapping screen, select your **Prometheus** data source for the
   `DS_PROMETHEUS` input.
4. Click **Import**.

Using the local `docker-compose.yml` stack: open http://localhost:3000
(admin/admin), add the Prometheus data source first (**Connections → Data
sources → Add data source → Prometheus**, URL `http://prometheus:9090`) if
you haven't already, then follow the steps above.

## Using the dashboard

- **Instance** (top-left variable): selects the `sqlserver_instance` label —
  populated from `label_values(sqlserver_up, sqlserver_instance)`. If it's
  empty, no target has reported `sqlserver_up` yet.
- **Database** (multi-select, defaults to all): filters per-database panels
  by `db_name`, sourced from `sqlserver_database_state`. Requires the
  `databases` collector.

## Troubleshooting

- **"No data" everywhere**: confirm the Prometheus data source URL is
  reachable from Grafana and that `sqlserver_up == 1` for the selected
  instance.
- **Instance dropdown is empty**: the exporter hasn't successfully scraped
  yet, or Prometheus isn't scraping the right target — check
  `http://localhost:9090/targets`.
- **Database-scoped panels empty but instance panels work**: the
  `databases` collector may be disabled (`--collector.databases=false` or
  `collectors.databases: false` in YAML), or the login is missing
  `VIEW ANY DEFINITION` (see the DMV reference table in the root README).

## Amazon RDS dashboard (exporter + CloudWatch)

`mssql-exporter-rds-cloudwatch-dashboard.json` is meant for SQL Server on
Amazon RDS, where there is no access to the host: CloudWatch (`AWS/RDS`)
covers the host side and the exporter covers the engine.

- **Overview:** SQL Server up, the current host behind the endpoint (it
  changes on stop/start and Multi-AZ failover), CPU, freeable memory, free
  storage, batch requests, blocked sessions and the age of the oldest last
  FULL backup.
- **Host — CloudWatch:** CPU and CPU credits (burstable classes), freeable
  memory, IOPS, latency, throughput, disk queue depth, free storage and
  network.
- **Engine vs host — correlation:** panels mixing both sources: CPU vs batch
  requests, disk writes vs SQL Server file writes, and CloudWatch
  `DatabaseConnections` vs the exporter's user connections.
- **Engine — SQL Server:** throughput, page life expectancy, buffer cache,
  top waits (idle/background waits excluded), sessions by login, blocking,
  long-running requests and open transactions.
- **Workload:** top queries and index fragmentation (opt-in collectors).
- **Backups & high availability:** time since the last backup per database
  and type, log used, and the Multi-AZ (Always On) replicas and databases.
  These show "No data" on Single-AZ instances.

Requirements and import:

1. A **CloudWatch** data source with read access to the `AWS/RDS` metrics
   (`cloudwatch:GetMetricData`, `cloudwatch:ListMetrics` and
   `ec2:DescribeRegions`), in addition to the Prometheus data source.
2. Import the JSON as above and map both inputs: `DS_PROMETHEUS` and
   `DS_CLOUDWATCH`.
3. Pick the region (`default` uses the data source's default region), the
   **RDS instance** (CloudWatch `DBInstanceIdentifier`) and the
   **sqlserver_instance** (exporter label).

Set `instance_label` in the exporter to the DB instance identifier (see the
root README): both selectors then use the same name, and the series do not
change when RDS moves the instance to another host. CloudWatch does not
publish `SwapUsage` for SQL Server, and RDS needs Enhanced Monitoring (not
used here) for OS process metrics.

## Updating

The dashboard was exported from Grafana with **Export for sharing
externally**, which keeps the data source as the `${DS_PROMETHEUS}` template
variable instead of a hardcoded UID. If you edit it in the Grafana UI and
re-export, use the same option so the JSON stays portable across Grafana
instances.
