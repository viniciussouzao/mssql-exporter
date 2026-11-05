# Validation

Besides the unit tests and the CI integration tests (SQL Server 2022 and
2025 in containers, every collector with the [documented grants](permissions.md)),
the exporter was validated by hand on each platform it supports, under load
from the load generator in `hack/loadgen`. Each environment was destroyed at
the end of its phase; a Windows notebook and a Raspberry Pi stay as a
permanent lab. This page summarises that work. The full log, with every
finding (HML-001 to HML-059), is in
[validation/BACKLOG-HOMOLOG.md](validation/BACKLOG-HOMOLOG.md), in
Portuguese.

## Environments

| Platform | SQL Server | Topology | When |
|---|---|---|---|
| Amazon EC2, Windows Server | 2019 CU32 → 2022 → 2025 RTM | 2-replica Always On AG, database mirroring, log shipping | 2026-10-08 to 2026-10-15 |
| Amazon RDS for SQL Server | 2025 Standard (`17.0.5005.3`, RTM-CU9), `db.t3.xlarge` | Single-AZ, converted to Multi-AZ | closed 2026-10-19 |
| Google Cloud SQL for SQL Server | 2022 Standard (`16.0.4262.2`) on Linux, private IP only | zonal, converted to regional HA | 2026-10-20 to 2026-10-26 |
| Azure SQL Managed Instance | General Purpose Gen5, 4 vCores, private endpoint | managed HA | closed 2026-10-27 |
| Azure SQL Database | serverless (reconnaissance only) | — | 2026-10-27 |
| Home lab | 2025 Developer on Windows 11 (pt-BR), Prometheus and Grafana on a Raspberry Pi 5 | standalone, exporter as a Windows service | since 2026-10-27 |
| Docker | 2022, 2025, 2025 Express | standalone | CI on every pull request; Express by hand |

On RDS, Cloud SQL and Managed Instance the exporter ran on a Linux bastion
with the systemd unit, and the servers were set to UTC-3.

## Scenarios

- **Availability Group** (EC2): failover, suspend and resume, send/redo
  queues and secondary lag (HML-007, HML-008).
- **Database mirroring and log shipping** (EC2): mirroring failover, partner
  down, suspend; log shipping backup, copy and restore times (HML-011).
- **Rolling upgrade** 2019 → 2022 → 2025 of both instances with the AG,
  mirroring and log shipping in place, with no query change (HML-017).
- **Failures** (EC2): database offline, SQL Server stopped, reboot, killed
  process, differential backup (HML-009).
- **Managed failovers**: RDS Multi-AZ conversion and failover, about 45 s of
  `sqlserver_up 0` and back on its own (HML-030); Cloud SQL conversion to
  regional (about 5 min 45 s down) and two failovers under load (about 77 s
  and 60 s, HML-036); Managed Instance failover under load, about 80 s of
  `sqlserver_up 0` and no scrape lost (HML-051).
- **A server that stops answering without closing the connection** (seen
  during the Cloud SQL failover): the exporter now answers within the query
  timeout plus 0.5 s (HML-037, v0.12.1).
- **Simultaneous scrapes**: two Prometheus servers or a dashboard next to
  Prometheus; from about 5 at once the pool saturated, fixed by sharing one
  collection (HML-034, v0.14.0).
- **Time zone**: servers in UTC-3, with backup and Agent job timestamps
  exported in UTC (HML-010).
- **TLS**: `encrypt=strict` on EC2 (SQL Server 2025) with a certificate on
  the instance (HML-016); `encrypt=true` and `encrypt=strict` on RDS with the
  AWS CA bundle (HML-027) and on Managed Instance with the public Microsoft
  CA (HML-050); `encrypt=true` only on Cloud SQL (HML-033).
- **Least privilege** per platform: msdb grants (HML-002), `CONNECT ANY
  DATABASE` not grantable on RDS (HML-024), `AS [CustomerDbRootRole]` and a
  closed msdb on Cloud SQL (HML-032), boxed-SQL Server grants on Managed
  Instance (HML-039).
- **Running as a service**: Windows service install, start, stop, reboot,
  restart after a kill, uninstall and Event Log (HML-006); systemd unit with
  sandboxing on Ubuntu (HML-025).

## Findings that changed the exporter

- Duplicate series in `volumes` made the whole `/metrics` response fail with
  HTTP 500 (HML-004, v0.1.1).
- Backup and Agent job timestamps were the server's local time read as UTC
  (HML-010, v0.3.1).
- `volumes` ran into the query timeout while a database changed state
  (HML-019, v0.13.4).
- On RDS the server name is a generic host name that changes on stop/start
  and failover: `instance_label` (HML-029, v0.10.0).
- Managed Instance differences: NULL server properties, instance storage
  instead of volumes, log counters by physical name, automatic backups with a
  NULL `server_name` and the `TdService` telemetry session (HML-042 to
  HML-046, v0.12.2 to v0.13.2).
- Azure SQL Database: only 10 of 19 collectors worked with a contained user,
  so it is not supported yet; the plan is in [ROADMAP.md](../ROADMAP.md)
  (HML-053).

## Not covered

RDS Enterprise, Express and Web editions, read replicas and
`rds.force_ssl = 1`; Cloud SQL Enterprise and Web, read replicas and the
Cloud SQL Auth Proxy; Managed Instance Business Critical, failover groups and
Managed Instance link; database mirroring with a witness; SQL Server 2016
and 2017. The integration tests cover a standalone SQL Server on Linux only.
