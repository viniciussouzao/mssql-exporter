# mssql-exporter

[![CI](https://github.com/viniciussouzao/mssql-exporter/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/viniciussouzao/mssql-exporter/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/viniciussouzao/mssql-exporter)](https://github.com/viniciussouzao/mssql-exporter/releases/latest)
[![License](https://img.shields.io/github/license/viniciussouzao/mssql-exporter)](LICENSE)
[![Go version](https://img.shields.io/github/go-mod/go-version/viniciussouzao/mssql-exporter)](go.mod)

A Prometheus exporter for Microsoft SQL Server, for DBAs and SREs who run SQL
Server on Windows, on Linux or as a managed cloud service and want instance
health, HA/DR state and workload signals in Prometheus and Grafana.

- **23 collectors** (17 enabled by default, 6 opt-in): sessions, waits, I/O,
  memory, backups, transaction log, blocking, Availability Groups, database
  mirroring, log shipping, transactional replication, Query Store, SQL Agent
  jobs and more
- **Validated** on SQL Server on Windows (Amazon EC2), Amazon RDS, Google
  Cloud SQL and Azure SQL Managed Instance, including failovers and a rolling
  upgrade 2019 → 2022 → 2025 of an Availability Group, database mirroring and
  log shipping
- **Remote collection**: it connects over TDS like any client, so nothing is
  installed on the database host; runs as a native Windows service or with the
  hardened systemd unit shipped in the release archives
- **Grafana dashboards** for each platform: general, Amazon RDS + CloudWatch,
  Cloud SQL + Cloud Monitoring, Managed Instance + Azure Monitor
- **Alerting rules** with `promtool` unit tests that run in CI
- **Least-privilege login**: the CI integration tests run every collector with
  exactly the [documented grants](docs/permissions.md)

## Why another SQL Server exporter?

Good options exist: the `mssql` collector of
[`windows_exporter`](https://github.com/prometheus-community/windows_exporter)
reads SQL Server performance counters on the Windows host, and generic SQL
exporters such as [`sql_exporter`](https://github.com/burningalchemist/sql_exporter)
run the queries you write. This one focuses on what they leave to you:
collectors for HA/DR (Availability Groups, mirroring, log shipping,
replication), the differences of managed platforms already handled (RDS,
Cloud SQL, Managed Instance), remote collection without an agent on the
database host, and dashboards and alerts that come with it.

## Supported platforms

| Platform | Validated | TLS | Notes |
|---|---|---|---|
| SQL Server on Windows (Amazon EC2) | 2019 CU32, 2022, 2025 | `encrypt=strict` (2022+) with a certificate configured on the instance | Availability Group, mirroring, log shipping, rolling upgrade |
| SQL Server on Linux / Docker | 2022 and 2025 (CI integration tests), Express | — | — |
| Amazon RDS for SQL Server | 2025 Standard, Single-AZ and Multi-AZ | `encrypt=true` or `encrypt=strict` with the AWS CA bundle | no `CONNECT ANY DATABASE`; set `instance_label` |
| Google Cloud SQL for SQL Server | 2022 Standard (Linux), zonal and regional HA | `encrypt=true` only (`strict` fails) | grants need `AS [CustomerDbRootRole]`; `agent_jobs` unavailable (msdb is closed) |
| Azure SQL Managed Instance | General Purpose (Gen5) | `encrypt=true` or `encrypt=strict` with the system CAs | run inside the virtual network (the gateway redirects to the node) |
| Azure SQL Database | not supported yet | — | see [ROADMAP.md](ROADMAP.md) |

Validated on SQL Server 2019, 2022 and 2025. 2016 SP2 and 2017 should work
(same DMVs) but are untested. Earlier versions are not supported:
`transaction_log`, enabled by default, needs 2016 SP2 (`sys.dm_db_log_stats`),
as do the opt-in `replication` (2016 SP2) and `query_store` (2016). Details
per platform: [Permissions](docs/permissions.md) and
[Encrypted connections](docs/tls.md).

<!-- screenshot: dashboards/img/overview.png -->

## Quick Start

1. Create the monitoring login (on Amazon RDS, Google Cloud SQL and for the
   opt-in collectors, see [Permissions](docs/permissions.md)):

   ```sql
   CREATE LOGIN [mssql_exporter] WITH PASSWORD = '<strong password>', CHECK_POLICY = ON;
   GRANT VIEW SERVER STATE TO [mssql_exporter];
   GRANT VIEW ANY DEFINITION TO [mssql_exporter];
   USE msdb;
   CREATE USER [mssql_exporter] FOR LOGIN [mssql_exporter];
   GRANT SELECT ON dbo.sysjobs TO [mssql_exporter];
   GRANT SELECT ON dbo.sysjobhistory TO [mssql_exporter];
   ```

2. Download the latest release and check it. On Linux (`ARCH=arm64` on ARM):

   ```bash
   VERSION=$(curl -fsSL https://api.github.com/repos/viniciussouzao/mssql-exporter/releases/latest | sed -n 's/.*"tag_name": *"v\([^"]*\)".*/\1/p')
   ARCH=amd64
   URL=https://github.com/viniciussouzao/mssql-exporter/releases/download/v${VERSION}
   curl -fsSLO ${URL}/mssql-exporter_${VERSION}_linux_${ARCH}.tar.gz
   curl -fsSLO ${URL}/checksums.txt
   sha256sum --check --ignore-missing checksums.txt
   tar xzf mssql-exporter_${VERSION}_linux_${ARCH}.tar.gz mssql-exporter
   ```

   On Windows, download `mssql-exporter_<version>_windows_amd64.zip` from the
   [releases page](https://github.com/viniciussouzao/mssql-exporter/releases/latest)
   and compare its SHA-256 (`Get-FileHash`) with `checksums.txt`. macOS
   archives are there too.

   Or use the container image (linux/amd64 and linux/arm64) instead of the
   archive: put the `SQL_SERVER_DSN=...` line of step 3 in
   `mssql-exporter.env`, run the container and check it with the `curl` of
   step 4:

   ```bash
   docker run -d --name mssql-exporter -p 10060:10060 --env-file mssql-exporter.env \
     ghcr.io/viniciussouzao/mssql-exporter:latest
   ```

   See [Running in Docker](docs/install.md#running-in-docker) for pinning a version and
   using a config file.

3. Point it at SQL Server (URL-encode special characters in the password,
   e.g. `@` as `%40`):

   ```bash
   export SQL_SERVER_DSN='sqlserver://mssql_exporter:<password>@sqlhost:1433?database=master&app+name=mssql-exporter'
   ```

4. Run it and check that it reaches SQL Server (`sqlserver_up 1`):

   ```bash
   ./mssql-exporter
   curl -s localhost:10060/metrics | grep sqlserver_up
   ```

   The default port is 10060, registered in the
   [Prometheus port registry](https://github.com/prometheus/prometheus/wiki/Default-port-allocations)
   (it was 9187 up to v0.17, see [Upgrading](docs/upgrading.md)).

5. Scrape it with Prometheus, with a `scrape_timeout` above the exporter's 5s
   query timeout:

   ```yaml
   scrape_configs:
     - job_name: 'sqlserver_exporter'
       scrape_timeout: 10s
       static_configs:
         - targets: ['exporter-host:10060']
   ```

6. Keep it running as a [Windows service](docs/install.md#running-as-a-windows-service) or a
   [Linux service (systemd)](docs/install.md#running-as-a-linux-service-systemd).

A `docker-compose.yml` for Prometheus and Grafana is included; it mounts a
local `prometheus.yml` that you create first (see
[Running With Prometheus and Grafana](docs/install.md#running-with-prometheus-and-grafana)).

## Documentation

- [Installation](docs/install.md): Windows service, systemd, Docker, and a
  local Prometheus + Grafana stack
- [Configuration](docs/configuration.md): flags, environment variables,
  config file and precedence, connection pool, TLS and basic auth on the
  metrics endpoint
- [Permissions](docs/permissions.md): grants per collector and per platform
- [Encrypted connections](docs/tls.md): `encrypt=true` and `encrypt=strict`,
  certificates per platform
- [Metrics](docs/metrics.md): collectors, self-metrics and every exposed
  metric
- [Alerting rules](docs/alerting.md) and [dashboards](dashboards/README.md)
- [Upgrading](docs/upgrading.md): changes to check per version
- [Validation](docs/validation.md): where and how the exporter was tested

## Troubleshooting

- Cannot connect / ping fails: verify host, port, authentication mode, and that the SQL Server accepts TCP connections.
- Permission errors or empty metrics: ensure the login has `VIEW SERVER STATE` and access to the queried DMVs.
- No metrics in Prometheus: confirm Prometheus can reach `:10060/metrics` and that the scrape job is correctly targeted.

## Roadmap

See [ROADMAP.md](ROADMAP.md) for the phased plan and the next candidates.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for the development setup, the tests,
the pull request checklist, releases and how to add a collector. To report a
vulnerability, see [SECURITY.md](SECURITY.md).

## License

Apache License 2.0 — see [LICENSE](LICENSE).
