# Homologation Backlog

English translation of [BACKLOG-HOMOLOG.md](BACKLOG-HOMOLOG.md), the
original in Portuguese. A summary is in [Validation](../validation.md).

Findings from the homologation of mssql-exporter in environments simulating
production (AWS EC2 Windows/Linux and RDS for SQL Server) and, from
2026-10-20, on other clouds (Google Cloud SQL and Azure).

**Status:** `open` · `in progress` · `resolved` · `deferred` (became a ROADMAP item) · `won't fix`
**Severity:** `high` (blocks use) · `medium` (hinders operation) · `low` (cosmetic/noise)

| ID | Title | Environment | Severity | Status |
|----|-------|-------------|----------|--------|
| HML-001 | `ERROR` log when `.env` does not exist | EC2 Windows | low | resolved (v0.1.1) |
| HML-002 | `agent_jobs` fails without permission on msdb (point of attention) | EC2 Windows | medium | resolved on EC2 (repeat on RDS) |
| HML-003 | Disk I/O collection on Windows fails with `Invalid class` | EC2 Windows | medium | resolved (v0.1.1) |
| HML-004 | Intermittent gaps: `/metrics` returns HTTP 500 because of a duplicate series in `volumes` | EC2 Windows | high | resolved (v0.1.1) |
| HML-005 | EC2 C: drive almost full during load on `StressDB` (environment) | EC2 Windows | high | resolved |
| HML-006 | Exporter does not run as a native Windows service (point of attention) | EC2 Windows | medium | resolved (v0.2.0; `uninstall` validated on 2026-10-15) |
| HML-007 | Validation of the `availability_groups` collector on a real AG (validation) | EC2 Windows (AG 01+02) | — | resolved (v0.3.0) |
| HML-008 | Send queue disappears from the chart with the replica suspended; `is_suspended` and `secondary_lag_seconds` missing | EC2 Windows (AG 01+02) | medium | resolved (v0.3.0) |
| HML-009 | Failure scenarios: database offline, SQL stopped, reboot, kill, differential backup (validation) | EC2 Windows (AG 01+02) | — | resolved |
| HML-010 | Backup and Agent job timestamps treat SQL Server local time as UTC | EC2 Windows (UTC-3) | medium | resolved (v0.3.1) |
| HML-011 | No coverage of database mirroring and log shipping | EC2 Windows (01+02) | medium | resolved (v0.4.0) |
| HML-012 | `backups` reports on the secondary the backups restored from the primary | EC2 Windows (01+02) | medium | resolved (v0.4.0) |
| HML-013 | Whole scrape fails (`up 0`) when SQL is slow: the example's `scrape_timeout` = the exporter's timeout | EC2 Windows (01+02) | low | resolved (v0.4.0) |
| HML-014 | SQL Server 2019 RTM → CU32 + KB5128494 update (validation) | EC2 Windows (01+02) | — | resolved |
| HML-015 | `blocking` fails sporadically on the AG secondary replica (error 976) | EC2 Windows (02) | low | resolved (v0.4.0) |
| HML-016 | `denisenkom/go-mssqldb` driver archived: migrate to `microsoft/go-mssqldb` | EC2 Windows (01+02) | medium | resolved (v0.5.0) |
| HML-017 | Rolling upgrade SQL Server 2019 → 2022 (02) → 2025 (01, 02) with AG, mirroring and log shipping (validation) | EC2 Windows (01+02) | — | resolved |
| HML-018 | Mirroring state disappears (no series) when the mirrored database is not accessible | EC2 Windows (01+02) | medium | resolved (documentation, v0.5.0) |
| HML-019 | `volumes` exceeds the query timeout when a database is changing state | EC2 Windows (01+02) | low | resolved (v0.13.4) |
| HML-020 | `blocking` fails with error 983 right after SQL Server starts (AG replica in `RESOLVING`) | EC2 Windows (02) | low | resolved (v0.5.0) |
| HML-021 | `sqlserver_sessions_by_login` with load from several logins (validation) | EC2 Windows (01) | — | resolved |
| HML-022 | `active_queries` collector: long requests and open transactions (validation) | EC2 Windows (01) | — | resolved |
| HML-023 | `top_queries` collector with obfuscated text (validation) | EC2 Windows (01) | — | resolved |
| HML-024 | `index_fragmentation` collector in the background and `CONNECT ANY DATABASE` (validation) | EC2 Windows (01) | — | resolved |
| HML-025 | Exporter as a Linux service (systemd) on the RDS phase bastion (validation) | Linux (bastion) | — | resolved (v0.9.1) |
| HML-027 | RDS Single-AZ baseline: permissions, collectors, load and TLS (validation) | RDS (SQL Server 2025 SE) | — | resolved |
| HML-028 | Backups on RDS: snapshots and the log every 5 min appear; creation backups written to `NUL` count as FULL | RDS (SQL Server 2025 SE) | medium | resolved (v0.9.3) |
| HML-029 | `sqlserver_instance` on RDS: generic host name, changes on every stop/start and is empty if the exporter starts before the database | RDS (SQL Server 2025 SE) | medium | resolved (v0.10.0) |
| HML-030 | RDS Single-AZ → Multi-AZ conversion and failover: Always On appears with no collection failures (validation) | RDS (SQL Server 2025 SE) | — | resolved |
| HML-031 | Demonstration dashboard for RDS: exporter + CloudWatch (validation) | RDS (SQL Server 2025 SE) | — | resolved |
| HML-026 | `database_log_used_percent` with hidden databases and, on RDS, without the system databases | RDS (SQL Server 2025 SE) | low | resolved (v0.9.2) |
| HML-032 | Cloud SQL baseline: permissions, collectors and load (validation) | Cloud SQL (SQL Server 2022 SE) | — | resolved |
| HML-033 | TLS on Cloud SQL: `encrypt=true` with the instance CA; `encrypt=strict` is not accepted | Cloud SQL (SQL Server 2022 SE) | low | resolved (README) |
| HML-034 | Concurrent scrapes saturate the 5-connection pool: collectors fail with `context deadline exceeded` | Cloud SQL (SQL Server 2022 SE) | medium | resolved (v0.14.0) |
| HML-035 | Dashboard for Cloud SQL: exporter + Cloud Monitoring (validation) | Cloud SQL (SQL Server 2022 SE) | — | resolved |
| HML-036 | Conversion to regional HA and Cloud SQL failover (validation) | Cloud SQL (SQL Server 2022 SE) | — | resolved |
| HML-037 | Whole scrape lost when the server stops answering without closing the connection | Cloud SQL (SQL Server 2022 SE) | medium | resolved (v0.12.1) |
| HML-038 | SQL Server Express without Agent (local validation) | Local Docker (SQL Server 2025 Express) | — | resolved |
| HML-039 | SQL Managed Instance baseline: permissions, collectors and load (validation) | Azure SQL Managed Instance (GP) | — | resolved |
| HML-040 | Azure Free Trial: Managed Instance quota only in `westus3`, no B-series VMs and 404 when reusing the resource group name (environment) | Azure SQL Managed Instance (GP) | — | resolved |
| HML-041 | SQL through the SSH tunnel does not reach the MI: the gateway requires the instance name and redirects to the node (environment) | Azure SQL Managed Instance (GP) | low | resolved (T-SQL from the bastion) |
| HML-042 | `server_info` fails entirely with `MachineName` NULL on MI | Azure SQL Managed Instance (GP) | medium | resolved (v0.12.2) |
| HML-043 | `volumes` on MI: `total_bytes` is the size of the files, not of the volume (available above 100%) | Azure SQL Managed Instance (GP) | medium | resolved (v0.13.0) |
| HML-044 | `database_log_used_percent` without the user databases on MI (counter by physical name) | Azure SQL Managed Instance (GP) | low | resolved (v0.13.1) |
| HML-045 | `backups` without any series on MI: `server_name` NULL in `backupset` | Azure SQL Managed Instance (GP) | medium | resolved (v0.13.2) |
| HML-046 | `requests_long_running` always 1 in every bucket on MI: `TdService` telemetry session | Azure SQL Managed Instance (GP) | low | resolved (v0.13.2) |
| HML-047 | OS collector logs a swap `ERROR` on every scrape on Windows without a pagefile | Windows 11 (home notebook, SQL Server 2025 Developer) | medium | resolved (v0.13.3) |
| HML-048 | The README's `icacls` fails on non-English Windows (`Administrators` does not exist in pt-BR) | Windows 11 (home notebook, SQL Server 2025 Developer) | low | resolved (README) |
| HML-049 | Per-database collectors on MI with `CONNECT ANY DATABASE` (validation) | Azure SQL Managed Instance (GP) | — | resolved |
| HML-050 | TLS on MI: `encrypt=true` and `encrypt=strict` with a public CA, no certificate file (validation) | Azure SQL Managed Instance (GP) | — | resolved (README) |
| HML-051 | MI managed failover under load: ≈80 s of `sqlserver_up 0`, no scrape lost (validation) | Azure SQL Managed Instance (GP) | — | resolved |
| HML-052 | Dashboard for Managed Instance: exporter + Azure Monitor (validation) | Azure SQL Managed Instance (GP) | — | resolved |
| HML-053 | Azure SQL Database reconnaissance: 10 of 19 collectors work; the rest depend on `sys.master_files`, `msdb` and server DMVs (validation) | Azure SQL Database (serverless GP, free offer) | — | deferred (ROADMAP) |
| HML-054 | Automatic instance label, disk queue and idle waits (validation) | Home lab + local Docker | — | resolved (v0.15.0) |
| HML-055 | `transaction_log` collector (`sys.dm_db_log_stats`): VLFs and log since the last log backup (validation) | Local Docker + home lab | — | resolved (v0.16.0) |
| HML-056 | `missing_indexes` collector (opt-in): optimizer suggestions (validation) | Local Docker + home lab | — | resolved (v0.17.0) |
| HML-057 | `replication` collector (opt-in): transactional replication agents on the Distributor (validation) | Local Docker (SQL Server 2025 with Agent) | — | resolved (v0.17.0) |
| HML-058 | `agent_jobs` and `replication` convert old times with the server's current offset | Any server with daylight saving time | low | won't fix (documented) |
| HML-059 | Coverage dashboard without the metrics of the newest collectors | Grafana (homologation) | low | open |

## Closing of the EC2 phase (2026-10-15)

EC2 Windows phase closed with v0.9.0 running on 01. Covered:

- **SQL Server versions:** 2019 CU32 → 2022 → 2025 RTM with a rolling upgrade
  of both instances (HML-017), with no query change.
- **HA/DR:** Always On AG with 2 replicas (failover, suspend/resume, lag),
  database mirroring (failover, partner down, suspend) and log shipping.
- **Connection:** `encrypt=strict` (TDS 8) with our own certificate, after the
  migration to `microsoft/go-mssqldb` (HML-016).
- **Operation:** native Windows service (install/start/stop/status, reboot,
  restart by the SCM after a kill, uninstall/reinstall), Event Log, failure
  scenarios (HML-009), UTC-3 time zone (HML-010), scrape timeout (HML-013).
- **ROADMAP phase 6 under `hack/loadgen` load:** sessions by login,
  `active_queries`, `top_queries` and `index_fragmentation` (HML-021 to HML-024).

Left for later:

- **RDS phase:** repeat the `msdb` permissions (HML-002) and test whether the
  master user can grant `CONNECT ANY DATABASE` (HML-024).
- **ROADMAP (Later):** state filter in the `volumes` query (HML-019) and
  a fix of `os_disk_*_queue_length` on Windows (pending from HML-003).
- **Lab limitations:** mirroring with a witness (no third instance); a
  virtual service account instead of `LocalSystem` (HML-006) remains a point
  to consider.

---

## Closing of the RDS phase (2026-10-19)

RDS for SQL Server phase closed with v0.10.0 running on the Linux bastion.
Environment: RDS SQL Server 2025 Standard (`17.0.5005.3`, RTM-CU9),
`db.t3.xlarge`, private subnets, UTC-3 time zone; exporter on the Ubuntu
26.04 bastion via systemd, read by the local Prometheus through an SSH
tunnel. Covered:

- **Linux installation:** sandboxed systemd unit shipped in the release
  (v0.9.1); installation, upgrade by replacing only the binary, host restart
  before the database (HML-025).
- **Permissions:** the same grants as on EC2 work when given by the master
  user, including the `msdb` ones (HML-002); `CONNECT ANY DATABASE` cannot be
  granted on RDS — a user per database instead (HML-024, README).
- **Collectors:** all 18 with `success 1` from the first collection, under
  `hack/loadgen` load and during a snapshot, the Multi-AZ conversion and a
  failover (HML-027, HML-030).
- **Fixes that came from RDS:** `log_used_percent` without hidden databases
  (v0.9.2, HML-026), backups to `NUL` ignored (v0.9.3, HML-028) and
  `instance_label` to pin the `sqlserver_instance` label (v0.10.0,
  HML-029).
- **Connection:** `encrypt=true` and `encrypt=strict` (TDS 8.0) with the AWS
  CA bundle, no server configuration; wrong CA refused (HML-027).
- **Backups:** manual and automated snapshot (FULL), LOG every 5 min with
  retention on, native backup through the procs (FULL and DIFFERENTIAL), all
  in correct UTC with the server in UTC-3 (HML-028).
- **High availability:** conversion to Multi-AZ (Basic AGs, one per database)
  and failover with ≈45 s of `sqlserver_up 0`, recovering on its own; stable
  label with `instance_label` (HML-029, HML-030).
- **Grafana:** exporter + CloudWatch demonstration dashboard for RDS
  (HML-031).

Left for later:

- **ROADMAP (Later):** `sys.dm_db_log_stats` metrics (HML-026), automatic
  label resolved on the first connection (HML-029) and an idle wait filter in
  `wait_stats` (HML-031).
- **Private repository:** the README's installation steps download the
  release from GitHub; while the repository is private, the binaries go
  through S3 (HML-025).
- **Not tested:** RDS Enterprise/Express/Web, read replicas,
  `rds.force_ssl = 1` and Enhanced Monitoring/Database Insights.

---

## Opening of the GCP phase — Cloud SQL for SQL Server (2026-10-20)

Second managed cloud, with the same plan as the RDS baseline (HML-027). The
main goal is the permission documentation per platform; Cloud SQL differs
from RDS because there is no `sysadmin` and every `GRANT` requires
`AS CustomerDbRootRole`. Azure comes after GCP (below).

**Planned environment**

- Project on the GCP free trial (credit valid until 2027-01-13). Resources
  created and destroyed with `gcloud`, with the commands recorded in the HMLs;
  a budget alert before the first resource.
- Region `us-central1`. Cloud SQL for SQL Server 2022 Standard (with Agent)
  and a short pass on Express. Private IP, UTC-3 time zone set at creation.
- Linux bastion in the same VPC with the exporter via systemd, read by the
  local Prometheus through an SSH tunnel, as on RDS.
- Cloud SQL stopped between test sessions (only the disk is billed).

**Plan** (each item becomes an HML when it is run)

1. Baseline: minimal login with `VIEW SERVER STATE` and `VIEW ANY DEFINITION`
   granted `AS CustomerDbRootRole`, and the alternative with
   `##MS_ServerStateReader##` / `##MS_DefinitionReader##`; default collectors
   with `success 1` under `hack/loadgen` load.
2. `msdb`: `backups` and `agent_jobs` with the minimal login. `msdb` is
   protected and only users created through the console get
   `CustomerDbRootRole`; document the recipe or the limit.
3. Per-database collectors (`index_fragmentation`, `query_store`):
   `CONNECT ANY DATABASE` or `##MS_DatabaseConnector##` instead of a user per
   database.
4. TLS: `encrypt=true` and `encrypt=strict` with the Cloud SQL server CA;
   the certificate does not have the IP in the SAN (`hostNameInCertificate`).
5. Platform: `@@SERVERNAME`, `sys.dm_os_host_info`, `volumes`, UTC-3 time
   zone in the timestamps (HML-010) and `instance_label`.
6. Regional HA and failover: `sqlserver_up` and recovery on its own, as on
   RDS (HML-030); check that no AG appears, since HA is by regional disk.
7. Cloud SQL automated backups and PITR in the `backups` collector.
8. Express: `msdb` without Agent, no license cost.
9. Pool concurrency (pending from the review): `hey -c 5` against the
   bastion's `/metrics`, watching `scrape_collector_duration_seconds`.

**Out of scope**

- Read replicas (Read Scale AG, Enterprise edition only): license cost; left
  as not tested, with the risk of the AG collector showing an AG the user does
  not manage.
- Exporter on GKE with a Helm chart: after the public release (ROADMAP).

**Next: Azure**

- SQL Managed Instance through the free offer (General Purpose, 12 months,
  public endpoint): permissions, remote storage in `volumes`/`io_file_stats`,
  time zone chosen at creation (UTC-3 to repeat HML-010).
- SQL Database through the free offer, as reconnaissance only: one exporter
  per database, incompatible collectors disabled, to size the new mode
  (ROADMAP).

**Pending items from the post-homologation review (2026-10-20)**

- `volumes`: filter `sys.databases.state IN (0, 1)` (HML-019, ROADMAP).
  `ONLINE` only would make the volume of secondaries in `RESTORING`
  disappear.
- Mirroring with a witness: remains a lab limitation.
- Before the public release: ~~scan the history with `gitleaks`~~ (2026-10-28,
  8.30.1 over all refs, merges included: nothing found; the DSNs only have
  example passwords and the CI container's; only names of resources already
  destroyed become public, such as the `EC2AMAZ-*` hosts and the MI DNS zone;
  CI now runs `make scan-secrets` on every PR); update the versioned
  `CLAUDE.md`; ~~review the versioned `prometheus.yml`~~ (no longer versioned
  since 2026-10-28, example in the README); ~~change port 9187 (allocated to
  postgres_exporter)~~ (the default became 10060 on 2026-10-28; 9100–9999 is
  fully allocated and 9433 belongs to the Knot exporter); register 10060 in
  the Prometheus port list after the repository becomes public (check first
  that it is still free); optional CHANGELOG.
- Already done: `ContinueOnError` on `/metrics` (HML-004, v0.1.1), tests of
  every collector through a pedantic registry and dashboards with
  `${DS_PROMETHEUS}` / `${DS_CLOUDWATCH}`.

---

## Closing of the GCP phase (2026-10-26)

Google Cloud SQL for SQL Server phase closed with v0.12.1 running on the
bastion. Environment: Cloud SQL SQL Server 2022 Standard (`16.0.4262.2`) on
Linux, `db-custom-2-7680`, private IP only, `ENCRYPTED_ONLY`, UTC-3 time
zone, later regional (HA); exporter on the Ubuntu 26.04 bastion via systemd,
SSH through IAP, read by the local Prometheus through a tunnel. Resources in
Terraform, outside the repository. Covered:

- **Permissions:** no `sysadmin`; `VIEW SERVER STATE` and `VIEW ANY
  DEFINITION` only with `AS [CustomerDbRootRole]`; `##MS_...##` roles,
  `CONNECT ANY DATABASE` and users in `msdb` are not allowed; `backups`
  works without a grant and `agent_jobs` stays disabled (HML-032, README).
- **Collectors:** 16 of 17 default ones and the opt-in `index_fragmentation`
  and `query_store` with `success 1`, under `hack/loadgen` load; first
  environment with SQL Server on Linux, with no query change (HML-032).
- **Connection:** `encrypt=true` with the instance CA (private IP in the
  SAN); `encrypt=strict` is not accepted by Cloud SQL (HML-033, README).
- **High availability:** conversion to regional (≈5 min 45 s down) and two
  failovers under load (≈77 s and ≈60 s of `sqlserver_up 0`); `@@SERVERNAME`
  stable across zones (HML-036).
- **Fix that came from GCP:** whole scrape lost when the server stops
  answering without closing the connection — the exporter now answers within
  the deadline (v0.12.1, HML-037).
- **Concurrency:** up to 2 simultaneous scrapes without failure; 5 or more
  saturate the pool (HML-034).
- **Grafana:** exporter + Cloud Monitoring dashboard, with keyless
  authentication through impersonation (HML-035).
- **Express:** validated locally in Docker, not on Cloud SQL (HML-038).

Left for later:

- **ROADMAP (Later):** singleflight for simultaneous scrapes and a
  configurable pool (HML-034); fallback label in `volumes` for SQL Server on
  Linux (HML-032, HML-038).
- **Not tested:** Cloud SQL Enterprise/Web, read replicas (read-scale AG),
  PITR restore, Cloud SQL Auth Proxy and logins created through the Cloud SQL
  API reading the Agent jobs.
- **Next phase:** Azure (SQL Managed Instance and SQL Database
  reconnaissance), as planned at the opening.

---

## Opening of the Azure phase — SQL Managed Instance (2026-10-27)

Third managed cloud, with the same plan as the previous ones. Managed
Instance is the first environment with `EngineEdition` 8: the exporter sees a
complete instance (Agent, `msdb`, server DMVs), but with remote storage and
physical names different from those of an installed SQL Server.

**Environment**

- Free Trial subscription (US$ 200 of credit, 30 days). Resources in
  Terraform (`azurerm` + `azapi`), outside the repository, with plan/apply run
  by the operator; US$ 50/month budget with alerts.
- Region `westus3` (HML-040). SQL Managed Instance through the free offer:
  General Purpose Gen5, 4 vCores, 64 GB, `SQLServer2022` format, private
  endpoint only (VNet), UTC-3 time zone (`E. South America Standard Time`),
  start/stop schedule Mon–Fri 13:00–21:00 to fit the 720 vCore-hours of the
  month.
- Ubuntu 26.04 bastion (`Standard_D2s_v4`) in the same VNet, with the
  exporter via systemd, shut down at 21:10; local Prometheus through the SSH
  tunnel, as on RDS and Cloud SQL. Ad hoc T-SQL runs on the bastion with
  `go-sqlcmd` (HML-041).

**Plan** (each item becomes an HML when it is run)

1. Baseline: minimal login, default collectors under `hack/loadgen` load
   (HML-039).
2. MI differences in the collectors: `server_info`, `volumes`,
   `log_used_percent`, `backups` and `active_queries` (HML-042 to HML-046).
3. Per-database collectors (`index_fragmentation`, `query_store`) with
   `CONNECT ANY DATABASE`.
4. TLS: `encrypt=true` (in use) and `encrypt=strict`.
5. Microsoft Entra ID authentication instead of the SQL login: the exporter
   does not support it; it went to the ROADMAP (Later), with no priority.
6. MI managed failover (`az sql mi failover`): `sqlserver_up` and recovery on
   its own, as on RDS (HML-030) and Cloud SQL (HML-036).
7. Exporter + Azure Monitor dashboard (HML-052).
8. SQL Database through the free offer, as reconnaissance only: one exporter
   per database, incompatible collectors disabled, to size the new mode
   (ROADMAP; HML-053).

---

## Closing of the Azure phase (2026-10-27)

Azure phase closed with v0.13.2 running on the bastion. Environment: Azure
SQL Managed Instance through the free offer (General Purpose Gen5, 4 vCores,
64 GB, private endpoint only, UTC-3 time zone) and, for the reconnaissance, a
serverless Azure SQL Database through the free offer; exporter on the Ubuntu
26.04 bastion via systemd, read by the local Prometheus through an SSH
tunnel. Resources in Terraform, outside the repository. Covered:

- **Permissions:** the MI SQL admin is `sysadmin`; the EC2 grants, including
  `CONNECT ANY DATABASE` and the user in `msdb`, work as on an installed SQL
  Server (HML-039, HML-049, README).
- **Collectors:** the 16 default ones and the opt-in `top_queries`,
  `index_fragmentation` and `query_store` with `success 1`, under
  `hack/loadgen` load and during the failover (HML-039, HML-049, HML-051).
- **Fixes that came from Azure:** NULL properties in `server_info`
  (v0.12.2, HML-042), instance storage instead of the volumes (v0.13.0,
  HML-043), `log_used_percent` by physical name (v0.13.1, HML-044), backups
  with NULL `server_name` and the `TdService` session left out of long
  requests (v0.13.2, HML-045, HML-046). All of them apply only with
  `EngineEdition` 8; on the other platforms the queries return the same as
  before.
- **Connection:** `encrypt=true` and `encrypt=strict` (TDS 8.0) with the
  public Microsoft CA, no certificate file; the exporter must run in the VNet
  because of the redirection to the node (HML-041, HML-050, README).
- **High availability:** managed failover under load, ≈80 s of
  `sqlserver_up 0`, no scrape lost and a stable `@@SERVERNAME` (HML-051).
- **Grafana:** exporter + Azure Monitor dashboard for MI (HML-052).
- **SQL Database:** reconnaissance with 10 of 19 collectors working and its
  own mode sized in the ROADMAP (HML-053).

Left for later:

- **ROADMAP (Later):** SQL Database mode (HML-053), Microsoft Entra ID
  authentication (no priority) and an idle wait filter in `wait_stats`, with
  the MI background waits (HML-039).
- **Not tested:** MI Business Critical and Next-gen General Purpose, failover
  groups, Managed Instance link, instance pools, the MI public endpoint;
  SQL Database Hyperscale, elastic pools, geo-replication and the real effect
  of the scrape on the serverless auto-pause.
- **Home lab:** the Windows notebook and the Raspberry Pi (below) remain the
  permanent environment to validate the next changes.

---

## Home lab — Windows + Raspberry Pi (2026-10-27)

Permanent environment to validate changes on Windows, in addition to local
Docker/Linux and the clouds:

- Windows 11 Pro **pt-BR** notebook (`INSPIRON`, i7-7500U, 16 GB) with SQL
  Server 2025 Developer RTM (`17.0.1000.7`), on the home network. Exporter as a
  Windows service (`C:\mssql-exporter`, config in
  `C:\ProgramData\mssql-exporter\config.yml`), OS collectors on,
  `instance_label` `notebook-win`; access through OpenSSH with a key.
- Raspberry Pi 5 (Debian 13 aarch64) with Prometheus 3.10 and Grafana 12.4 in
  Docker: `sqlserver` job scraping the notebook every 15 s.

---

## HML-001 — `ERROR` log when `.env` does not exist

- **Environment:** EC2 Windows, release v0.1.0 (`mssql-exporter_0.1.0_windows_amd64.zip`)
- **Found on:** 2026-10-08
- **Severity:** low
- **Status:** resolved (v0.1.1)

**Reproduction**

```powershell
.\mssql-exporter.exe --config.file config.yml
```

**Observed**

```
ERROR Error loading .env file error="open .env: The system cannot find the file specified."
```

The exporter keeps working normally (connects, pings, exposes `:9187/metrics`).

**Cause**

`main.go` calls `godotenv.Load()` unconditionally and logs any failure at
`ERROR` level. `.env` is a local development convenience; in production the
configuration comes from `--config.file` or environment variables, so a
missing file is the normal case. Running as a service, the working directory
is `C:\Windows\System32` and the message appears on every startup.

**Impact**

False positives in alerts based on log level (CloudWatch Logs, SIEM) and
confusion during operation.

**Proposed fix**

Ignore `fs.ErrNotExist` and downgrade the other errors to `WARN`:

```go
if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
    slog.Warn("failed to load .env file", slog.Any("error", err))
}
```

Suggested branch: `fix/dotenv-missing-log` (label `release:patch`).

**Fix (2026-10-08)** — branch `fix/homolog-windows-findings`, test build
`0.1.1-homolog`. Validated locally: without `.env` nothing is logged; a
malformed `.env` produces a `WARN`. Still to validate on EC2.

---

## HML-002 — `agent_jobs` fails without permission on msdb (point of attention)

- **Environment:** EC2 Windows, release v0.1.0
- **Found on:** 2026-10-08
- **Severity:** medium
- **Status:** resolved on EC2 (v0.1.1; repeat the validation on RDS)

**Observed**

```
sqlserver_scrape_collector_success{collector="agent_jobs"} 0
... The SELECT permission was denied on the object 'sysjobhistory', database 'msdb', schema 'dbo'.
```

The other collectors are not affected.

**Cause**

The collector reads `msdb.dbo.sysjobs` and `msdb.dbo.sysjobhistory` directly,
and the monitoring login has no access to those tables.

**Permission options**

1. **Grants on the two tables only** (least privilege, already documented in the README):

   ```sql
   USE msdb;
   CREATE USER [exporter] FOR LOGIN [exporter];
   GRANT SELECT ON dbo.sysjobs TO [exporter];
   GRANT SELECT ON dbo.sysjobhistory TO [exporter];
   ```

2. **`SQLAgentReaderRole` role**:

   ```sql
   USE msdb;
   CREATE USER [exporter] FOR LOGIN [exporter];
   ALTER ROLE SQLAgentReaderRole ADD MEMBER [exporter];
   ```

   **Note:** the SQL Agent roles give access through procedures and views
   (`sp_help_job`, `sysjobs_view`), not through `SELECT` on the base tables the
   collector queries. The role alone probably does **not** solve it. Validate
   in the homologation before recommending it.

**Validation result (2026-10-08, EC2 Windows)**

The **grants on `dbo.sysjobs` and `dbo.sysjobhistory` only** (option 1)
solved it: `sqlserver_scrape_collector_success{collector="agent_jobs"}` went
to `1` and `sqlserver_agent_job_enabled` started being emitted. The
`sqlserver_agent_job_last_run_*` metrics do not appear yet because no job
has completed a run in the environment.

Option 2 (`SQLAgentReaderRole`) was not tested. **Recommendation: option 1**,
which is the least privileged and the one the README already documents.

**Least privilege (2026-10-13)**

The homologation login (`usr_exporter`) has exactly the README grants:
`VIEW SERVER STATE` and `VIEW ANY DEFINITION` on the server, and in msdb a
user with `SELECT` on `dbo.sysjobs` and `dbo.sysjobhistory`. With it, all 14
SQL collectors had `success 1` on both replicas, including with AG, failover
and suspension. Least privilege validated; no other login is needed.

**Points to resolve**

- A comment in `internal/collector/agent_jobs.go` mentions `SQLAgentReaderRole`,
  while the README documents the grants on the tables. Update the comment to
  the validated grants.
- Run an Agent job to completion to validate the
  `sqlserver_agent_job_last_run_*` metrics.
- In environments without SQL Agent (Express) or with no interest in jobs,
  disable it with `--collector.agent_jobs=false` or `collectors.agent_jobs: false` in YAML.
- Repeat the validation on RDS for SQL Server, where msdb access has its own
  restrictions.

---

## HML-003 — Disk I/O collection on Windows fails with `Invalid class`

- **Environment:** EC2 Windows, release v0.1.0, `collectors.os_collectors: true`
- **Found on:** 2026-10-08
- **Severity:** medium
- **Status:** resolved (v0.1.1)

**Observed** (on every scrape)

```
ERROR failed to query disk performance counters error="Exception occurred. (Invalid class )"
```

No I/O `os_disk_*` metric is emitted on Windows.

**Cause (confirmed in the code)**

`internal/oscollector/disk_io_windows.go` builds the query with
`wmi.CreateQuery(&disks, "")`. Without the third argument, the
`yusufpapurcu/wmi` library uses the **Go type name** as the WMI class. The
generated query is `SELECT ... FROM win32RawDiskPhysicalDisk`, which does not
exist. The correct class is `Win32_PerfRawData_PerfDisk_PhysicalDisk`.

**Proposed fix**

Pass the class explicitly:

```go
query := wmi.CreateQuery(&disks, "", "Win32_PerfRawData_PerfDisk_PhysicalDisk")
```

**Check together with the fix**

- `PercentDiskTime` is of type `PERF_PRECISION_100NS_TIMER` (100 ns unit).
  Dividing by `Frequency_PerfTime` only gives the right value when the
  frequency is 10 MHz. Check `os_disk_io_time_ms_total` against Performance
  Monitor.
- `AvgDiskReadQueueLength` and `AvgDiskWriteQueueLength` are raw accumulators
  in the `PerfRawData` class, but are exposed as gauges. The values will grow
  without end.
- The OS collectors expose no per-scrape success metric (unlike
  `sqlserver_scrape_collector_success`), so the failure only shows in the log.
- There are no tests for the collectors of `internal/oscollector`.

Suggested branch: `fix/windows-disk-io-wmi-class` (label `release:patch`).

**Fix (2026-10-08)** — branch `fix/homolog-windows-findings`, test build
`0.1.1-homolog`:

- WMI class passed explicitly (`Win32_PerfRawData_PerfDisk_PhysicalDisk`).
- `os_disk_io_time_ms_total` now converts `PercentDiskTime` from 100 ns ticks
  to ms, without depending on `Frequency_PerfTime`.
- **Pending:** `os_disk_read_queue_length`/`os_disk_write_queue_length`
  remain gauges over raw accumulators. Fixing it requires changing the metric
  type and name (and aligning with Linux); handle it in its own item.

Validated on EC2 on 2026-10-08 with build `0.1.1-homolog`: I/O `os_disk_*` metrics showing in the charts.

**Update (2026-10-28, v0.15.0):** `os_disk_read_queue_length` and
`os_disk_write_queue_length` were gauges over cumulative counters
(`PERF_COUNTER_100NS_QUEUELEN_TYPE`) and only grew. They were replaced by
`os_disk_read_time_seconds_total` and `os_disk_write_time_seconds_total`
(counters; `rate()` is the average queue), on Linux too; see HML-054.

---

## HML-004 — Intermittent gaps: `/metrics` returns HTTP 500 because of a duplicate series in `volumes`

- **Environment:** EC2 Windows, build `0.1.1-homolog`, with load on `StressDB`
- **Found on:** 2026-10-08
- **Severity:** high (the whole scrape is lost)
- **Status:** resolved (v0.1.1)

**Observed**

Gauge charts (PLE, buffer cache hit ratio, sessions, connections) with
intermittent gaps, while `rate()` charts (batch requests, compilations) look
continuous: `rate()` over the window "skips" missing samples and hides the
problem. Nothing appears in the exporter log.

In Prometheus, the gaps match `up == 0` and `scrape_samples_scraped == 0`;
the target reports `server returned HTTP status 500 Internal Server Error`.
The exporter's response:

```
An error has occurred while serving metrics:
2 error(s) occurred:
* collected metric "sqlserver_volume_total_bytes" { ... volume_mount_point="C:\\" ... } was collected before with the same name and label values
* collected metric "sqlserver_volume_available_bytes" { ... volume_mount_point="C:\\" ... } was collected before with the same name and label values
```

**Cause (two combined failures)**

1. **`volumes` emits duplicate series.** The query uses `SELECT DISTINCT` over
   `(mount_point, total_bytes, available_bytes)`. `sys.dm_os_volume_stats` is
   evaluated per database file; under writes, files on the same volume return
   different `available_bytes` in the same run, the `DISTINCT` does not
   collapse the rows and the same `C:\` comes out twice.
2. **One invalid metric takes down the whole scrape.** `promhttp.Handler()`
   uses `HTTPErrorOnError` and has no `ErrorLog`: any collection error becomes
   HTTP 500 for **all** metrics, without any log line.

**Fix** — branch `fix/homolog-windows-findings`, test build `0.1.2-homolog`:

- `internal/collector/volumes.go`: `GROUP BY` mount point with
  `MAX(total_bytes)`/`MIN(available_bytes)` (one row per volume, conservative
  free space value).
- `main.go`: handler with `promhttp.ContinueOnError` and `ErrorLog` through
  `slog`. An invalid series is dropped and logged at `ERROR`; the rest is
  served.

**Validation**

- On EC2, under load: `up` must stay at `1` continuously and the gaps must go away.
- Query to follow: `min_over_time(up{job="sqlserver_exporter_homolog"}[1h])`.
- **Validated on 2026-10-08** with `0.1.2-homolog` under load: 41 scrapes in
  10 min, minimum `up` 1, no HTTP 500, all collectors with `success 1` and
  continuous gauge series (40/40 PLE samples). Before: ~7% of the scrapes
  with HTTP 500.

**Points to consider**

- Add an `up` / `scrape_samples_scraped` panel to the coverage dashboard:
  `sqlserver_up` does not detect this kind of failure, since it also
  disappears on the 500.
- `backups` could, in theory, produce a duplicate if two backups of the same
  database and type finish at the same `backup_finish_date`. With
  `ContinueOnError` this no longer takes down the scrape, but the query is
  worth reviewing.

---

## HML-005 — EC2 C: drive almost full during load on `StressDB` (environment)

- **Environment:** EC2 Windows (`AWSDBSERVER01`), 59.5 GiB C: volume
- **Found on:** 2026-10-08
- **Severity:** high (risk of autogrowth failure on `StressDB`/`tempdb`)
- **Status:** resolved (space freed manually)
- **Type:** homologation environment, not an exporter defect

**Observed**

During the HML-004 validation, `os_filesystem_free_bytes{mountpoint="C:"}` and
`sqlserver_volume_available_bytes{volume_mount_point="C:\\"}` dropped to
~9 MB free (they were ~1.7 GiB around 13:50). `StressDB` had 7.7 GiB of data
and was growing under SqlQueryStress.

**Action**

Files removed from the EC2 to free space.

**Points to consider**

- Positive for the homologation: the exporter detected the situation through
  both paths (`volumes` through SQL Server and `disk_usage` through the OS
  collectors).
- For the next load rounds, put data/log on a dedicated EBS volume or limit
  `StressDB` growth (`MAXSIZE`).
- Example dashboard: consider a free space alerting rule
  (e.g. `sqlserver_volume_available_bytes / sqlserver_volume_total_bytes < 0.1`).

---

## HML-006 — Exporter does not run as a native Windows service (point of attention)

- **Environment:** EC2 Windows (instances 01 and 02)
- **Found on:** 2026-10-08
- **Severity:** medium
- **Status:** resolved (v0.2.0; `uninstall` validated on 2026-10-15)

**Observed**

In the homologation the exporter runs in the foreground in a PowerShell
console. The EC2s are stopped/started often (to save credits), and on every
start the exporter has to be started by hand; if the console is closed the
process dies and the target goes `down` without any warning.

**Cause**

The binary does not implement the Service Control Manager protocol
(`golang.org/x/sys/windows/svc`). Registering the `.exe` directly with
`sc.exe create` fails with error 1053 after ~30 s; today the only option is a
wrapper (WinSW/NSSM), with manual configuration (XML, recovery, logs).

**Fix** — branch `feat/windows-service` (on top of `fix/homolog-windows-findings`),
test build `0.2.0-homolog`:

- `mssql-exporter service install|uninstall|start|stop|status` subcommand
  (`--config.file`, `--name`).
- `install` validates the config file, registers with automatic (delayed)
  start and recovery (restarts after 5 s, 30 s, 60 s; also on exit with an
  error) and registers the source in the Event Log.
- Running as a service: logs in the Event Log (Application), working
  directory = the config file's folder, clean stop (`Stop`/`Shutdown`).
- Outside the service: Ctrl+C/SIGTERM now also shut down cleanly.

**Validation on EC2 (2026-10-12, release v0.2.0)**

- [x] `service install` / `start` / `status` on instances 01 and 02: installed
  at the first try, `Automatic (Delayed Start)`, running as `LocalSystem`.
- [x] Prometheus: both targets `up 1`, `sqlserver_up 1`, version `0.2.0`,
  no collector failing.
- [x] EC2 reboot → target comes back on its own (02, 2026-10-12; details in HML-009).
- [x] `Stop-Process -Force` → the SCM restarts it (5 s on the 1st failure, 30 s on the 2nd; details in HML-009).
- [x] `service stop` / `start`: used in the upgrades to 0.2.1-homolog and
  0.3.0 on both replicas, with no orphan process. `uninstall` is left for the
  end of the homologation.
- [x] Event Log with the exporter's entries (details in HML-009, item 4).

Steps:

```powershell
.\mssql-exporter.exe service install --config.file C:\mssql-exporter\config.yml
.\mssql-exporter.exe service start
.\mssql-exporter.exe service status
Get-EventLog -LogName Application -Source mssql-exporter -Newest 10
```

- Restart the EC2 and confirm that the target comes back on its own in Prometheus.
- Kill the process (`Stop-Process -Name mssql-exporter -Force`) and confirm
  that the SCM restarts it in ~5 s.
- `service stop` / `service uninstall` without leaving an orphan process.

**`uninstall` validation (2026-10-15, v0.9.0, EC2 01)**

- [x] `service uninstall` with the service running: stopped and removed it
  (`Service "mssql-exporter" removed.`); `Get-Service` does not find the
  service, no `mssql-exporter` process, port 9187 free and the Event Log
  source removed from the registry.
- [x] `service install` + `start` again: `automatic delayed start, restarts
  on failure`, running. In Prometheus, `up 0` only between 18:01:10 and
  18:01:25; back at 18:01:40 with the 19 collectors at `success 1`.

**Points to consider**

- Service account: `LocalSystem`. Consider a virtual account
  (`NT SERVICE\mssql-exporter`) for least privilege — it requires an ACL on
  the config file and validating access to the WMI counters of the OS
  collectors.
- The firewall rule is still manual.

---

## HML-007 — Validation of the `availability_groups` collector on a real AG (validation)

- **Environment:** AG `MSSQL_AG` between `AWSDBSERVER01` and `AWSDBSERVER02`
  (SQL Server 2019, EC2 Windows), database `AlwaysOnDB`, exporter v0.2.0 as a
  service on both replicas
- **Started on:** 2026-10-12
- **Type:** validation (first run against a real Always On)
- **Status:** resolved (v0.3.0)

**Initial result (AG just created, no load)**

| Metric | Seen by 01 (primary) | Seen by 02 (secondary) |
|---|---|---|
| `sqlserver_is_hadr_enabled` | 1 | 1 |
| `sqlserver_hadr_replica_role` | 01 = 1 (primary), 02 = 2 | 02 = 2 |
| `sqlserver_hadr_replica_synchronization_health` | 2 / 2 | 2 |
| `sqlserver_hadr_database_synchronization_state` | 2 / 2 (synchronized) | 2 |
| `sqlserver_hadr_database_log_send_queue_kb` | 0 (02 series only) | 0 |
| `sqlserver_hadr_database_redo_queue_kb` | 0 (02 series only) | 0 |

`sqlserver_scrape_collector_success{collector="availability_groups"} 1` on both.

**Expected behaviors (not defects)**

- The secondary sees only its own replica in the `sys.dm_hadr_*` DMVs; the
  primary sees all of them. The same series (`replica_server="AWSDBSERVER02"`)
  comes from both exporters, told apart by `sqlserver_instance`. In panels
  and alerts: `max by (ag_name, replica_server, db_name)`.
- Queues only exist for secondary replicas; the primary's null values are
  dropped by the collector.

**Remaining steps**

- [x] Write load on `AlwaysOnDB` (SqlQueryStress, `INSERT dbo.EventLog`
  with `REPLICATE('x', 200)`, 32 threads × 10k iterations, no delay).
  19:44–19:46, peak of ~13.3k batch requests/s on the primary:
  - `log_send_queue_kb` stable at ~60 KB; `redo_queue_kb` between 296 and 432 KB.
  - `synchronization_state` 2 and `synchronization_health` 2 the whole time.
  - Consistent with synchronous commit: each commit waits for hardening on the
    secondary, so the send queue does not build up.
  - Earlier peak at 19:37, **without load** (~2 batch/s): `log_send_queue_kb`
    up to ~349 MB and `redo_queue_kb` up to ~33 MB, back to zero in ~1 min.
    Probably the AG's initial synchronization right after the setup. To confirm.
  - Values read straight from the DMV (already in KB), with no conversion in
    the collector.
- [x] Load in `ASYNCHRONOUS_COMMIT` (same script). At 20:32:13
  `synchronization_state` went from 2 (synchronized) to **1
  (synchronizing)**, as expected in asynchronous mode; `health` stayed at 2.
  Throughput went up to ~31.7k batch/s (twice the synchronous one: the commit
  no longer waits for the secondary). Queues stayed small (send ≤ 60 KB,
  redo ≤ 548 KB): the network and the secondary kept up with the load.
- [x] Failover to 02 (~21:04–21:05). Sequence seen in Prometheus (samples
  every 15 s, no scrape lost, `up`/`sqlserver_up` = 1 the whole time):
  - 21:04:00 — 02 back to `SYNCHRONOUS_COMMIT`: 02 health = 1
    (partially healthy) for one sample; 21:04:15 `synchronization_state` = 2.
  - 21:05:15 — **roles swapped on both exporters**: 02 = 1 (primary),
    01 = 2 (secondary). Reflected in a single sample (≤ 15 s).
  - 21:05:15 — replica 01 (former primary) with `is_suspended` = 1,
    **`suspend_reason` = 1 (SUSPEND_FROM_PARTNER)**, `synchronization_state`
    = 0 and health 0, seen by 01 and 02. According to the documentation,
    reason 1 is the suspension of the replica after a forced failover —
    confirmed: the failover was done with the SSMS wizard script (SQLCMD
    mode), which runs `SET (ROLE = SECONDARY)` on 01 and
    `FORCE_FAILOVER_ALLOW_DATA_LOSS` on 02. A manual RESUME of 01 is the
    expected step after a forced failover.
  - 21:07:00 — manual RESUME on 01: `is_suspended` 0, sync 2, health 2.
    Final state: 02 primary, 01 healthy secondary.
  - The `role` series of 01 seen by 01 went from 1 to 2 with no gap; the
    series that only exist on the primary (e.g. the remote replica's queues)
    changed exporter (they came from 01, now from 02), as expected.
- [ ] Secondary unavailable (stop 02's SQL Server): the primary should
  report `synchronization_health` < 2 and `synchronization_state` != 2.

---

## HML-008 — Send queue disappears from the chart with the replica suspended; `is_suspended` and `secondary_lag_seconds` missing

- **Environment:** AG `MSSQL_AG` (01 primary, 02 secondary in `ASYNCHRONOUS_COMMIT`), v0.2.0
- **Found on:** 2026-10-12, during HML-007
- **Severity:** medium (the AG's most serious problem shows up as "no data")
- **Status:** resolved (v0.3.0)

**Observed**

With `ALTER DATABASE AlwaysOnDB SET HADR SUSPEND` on 02 and load on the
primary (20:35–20:40), the `sqlserver_hadr_database_log_send_queue_kb` panel
was empty. Seen by the primary:

| Time | Load | send queue | redo queue | state | health | log used |
|---|---|---|---|---|---|---|
| 20:35–20:40 (suspended) | up to 33k batch/s | **no series** | 0 | 0 | 0 | 27% → 98% |
| 20:41 (RESUME) | — | 2.5 GB | 11 MB | 1 | 2 | 98% |
| 20:42 | — | 1.8 GB | 72 MB | 1 | 2 | 98% |

**Cause**

With data movement suspended SQL Server returns `NULL` in
`log_send_queue_size`; the collector omits `NULL` on purpose (zero would make
the stopped replica look healthy). `synchronization_state`/`health` = 0
already flagged the problem, but there was no metric saying *suspended* nor
the delay in time.

**Fix** — branch `feat/ag-suspend-and-lag`:

- `sqlserver_hadr_database_is_suspended` (0/1), present even with `NULL` queues.
- `sqlserver_hadr_database_suspend_reason` (only while suspended; 0 = user,
  1 = partner, 2 = redo, ...).
- `sqlserver_hadr_database_secondary_lag_seconds` (SQL Server 2016+).
- Unit test reproducing the observed format; README and coverage dashboard
  updated (105 metrics).

**Validation on EC2 (2026-10-12, build `0.2.1-homolog` on both replicas)**

SUSPEND run on the primary (01) at ~20:55, followed by load with large
transactions (50k rows per transaction, 4 threads):

| Series (seen by 01) | Value |
|---|---|
| `is_suspended{replica_server="AWSDBSERVER01"}` | **1** |
| `suspend_reason{replica_server="AWSDBSERVER01"}` | **0** (user) |
| `is_suspended{replica_server="AWSDBSERVER02"}` | 0 |
| `synchronization_state` (01 and 02) | 0 (not synchronizing) |
| `log_send_queue_kb` | missing (`NULL`), as before |
| `secondary_lag_seconds{replica_server="AWSDBSERVER02"}` | **0** |
| `AlwaysOnDB` log used on 01 | 8% → 46% |

- [x] `is_suspended` / `suspend_reason` work. With SUSPEND on the primary,
  SQL Server marks **the primary's** replica as suspended (the secondary
  stays at `is_suspended` 0, but `synchronization_state` 0). Alerts should
  use `max by (ag_name, db_name) (sqlserver_hadr_database_is_suspended) == 1`,
  without filtering by replica.
- [x] **`secondary_lag_seconds` cannot detect a suspension:** it stayed at 0
  during the whole suspension, even with the log growing. Use it together
  with `is_suspended` and `synchronization_state`, never alone. Document it
  in the README.
- [x] RESUME: `is_suspended` went back to 0 in the three series,
  `suspend_reason` disappeared, `log_send_queue_kb` reappeared (0) and the
  log used went back to 0%.
- [x] `secondary_lag_seconds` > 0 with data movement active: 21:01:15–21:01:45,
  02 in asynchronous mode catching up with the primary after the load with
  large transactions — lag 71 → 86 → 101 s, `log_send_queue_kb` 1.7 GB → 0,
  `redo_queue_kb` up to ~255 MB; back to zero at 21:02. The documentation
  confirms that the lag stays at 0 while the replica is suspended.
- [x] `suspend_reason` = 1 (partner) observed after the failover (see HML-007).
- Fix: the description of the `suspend_reason` metric had codes 3–8 shifted;
  fixed in the collector, README and dashboard according to the
  documentation of `sys.dm_hadr_database_replica_states`.

**Test note**

During the suspension the `AlwaysOnDB` log went from 27% to 98%: the log does
not truncate while the secondary does not receive. In production, a
suspended replica = risk of a full log (error 9002) on the primary. A good
candidate for an alert, together with `is_suspended`.

---

## HML-009 — Failure scenarios: database offline, SQL stopped, reboot, kill, differential backup (validation)

- **Environment:** AG `MSSQL_AG` (01 primary, 02 secondary), exporter v0.3.0
  as a service on both replicas, `usr_exporter` login with least privilege
- **Started on:** 2026-10-13
- **Type:** validation
- **Status:** resolved (all 5 scenarios validated)

### 1. Database `dbAdmin` OFFLINE on 02 — validated

- 21:30:47: `sqlserver_database_state{db_name="dbAdmin"}` went from
  `state="ONLINE"` to `state="OFFLINE"` in one scrape (the ONLINE series stops
  being emitted and the OFFLINE one is born; alerts:
  `sqlserver_database_state{state!="ONLINE"}`).
- No collector failed; `up` 1, no duplicate series.
- Still emitted: `data_size_bytes`/`log_size_bytes` (from
  `sys.master_files`) and `recovery_model`.
- No longer emitted (expected): `log_used_percent` and the
  `sqlserver_io_file_*` series of `dbAdmin` — the DMVs do not return offline
  databases.

### 2. SQL Server stopped on the secondary (02) — validated

02's SQL Server service stopped at ~21:37:45:

| Series | Before | After |
|---|---|---|
| `up{instance=<02>}` | 1 | **1** (exporter alive) |
| `sqlserver_up{sqlserver_instance="AWSDBSERVER02"}` | 1 | **0** from 21:37:51 |
| `scrape_samples_scraped` (02) | ~3,667 | ~108 (OS collectors + self-metrics only) |
| `scrape_duration_seconds` (02) | ~0.2 s | 0.16–2.15 s |
| Primary: `replica_synchronization_health{replica_server="AWSDBSERVER02"}` | 2 | **0** from 21:38:06 |
| Primary: `database_synchronization_state` (02) | 2 | **0** |
| Primary: `database_is_suspended` (02) | 0 | **0** |
| Primary: `log_send_queue_kb` (02) | 0 | missing (`NULL`) |

- The exporter does not go down when SQL Server goes down: `up` 1 +
  `sqlserver_up` 0 tells "SQL down" from "exporter down". Scrapers are
  skipped; no error.
- The primary detected the replica down in ≤ 15 s.
- `is_suspended` 0 with `synchronization_state` 0 tells a **disconnected
  replica** from a **suspended replica** (`is_suspended` 1). Useful for alerts.
- Recovery (SQL Server started, exporter **not** restarted): at 21:41:13, in
  a single scrape, `sqlserver_up` went back to 1, ~3,600 samples per scrape,
  no collector failing, and the primary again reported 02 with health/sync 2
  and queues 0. Automatic reconnection by the pool, label
  `sqlserver_instance="AWSDBSERVER02"` preserved (it was resolved at the
  exporter's start, when SQL was up).

### 3. Reboot of EC2 02 — validated (also closes an item of HML-006)

| Event | Time |
|---|---|
| Last OK scrape before the reboot | 21:45:49 |
| `os_boot_time_seconds` (Windows back) | 21:45:57 |
| Primary sees 02 with health/sync 0 | 21:46:21 (one sample) |
| Primary sees 02 healthy again (SQL Server already up) | 21:46:36 |
| Exporter's `process_start_time_seconds` | **21:48:07** (boot + ~2 min 10 s) |
| First OK scrape, `sqlserver_up` 1, no collector failing | 21:48:19 |

- The service started on its own, with no console.
- **`sqlserver_instance="AWSDBSERVER02"` filled in:** the delayed start
  (`Automatic (Delayed Start)`, ~2 min after boot) made the exporter start
  after SQL Server, so the label was resolved correctly.
- Cost: ~2 min 30 s without metrics from the machine on every boot, almost
  all of it because of the delay. Possible improvement: resolve
  `sqlserver_instance` on the first successful ping (not only at start),
  which would allow automatic start without delay and remove the known
  empty-label limitation.

### 4. Process kill — validated (also closes an item of HML-006)

`Stop-Process -Name mssql-exporter -Force` on 02. `process_start_time_seconds`
recorded two automatic restarts by the SCM, with no intervention:

| New process | Consistent with |
|---|---|
| 21:50:20 | 1st kill → recovery action "restart after 5 s" |
| 21:51:09 | 2nd kill (done on purpose right after) → "restart after 30 s" (21:50:56 scrape with `up` 0) |

- Right after the kill, `service status` shows `stopped` / PID 0: expected,
  it is the recovery action's waiting window.
- The 5 s → 30 s (→ 60 s on the 3rd) escalation follows `install`; the
  failure counter resets after 24 h without failures.
- After the restart: `sqlserver_up` 1, label filled in, no intervention.
- Event Log (`Application`, source `mssql-exporter`): complete startup
  sequence at 00:50:20 and 00:51:09 UTC (`starting mssql-exporter` with
  version 0.3.0, `successfully connected and pinged sql server`,
  `sql server instance instance=AWSDBSERVER02`, `os collectors enabled`,
  `Listening on`), all as *Information*. There is no entry for the kill itself
  (expected: `Stop-Process -Force` ends it with no chance to log).

### 5. Differential backup — validated

Database `diff_test` (SIMPLE recovery) on 02, full and differential backup
with `COMPRESSION` to `D:\BACKUP`:

| Metric | full | differential |
|---|---|---|
| `sqlserver_backup_last_completed_timestamp_seconds` | 21:58:07 | 21:58:11 |
| `sqlserver_backup_last_size_bytes` | 433,994 (~424 KB) | 59,482 (~58 KB) |
| `sqlserver_backup_last_duration_seconds` | 0 | 0 |

- `backup_type="differential"` appears correctly.
- The size is the compressed one (`compressed_backup_size`), consistent with
  `COMPRESSION`.
- Duration 0: `backupset` stores the dates with a precision of seconds and
  the empty database is backed up in under 1 s. Expected.
- Times correct **because Windows on the EC2s is in UTC** — see HML-010.

---

## HML-010 — Backup and Agent job timestamps treat SQL Server local time as UTC

- **Environment:** any SQL Server whose time zone is not UTC (does not
  reproduce on the homologation EC2s, which are in UTC)
- **Found on:** 2026-10-12, reviewing the code during HML-009 (item 5)
- **Severity:** medium ("late backup" and "job did not run" alerts are off
  by the size of the time zone)
- **Status:** resolved (v0.3.1)

**Cause**

- `backups`: `msdb.dbo.backupset.backup_finish_date` is a `datetime` in the
  server's **local** time, without an offset. The driver returns a
  `time.Time` without a zone (treated as UTC) and the collector emits
  `finishDate.Unix()`.
- `agent_jobs`: `run_date`/`run_time` of `sysjobhistory` are local time too;
  `DATETIMEFROMPARTS` + `.Unix()` has the same problem.

**Impact (example)**

Server in `UTC-03:00`: a backup at 21:00 local (00:00 UTC) becomes 21:00 UTC
in the timestamp, i.e. 3 h in the past.
`time() - sqlserver_backup_last_completed_timestamp_seconds` shows the backup
3 h older than it is. In positive zones (`UTC+X`) the timestamp is in the
future. Durations are not affected (difference between two local dates).

**Proposed fix**

- `backups`: convert in the query with the `backupset.time_zone` column (the
  server's offset in 15-min intervals **at the time of the backup**, correct
  even with daylight saving time):
  `DATEADD(MINUTE, -15 * bs.time_zone, bs.backup_finish_date)`;
  `time_zone = 127` means unknown → fall back to the current offset.
- `agent_jobs`: convert with the server's current offset
  (`DATEDIFF(MINUTE, GETDATE(), GETUTCDATE())`).
- Unit tests with a non-UTC zone.

**Fix applied**

- `backups`: a `CROSS APPLY` computes the offset per row
  (`bs.time_zone * 15` when between -48 and 48; otherwise
  `DATEPART(TZOFFSET, SYSDATETIMEOFFSET())`) and the start/finish dates are
  converted with `DATEADD(MINUTE, -offset, ...)`. Duration unchanged.
- `agent_jobs`: `DATEADD(MINUTE, -DATEPART(TZOFFSET, SYSDATETIMEOFFSET()), ...)`.
- The test regexes pin the conversion; mutation test: removing it, 10 tests
  fail.

**Validation 1 (2026-10-12, `0.3.1-homolog`, EC2 01 and 02 in UTC-3 via `tzutil`)**

- [x] Correct conversion for backups written before the change
  (`time_zone` = 0): `diff_test` full/diff of 21:58 −03 → `00:58Z`.
- [ ] **New problem:** the backups taken *after* the change did not appear,
  and 01's `JB_BACKUP_FULL` went to `04:00Z` (3 h in the future). Cause: the
  "last" record was chosen by **local** time
  (`MAX(backup_finish_date)` / `ORDER BY run_date, run_time`). After moving
  the zone back 3 h, the old records (01:00 in UTC) look newer than the new
  ones (22:05 in −03) and the collector stays stuck on them for up to 3 h. It
  also affects the daylight saving change (fall-back).

**Fix 2 (`0.3.2-homolog`)**

- `backups`: `ROW_NUMBER() OVER (PARTITION BY database_name, type ORDER BY backup_set_id DESC)`.
- `agent_jobs`: `ORDER BY h.instance_id DESC`.
- Both are msdb identities: they only grow, regardless of the time zone.
- Remaining limitation: the job history does not keep the offset, so a run
  before a time zone change is converted with the current offset. It goes
  away as soon as the job runs again.

**Validation 2 (2026-10-12 22:12, `0.3.2-homolog`, EC2 in UTC-3, Agent running)**

| Series | Time (−03) | Age | Real? |
|---|---|---|---|
| 02 `diff_test` full | 22:07:30 | 5.1 min | ✅ |
| 02 `diff_test` differential | 22:08:00 | 4.6 min | ✅ |
| 01 `AlwaysOnDB` full | 22:08:37 | 4.0 min | ✅ |
| 01 `AlwaysOnDB` log | 22:08:35 | 4.0 min | ✅ |
| 01 `JB_BACKUP_FULL` (status 1) | 22:07:44 | 4.8 min | ✅ |
| 01 `JB_BACKUP_LOG` (status 1) | 22:08:32 | 4.0 min | ✅ |

- [x] Backups and jobs run after the time zone change appear with the real
  time; no timestamp in the future; no collector failing.
- [x] An old backup (02 `AlwaysOnDB` full, 19:33 −03, written in UTC) stays
  correct: the per-row conversion uses each backup's `time_zone`.

**How to reproduce in the homologation**

Change 02's time zone (`tzutil /s "E. South America Standard Time"`), restart
SQL Server, take a backup and compare
`sqlserver_backup_last_completed_timestamp_seconds` with the real time.

---

## HML-011 — No coverage of database mirroring and log shipping

- **Environment:** EC2 01 + 02 (SQL Server 2019), outside the `MSSQL_AG` AG
- **Found on:** 2026-10-13, HA/DR coverage review
- **Severity:** medium (legacy environments with these features have no HA/DR monitoring)
- **Status:** resolved (v0.4.0)

**Observed**

The exporter only covers Always On. Environments that still use database
mirroring (deprecated since 2012, but present in the current versions) or
log shipping have no state, queue or delay metric. The last log backup
(`backups`) and the result of the LSBackup/LSCopy/LSRestore jobs
(`agent_jobs`) already showed indirectly, but not the secondary's delay.

**Proposed fix**

- New `database_mirroring` collector (on by default; no new grant):
  - `sys.database_mirroring` (`mirroring_guid IS NOT NULL`): role, state,
    safety level, witness state and the `partner_instance` in an `_info`
    metric.
  - `Database Mirroring` counters from `sys.dm_os_performance_counters`:
    send/redo queues, log bytes sent/received/redone, Transaction Delay and
    Mirrored Write Transactions.
- New `log_shipping` collector (**off by default**,
  `--collector.log_shipping`): `msdb.dbo.log_shipping_monitor_primary` and
  `_secondary`. The `*_date_utc` columns are already UTC (without the HML-010
  problem). Needs `GRANT SELECT` on both tables.
- Support for collectors off by default (`ScraperEnabledByDefault`), with the
  default reflected in the flag and in YAML.

**Validation plan — mirroring**

1. New database (e.g. `MirrorDB`, FULL recovery) mirrored between 01 and 02
   using the `DATABASE_MIRRORING` endpoint the AG already uses.
2. Check `role`, `state` = 4 and `safety_level` on both instances, and the
   queues under load on the principal.
3. Failures: `SET PARTNER SUSPEND` / `RESUME`, `SAFETY OFF` (async), stop the
   mirror's SQL (state 1), manual failover (`SET PARTNER FAILOVER`, role swap).
4. Confirm that `AlwaysOnDB` (from the AG) does **not** appear in the collector.
5. Check whether `transaction_delay_ms_total` and `write_transactions_total`
   are monotonic (cumulative type in the DMV).

**Validation 1 (2026-10-13, `0.4.0-homolog`, `MirrorDB` 01 → 02, SAFETY FULL, no witness)**

| Metric | 01 (principal) | 02 (mirror) |
|---|---|---|
| role / state / safety | 1 / 4 / 2 | **no series** |
| `mirroring_info` | partner = AWSDBSERVER02 | **no series** |
| log sent / received (bytes) | 69632 / 0 | 0 / 69632 |
| redo bytes | 0 | 81920 |
| send / redo queues (KB) | 0 / 0 | 0 / 0 |

- [x] `database_mirroring` with `success 1` on both; `AlwaysOnDB` (AG) left
  out of the collector.
- [x] Counters consistent with each side's role.
- [ ] **Witness:** without a witness the principal reported `witness_state 0`
  (unknown) — the documentation says NULL. In practice it is 0 with an empty
  `mirroring_witness_name`. **Fix (`0.4.1-homolog`):** only emit it when
  `mirroring_witness_name <> ''`. The case with a witness is not validated in
  the lab (it would need a third instance).
- [x] **Mirror without state (limitation, won't be fixed):** on 02
  `usr_exporter` sees the `MirrorDB` row with all `mirroring_*` columns NULL.
  The documentation confirms: values on the mirror only for **sysadmin**.
  Tested `GRANT SELECT ON sys.database_mirroring` with a user in master:
  still NULL. Decision: do not require sysadmin; the state belongs to the
  session and the principal reports it (including "disconnected"), and after
  a failover the new principal takes over. Documented in the README.

**Validation 2 — load on `MirrorDB` (2026-10-13 11:50–11:54, `0.4.1-homolog`)**

| Time | send queue 01 / 02 | redo queue 01 / 02 | write tx (cum.) | delay ms (cum.) | log used 01 | state |
|---|---|---|---|---|---|---|
| 11:49:47 | 0 / 0 | 0 / 0 | 1 | 1 | 11% | 4 |
| 11:51:47 | 700 MB / 652 MB | 0 / 0 | 22 | 98 424 | 100% | 4 |
| 11:52:17 | 96 MB / 352 MB | 286 MB / 159 MB | 23 | 129 851 | 96% | 4 |
| 11:52:32 | 0 / 0 | 0 / 0 | 28 | 163 782 | 96% | 4 |
| 11:53:32 | 0 / 0 | 0 / 0 | 160 029 | 440 625 | 99% | 4 |

- [x] Queues, throughput (peak of ~23 MB/s sent by 01 and ~19 MB/s
  received/redone by 02) and transactions followed the load.
- [x] `transaction_delay_ms_total` and `write_transactions_total` monotonic
  (`resets()` = 0): `CounterValue` is right.
- [x] Average delay consistent with the kind of load: ~37 s per transaction
  in the phase of few large transactions (11:50–11:52, 700 MB of log in ~7
  commits) and ~1.7 ms in the phase of many small transactions (11:52:47
  onwards).
- [x] **Documentation fix:** both replicas report both queues (each from its
  own view), not only the principal (send) / the mirror (redo) as the help
  and the README said. Fixed in the help, README and dashboard.
- Environment note: the `MirrorDB` log on 01 stayed at 96–100% (FULL recovery
  without log backups).

**Validation 3 — SUSPEND under load (11:58) and RESUME (12:00)**

| Time | state (01) | send queue 01 / 02 | write tx (cum.) | log used 01 |
|---|---|---|---|---|
| 11:58:08 | 4 | 0 / 0 | 480 035 | 2% |
| 11:58:23 | **0** (suspended) | 0 / 0 | 480 035 | 2% |
| 11:59:23 | 0 | 1.4 GB / 1.7 GB | 480 035 | 54% |
| 12:00:38 | 0 | 3.3 GB / 3.2 GB | 480 035 | 99% |
| 12:00:53 | **2** (synchronizing) | 3.1 GB / 2.9 GB | 480 035 | 99% |

- [x] SUSPEND visible within 15 s (state 0) and the send queue growing in
  both views. Unlike the AG (HML-008), the queue does **not** disappear while
  suspended.
- [x] `write_transactions_total` and `transaction_delay_ms_total` stop growing
  while suspended (no transaction waits for the mirror), so the average delay
  panel is empty in that period — expected.
- [x] RESUME: state 2 and the queue draining (~9 MB/s of redo on 02).
- The `MirrorDB` log on 01 grew to 3.3 GB during the suspend (it does not
  truncate until the mirror receives).

- [x] Resynchronized at 12:02:31 (state 4): ~3.3 GB drained in ~2 min, with a
  peak of ~43 MB/s of redo on 02.

**Validation 4 — 02's SQL Server stopped (12:03)**

| Time | state (01) | `up` 02 | `sqlserver_up` 02 | `database_mirroring` 01 |
|---|---|---|---|---|
| 12:03:01 | 4 | 1 | 1 | success 1 |
| 12:03:16 | **1** (disconnected) | 1 | 1 | success 1 |
| 12:03:31 | 1 | 1 | **0** | success 1 |

- [x] The principal reports the disconnection within 15 s; 02's exporter
  stays up with `sqlserver_up 0` and without the mirroring series.

**Validation 5 — 02's SQL back (12:04) and manual failover (12:05)**

| Time | `sqlserver_up` 02 | role / state 01 | role / state 02 | `mirroring_info` |
|---|---|---|---|---|
| 12:04:24 | 0 | 1 / **4** | — | 01 → partner 02 |
| 12:04:39 | **1** | 1 / 4 | — | 01 → partner 02 |
| 12:05:24 | 1 | 1 / 4 (last scrape before) | **1 / 4**, safety 2 | — |
| 12:05:39 | 1 | — (becomes the mirror) | 1 / 4 | 02 → partner 01 |

- [x] Reconnection: state goes back to 4 without visibly going through 2
  (little pending log).
- [x] `SET PARTNER FAILOVER`: 02 starts reporting role, state, safety and
  `_info` (partner AWSDBSERVER01); 01 keeps only the counters, and
  `sqlserver_database_state` shows 02 ONLINE / 01 RESTORING.
- For one scrape (12:05:24) both appear as principal: it is just the offset
  between the scrapes of the two instances (01 was read before the
  failover). In "two principals" alerts, use a `for:` of at least 1 min.

**Result:** `database_mirroring` collector validated (load, suspend/resume,
partner down and failover). Only the case with a witness is pending (no
third instance in the lab).

**Validation 6 — log shipping `LogShipDB` 01 → 02 (2026-10-14, `0.4.2-homolog`)**

- [x] Grants on the `log_shipping_monitor_*` tables and `log_shipping: true`:
  `success 1` on both.
- [x] Correct thresholds: backup 3600 s (01), restore 2700 s (02).
- [ ] **Timestamps frozen at the configuration time (15:43:16)** although
  `LSBackup` generates a `.trn` every 3 min and `LSCopy`/`LSRestore` run.
  `sp_help_log_shipping_monitor` itself and even
  `log_shipping_primary_databases` (`last_backup_file`/`date` NULL) agree: the
  exporter reflects SQL Server.
- **Cause (environment):** SQL Server 2019 RTM/CU1 bug
  ([KB4537869](https://support.microsoft.com/help/4537869)) — `sqllogship`
  does not write history/monitor: "Could not log history/error message ...
  Failed to convert parameter value from a SqlGuid to a String". The EC2s are
  on 15.0.2000.5 (RTM). Fixed in CU2. Documented in the README.

**Validation 7 — log shipping after CU32 (2026-10-14 16:22–16:31, `0.4.2-homolog`)**

- [x] Timestamps moving after the patch: `primary_last_backup` 16:22:00,
  `secondary_last_copied` / `last_restored` 16:18:07, latency 240 s.
- [x] **`LSRestore` disabled + restore threshold at 2 min:** at 16:30,
  `last_restored` stopped at 16:27:01 while `last_copied` kept going
  (16:30:01); `restore_threshold_seconds` = 120 reflected the change; the
  alert expression `time() - last_restored > restore_threshold` fired (239 s).
- [x] **Grant removed on `log_shipping_monitor_primary` on 01:**
  `success{collector="log_shipping"} 0` only on that collector; the primary's
  series disappear and the other collectors stay at `success 1`.

**Validation plan — log shipping**

1. Grants on the `log_shipping_monitor_*` tables for `usr_exporter` on both
   instances and `log_shipping: true` in `config.yml`.
2. Log shipping of a third database (e.g. `LSDB`) from 01 to 02.
3. Check the timestamps (UTC) against the jobs and the restore latency.
4. Failures: disable the restore job and watch `time() - last_restored` pass
   the threshold; remove the grant and see `success 0` only on `log_shipping`.

---

## HML-012 — `backups` reports on the secondary the backups restored from the primary

- **Environment:** EC2 02 (secondary of the `LogShipDB` log shipping and mirror of `MirrorDB`), `0.4.1-homolog`
- **Found on:** 2026-10-14, during HML-011
- **Severity:** medium (the secondary looks like it has a recent backup without ever taking one)
- **Status:** resolved (v0.4.0)

**Observed**

02 reported `sqlserver_backup_last_completed_timestamp_seconds{backup_type="full"}`
for `LogShipDB` (2026-10-14 15:43:15) and `MirrorDB` (2026-10-13 11:20:40),
exactly the same seconds as the backups taken on 01. Query on 02's
`msdb.dbo.backupset`:

| server_name | database_name | type | backup_finish_date |
|---|---|---|---|
| AWSDBSERVER01 | MirrorDB | D | 2026-10-13 11:20:40 |
| AWSDBSERVER01 | LogShipDB | D | 2026-10-14 15:43:15 |

**Cause**

`RESTORE` writes the restored backup into the local `backupset`, with the
origin `server_name`. With log shipping running, each log restore would show
on 02 as a recent log backup, hiding that the secondary has no backup of its
own (and masking a "no backup" alert after a failover/role change).

**Fix**

- Filter `AND server_name = @@SERVERNAME` in the `backupset` subquery: only
  backups taken by the instance itself count. No new grant.
- The test regex pins the filter; mutation test: removing it, 5 tests fail.
- README updated.

**Validation (2026-10-14 15:58, `0.4.2-homolog`)**

- [x] On 02, the `backups` series of `LogShipDB` and `MirrorDB` disappeared
  (there is no local backup).
- [x] On 01, nothing changed: `LogShipDB` full 15:43:15 / log 15:58:00 and
  `MirrorDB` full / log are still there.

---

## HML-013 — Whole scrape fails (`up 0`) when SQL is slow: the example's `scrape_timeout` = the exporter's timeout

- **Environment:** local Prometheus (the repository's `prometheus.yml`), EC2 02, during HML-014
- **Found on:** 2026-10-14
- **Severity:** low (only at moments when SQL is slow; loses the whole scrape instead of part of it)
- **Status:** resolved (v0.4.0)

**Observed**

On 02, two scrapes failed (16:12:30 and 16:15:15) with `up 0`, 0 samples and
a duration of exactly 5.0 s, while the exporter kept running (same
`process_start_time_seconds`). The first one coincides with SQL coming back
from the CU.

**Cause**

The example `prometheus.yml` has `global.scrape_interval: 5s`; without an
explicit `scrape_timeout`, Prometheus uses 5 s for the exporter jobs too
(confirmed in `/api/v1/status/config`). The exporter uses 5 s
(`ConnTimeout`) as the deadline of the ping + collectors: if SQL is slow, the
answer arrives after 5 s and Prometheus discards everything.

**Fix**

- Example `prometheus.yml`: `scrape_timeout: 10s` in the `sqlserver_exporter`
  job (also applied to the local homologation job).
- README: guidance to keep `scrape_timeout` above 5 s.
- Possible future improvement: the exporter bounding its own deadline by the
  `X-Prometheus-Scrape-Timeout-Seconds` header, returning partial data instead
  of exceeding the Prometheus timeout.

---

## HML-014 — SQL Server 2019 RTM → CU32 + KB5128494 update (validation)

- **Environment:** EC2 01 + 02, `0.4.2-homolog`; patch on both in parallel, no failover
- **Date:** 2026-10-14 16:06–16:22
- **Status:** resolved

**Timeline (Prometheus)**

| Time | 01 | 02 |
|---|---|---|
| 16:10 | `sqlserver_up 0` (CU32) | `sqlserver_up 0` (CU32) |
| 16:12:45–16:13 | `sqlserver_up 1`, version **15.0.4430.1** | `sqlserver_up 1`, version **15.0.4430.1** |
| 16:12:30 / 16:15:15 | — | scrape failed on timeout (HML-013) |
| 16:16:31 | reboot (`os_boot_time_seconds`), `up 0` | reboot, `up 0` |
| 16:18:42 | exporter starts (delayed start) | exporter starts |
| 16:19:30–16:21:45 | `sqlserver_up 0` (KB5128494) | `sqlserver_up 0` (KB5128494) |
| 16:21:45 | version **15.0.4500.2** | version **15.0.4500.2** |

- [x] `sqlserver_product_version` recorded both version changes on both
  instances (one series per version, with no overlap).
- [x] The exporter survived the SQL stops (no restart) and came back on its
  own after the reboot; `sqlserver_instance` resolved correctly (SQL was up
  at start, at 16:18:42).
- [x] After the patch: no collector with `success 0`; AG `MSSQL_AG` healthy
  (sync 2, health 2); `MirrorDB` with 01 as principal, state 4.
- [x] Log shipping writes the monitor again (KB4537869 fixed):
  `primary_last_backup` 16:22:00, `secondary_last_copied`/`last_restored`
  16:18:07, latency 240 s.

---

## HML-015 — `blocking` fails sporadically on the AG secondary replica (error 976)

- **Environment:** EC2 02 (secondary of `MSSQL_AG`, `ALLOW_CONNECTIONS = ALL`), `0.4.2-homolog`
- **Found on:** 2026-10-14, during HML-011
- **Severity:** low (≈1.4% of 02's scrapes with `success 0` on `blocking`; 01 never failed)
- **Status:** resolved (v0.4.0)

**Observed**

`sqlserver_scrape_collector_success{collector="blocking"} 0` on 02 in
isolated scrapes (15:22, 15:37, 15:39, 16:05, 16:38). Event Log:

```
collector=blocking error="mssql: The target database, 'AlwaysOnDB', is participating
in an availability group and is currently not accessible for queries. ..."
```

The message without a prefix comes from `rows.Err()` of
`SELECT status, COUNT(*) FROM sys.dm_exec_requests GROUP BY status`: the DMV
is server-scoped and the connection is on `master` (`database=master` in the
DSN), but SQL Server aborts the read halfway. Hypothesis: a per-request
visibility check that touches `AlwaysOnDB` when it is momentarily unreadable
on the secondary. The same error is known in session snapshot jobs in DR
environments (user report). A loop of 300 runs of the query as
`usr_exporter` (16:40–16:45) did not reproduce it; the exact cause remained
open.

**Fix (decision: waiting_tasks + retry)**

- `sqlserver_blocked_sessions` / `_max_wait_ms` now come from
  `sys.dm_os_waiting_tasks` (no database context):
  `COUNT(DISTINCT session_id)`, `MAX(wait_duration_ms)`,
  `blocking_session_id > 0 AND blocking_session_id <> session_id`.
- `sqlserver_active_requests{status}` stays on `dm_exec_requests`, with **one**
  retry only for error 976 (detected through the driver's
  `SQLErrorNumber()`). The counts are accumulated and emitted only at the end,
  so the failed attempt does not produce duplicate series (HML-004). If it
  fails again, `success 0` as before.
- Tests: successful retry, 976 twice, another error without retry. Mutation
  test: without the retry, 2 tests fail; with a retry for any error, 3.

**Validation (2026-10-14, `0.4.3-homolog`)**

- [x] Blocking caused on 01 (`##blk`, 16:56–17:01): `blocked_sessions`
  0 → 1 → 0 and `blocked_max_wait_ms` rising ~30 s every 30 s (7,964 →
  277,967 ms) until the `ROLLBACK`.
- [x] 02 without `success 0` on `blocking` from 17:03 to 17:49 (184 scrapes;
  before the fix there would be ~2–3 failures in the period). Window closed by
  the user's decision.
- Note: after the upgrade, 02's service stayed stopped (16:55–17:02) until it
  was started by hand — `service start` did not bring the process up; no sign
  of an exporter crash.

---

## HML-016 — `denisenkom/go-mssqldb` driver archived: migrate to `microsoft/go-mssqldb`

- **Environment:** EC2 01 + 02 (SQL Server 2019 CU32), before the upgrade to 2022/2025
- **Found on:** 2026-10-14, planning the version upgrade
- **Severity:** medium (unmaintained driver; no TDS 8.0 / `encrypt=strict`)
- **Status:** resolved (v0.5.0)

**Context**

`denisenkom/go-mssqldb` v0.12.3 is archived; `microsoft/go-mssqldb` is the
maintained fork and the official Go driver. Before homologating SQL Server
2022 and 2025, the exporter switches to the new driver.

**Change**

- Import and `go.mod`: `github.com/microsoft/go-mssqldb` v1.11.2 (Go 1.25+,
  the project is already on 1.25.0).
- Same driver name (`sqlserver`) and same DSN formats; same encryption
  default (login packet only). New: `encrypt=strict` (TDS 8.0).
- Windows binary +0.5 MB (15.06 → 15.58 MB); nothing from the Azure SDK goes
  into the binary.

**Validation**

- [x] `0.5.0-homolog` on both EC2s (2019 CU32, 2026-10-14 17:55):
  `sqlserver_up 1`, the 16 collectors with `success 1`, the same metric names
  as `0.4.3-homolog` (= v0.4.0): 01 156 names / 3,779 series (identical); 02
  153 names / 3,802 series (+5: the SQL 2022 ISO mounted on `E:` (CDFS)
  appeared in `os_filesystem_*`). Average scrape duration 0.29 s → 0.31 s
  (within noise).
- [x] SQL Server 2022 RTM and 2025 RTM (HML-017): the 16 collectors with
  `success 1` on both versions, no metric disappeared.
- [x] `encrypt=strict` (TDS 8.0) on 01 (2025), 2026-10-14 19:45:
  - Without a certificate configured on SQL Server the TDS 8 handshake is
    refused: `openssl s_client -alpn tds/8.0` gets EOF with 0 bytes (TLS 1.2
    and 1.3) and the exporter stays at `sqlserver_up 0` (process up,
    `/metrics` answering).
  - The driver **ignores `trustservercertificate=true` in strict mode**
    (`msdsn/conn_str.go`, `parseTLS`: `if encrypt == "strict" { trustServerCert = false }`,
    the same behavior as Microsoft.Data.SqlClient): still
    `x509: certificate signed by unknown authority`.
  - With a self-signed `CN=AWSDBSERVER01` certificate configured on the
    instance and `certificate=C:/.../sql01.pem` in the DSN: `sqlserver_up 1`,
    16/16 collectors, scrape 0.30 s. In SQL, the pool's 5 connections with
    `encrypt_option TRUE` and `protocol_version 134217728` (0x08000000 =
    TDS 8.0).
  - Tested from the Mac with a fake login (`Login failed` = TLS ok):
    `certificate=<pem>` fails connecting by IP (no IP SAN) and works with
    `hostnameincertificate=AWSDBSERVER01`; `servercertificate=<pem>` (byte
    comparison) works by IP.
  - Documented in the README ("Encrypted connections").

---

## HML-017 — Rolling upgrade SQL Server 2019 → 2022 (02) → 2025 (01, 02) with AG, mirroring and log shipping (validation)

- **Environment:** EC2 01 + 02, exporter `0.5.0-homolog` (`microsoft/go-mssqldb` driver)
- **Found on:** 2026-10-14
- **Severity:** — (validation)
- **Status:** in progress

**Plan**

Rolling upgrade from 2019 CU32 (`15.0.4500.2`): 02 (AG secondary, mirroring
mirror, log shipping secondary) to 2022; AG and mirroring failover to 02; 01
straight to 2025 (a decision to save time, without bringing both to 2022);
then 02 to 2025.

**Step 1 — upgrade of 02 to 2022 RTM (`16.0.1000.6`), 2026-10-14**

| Time | Event (seen in Prometheus) |
|---------|------------------------------|
| 18:03:37 | Windows reboot (`os_boot_time_seconds`); 01 sees `mirroring_state` 1 (disconnected) |
| 18:05:50 | the exporter service starts on its own after boot (`process_start_time_seconds`) |
| 18:08–18:10, 18:15, 18:19–18:22 | SQL restarts during setup: 02 `sqlserver_up 0`, 01 mirroring disconnected |
| 18:22:06 | `product_version` `16.0.1000.6`, `sqlserver_up 1` |
| 18:22:13 | mirroring back to 4 (synchronized) |

- The exporter process **did not need to be restarted**: it reconnected on
  its own at every SQL restart; `sqlserver_instance` stayed correct.
- Only error: `volumes` with `success 0` at 18:22:02, the first scrape after
  SQL came up (databases in recovery). It took 4.98 s (5 s query timeout);
  the whole scrape took 5.17 s, below `scrape_timeout: 10s` (HML-013), so
  Prometheus got the partial data. The next scrape was already `success 1`.
- `AlwaysOnDB` on 02 went from `ONLINE` to `RECOVERING`: a secondary on a
  higher version is not readable until the end of the upgrade. This is the
  error 976 scenario; `blocking` did not fail even once after the upgrade
  (HML-015 fix confirmed).
- AG synchronized/healthy, queues at zero; mirroring synchronized; log
  shipping 2019 (01) → 2022 (02) restoring normally.
- Differences in 02's metrics (3,802 → 4,532 series), no metric disappeared:
  - wait types 1,027 → 1,266 (+239 new in 2022) in `sqlserver_wait_time_ms`,
    `sqlserver_max_wait_time_ms` and `sqlserver_waiting_tasks`;
  - `sqlserver_database_log_used_percent` gained `model_msdb` and
    `model_replicatedmaster` (internal model databases of 2022 for Contained
    AG, only in the performance counters, like `mssqlsystemresource`). Low
    cardinality, no action;
  - `sqlserver_memory_clerk_used_kb` +1 clerk.
- Pre-existing, not related to the upgrade: `LSAlert_AWSDBSERVER02` with
  `last_run_status 0` since before (restore threshold still at 2 min from the
  HML-011 test).

**Step 2 — AG and mirroring failover to 02 (2022), 18:39:33**

- AG: `replica_role` swapped (02 primary); `AlwaysOnDB` was upgraded on 02
  (`ONLINE`) and stayed in `RECOVERING` on 01 (Not Synchronizing). Both
  instances report 01's replica with `is_suspended 1`, `suspend_reason 7`
  (revalidation), `synchronization_state 0` and `synchronization_health 0`;
  the README alert
  (`max by (ag_name, db_name) (sqlserver_hadr_database_is_suspended) == 1`)
  would fire.
- Mirroring: `MirrorDB` stayed `RECOVERING` on 02 (new principal) and
  `RESTORING` on 01 — the principal on 2022 does not leave recovery with the
  mirror on 2019. The state series disappeared: see HML-018.
- No collector with `success 0`.

**Step 3 — upgrade of 01 to 2025 RTM (`17.0.1000.7`), 18:52–18:56**

- 18:47:06 Windows reboot; 18:49:18 the exporter service starts on its own;
  SQL restarts 18:52–18:53 and 18:54–18:55; 18:55:50 new version,
  `sqlserver_up 1`, 16/16 collectors. No collector failed.
- 01's metrics (3,777 → 5,143 series): wait types 1,027 → 1,475 (+209 over
  2022), +15 memory clerks, the same `model_msdb`/`model_replicatedmaster`.
- AG: 01's `AlwaysOnDB` stayed suspended until the manual `SET HADR RESUME`
  (18:59:00), then synchronized; still `RECOVERING` on 01 (a secondary on a
  higher version than the primary is not readable).
- Log shipping broke as expected: at 19:00:15 02's `LSRestore` (2022) with
  `last_run_status 0`, and the time since the last restore, which oscillated
  0–180 s (3-min cycle), started growing without end, above the threshold.
- Mirroring: 02's `MirrorDB` got stuck in recovery. 02's ERRORLOG: 18:39:56
  MIRROR → PRINCIPAL switch, "Converting database 'MirrorDB' from version 904
  to the current version 957" and a stop at the first step (904 → 905): the
  step generates log that must be confirmed by the mirror in synchronous
  mode, and the mirror on 2019 did not apply log from 2022. `DB MIRROR`
  threads idle on `DBMIRROR_EVENTS_QUEUE`; 01 (sysadmin) saw `MIRROR` /
  `SUSPENDED`. `SET PARTNER RESUME` on 01 took the session to
  `SYNCHRONIZING`, but 02 only got unstuck with the restart of step 4. For a
  real rolling upgrade: do not fail mirroring over to the upgraded partner
  while the other one is on a lower version, or suspend/remove the session
  first.

**Step 4 — upgrade of 02 to 2025, failovers and failbacks**

| Time | Event |
|---------|--------|
| 19:08:21 | AG failover to 01 (2025 as primary); 02's `AlwaysOnDB` (2022) suspended |
| 19:17:51–19:21:51 | SQL restarts on 02 during setup (no reboot, no exporter restart) |
| 19:18:41 | `blocking` with error 983 on 02 (HML-020) |
| 19:18:51 | 02's `MirrorDB` in `RECOVERY_PENDING` |
| 19:21:51 | 02 on `17.0.1000.7`; `MirrorDB` `ONLINE` as principal on 02, state series back (`role 1`, `state 4`) |
| 19:23:06 | AG failover to 02; 01's `volumes` with a timeout (HML-019) |
| 19:24:21 | `LSRestore` back to `last_run_status 1` (log shipping secondary on 2025) |
| 19:24:36 | 01's `AlwaysOnDB` suspended with `suspend_reason 1` (from the partner, typical of a forced failover) |
| 19:24:51 | mirroring failback to 01: principal 2025, synchronized |
| 19:28:10 | `SET HADR RESUME` on 01, `is_suspended 0` on both |

- Both on 2025: 01 5,129 / 02 5,136 series, 1,475 wait types each, scrape
  0.29 s.

**Final coverage**

| Role | 2019 CU32 | 2022 RTM | 2025 RTM |
|-------|-----------|----------|----------|
| AG primary | ✓ | ✓ | ✓ |
| AG secondary | ✓ | ✓ | ✓ |
| Mirroring principal (`ONLINE`) | ✓ | — (stuck in recovery) | ✓ |
| Mirroring mirror | ✓ | ✓ | ✓ |
| Log shipping primary | ✓ | — | ✓ |
| Log shipping secondary | ✓ | ✓ | ✓ |

The exporter never needed to be restarted because of an upgrade and
reconnected on its own at every SQL restart; the service starts on its own
after a reboot. Collector failures seen: only those of HML-019 and HML-020,
one collection each.

---

## HML-018 — Mirroring state disappears (no series) when the mirrored database is not accessible

- **Environment:** EC2 01 + 02, during HML-017 (principal 2022, mirror 2019)
- **Found on:** 2026-10-14 18:39
- **Severity:** medium (an alert on `sqlserver_mirroring_state` does not fire)
- **Status:** resolved (documentation, v0.5.0)

**Symptom**

After the mirroring failover to 02 (2022), `sqlserver_mirroring_info`,
`_role` and `_state` stopped existing on both instances. Only the performance
counters (`_log_send_queue_kb`, `_redo_queue_kb`, `*_total`) remained.

**Cause**

With `MirrorDB` in `RECOVERING` on 02, `sys.database_mirroring` returns
`mirroring_guid`, `mirroring_role_desc` and `mirroring_state_desc` NULL
**even for sysadmin** (validated on 02 as sysadmin and with
`EXECUTE AS LOGIN = 'usr_exporter'`: same result). The collector filters
`mirroring_guid IS NOT NULL`, so the row disappears. On 01 (mirror) the login
without sysadmin already got NULL (HML-011 limitation). A SQL Server
limitation, not a permission one: there is no other source for the session
state.

**Impact**

An alert `sqlserver_mirroring_state != 4` does not fire: the series
disappears instead of changing value.

**Proposal**

Document in the README an alert on absence, using the performance counters
(which remain present for every mirrored database) as reference, and note
that `sqlserver_database_state{state="RECOVERING"}` also covers the case.

**Additional evidence (step 3 of HML-017)**

With the session `SUSPENDED`, only a sysadmin on the mirror (01) saw the
state; for `usr_exporter`, 02 returned no row (database in recovery) and 01
returned NULL (mirror without sysadmin). In the most serious failure, the
state disappeared on both ends. When `MirrorDB` went back to `ONLINE` on 02
(19:21:51), the series came back on their own.

**Resolution**

Documented in the README (Database mirroring section), with the expression
tested in Prometheus (it returned `MirrorDB` during the whole period without
state):
`count by (db_name) (sqlserver_mirroring_log_send_queue_kb) unless on (db_name) count by (db_name) (sqlserver_mirroring_state)`.
Do not emit an "unknown" state in the collector: the expression covers the
case without inventing a value SQL Server does not report.

---

## HML-019 — `volumes` exceeds the query timeout when a database is changing state

- **Environment:** EC2 01 + 02, during HML-017
- **Found on:** 2026-10-14
- **Severity:** low (one partial collection per event)
- **Status:** resolved (v0.13.4)

**Observed**

`volumes` with `success 0` twice, always with a duration of ≈5 s (5 s query
timeout):

- 18:22:02 on 02, the first collection after SQL came up on 2022 (databases in recovery);
- 19:23:06 on 01, in the AG failover (01's `AlwaysOnDB` in transition).

`sys.dm_os_volume_stats(database_id, file_id)` is applied to every file of
`sys.master_files` and waits while some database is changing state.
`scrape_timeout: 10s` (HML-013) absorbed both cases: the scrape took ≈5.2 s,
Prometheus got the partial data and the next collection already had
`success 1`. Databases stable in `RECOVERING` (unreadable secondary) or
`RESTORING` did not cause failures.

**Options**

Filter the query by `sys.databases.state IN (0, 1)` (ONLINE/RESTORING),
which narrows the window but does not remove the race if the transition
starts during the query. Small gain; decide later.

**Fix (2026-10-27)** — branch `fix/quick-fixes`:

- `volumes` only considers files of `ONLINE` (0) or `RESTORING` (1)
  databases, to keep the volumes of AG and log shipping secondaries. A volume
  that hosts only databases in another state (`RECOVERING`, `SUSPECT`,
  `OFFLINE`…) no longer appears while the database is in that state.
- Validated on the `mssql-2025-agent` container (Linux): with `QSOffTest`
  `OFFLINE`, the query reads 17 of the 19 files and the volume keeps the same
  bytes; the database went back to `ONLINE` afterwards.
- The race in the middle of a transition is still possible; a
  `scrape_timeout` larger than the query timeout (HML-013) keeps absorbing
  that case.

---

## HML-020 — `blocking` fails with error 983 right after SQL Server starts (AG replica in `RESOLVING`)

- **Environment:** EC2 02, during HML-017 (restart of the 2025 setup)
- **Found on:** 2026-10-14 19:18:41
- **Severity:** low (one collection)
- **Status:** resolved (v0.5.0)

**Observed**

```
collector=blocking error="mssql: Unable to access availability database 'AlwaysOnDB'
because the database replica is not in the PRIMARY or SECONDARY role. ... (983)"
```

Same query as HML-015 (`sys.dm_exec_requests`), with the replica still
resolving its role right after SQL came up. The HML-015 retry only
recognized 976.

**Fix**

`isAGDatabaseNotAccessible` recognizes 976 and 983; the retry test runs for
both codes. Mutation (removing 983 from the `case`) breaks
`TestBlockingScrapeRequestsRetriesOnceOnAGNotAccessible/983`. `RESOLVING`
lasts a few seconds, so the immediate retry does not always get past it; it
reduces noise without changing the essential behavior.

**Validation**

By unit test, without reproduction on the EC2s (user's decision): the
`RESOLVING` window is seconds long and the retry path is the same one
already validated live in HML-015. Build `0.5.1-homolog` published on S3.

---

## HML-021 — `sqlserver_sessions_by_login` with load from several logins (validation)

- **Environment:** EC2 01 (SQL Server 2025 RTM), `0.6.0-homolog`; AG,
  mirroring and log shipping torn down, 02 shut down
- **Found on:** 2026-10-15
- **Severity:** — (validation of item 1 of the roadmap's Phase 6)
- **Status:** resolved

**Scenario**

`hack/loadgen` running from the Mac against 01: 10 logins `loadgen_01..10`
(`setup.sql`), 2 workers each, a mix of reads, scans, updates, long waits and
deliberate blocking.

**Result**

- `sqlserver_sessions_by_login` = 2 per `loadgen_NN`, plus
  `AWSDBSERVER01\usr_dbsvc` (SSMS), `usr_exporter` (the exporter's pool, 5)
  and `NT SERVICE\SQLTELEMETRY`; the sum matches `sqlserver_user_connections`.
- A first test showed 4 sessions per login: there were two loadgen processes
  at the same time (one started from the UI). Reproduced on a local SQL
  Server 2025 in a container with one process: exactly 2 sessions, always on
  the same 2 TCP ports.
- Without RCSI on `LoadTestDB`, the `status_report` scan waited for the
  `blocker` and 38 sessions were blocked (`blocking` reported it correctly,
  maximum wait 54.7 s); with RCSI and rebalanced weights: ~42 batch
  requests/s, blocking only in bursts and the tempdb version store in use.
- 16/16 collectors with `success 1` during the whole load; maximum scrape
  1.06 s.

---

## HML-022 — `active_queries` collector: long requests and open transactions (validation)

- **Environment:** EC2 01 (SQL Server 2025 RTM), `0.7.0-homolog`
- **Found on:** 2026-10-15
- **Severity:** — (validation of item 2 of the roadmap's Phase 6)
- **Status:** resolved

**Scenario**

`hack/loadgen` (10 logins × 2 workers, from 16:32) plus two tests in SSMS
around 16:32: a `BEGIN TRAN` + `UPDATE` without `COMMIT` (idle session with
an open transaction) and a `WAITFOR DELAY '00:02:00'` (long request).
`ROLLBACK` at 16:44.

**Result**

| Metric | Observed |
|---------|-----------|
| `sqlserver_request_max_elapsed_seconds{command="WAITFOR"}` | 27 → 117 s, +15 s per scrape; after the SSMS `WAITFOR` ended, only the loadgen's `long_wait` (up to ~30 s) |
| `sqlserver_requests_long_running{over_seconds="60"}` | 0 → 1 at 16:33:20 (the `WAITFOR` went from 57 to 72 s), back to 0 at 16:34:20 |
| `sqlserver_requests_long_running{over_seconds="10"}` | 1–2 (the SSMS `WAITFOR` + the loadgen's `long_wait`/`blocker`) |
| `sqlserver_sleeping_sessions_with_open_transaction` | 1 during the whole test, 0 at 16:45:11 |
| `sqlserver_oldest_open_transaction_seconds` | 32 → 767 s in a straight line; 17 s at 16:45:11 (only the loadgen's `blocker`) |
| `sqlserver_request_max_elapsed_seconds` of `SELECT`/`UPDATE` | 0 s (the loadgen's queries are fast) |

- `active_queries` always with `success 1`, ≈27 ms per scrape.

---

## HML-023 — `top_queries` collector with obfuscated text (validation)

- **Environment:** EC2 01 (SQL Server 2025 RTM), `0.8.0-homolog`, with
  `collectors: {top_queries: true}` and `top_queries: {include_text: true}`
- **Found on:** 2026-10-15
- **Severity:** — (validation of item 3 of the roadmap's Phase 6)
- **Status:** resolved

**Before the EC2**

Query and full exporter validated against a local SQL Server 2025 in a
container: `query_hash` in hexadecimal, `db_name` from the plan attributes,
text per statement (not the whole batch), obfuscated literals and truncation.

**Result (`hack/loadgen` load from 17:01)**

- `top_queries` with `success 1` in ≈19 ms per scrape; 50 series per metric
  (= the default `limit`) and 50 `sqlserver_top_query_info`.
- Top by CPU: the loadgen's `status_report` (table scan; 450 s of CPU in
  7,691 executions, ≈58 ms each) by far, followed by the exporter's own
  queries in `master` (documented: on an idle server the top is the
  exporter).
- Text: `TOP (50)` → `TOP (?)`, strings of the exporter's queries → `?`,
  `@p1` parameters kept, cut at 120 characters with `...`.
- Top 5 by cumulative CPU (hash, CPU, executions) checked in SSMS with
  `sys.dm_exec_query_stats` grouped by `query_hash`: same hashes, same order.

## HML-024 — `index_fragmentation` collector in the background and `CONNECT ANY DATABASE` (validation)

- **Environment:** EC2 01 (SQL Server 2025 RTM), `0.9.0-homolog`, with
  `collectors: {index_fragmentation: true}` and
  `index_fragmentation: {interval: 2m, min_page_count: 100}`
- **Found on:** 2026-10-15
- **Severity:** — (validation of item 4 of the roadmap's Phase 6)
- **Status:** resolved

**Before the EC2**

Validated against a local SQL Server 2025 in a container: a table with a
`NEWID()` key and 20,000 row-by-row inserts took `PK_Guids` from 0.05% to
99.03% in the next cycle (1-min interval); a database named `Odd]Name` read
without error (name in brackets); scrape in microseconds.

**Result**

- First start (17:31) without the collector: the `index_fragmentation:`
  section with the options was in the file, but `index_fragmentation: true`
  inside `collectors:` was missing — the options section does not enable the
  collector (same rule as `top_queries`). `config.example.yml` now says so
  explicitly.
- 17:37, **without a new grant**: `last_collection_success 1`,
  `last_collection_databases 0` in 0.13 s — `usr_exporter` has no user in the
  databases and the `HAS_DBACCESS` filter skips them without error. Since
  `success 1` with zero databases hides that nothing is collected, the README
  documents the `last_collection_databases == 0` alert.
- After `GRANT CONNECT ANY DATABASE TO usr_exporter` (≈17:38), the next cycle
  read **3 databases** in 0.36 s: `LoadTestDB.dbo.Guids.PK_Guids` at
  **98.9%** (1,614 pages, created by the random inserts script), `Orders` at
  0.02% / 0.86%; `MirrorDB` with indexes of 277,778 and 286,838 pages
  (~4.4 GB) and fragmentation < 0.3%. `LIMITED` mode reads only the level
  above the leaf: cheap even on GB-sized indexes.
- The collector's scrape reports duration 0 (it serves the cache);
  `scrape_collector_success 1` the whole time.

**RDS phase (2026-10-18):** the RDS master user **cannot** grant it
(`Grantor does not have GRANT permission`). Alternative documented in the
README and validated in HML-026: create a user for the login in each
database to be read. `rdsadmin` is already accessible to any login, so even
with nothing the collector reads 1 database on RDS (no indexes above 100
pages).

---

## HML-025 — Exporter as a Linux service (systemd) on the RDS phase bastion (validation)

- **Environment:** Linux bastion of the RDS phase (the exporter runs on it
  and reaches RDS through the VPC; the local Prometheus reads the exporter
  through an SSH tunnel)
- **Found on:** 2026-10-15
- **Severity:** — (until now Linux had no documented way to run as a
  service; the `service` subcommand is Windows-only)
- **Status:** resolved (v0.9.1)

**Delivery**

- Unit `contrib/systemd/mssql-exporter.service`, included in the release
  archives from v0.9.1 together with `config.example.yml`: system user
  `mssql-exporter` without a shell, config in `/etc/mssql-exporter`
  (`0640 root:mssql-exporter`), `Restart=on-failure`, systemd sandbox.
- README: "Running as a Linux service (systemd)" section with the
  installation steps from the release.

**Local validation (2026-10-15, Debian 12 container with systemd, snapshot build)**

- [x] README steps from the tarball: checksum OK, binary, user, config and
  unit installed with the expected permissions.
- [x] `systemctl enable --now`: service active as `mssql-exporter`,
  `sqlserver_up 1` and the 16 default collectors + `index_fragmentation` with
  `success 1` against the local SQL Server 2025.
- [x] With the sandbox, the OS collectors (CPU, memory, disk, network, boot)
  work; `os_filesystem_*` and `os_process_*` do not appear in the container
  even without the sandbox (no physical partition nor SQL Server process) —
  check `os_filesystem_*` on a VM.
- [x] `kill -9` → restarted in 5 s (`NRestarts=1`); `systemctl stop` →
  `shutting down` in the journal, status 0; container restart → service comes
  back on its own (`enabled`).
- [x] `systemd-analyze security`: exposure 3.1 ("OK").

**Validation on the bastion (2026-10-18, Ubuntu 26.04 `amd64`, release v0.9.1)**

- [x] README steps followed on the bastion, with one difference: the
  repository is still private, so the `curl` of the GitHub release does not
  work; the tarball and the `.sha256` went to S3 and were downloaded with
  `aws s3 cp`. Once public, the README steps apply as they are.
- [x] `systemctl enable --now` at 23:14:49 UTC: active at the first try, no
  error in the journal; `web.listen_address: "127.0.0.1:9187"` and SSH tunnel
  (`-L 19187:127.0.0.1:9187`) to the local Prometheus, target `up` at 23:15.
- [x] Upgrades to `0.9.2-homolog` and `0.9.2` replacing only the binary
  (`install` + `systemctl restart`), without touching config or unit.
- `cd /etc/mssql-exporter` as `ubuntu` gives *Permission denied* (folder
  `0750 root:mssql-exporter`, as planned); edit with `sudo`.

---

## HML-026 — `database_log_used_percent` with hidden databases and, on RDS, without the system databases

- **Environment:** RDS for SQL Server 2025 Standard (`17.0.5005.3`, Single-AZ),
  exporter `0.9.1` on the bastion; reproduced on the local SQL Server 2025
- **Found on:** 2026-10-18
- **Severity:** low
- **Status:** resolved (v0.9.2)

**Observed**

On RDS, `sqlserver_database_log_used_percent` came for `model_msdb`,
`model_replicatedmaster`, `mssqlsystemresource`, `tempdb` and the user
database `StressDB`, but not for `master`, `model`, `msdb` and `rdsadmin`. No
other metric knows the first three: they are not in `sys.databases`.

**Cause**

The metric passes through every instance of the `Percent Log Used` counter
(`sys.dm_os_performance_counters`). The counter has instances for hidden
databases — `mssqlsystemresource` on any version and, on SQL Server 2025,
`model_msdb`/`model_replicatedmaster` (the local container shows the same) —
and, on RDS only, has no instance for the system databases. It is not a
permission: `usr_exporter` and `usr_dba` see the same list.

Alternative sources tested on RDS as `usr_exporter`:

| Source | Databases | Missing |
|---|---|---|
| `Percent Log Used` counter | StressDB, tempdb + 3 hidden | master, model, msdb, rdsadmin |
| `DBCC SQLPERF(LOGSPACE)` | master, tempdb, msdb, rdsadmin + 2 hidden | **model and StressDB** |
| `sys.dm_db_log_stats` | the 6 of `sys.databases` | — |

`dm_db_log_stats` is complete, but `active_log_size_mb` measures something
else (active VLFs): master 10.6% against 44% from DBCC, msdb 0.5% against
10.9%. Mixing the sources would make the same metric mean different things
per environment.

**Fix** — branch `fix/log-used-percent-ghost-databases`

- Counter joined with `sys.databases` (`COLLATE DATABASE_DEFAULT` on the name,
  so as not to depend on the DMV's collation): the series of the hidden
  databases go away in any environment.
- The README documents that on RDS the system databases do not have the
  metric (the user ones do).
- ROADMAP (Later): metrics of their own from `sys.dm_db_log_stats` (VLFs,
  reason the log does not truncate, log since the last backup), which work on
  RDS for every database.

Validated on the local SQL Server 2025: before, 9 series (with `model_msdb`,
`model_replicatedmaster` and `mssqlsystemresource`), after, only the 6
databases of `sys.databases`.

**Regression check on other platforms (2026-10-18)**

- EC2 01 (2025) history in Prometheus: `log_used_percent` had the 10 real
  databases **+** `model_msdb`, `model_replicatedmaster` and
  `mssqlsystemresource`; with the join exactly the 10 remain (`RESTORING`
  databases of mirroring/log shipping stay, they are in `sys.databases`). On
  2019/2022 only `mssqlsystemresource` goes away.
- Hard names, local SQL Server 2025 (Linux): 11 test databases — `]`, `/`,
  `\`, `#`, `%`, `'`, `.`, `(`/`)`, space, accents and a 128-character name.
  Counter instance identical to `sys.databases` in all of them; the exporter
  with the fix emitted the 11 series. Not checked on Windows (the DMV reads
  SQL Server's own counters, not PerfMon's, so no character replacement is
  expected; the EC2 databases had no special names).

**Validation on RDS (2026-10-18, `0.9.2-homolog` on the bastion, restarted 23:32 UTC)**

- `log_used_percent` only for `StressDB` (17) and `tempdb` (9); the three
  hidden databases went away. `databases` with `success 1` in 22 ms; the 18
  collectors with `success 1`.
- Alternative to `CONNECT ANY DATABASE` (HML-024): with
  `CREATE USER usr_exporter FOR LOGIN usr_exporter` in `StressDB`, the next
  `index_fragmentation` cycle read 2 databases (`rdsadmin` + `StressDB`) in
  0.46 s: `PK_Orders` with 277,778 pages (0.01%) and `cx_Numbers` (0%).

---

## HML-027 — RDS Single-AZ baseline: permissions, collectors, load and TLS (validation)

- **Environment:** RDS for SQL Server 2025 Standard (`17.0.5005.3`, RTM-CU9),
  `db.t3.xlarge` Single-AZ, `mssql-exporter-hml`, UTC-3 time zone, private
  subnets; exporter `0.9.2` on the bastion (HML-025)
- **Found on:** 2026-10-18
- **Severity:** — (validation)
- **Status:** resolved

**Permissions**

The same grants as on EC2, given by the master user (`usr_dba`):
`VIEW SERVER STATE`, `VIEW ANY DEFINITION` and `SELECT` on
`msdb.dbo.sysjobs`/`sysjobhistory` (HML-002 repeated on RDS: OK).
`CONNECT ANY DATABASE` **cannot** be granted (HML-024); a user per database
instead.

**Collectors (default + `top_queries` and `index_fragmentation`)**

- The 18 collectors with `success 1` from the first collection; per-collector
  medians the same as on EC2 (`perf_counters` ≈ 0.13 s, the others < 0.09 s).
- `server_info`: Standard Edition, 4 CPUs, 16 GB, `is_hadr_enabled 1` even on
  Single-AZ.
- `volumes`: RDS's `D:\` (21.4 GB). `io_file_stats`: 13 files.
- `backups`: FULL of `model`, `msdb` and `rdsadmin` taken by RDS when the
  instance was created (backup retention 0).
- `agent_jobs`: `syspolicy_purge_history`.
- `sqlserver_instance` = `EC2AMAZ-UU69199` (the name of the Windows host
  behind RDS), not the instance identifier — review on the Multi-AZ failover.
- `rdsadmin` is accessible to any login: it shows up in `databases`,
  `index_fragmentation` etc.
- `log_used_percent`: see HML-026.

**Load (`hack/loadgen`, 10 logins × 2 workers, 23:52–00:12 UTC, through the tunnel)**

- 55,020 statements, no error in the loadgen; ≈33 batch requests/s; RDS CPU
  24% on average (peak 29%).
- `sessions_by_login`: 2 sessions per `loadgen_NN`, plus `usr_exporter`,
  `usr_dba` and the RDS service accounts (`NT AUTHORITY\SYSTEM`,
  `NETWORK SERVICE`).
- `active_queries`: up to 14 requests above 10 s, `UPDATE` with up to 18 s
  and up to 13 blocked sessions — all in the queue of the `blocker` (holds
  the hot row for 20 s) behind `hot_update`; checked in
  `sys.dm_exec_requests` (`LCK_M_X` chain up to the `blocker` session).
  Oldest open transaction up to 36 s.
- `top_queries`: 1st place the loadgen's `status_report` (223 s of CPU in
  1,405 executions, ≈159 ms each; on EC2 ≈58 ms), text with obfuscated
  literals; next, the exporter's own queries in `master`.
- `index_fragmentation` with 3 databases (`rdsadmin`, `StressDB`,
  `LoadTestDB`, the latter with a user created for `usr_exporter`).
- No collector with `success 0` during the load.

**TLS (bastion → RDS, 23:58–23:59 UTC)**

| DSN | Result |
|---|---|
| no `encrypt` | connects (only the login encrypted) |
| `encrypt=true&certificate=global-bundle.pem` | OK, `encrypt_option` TRUE |
| `certificate=` with a CA that is not AWS's (`.pem`) | `sqlserver_up 0`, `x509: certificate signed by unknown authority` (the validation happens) |
| `certificate=/etc/ssl/certs/ca-certificates.crt` | the exporter exits at once: `certificate type .crt is not supported` (the driver only accepts `.pem`/`.der`); systemd restarts it every 5 s until fixed |
| `encrypt=strict&certificate=global-bundle.pem` | OK, `protocol_version` 0x08000000 (TDS 8.0), 18 collectors with `success 1` |

On RDS, `encrypt=strict` works without any server configuration (on EC2 a
certificate had to be installed on the instance). The bastion kept
`encrypt=strict`. README: TLS section with the RDS case and the file
extension restriction.

---

## HML-028 — Backups on RDS: snapshots and the log every 5 min appear; creation backups written to `NUL` count as FULL

- **Environment:** RDS for SQL Server 2025 SE (`mssql-exporter-hml`), UTC-3
  time zone, exporter `0.9.2`
- **Found on:** 2026-10-18/19
- **Severity:** medium (the `NUL` finding makes an instance without backups
  look like it has a FULL backup)
- **Status:** resolved (v0.9.3)

**1. Manual snapshot with retention 0 (2026-10-19, 01:16 UTC)**

`aws rds create-db-snapshot` recorded in `msdb.backupset` a FULL of each of
the 6 databases with `is_snapshot = 1`, `copy_only = 0` and a virtual device
(`device_type` 7), besides regular FULLs of `model`/`msdb` just before. The
collector started showing `backup_type="full"` at 01:16:47 UTC for all —
time zone converted correctly (HML-010 holds on RDS). The I/O suspension of
the Single-AZ snapshot did not affect any collector.

- `backup_last_duration_seconds` = 1 s (the snapshot is instantaneous).
- `backup_last_size_bytes` = database size (`StressDB` 2.3 GB), not what the
  snapshot takes on S3.

**2. Retention 0 → 1 (2026-10-19, `apply` at 13:57 UTC, no restart)**

RDS events: `Enabled automated backups` 13:57:25, snapshot finished
14:01:37, `Emergent Snapshot Request` (RDS forced `model`, `msdb` and
`rdsadmin` back to FULL and took another snapshot) finished 14:03:48.

- `databases`: recovery model of the user databases from `SIMPLE` to `FULL`
  (with retention 0 RDS keeps them in `SIMPLE`; `master` stays `SIMPLE`).
- `msdb.backupset`: FULL of `master` to local disk (`D:\rdsdbdata\BACKUP`),
  two snapshots (10:59 and 11:02 local), a real FULL of `StressDB` (41 s) and
  `LoadTestDB` to start the log chain and, then, **BACKUP LOG every 5 min**
  (virtual device) for all databases in FULL.
- Collector: `backup_type="full"` at 14:02:56 UTC (the last snapshot) and
  `backup_type="log"` at 14:07:01 and 14:12:02 UTC — the 5-min interval shows
  as expected. "No log backup for more than N minutes" alerts work on RDS
  (no log for `master`, in `SIMPLE`).

**3. Creation backups to `NUL`**

When creating the instance (retention 0) RDS runs `BACKUP DATABASE ... TO
DISK = 'nul'` of `model`, `msdb` and `rdsadmin` (only to start the chain; the
backup is discarded). The collector reports those records as the last FULL.
With retention at 0 they would be the instance's only backups forever: a
panel would show backups that do not exist and a "no FULL for more than 25 h"
alert would stay silent on the first day. The same applies outside RDS to
anyone who runs `BACKUP ... TO DISK = 'NUL'` (a common trick to cut the log
chain).

**Proposal**

Ignore in the `backups` collector the records whose `physical_device_name`
is `NUL` (case-insensitive). Snapshots (`is_snapshot = 1`) still count: RDS
restores from them.

**4. Native backup through the RDS procs (2026-10-19, ≈18:14 UTC)**

Option group with `SQLSERVER_BACKUP_RESTORE` and IAM role
`rds-backup-restore-role`; `msdb.dbo.rds_backup_database` to S3 (automated
backup retention at 0 at that moment — the 18:03 `apply` turned the
automated backups off and RDS put all databases back in `SIMPLE`).

| Task | Database | Type | S3 | `msdb` / collector |
|---|---|---|---|---|
| 1 | `LoadTestDB` | FULL | 56.6 MB | `full` 18:14:53 UTC, 1 s |
| 2 | `StressDB` | FULL | 2.1 GB | `full` 18:17:28 UTC, **36 s**, 2.3 GB |
| 3 | `LoadTestDB` | DIFFERENTIAL | 537 KB | `differential` 18:16:52 UTC |

- The three tasks ran in parallel, all `SUCCESS`. The task waits in the RDS
  queue before starting (created 18:13:55, started 18:14:52).
- Unlike the snapshot, duration and size are real: backup duration alerts
  make sense for the native backup.
- In `msdb` the native backup is the same as the FULL RDS takes on its own:
  a virtual device (`device_type` 7) with a GUID in `physical_device_name`,
  not the S3 path. One cannot be told from the other — and there is no need:
  both are restorable.
- `rds_backup_database` is **not** `COPY_ONLY`: a native FULL becomes the new
  base of the differentials.
- With this the three types are validated on RDS: FULL (snapshot, automated
  and native), DIFFERENTIAL (native) and LOG (automated, every 5 min).

**Fix** — branch `fix/backups-ignore-nul-device`

- Before choosing the last backup per database/type, the collector drops the
  backup sets with any family on `NUL` or `NUL:` (`UPPER(RTRIM(...))`; a
  striped backup with one part on `NUL` does not restore either). So the
  previous real backup is the one reported. `/dev/null` is not even accepted
  by SQL Server on Linux (error 3634).
- `msdb.dbo.backupmediafamily` has `SELECT` for `public`, like `backupset`,
  locally and on RDS: no new grant.
- README: the filter, how the RDS automated backups show up and the new
  table read.

**Validation**

- Local SQL Server 2025, with `BACKUP DATABASE model TO DISK = 'NUL'`
  (written as `NUL`, `nul` and `NUL:`): the previous version reported the
  FULL on `NUL`; the fixed one reports none. With a real backup at 14:16:56
  followed by one to `NUL` at 14:16:58, the previous one reports 14:16:58 and
  the fixed one 14:16:56.
- RDS (automated backups on): both versions emit the same 11 series with the
  same values.
- 4 mutations in the query (`EXISTS`, only `'NUL'`, no `UPPER/RTRIM`, wrong
  join with the media) caught by the tests.

---

## HML-029 — `sqlserver_instance` on RDS: generic host name, changes on every stop/start and is empty if the exporter starts before the database

- **Environment:** RDS for SQL Server 2025 SE, exporter `0.9.2`/`0.9.3` on the bastion
- **Found on:** 2026-10-19
- **Severity:** medium (series change name and the label does not identify
  the instance)
- **Status:** resolved (v0.10.0)

**Observed**

- The label is `@@SERVERNAME`, which on RDS is the name of the Windows host
  behind the instance: `EC2AMAZ-UU69199` on 2026-10-18 and `EC2AMAZ-JFK9UC6`
  on 2026-10-19, just from stopping and starting RDS (no failover). Each
  change breaks the continuity of series and dashboards, and with several RDS
  instances the name does not tell which is which.
- The bastion started at 13:39 UTC and RDS only at 13:54: the exporter came
  up without connecting and, as documented, stayed with
  `sqlserver_instance=""` until it was restarted (14:20).

**Fix** — branch `feat/instance-label`

- `instance_label` option (YAML), `--sqlserver.instance-label` (flag) and
  `SQL_SERVER_INSTANCE_LABEL` (env), with the usual precedence. When set, the
  value is used as is and `@@SERVERNAME` is not even read: the label stays
  stable across stop/start and failover and exists even with the database
  down at startup. Empty or absent, the previous automatic behavior. A value
  of only spaces is refused at startup.
- A single option (not `auto_label` + `custom_label`): set means fixed, with
  no contradictory combination to validate.
- README (with the recommendation to use the instance identifier on RDS) and
  `config.example.yml`.
- ROADMAP (Later): in automatic mode, resolve the name on the first
  successful connection instead of only at startup.

**Local validation (SQL Server 2025)**

| Scenario | `sqlserver_instance` |
|---|---|
| automatic, database up | `mssql2025` |
| `--sqlserver.instance-label=orders-prod`, database up | `orders-prod` (`sqlserver_servername` stays `mssql2025`) |
| fixed, database down at startup | `orders-prod`, with `sqlserver_up 0` |
| automatic, database down at startup | empty, with the warning suggesting `instance_label` |

Precedence tests (YAML < env < flag, an empty flag goes back to automatic)
and blank value tests; 5 mutations in `config.go` caught.

**Validation on RDS (`0.9.4-homolog`, `instance_label: "mssql-exporter-hml"`,
15:14 UTC):** `sqlserver_instance="mssql-exporter-hml"` on every series,
`sqlserver_servername` with the real host. In the HML-030 failover
`servername` went from `EC2AMAZ-JFK9UC6` to `EC2AMAZ-94TRL9U` and the label
stayed `mssql-exporter-hml`: the series continued without a break.

**Update (2026-10-28, v0.15.0):** without `instance_label` and with SQL
Server down at startup, the exporter tries to read `@@SERVERNAME` every 15 s
and sets the label without restarting; see HML-054.

---

## HML-030 — RDS Single-AZ → Multi-AZ conversion and failover: Always On appears with no collection failures (validation)

- **Environment:** RDS for SQL Server 2025 SE (`db.t3.xlarge`), exporter
  `0.9.3` on the bastion; samples every 30 s
- **Found on:** 2026-10-19
- **Severity:** — (validation)
- **Status:** resolved

**Conversion (`apply` at 14:20 UTC, `available` with Multi-AZ at 15:02)**

| UTC | RDS | Exporter |
|---|---|---|
| 14:20 | `modifying`, Multi-AZ false | no `sqlserver_hadr_*` series |
| 14:55 | `modifying` | the `RDSAG0` AG appears (database `rdsadmin`) |
| 14:56 | Multi-AZ true | — |
| 14:57–14:58 | `modifying` | two AGs with a GUID name appear, one per user database (`StressDB`, `LoadTestDB`) |
| 14:59 | `backing-up` | — |
| 15:02 | `available`, secondary in `us-east-1b` | 33 `sqlserver_hadr_*` series |

- `sqlserver_up 1` the whole time, no collector with `success 0`, the label
  and `servername` (`EC2AMAZ-JFK9UC6`) did not change.
- RDS Standard uses **Basic Availability Groups**: one AG per database, with
  a GUID name (`762b6dfb-…`, `de928b35-…`), plus `RDSAG0` for `rdsadmin`. In
  the dashboards `ag_name` is a GUID; `db_name` is what identifies it.
- The primary sees both replicas: `replica_role` 1 for `EC2AMAZ-JFK9UC6` and
  2 for `EC2AMAZ-94TRL9U` (the new one), `synchronization_health` 2 on both,
  databases `SYNCHRONIZED` (2), no suspension, redo queue 0 and send queue of
  60 KB.

**Failover (`aws rds reboot-db-instance --force-failover`, 15:15 UTC)**

RDS events: reboot 15:15:31, failover started 15:15:35 and completed
15:15:40, "DNS propagation may take a few minutes" 15:16:04.

| UTC (15 s scrapes) | `sqlserver_up` | Note |
|---|---|---|
| until 15:15:01 | 1 | primary `EC2AMAZ-JFK9UC6` |
| 15:15:16–15:15:46 | 0 | `connection reset by peer` to the old IP (10.0.1.249) on the pool connections |
| from 15:16:01 | 1 | `servername` = `EC2AMAZ-94TRL9U`, new connections already on the new IP |

- Unavailable for ≈45 s (3 scrapes); came back on its own, without restarting
  the exporter: new connections resolve the endpoint's DNS again.
- No collector with `success 0` with the database up; in the scrapes with
  `sqlserver_up 0` the collectors are skipped, as designed.
- Roles swapped in the 3 AGs: `EC2AMAZ-94TRL9U` primary (1), `EC2AMAZ-JFK9UC6`
  secondary (2), health 2 on both, the 6 database states `SYNCHRONIZED`.
- With `instance_label` (HML-029) the label did not change; without it the
  series would only get `sqlserver_instance="EC2AMAZ-94TRL9U"` after an
  exporter restart (until then they would keep the old primary's name).
- The RDS API kept showing `us-east-1a`/`us-east-1b` as primary/secondary
  minutes after the failover (AWS's own update delay); the exporter's
  `servername` is the fastest source.

---

## HML-031 — Demonstration dashboard for RDS: exporter + CloudWatch (validation)

- **Environment:** local Grafana 13.1 with Prometheus (exporter `0.10.0` on
  the bastion) and CloudWatch (`us-east-1`), RDS `mssql-exporter-hml`
- **Found on:** 2026-10-19
- **Severity:** — (delivery requested in the RDS phase)
- **Status:** resolved

**Delivery**

`dashboards/mssql-exporter-rds-cloudwatch-dashboard.json`: CloudWatch
(`AWS/RDS`) for the host and the exporter for the engine, with panels that
mix both sources (CPU × batch requests, disk writes × writes to the SQL
Server files, `DatabaseConnections` × user connections). Variables: region,
RDS instance (`DBInstanceIdentifier` dimension) and `sqlserver_instance`;
with `instance_label` equal to the RDS identifier (HML-029) both selectors
use the same name.

**Validation**

- Every query of the 34 panels run through the Grafana API for the last
  6 h: all with data, except `SwapUsage` (CloudWatch does not publish that
  metric for SQL Server — removed) and the Always On panels (empty on
  Single-AZ; they had data during HML-030).
- Checked visually in the browser: UP, current host `EC2AMAZ-V7Q122H`, CPU
  ≈7%, free memory 12.6 GB, free storage 17.1 GB, last FULL ≈30 min ago.
- Adjustments after the first look: lines broken on gaps of more than 10 min
  (RDS was stopped between 15:35 and 17:58 UTC and the CloudWatch series were
  joined by a straight line) and "Top waits" without the idle waits
  (`LAZYWRITER_SLEEP`, `CHECKPOINT_QUEUE`, `HADR_WORK_QUEUE`, …), which
  dominated the panel; the exporter only filters `SLEEP%`, `XE%` and
  `BROKER_%`.

---

## HML-032 — Cloud SQL baseline: permissions, collectors and load (validation)

- **Environment:** Cloud SQL for SQL Server 2022 Standard (`16.0.4262.2`,
  RTM-CU25-GDR) **on Linux** (Ubuntu 22.04), `db-custom-2-7680` (2 vCPUs,
  7.5 GB) zonal in `us-central1-a`, private IP only (Private Service Access),
  `ssl_mode = ENCRYPTED_ONLY`, UTC-3 time zone (`E. South America Standard
  Time`); exporter `0.12.0` via systemd on the Ubuntu 26.04 bastion
  (`e2-small`, no public IP, SSH through IAP), read by the local Prometheus
  through a tunnel. Resources in Terraform, outside the repository
- **Found on:** 2026-10-26
- **Severity:** — (validation)
- **Status:** resolved (README: permissions on Cloud SQL)

**Permissions**

There is no `sysadmin`. The `sqlserver` user is a member of
`CustomerDbRootRole`, which has `VIEW SERVER STATE` and `VIEW ANY DEFINITION`
`WITH GRANT OPTION`. Tests run as `sqlserver`, each in a transaction undone
with `ROLLBACK`:

| Command | Result |
|---|---|
| `GRANT VIEW SERVER STATE TO [usr_exporter]` | error 4613: `Grantor does not have GRANT permission` |
| `GRANT VIEW SERVER STATE … AS [CustomerDbRootRole]` (in `master`) | OK |
| `GRANT VIEW ANY DEFINITION … AS [CustomerDbRootRole]` | OK |
| `ALTER SERVER ROLE [##MS_ServerStateReader##]` / `##MS_DefinitionReader##` / `##MS_DatabaseConnector##` `ADD MEMBER` | error 15151: the roles exist, but do not accept members |
| `GRANT CONNECT ANY DATABASE … AS [CustomerDbRootRole]` | error 4613: a user per database, as on RDS (HML-024) |
| `CREATE USER` in `msdb` | error 15247: `msdb` is protected |

- `backups` works without a grant: `guest` enabled in `msdb` and `SELECT` on
  `backupset`/`backupmediafamily` for `public`.
- `agent_jobs`: `SELECT` on `sysjobs`/`sysjobhistory` only for `sqlserver`
  itself; the minimal login gets error 229 (`success 0`). Disabled in the
  config. Logins created through the Cloud SQL API join
  `CustomerDbRootRole` (Google's documentation, not tested) and would read
  the jobs, but with too much privilege (`ALTER ANY LOGIN`,
  `CREATE ANY DATABASE`). Cloud SQL publishes `sqlserver/sql_agent/jobs` in
  Cloud Monitoring.
- `hack/loadgen`'s `setup.sql` runs as `sqlserver` without changes
  (`CREATE ANY DATABASE` and `ALTER ANY LOGIN`).

**Collectors**

- First collection: 16 of the 17 default collectors with `success 1`, only
  `agent_jobs` with `success 0`; scrape of 0.78 s and 4,708 samples.
- First environment of the homologation with SQL Server on **Linux**: no
  query needed a change.
- `@@SERVERNAME` = `MachineName` = `3da04acc8e845da`, a generated host name:
  `instance_label` recommended, as on RDS (HML-029).
- `volumes`: `sys.dm_os_volume_stats` returns `NULL` in
  `volume_mount_point`, `logical_volume_name` and `file_system_type`; the
  bytes are right (20 GB disk), but the series comes out with
  `volume_mount_point=""`. Point to consider: a fallback label in the
  collector. Resolved in v0.13.4: the label becomes the directory of the
  files (`/var/opt/mssql/data/`).
- Internal database `gcloud_cloudsqladmin` in `databases`, `backups` and
  `log_used_percent`, like `rdsadmin` on RDS. In `sessions_by_login` the
  Cloud SQL accounts show up: `CloudDbSqlRoot`, `CloudDbSqlRoot_2`,
  `CloudDbSqlAgent` and `NT AUTHORITY\NETWORK SERVICE`.
- `backups`: FULL of the user databases and LOG every 5 min with PITR on
  (written to `/var/opt/mssql/data/pitr/tmptlog/…`); with the server in UTC-3
  the latest LOG showed up 8 s old (HML-010 OK).
- `index_fragmentation` and `query_store` (opt-in) with a user per database:
  `success 1` on 2 databases (`LoadTestDB` and `StressDB`), Query Store in
  `READ_WRITE` with a 1 GB limit.

**Load (`hack/loadgen` on the bastion, 10 logins × 2 workers, 19:25–19:44 UTC)**

- ≈56,700 statements without error in the loadgen; ≈61 batch requests/s on
  average, peak of ≈101/s.
- Up to 14 blocked sessions (maximum wait of 28 s), 10 requests above 10 s
  and an oldest open transaction of 29 s, all in the queue of the `blocker`.
- `perf_counters` median at 0.64 s under load (0.13 s on RDS). The
  instance's CPU reached ≈98% (Cloud Monitoring, HML-035) with the loadgen and
  the HML-034 test together.

---

## HML-033 — TLS on Cloud SQL: `encrypt=true` with the instance CA; `encrypt=strict` is not accepted

- **Environment:** that of HML-032; tests through the tunnel (Mac → bastion →
  Cloud SQL) with `hostnameincertificate` in place of the tunnel host
- **Found on:** 2026-10-26
- **Severity:** low (platform limitation; `encrypt=true` covers the case)
- **Status:** resolved (README: TLS section)

Server certificate: `CN=<project>:<instance>`, issued by the instance's
`Cloud SQL Server CA` (10-year validity), with the private IP and
`<uuid>.us-central1.sql.goog` in the SAN. The concern that the IP is not in
the certificate does not hold for the private IP.

| DSN | Result |
|---|---|
| `encrypt=true&certificate=<CA>` by IP or by the `sql.goog` name | OK, `encrypt_option` TRUE, TDS 7.4 |
| `encrypt=true&certificate=<CA>&hostnameincertificate=wrong.example.com` | refused (`x509: certificate is valid for …`) |
| `encrypt=true` without the CA (system CAs) | refused |
| `encrypt=strict` with the CA, with the name or with `trustservercertificate=true` | **fails** in the handshake with `EOF` |
| `encrypt=disable` | refused (`ENCRYPTED_ONLY`) |

RDS accepted `encrypt=strict` without configuration (HML-027); Cloud SQL
does not accept TDS 8.0. The bastion's exporter with `encrypt=true` and the
CA in `/etc/mssql-exporter/cloudsql-server-ca.pem`.

---

## HML-034 — Concurrent scrapes saturate the 5-connection pool: collectors fail with `context deadline exceeded`

- **Environment:** that of HML-032, during the loadgen load; `hey` on the
  bastion against the local `/metrics`, with Prometheus scraping every 15 s
- **Found on:** 2026-10-26
- **Severity:** medium (partial and silent data: HTTP 200 with collectors
  missing)
- **Status:** deferred (ROADMAP)

Each scrape runs all collectors in parallel over the fixed pool of 5
connections, with a single 5 s timeout (pending from the 2026-10-20 review).
Concurrency ladder, 100 requests per level:

| Concurrency | p50 | p90 | p99 | Failed collections (Prometheus) |
|---|---|---|---|---|
| 1 | 0.55 s | 1.79 s | 2.18 s | 0 of 85 |
| 2 | 1.12 s | 2.22 s | 3.39 s | 0 of 68 |
| 5 | 4.57 s | 5.06 s | 5.22 s | 2 of 85 |
| 10 | 5.03 s | 5.07 s | 5.43 s | 14 of 68 (21%) |

- Up to 2 simultaneous scrapes (one Prometheus or an HA pair) there is no
  failure.
- From 5 on, the collectors wait for a connection, the median touches 5 s and
  the collectors fail with `context deadline exceeded`; every response is
  HTTP 200 (`ContinueOnError`), only `scrape_collector_success` shows it.
- The 2-vCPU instance had its CPU at ≈98% in the period, so part of the wait
  is SQL Server's own.
- Cloud SQL also ran a `BACKUP_VOLUME` from 19:31 to 19:34 UTC, at the start
  of the ladder, which may have weighed on the first levels.

Candidate fixes (ROADMAP, Later): simultaneous scrapes wait for the
collection in progress and get the same result (singleflight), and a
configurable pool size.

**Fix (2026-10-27)** — branch `feat/scrape-singleflight`:

- `Exporter.Collect`: a scrape that arrives while another one is collecting
  waits for that collection and exposes the same metrics (const metrics,
  immutable), instead of running all collectors again. Scrapes that do not
  overlap collect normally; there is no cache.
- `--sqlserver.max-connections` (env `SQL_SERVER_MAX_CONNECTIONS`, YAML
  `max_connections`, default 5, from 1 to 100) sets the pool size.

**Validation** — `v0.13.4` binary (before) against the branch's (after),
running on the Mac against the home notebook's SQL Server 2025 Developer, 16
default collectors; a load tool that counts the responses with any
`scrape_collector_success 0`:

| Simultaneous | Before: p50 / p99 | Before: failed responses | After: p50 / p99 | After: failed responses |
|---|---|---|---|---|
| 1 | 0.13 s / 0.16 s | 0 of 40 | 0.13 s / 0.16 s | 0 of 40 |
| 5 | 0.49 s / 0.78 s | 0 of 40 | 0.13 s / 0.15 s | 0 of 40 |
| 20 | 1.73 s / 2.98 s | 0 of 40 | 0.13 s / 0.13 s | 0 of 40 |
| 60 | 5.02 s / 5.04 s | 112 of 120 (471 collectors) | 0.16 s / 0.19 s | 0 of 120 |

- Before, latency grows with concurrency until it touches the 5 s timeout;
  with an idle SQL Server the failure only shows with 60 simultaneous (on
  Cloud SQL under load, from 5). After, latency stays the same as a single
  scrape's.
- A single scrape with `max-connections` 1, 5 and 20: p50 of 0.34 s, 0.13 s
  and 0.13 s. The default pool already suffices for the 16 default
  collectors on that machine; the configurable value is for instances with
  more collectors or a slower network, or to limit the open sessions.

---

## HML-035 — Dashboard for Cloud SQL: exporter + Cloud Monitoring (validation)

- **Environment:** local Grafana 13.1 with Prometheus (exporter `0.12.0` on
  the bastion) and the Google Cloud Monitoring data source
- **Found on:** 2026-10-26
- **Severity:** — (delivery)
- **Status:** resolved (PR #35)

**Delivery**

`dashboards/mssql-exporter-cloudsql-dashboard.json`, the Cloud SQL version
of HML-031: Cloud Monitoring (`cloudsql.googleapis.com`) for the host, the
exporter for the engine, correlation panels and a row with the counters that
Cloud SQL also publishes (batch requests, page life expectancy and
deadlocks) from both sources; plus the log and Query Store panels of Phase 7.

**Keyless authentication**

The project's organization enforces `iam.disableServiceAccountKeyCreation`
(Google's default for new organizations). The data source uses a service
account with `roles/monitoring.viewer` only, through impersonation: ADC
generated with
`gcloud auth application-default login --impersonate-service-account=…` in a
separate gcloud folder made available to the Grafana process, with the "GCE
Default Service Account" authentication. The data source test button answers
`400` (it looks for the GCE metadata server), but the queries work.

**Validation**

- Every query run through the Grafana API in the load window (19:20–19:50
  UTC): all with data, except Top queries (`top_queries` off) and the Always
  On tables (no read replicas).
- Both sources agree between 19:27 and 19:42 UTC: batch requests averaging
  75.3/s on the exporter (`rate[5m]`) and 81.4/s on Cloud Monitoring (1-min
  alignment); PLE of 4,888 s and 4,909 s.
- Checked in the browser: variables resolved and panels rendered. A Grafana
  tab opened before the data source was created shows
  `Datasource … was not found` until it is reloaded.

---

## HML-036 — Conversion to regional HA and Cloud SQL failover (validation)

- **Environment:** that of HML-032, with `availability_type` `ZONAL` →
  `REGIONAL` through Terraform; `hack/loadgen` load during the failovers
- **Found on:** 2026-10-26
- **Severity:** — (validation)
- **Status:** resolved

**Conversion to regional (22:13–22:23 UTC)**

The instance is unavailable for ≈5 min 45 s (22:15:48–22:21:30): `Read:
EOF`, then `i/o timeout` and `connection refused` on the private IP. Primary
in `us-central1-a`, standby in `us-central1-f`. Unlike RDS, where the
conversion to Multi-AZ did not take the instance down (HML-030). The
exporter did not restart and came back on its own; `query_store` logged a
`WARN` and came back on the next collection.

**Failovers (`gcloud sql instances failover`, under load)**

| Failover | Command | `sqlserver_up 0` | Note |
|---|---|---|---|
| 1 (a → f) | 22:39:41–22:41:01 | ≈77 s (22:39:58–22:41:15) | server stopped answering without closing the connections: 2 scrapes lost (HML-037) |
| 2 (f → a) | 22:51:24–22:52:45 | ≈60 s (22:52:04–22:53:15) | connections dropped with `connection reset by peer`; no scrape lost |

- `@@SERVERNAME` stays `3da04acc8e845da` in both zones: on Cloud SQL the
  name follows the instance, not the VM (on RDS it changes on every
  failover, HML-029). `instance_label` is still recommended, because the name
  is generated.
- No AG appears: Cloud SQL HA is by regional disk.
- The loadgen gets `Neither the database "LoadTestDB" … could be opened`
  while the database runs recovery on the new primary (168 errors on the 1st
  failover, 1,284 on the 2nd) and goes on afterwards without intervention.

---

## HML-037 — Whole scrape lost when the server stops answering without closing the connection

- **Environment:** that of HML-036, on the 1st failover; reproduced on the
  bastion with `iptables` dropping the traffic to port 1433 for 45 s
- **Found on:** 2026-10-26
- **Severity:** medium (the unavailability window disappears: not even
  `sqlserver_up 0` is exposed)
- **Status:** resolved (v0.12.1)

**Observed**

On the 1st failover, two scrapes in a row exceeded the 10 s
`scrape_timeout` (`up` 0, no samples) and the exporter logged:

```
sql server is unreachable error="Invalid TDS stream: did not get cancellation confirmation from the server (current response: context deadline exceeded)"
```

**Cause**

The ping exceeds the 5 s deadline, the driver sends the cancellation and
keeps waiting for the confirmation from a server that no longer answers,
without honoring the context. `Collect` waited for the ping (and the
collectors) without a bound of its own, so the answer went past Prometheus's
10 s. An alert on `sqlserver_up == 0` does not fire in that window; only one
on `up == 0`.

**Fix** — branch `fix/scrape-deadline`, test build `0.12.1-homolog`

`Collect` waits for the ping and each collector only until the deadline plus
500 ms: a ping that did not return exposes `sqlserver_up 0`; a collector that
did not return exposes `success 0` and the others are still exposed; what it
sends later goes to a drain, not to the registry's channel (already closed).
Tests with a ping and a collector that ignore the context; mutation test:
without the bound, both fail.

**Validation (A/B on the bastion, loadgen load, 45 s without an answer from the server)**

| Version | Scrapes during the block | `up` | Log |
|---|---|---|---|
| `0.12.0` | 10 s (timeout), 0 samples, 2 in a row | 0 | `did not get cancellation confirmation…` |
| `0.12.1-homolog` | 5.7 s, 50 samples | 1 (`sqlserver_up 0`) | `ping did not return before the scrape deadline` |

In both versions the exporter comes back on its own ≈50 s after the traffic
is released, when the driver discards the stuck connections.

---

## HML-038 — SQL Server Express without Agent (local validation)

- **Environment:** SQL Server 2025 Express (`17.0.4025.3`) in Docker on the
  Mac (`mssql-2025-express` container, emulated amd64 image), local exporter
  `0.12.1`, local Prometheus
- **Found on:** 2026-10-26
- **Severity:** — (validation)
- **Status:** resolved

Done locally instead of on Cloud SQL: the Express differences (no Agent,
CPU, memory and size limits) belong to the edition, not to the platform, and
the exporter has no per-edition logic.

- Login with the README grants (`VIEW SERVER STATE`, `VIEW ANY DEFINITION`,
  `CONNECT ANY DATABASE`, `SELECT` on `sysjobs`/`sysjobhistory`).
- The 19 collectors (default + `top_queries`, `index_fragmentation` and
  `query_store`) with `success 1` from the first collection; no `ERROR` or
  `WARN` in the exporter log.
- `agent_jobs`: `SQL Server Agent` `Stopped` and `msdb.dbo.sysjobs` empty;
  the collector reports `success 1` without series.
- `sqlserver_edition{edition="Express Edition (64-bit)"}`;
  `schedulers_online` = 4 with 8 CPUs in the container (Express's 4-core
  limit).
- `volumes`: `volume_mount_point=""` here too — the `NULL` from
  `sys.dm_os_volume_stats` belongs to SQL Server on Linux, not to Cloud SQL
  (HML-032).
- `hack/loadgen` load (10 logins × 2 workers, 10 min): ≈45,200 statements
  without error, average of 78 batch requests/s (peak 118), up to 16 blocked
  sessions and 16 requests above 10 s; maximum scrape of 0.23 s, no collector
  with `success 0`; `top_queries` with 50 queries.

---

## HML-039 — SQL Managed Instance baseline: permissions, collectors and load (validation)

- **Environment:** Azure SQL Managed Instance through the free offer (General
  Purpose Gen5, 4 vCores, 64 GB, `SQLServer2022` format) in `westus3`,
  private endpoint only, UTC-3 time zone; Windows Server 2025 host with 22 GB
  and 4 visible CPUs. Exporter via systemd on the Ubuntu 26.04 bastion in the
  same VNet, DSN with `encrypt=true`, read by the local Prometheus through a
  tunnel. Resources in Terraform, outside the repository
- **Found on:** 2026-10-27
- **Severity:** — (validation)
- **Status:** resolved

**Permissions**

The instance's SQL admin (`sqladmin`) is `sysadmin`, so the EC2 grants work
without adaptation, including in `msdb`:

```sql
CREATE LOGIN [usr_exporter] WITH PASSWORD = '…', CHECK_POLICY = OFF, CHECK_EXPIRATION = OFF;
GRANT VIEW SERVER STATE TO [usr_exporter];
GRANT VIEW ANY DEFINITION TO [usr_exporter];
USE msdb;
CREATE USER usr_exporter FOR LOGIN usr_exporter;
GRANT SELECT ON sysjobs TO usr_exporter;
GRANT SELECT ON sysjobhistory TO usr_exporter;
```

Without the `msdb` grants, `agent_jobs` stays at `success 0`, as on EC2
(HML-002).

**Platform**

- `SERVERPROPERTY`: `EngineEdition` 8, `Edition` `SQL Azure`,
  `ProductVersion` `12.0.2000.8` (fixed, it does not follow the engine
  version); `@@SERVERNAME` is the instance's FQDN; `MachineName`,
  `ComputerNamePhysicalNetBIOS` and `InstanceName` are NULL (HML-042).
- System and user databases are on remote blob storage (`physical_name`
  `https://…blob.core.windows.net/…` in `io_file_stats`); only `tempdb` is on
  the local SSD (`C:\`, 12 files of 16 MB). See HML-043.
- `sessions_by_login` shows the service accounts: `NT AUTHORITY\SYSTEM` and
  `User Manager\ContainerAdministrator`.
- `availability_groups` shows nothing (General Purpose HA does not appear as
  an AG) and `agent_jobs` has no series (no job created).
- `wait_stats`: 1,499 wait types exposed. At the top, by cumulative time,
  only the service's background waits: `SOS_WORK_DISPATCHER`,
  `HADR_WORK_QUEUE`, `DISPATCHER_QUEUE_SEMAPHORE` and `LOGMGR_QUEUE`; the
  first load wait is `LCK_M_X`. It reinforces the ROADMAP's idle wait filter
  (HML-031).

**Collectors**

On the first collection, 14 of the 16 default collectors were at
`success 1`. The failures were `server_info` (HML-042) and `agent_jobs`,
which ran before the `msdb` grants. After the grants and the fixes, the 16
at `success 1`. `volumes`, `log_used_percent`, `backups` and
`requests_long_running` needed their own analysis (HML-043 to HML-046).

**Load (`hack/loadgen` on the bastion, 10 logins × 2 workers, 30 min, 14:55–15:25 UTC)**

- ≈115,300 statements without error in the loadgen; ≈67 batch requests/s on
  average.
- Up to 16 blocked sessions (maximum wait of 28.6 s) and an oldest open
  transaction of 35 s, in the queue of the `blocker`; no deadlock.
- Maximum scrape of 1.31 s; the slowest collectors were `perf_counters`
  (0.77 s) and `tempdb` (0.46 s); no collector with `success 0`.
- `LoadTestDB` write latency: 33 ms on data and 4.3 ms on the log, because of
  General Purpose's remote storage.
- `LoadTestDB`'s `log_reuse_wait` at `OLDEST_PAGE` during the load.

---

## HML-040 — Azure Free Trial: Managed Instance quota only in `westus3`, no B-series VMs and 404 when reusing the resource group name (environment)

- **Environment:** Free Trial subscription (`FreeTrial_2014-09-01`), Terraform
  with `azurerm` 5.11 and `azapi` 2.13
- **Found on:** 2026-10-27
- **Severity:** — (environment)
- **Status:** resolved

- **MI quota:** 0 vCores in `eastus`, `eastus2`, `centralus`,
  `southcentralus`, `westus2`, `brazilsouth`, `canadacentral`, `northeurope`
  and `westeurope`; 16 in `westus3`. The environment went to `westus3`.
- **Bastion VM:** the subscription does not offer the B series (not even the
  free B1s); the cheapest available was `Standard_D2s_v4` (US$ 0.096/h), with
  automatic shutdown at 21:10.
- **Resource providers:** the first run fails until Compute, Consumption,
  DevTestLab, Network and Sql are registered in the subscription.
- **MI free offer:** `azurerm` 5.11 does not expose the free offer; the MI was
  created through `azapi` with `pricingModel = "Freemium"`
  (`Microsoft.Sql/managedInstances@2025-01-01`), together with the start/stop
  schedule. Created in 6 min 14 s, with the new subnet.
- **Resource group name:** recreating the group with the same name in
  another region (from `eastus2` to `westus3`) made ARM answer 404 for
  resources that had just been created (`Root object was present, but now
  absent`): NSG, NIC and VNet disappeared from the state intermittently.
  Solved with the region in the name (`mssql-exporter-hml-westus3`). The
  subnets also needed `replace_triggered_by` on the VNet, because their ID
  does not change when the VNet is recreated.

---

## HML-041 — SQL through the SSH tunnel does not reach the MI: the gateway requires the instance name and redirects to the node (environment)

- **Environment:** that of HML-039; Mac → SSH tunnel on the bastion
  (`-L 14332:<fqdn>:1433`) → MI private endpoint
- **Found on:** 2026-10-27
- **Severity:** low (affects only ad hoc access; the exporter in the VNet
  connects normally)
- **Status:** resolved (T-SQL from the bastion)

- Connecting to `localhost:14332`, the MI gateway refuses the login with
  error 40532, because it does not know which instance to route to.
- With the login in the `user@instance` format, the gateway accepts it, but
  answers with a redirection to the node's IP in the MI subnet
  (`10.20.1.x:1433`), which the Mac does not reach through the tunnel.
- The exporter on the bastion does not have the problem: it resolves the
  FQDN and follows the redirection inside the VNet.
- Test and administration T-SQL runs on the bastion with `go-sqlcmd`, from
  `.sql` files copied with `scp` (`sqlcmd -i`).

---

## HML-042 — `server_info` fails entirely with `MachineName` NULL on MI

- **Environment:** that of HML-039, exporter `0.12.1`
- **Found on:** 2026-10-27
- **Severity:** medium (loses the instance's edition, version and
  `engine_edition`)
- **Status:** resolved (v0.12.2)

On MI, `SERVERPROPERTY('MachineName')`, `ComputerNamePhysicalNetBIOS` and
`InstanceName` are NULL. The properties row was read into Go strings, and a
single NULL took down the whole row:

```
ERROR collector scrape failed collector=server_info error="failed to query server properties: sql: Scan error on column index 0, name \"hostname\": converting NULL to string is unsupported"
```

`sqlserver_servername`, `sqlserver_edition`, `sqlserver_product_version`,
`sqlserver_engine_edition` and `sqlserver_is_hadr_enabled` disappeared, and
the collector stayed at `success 0` on every collection. CPU and memory
remained, because they come from other queries.

**Fix:** properties read as `sql.NullString`/`sql.NullInt64`; only the metric
of the NULL property is omitted (on MI, only `sqlserver_hostname`). A test
with a row like MI's; the README documents the behavior.

---

## HML-043 — `volumes` on MI: `total_bytes` is the size of the files, not of the volume (available above 100%)

- **Environment:** that of HML-039, exporter `0.12.2`
- **Found on:** 2026-10-27
- **Severity:** medium (disk panel and alert meaningless on MI)
- **Status:** resolved (v0.13.0)

On MI, `sys.dm_os_volume_stats` does not describe the volume: `total_bytes`
is the sum of the files on the volume and is smaller than
`available_bytes`.

| Volume | What it is | `total_bytes` | `available_bytes` |
|---|---|---|---|
| `http://` | remote blob storage (system and user databases) | 304 MB | 63.7 GB |
| `C:\` | local SSD (`tempdb` only) | 192 MB | 95.8 GB |

The "Volume Space Available (%)" panel showed ≈21,000% and ≈51,000%, and a
low-space alert built on those series would never fire.

**Fix:** `volumes` queries `SERVERPROPERTY('EngineEdition')`; on MI (8) it
does not emit the volume series and reports the instance storage from the
latest sample of `master.sys.server_resource_stats` (covered by
`VIEW SERVER STATE`): `sqlserver_instance_storage_reserved_bytes` (64 GiB)
and `sqlserver_instance_storage_used_bytes` (336 MiB in the validation;
99.49% available in the new panel). The view is published a few minutes
late and is empty right after the instance is created; without a row,
nothing is emitted. Dashboard with "Instance Storage Available (%)" and
"Instance Storage Used vs Reserved". On the other platforms, only one more
`SERVERPROPERTY` query per scrape (validated on SQL Server 2025 in Docker).

---

## HML-044 — `database_log_used_percent` without the user databases on MI (counter by physical name)

- **Environment:** that of HML-039, exporter `0.13.0`, during the load
- **Found on:** 2026-10-27
- **Severity:** low (the system databases remain; the log usage of the user
  databases disappears)
- **Status:** resolved (v0.13.1)

On MI, the `Percent Log Used` counter instance of a user database is its
`physical_database_name` (a GUID), not the name: `LoadTestDB` shows up as
`17f319d1-1099-430f-8903-7c77b206bf6b`. The `databases` join by name dropped
every user database. The counter also has instances for the physical names
of `master` (`replicatedmaster`) and `model` (`managed_model`), with values
different from those of `master` and `model`.

**Fix:** on MI only, the counter is matched by the name or by
`physical_database_name`, preferring the name (`CROSS APPLY … TOP (1)`), so
that `master` and `model` do not take the value of the physical instance. On
the other platforms the query does not change; a test ensures it does not use
`physical_database_name`, which only exists from SQL Server 2019. The
`EngineEdition` query became a helper shared with `volumes`. Validated on MI
(`LoadTestDB` 1, `master` 44, `model` 19) and on SQL Server 2025 in Docker
(the same series as before).

---

## HML-045 — `backups` without any series on MI: `server_name` NULL in `backupset`

- **Environment:** that of HML-039, exporter `0.13.1`
- **Found on:** 2026-10-27
- **Severity:** medium (without `sqlserver_backup_*` there is no late backup
  alert on MI)
- **Status:** resolved (v0.13.2)

MI's automated backups show up in `msdb.dbo.backupset` (FULL of `LoadTestDB`
at creation and LOG every ≈5 min during the load, then every ≈1 h), with a
correct `time_zone` of -12 (UTC-3). But `server_name` and `machine_name` come
NULL, and so does `physical_device_name` (`device_type` 9, URL). The
collector filters `server_name = @@SERVERNAME` so as not to report backups
restored from another server (HML-012), so it drops every row and stays at
`success 1` without series. The system databases have no history in `msdb`.

**Fix:** on MI only (`EngineEdition` 8), a NULL `server_name` counts as the
instance itself; on the other platforms the filter does not change.
Validated on MI (FULL and LOG of `LoadTestDB`, with the FULL of 11:59:01
local at 14:59:01 UTC) and on SQL Server 2025 in Docker (backup series the
same as with `0.13.1`).

---

## HML-046 — `requests_long_running` always 1 in every bucket on MI: `TdService` telemetry session

- **Environment:** that of HML-039, exporter `0.13.1`, without load
- **Found on:** 2026-10-27
- **Severity:** low (an alert on
  `sqlserver_requests_long_running{over_seconds="1800"} > 0` would fire
  forever on MI)
- **Status:** resolved (v0.13.2)

`active_queries` already excludes the system sessions
(`is_user_process = 0`, 98 requests on MI). One service session marked as a
user one is left: `program_name` `TdService`, login `NT AUTHORITY\SYSTEM`, a
`SELECT` in `suspended` with a `XE_LIVE_TARGET_TVF` wait (continuous read of
an Extended Events session), open since the instance started (15,292 s at
the time of the query). It counts 1 in the 10, 60, 300 and 1,800 s buckets
all the time.

**Fix:** on MI only, `active_queries` ignores the requests with
`program_name` `TdService` and login `NT AUTHORITY\SYSTEM` (the wait is not
part of the filter, so that the session does not count again while it is
running). This also removes the 15-thousand-second `SELECT` from
`sqlserver_request_max_elapsed_seconds`. Validated on MI: the buckets went
to 0 and a 40 s `WAITFOR` from `sqladmin` showed up in `over_seconds="10"`
and in `request_max_elapsed_seconds{command="WAITFOR"}`.

---

## HML-047 — OS collector logs a swap `ERROR` on every scrape on Windows without a pagefile

- **Environment:** home notebook (Windows 11 Pro pt-BR, no pagefile),
  exporter `0.13.2` as a service, scraped by the Raspberry Pi's Prometheus
  every 15 s
- **Found on:** 2026-10-27
- **Severity:** medium (≈5,760 errors a day in the Event Log, hiding real
  errors; the swap metrics disappear)
- **Status:** resolved (v0.13.3)

With the pagefile off, the `\Paging File(_Total)\% Usage` counter has no data
and gopsutil's `mem.SwapMemory()` fails on every collection:

```
level=ERROR msg="failed to collect swap memory metrics" error="A operação foi concluída com êxito.: this counter has not data"
```

18 errors in 5 minutes in the Event Log. It did not show on EC2 because there
was a pagefile.

**Fix:** when `SwapMemory` fails and Windows lists no pagefile
(`mem.SwapDevices()` empty), the collector emits zeroed swap without logging
an error, as on Linux without swap. With a pagefile, the error is still
logged. Validated on the notebook with build `0.13.3-homolog` on a separate
port: 4 scrapes without any `ERROR` and `os_memory_swap_total_bytes 0`.

---

## HML-048 — The README's `icacls` fails on non-English Windows (`Administrators` does not exist in pt-BR)

- **Environment:** that of HML-047
- **Found on:** 2026-10-27
- **Severity:** low (the command fails and the file is left without the restriction)
- **Status:** resolved (README)

The README restricted `config.yml` (which holds the DSN) with
`icacls … /grant:r "SYSTEM:R" "Administrators:F"`. On Portuguese Windows the
groups are `AUTORIDADE NT\SISTEMA` and `BUILTIN\Administradores`, and
`icacls` answers `Não foi feito mapeamento entre os nomes de conta e as
identificações de segurança` ("No mapping between account names and
security IDs was done"), without changing anything. The same happens with
OpenSSH's `administrators_authorized_keys`.

**Fix:** the README uses the fixed SIDs (`*S-1-5-18` for SYSTEM and
`*S-1-5-32-544` for Administrators), which work in any language.

---

## HML-049 — Per-database collectors on MI with `CONNECT ANY DATABASE` (validation)

- **Environment:** that of HML-039, exporter `0.13.2` with `top_queries`,
  `index_fragmentation` and `query_store` on
- **Found on:** 2026-10-27
- **Severity:** — (validation)
- **Status:** resolved

- **Without the grant:** `index_fragmentation` and `query_store` with
  `success 1` and `last_collection_databases 0`. `LoadTestDB`, without a user
  for the login, is skipped without error, as documented.
- **With `GRANT CONNECT ANY DATABASE TO [usr_exporter]`** (given by
  `sqladmin`, which is `sysadmin`), after restarting the exporter
  (`index_fragmentation` collects every 6 h): both collect `LoadTestDB`.
  - `index_fragmentation`: `PK_Orders` with 95.6% fragmentation after the
    loadgen load.
  - `query_store`: `READ_WRITE` (on by default on MI), 100 MB limit.
- `top_queries`: 50 queries with only `VIEW SERVER STATE`.

Unlike RDS (HML-024) and Cloud SQL (HML-032), where the grant cannot be given
and a user per database is needed.

---

## HML-050 — TLS on MI: `encrypt=true` and `encrypt=strict` with a public CA, no certificate file (validation)

- **Environment:** that of HML-039; a temporary exporter per DSN on the
  bastion, with its own `app name`, checked in `sys.dm_exec_connections`
- **Found on:** 2026-10-27
- **Severity:** — (validation)
- **Status:** resolved (README: TLS section)

Server certificate: `CN=65cf56cd264a.database.windows.net` (the instance's
DNS zone), with `*.<zone>.database.windows.net` in the SAN, issued by
`Microsoft TLS G2 RSA CA OCSP 10`, a public CA present in the system CAs.

| DSN | Result |
|---|---|
| `encrypt=true` (the service's) | OK, `encrypt_option` TRUE, TDS 7.4 (`0x74000004`) |
| `encrypt=strict` | OK, TDS 8.0 (`0x08000000`) |
| `encrypt=strict&hostnameincertificate=wrong.example.com` | refused (`x509: certificate is valid for *.65cf56cd264a.database.windows.net, …`) |
| `encrypt=disable` | refused (`EOF` in the handshake): MI requires TLS |

No CA file and no server configuration. The second managed environment to
accept TDS 8.0, after RDS (HML-027); Cloud SQL does not accept it (HML-033).

---

## HML-051 — MI managed failover under load: ≈80 s of `sqlserver_up 0`, no scrape lost (validation)

- **Environment:** that of HML-039, exporter `0.13.2`, `hack/loadgen` load
  (10 logins × 2 workers) and a `sqlserver_up` monitor every 2 s on the
  bastion; `az sql mi failover --replica-type Primary`
- **Found on:** 2026-10-27
- **Severity:** — (validation)
- **Status:** resolved

| Time (UTC) | Event |
|---|---|
| 19:39:59 | failover requested (`az` answered in 1 min 36 s) |
| 19:40:01 | connections dropped (`connection reset by peer`); `tempdb`, `perf_counters` and `database_mirroring` with `success 0` in that scrape |
| 19:40:03 – 19:41:18 | `sqlserver_up 0`: `i/o timeout` to the node and then `Database 'master' … is not currently available` |
| 19:41:18 | new `sqlserver_start_time` (10:53 → 16:41 local time) |
| ≈19:41:24 | `sqlserver_up 1`, 19 collectors with `success 1`; no error since 19:41:30 |

- The local Prometheus saw `up 1` the whole time: during the outage each
  scrape took ≈5.2 s and answered within the deadline (HML-037 fix).
  `sqlserver_up` stayed at 0 for 6 points of 15 s, so an alert on
  `sqlserver_up == 0` fires.
- `@@SERVERNAME` and `instance_label` the same before and after. The
  redirection node kept its IP (`10.20.1.12`): on General Purpose the
  failover restarted the SQL Server process.
- `query_store` (in the background) logged a `WARN` during the outage and
  collected normally on the next round.
- Loadgen: 285 statements with errors between 19:40:16 and 19:41:23; it went
  on without intervention.

Comparison: RDS Multi-AZ ≈45 s (HML-030), Cloud SQL HA ≈60–77 s (HML-036),
MI General Purpose ≈80 s.

---

## HML-052 — Dashboard for Managed Instance: exporter + Azure Monitor (validation)

- **Environment:** local Grafana 13.1 with Prometheus (exporter `0.13.2` on
  the bastion, through the tunnel) and an Azure Monitor data source
  authenticated by an app registration (Terraform: client secret,
  `Monitoring Reader` on the resource group only)
- **Found on:** 2026-10-27
- **Severity:** — (validation)
- **Status:** resolved (`dashboards/mssql-exporter-azure-mi-dashboard.json`)

Same model as the RDS (HML-031) and Cloud SQL (HML-035) dashboards. Azure
Monitor publishes only 7 metrics for MI (`avg_cpu_percent`,
`virtual_core_count`, `storage_space_used_mb`, `reserved_storage_mb`,
`io_requests`, `io_bytes_read`, `io_bytes_written`): no memory, connections
or latency, which are left to the exporter.

- The subscription, resource group and instance variables (queries of the
  data source itself) resolve on their own.
- The 7 Azure Monitor metrics bring data. Of the Prometheus queries, 52 with
  data and 10 empty, all expected: `top_query_info` (text off), Always On
  (General Purpose does not use an AG) and log held for a non-routine reason
  (none at the moment).
- The correlations show the load and the HML-051 failover in both sources.
  The storage used from Azure Monitor and from the exporter overlap
  (≈369 MB), because both read `sys.server_resource_stats`.
- Azure Monitor sends the metric's unit (`Count` for the storage in MB),
  which overrides the panel's; the dashboard pins the unit on the Azure
  Monitor series.
- Data source health check: Azure Monitor and Resource Graph OK; "No Log
  Analytics workspaces found" is expected (we do not use Log Analytics).

---

## HML-053 — Azure SQL Database reconnaissance: 10 of 19 collectors work; the rest depend on `sys.master_files`, `msdb` and server DMVs (validation)

- **Environment:** logical server `mssql-exporter-hml-sqldb-*` in `westus3`,
  public endpoint open only to the bastion and the operator; serverless
  database `hml` `GP_S_Gen5_2` through the free offer (pauses when the free
  limit runs out). `EngineEdition` 5, `ProductVersion` `12.0.2000.8`,
  `@@SERVERNAME` without the domain. Temporary exporter `0.13.2` on the
  bastion, with the opt-in collectors on
- **Found on:** 2026-10-27
- **Severity:** — (reconnaissance)
- **Status:** deferred (ROADMAP: SQL Database mode)

**Logins tested** (Azure requires a complex password: the `123` of the other
environments is refused)

- Contained user `usr_exporter` in the database, with `VIEW DATABASE STATE`
  and `VIEW DEFINITION`.
- Login `usr_exporter_srv` in `master`, member of `##MS_ServerStateReader##`,
  with a user in the database: the only difference is `tempdb`, which the
  contained user cannot read (`VIEW DATABASE PERFORMANCE STATE permission
  denied in database 'tempdb'`).

**Collectors**

| Situation | Collectors |
|---|---|
| `success 1` | `sessions`, `active_queries`, `blocking`, `perf_counters`, `schedulers`, `memory_clerks`, `wait_stats`, `top_queries`, `query_store`, `index_fragmentation` |
| `success 1` only with the server login | `tempdb` |
| `sys.master_files` does not exist (208) | `volumes`, `io_file_stats`, `databases` (state and recovery model come out before the failure) |
| `sys.dm_os_sys_memory` does not exist (208) | `server_info` (edition, version and servername come out) |
| object does not exist (208) | `database_mirroring`, `availability_groups` |
| `msdb` inaccessible (40515: `Reference to database and/or server name … is not supported`) | `backups`, `agent_jobs` |

- ≈5,900 series for an almost empty database: `wait_stats` reads
  `sys.dm_os_wait_stats` and emits 1,760 wait types (×3 metrics);
  `memory_clerks`, 324 series, many with a GUID name.
- Not measured: the serverless auto-pause (60 min without a connection). A
  scrape every 15 s connects all the time, so the database does not pause and
  consumes the free limit.

**Equivalents read by the contained user**

| Instead of | Use |
|---|---|
| `sys.master_files` | `sys.database_files` + `FILEPROPERTY(name, 'SpaceUsed')` (data 32 MB/20 MB used, log 8 MB/1 MB) |
| `sys.dm_io_virtual_file_stats(NULL, NULL)` + `sys.master_files` | `sys.dm_io_virtual_file_stats(DB_ID(), NULL)` (2 files) |
| `sys.dm_os_wait_stats` | `sys.dm_db_wait_stats` (11 types with a value against 106 in `dm_os`) |
| host CPU/memory | `sys.dm_db_resource_stats`: CPU, data IO, log write, memory, workers and sessions as % of the service objective's limit, every 15 s, in UTC |

The design of the SQL Database mode went to the ROADMAP (Later).

---

## HML-054 — Automatic instance label, disk queue and idle waits (validation)

- **Environment:** `mssql-2025-agent` container (SQL Server 2025 on Linux),
  the home notebook (Windows 11 pt-BR, SQL Server 2025 Developer) and the
  Raspberry Pi 5 (Debian 13 arm64, OS collectors only); binary of the
  `feat/label-waits-diskqueue` branch on a separate port, without touching
  the service
- **Found on:** 2026-10-28
- **Severity:** — (validation)
- **Status:** resolved (v0.15.0)

**Automatic label (HML-029)**

Container stopped with `docker stop -t 60`, exporter started without
`instance_label`: `sqlserver_up{sqlserver_instance=""} 0`. Container started
again at 10:41:01; at 10:41:09 the log records `sql server instance
resolved` and, ≈12 s after starting it again,
`sqlserver_up{sqlserver_instance="6af7e96149bc"} 1`, with no series left with
an empty label and without restarting the exporter.

**Disk queue (HML-003)**

- Windows: with continuous writes of 32 MB to a temporary file, in the same
  10.6 s window, `rate(os_disk_write_time_seconds_total{device="0 C:"})` gave
  0.7154 and the Windows formula over
  `Win32_PerfRawData_PerfDisk_PhysicalDisk` (`ΔAvgDiskWriteQueueLength /
  ΔTimestamp_Sys100NS`) gave 0.7092. On pt-BR Windows `Get-Counter` requires
  the translated counter names; the raw WMI class does not.
- Linux (Pi): `os_disk_read_time_seconds_total`,
  `os_disk_write_time_seconds_total` and `os_disk_queue_length` on
  `mmcblk0`, counters growing between collections.

**Idle waits**

| Environment | Wait types (default) | With `include_idle` |
|---|---|---|
| Docker (2025, Linux) | 74 | 1,475 |
| Notebook (2025 Developer) | 77 | 1,475 |

With `include_idle` the notebook's top is `SOS_WORK_DISPATCHER`,
`LOGMGR_QUEUE`, `DISPATCHER_QUEUE_SEMAPHORE` and other background waits; by
default it becomes `PREEMPTIVE_OS_CRYPTOPS`, `PREEMPTIVE_OS_DEVICEOPS`,
`SOS_SCHEDULER_YIELD` and `ASYNC_NETWORK_IO`.

## HML-055 — `transaction_log` collector (`sys.dm_db_log_stats`): VLFs and log since the last log backup (validation)

- **Environment:** `mssql-2025-agent` container (SQL Server 2025 on Linux, as
  `sa`) and the home notebook (SQL Server 2025 Developer, `usr_exporter`
  login with the README grants); binary of the `feat/transaction-log-stats`
  branch
- **Found on:** 2026-10-28
- **Severity:** — (validation)
- **Status:** resolved (v0.16.0)

A ROADMAP item born from HML-026 (on RDS the `Percent Log Used` counter has
no instance for the system databases). New `transaction_log` collector, on
by default:

| Metric | Source |
|---|---|
| `sqlserver_database_log_vlf_count` / `_active_vlf_count` | `total_vlf_count` / `active_vlf_count` |
| `sqlserver_database_log_active_bytes` | `active_log_size_mb` |
| `sqlserver_database_log_since_last_log_backup_bytes` | `log_since_last_log_backup_mb` (`NULL` without a log backup chain → series absent) |
| `sqlserver_database_log_since_last_checkpoint_bytes` | `log_since_last_checkpoint_mb` |

- Only `ONLINE` databases with `HAS_DBACCESS = 1`; local AG secondary
  replicas are left out (not validated on a real AG in this phase: the AWS
  environments were already destroyed).
- Container: `LogReuseTest` (FULL, reason `LOG_BACKUP`) with 12 VLFs, 8
  active and ≈87 MB since the last log backup; after a `BACKUP LOG` (file
  `/var/opt/mssql/data/LogReuseTest_hml055.trn` in the container), ≈76 KB
  since the backup and 1 active VLF. `master` (SIMPLE) without the
  `since_last_log_backup` series, as expected.
- Notebook: `success 1` with the exporter's login, 5 databases, 0.43 s.
- Main dashboard: "VLF Count by Database" and "Log Since Last Log Backup"
  panels in the Databases section.

## HML-056 — `missing_indexes` collector (opt-in): optimizer suggestions (validation)

- **Environment:** `mssql-2025-agent` container (SQL Server 2025 on Linux)
  and the home notebook; binary of the `feat/missing-indexes-replication`
  branch
- **Found on:** 2026-10-28
- **Severity:** — (validation)
- **Status:** resolved (v0.17.0)

Test database `MissingIdxTest` in the container (table `dbo.Orders`, 200
thousand rows, primary key only). Queries with a `TRIVIAL` plan produce no
suggestion; with `GROUP BY`/`ORDER BY` and a self-join (full optimization),
20 executions of each produced two:

| Suggested columns | Seeks | Impact | Improvement measure |
|---|---|---|---|
| equality `[CustomerID]` | 20 | 97.3 % | 93.5 |
| equality `[Status]`, inequality `[OrderDate]`, includes `[CustomerID], [Amount]` | 20 | 95.9 % | 92.5 |

- The collector exposed both, with `sqlserver_missing_index_info` carrying
  the columns, `success 1` in 3 ms.
- Permission: `VIEW SERVER STATE` alone is enough for the rows, but
  `OBJECT_NAME`/`OBJECT_SCHEMA_NAME` return `NULL` without access to the
  database; with `CONNECT ANY DATABASE` the table name shows up (README).
- Notebook (`usr_exporter`): `success 1`, no suggestion.

---

## HML-057 — `replication` collector (opt-in): transactional replication agents on the Distributor (validation)

- **Environment:** `mssql-2025-agent` container (SQL Server 2025 on Linux,
  with Agent) configured as a local Distributor, Publisher and Subscriber:
  `distribution` database, transactional publication `ItemsPub`
  (`ReplPub.dbo.Items`) and a push subscription to `ReplSub`; binary of the
  `feat/missing-indexes-replication` branch
- **Found on:** 2026-10-28
- **Severity:** — (validation)
- **Status:** resolved (v0.17.0)

**Setup** — `sp_adddistributor`, `sp_adddistributiondb`,
`sp_adddistpublisher` (folder `/var/opt/mssql/ReplData`), publication with
`immediate_sync`, Log Reader and Snapshot Agent with SQL authentication (on
Linux the agents do not use Windows auth). The initial snapshot and 500 rows
reached the subscriber.

**Findings that shaped the collector's design**

- `immediate_sync` creates two "virtual" subscriptions (`subscriber_id` < 0,
  `subscriber_db = 'virtual'`) with "pending" commands that are never
  delivered: they are left out.
- The publisher/subscriber ids of `MSdistribution_agents` match
  `distribution.dbo.MSreplservers`, not `master.sys.servers` (there id 1 is
  `repl_distributor`).
- The `replmonitor` role of `distribution` does not give SELECT on the tables
  (only on the monitoring procedures): the README asks for a user with SELECT
  on the six tables/views, validated with a `usr_repl_probe` test login.
- A stopped agent keeps its last status (2 = succeeded after an
  `sp_stop_job`), so the alert must use the pending commands and the age of
  the last history.

**Validation**

| Moment | Distribution agent: status / latency | Pending |
|---|---|---|
| Stable | 4 (idle) / 1.7 s | 0 |
| Agent stopped (`sp_stop_job`) + 300 rows inserted | 2 | 300 |
| Agent started again (`sp_start_job`), 25 s later | 3 (in progress) / 14.8 s | 0, 800 rows on the subscriber |

- Log Reader: status 4/3, latency 4 s, without the pending series.
- `success 1` with `sa` and with the least-privileged login; on the notebook
  (not a Distributor) `success 1` without series.
- Not tested: merge and snapshot, remote Distributor, Oracle publications.

---

## HML-058 — `agent_jobs` and `replication` convert old times with the server's current offset

- **Environment:** any server whose time zone has daylight saving time, or
  that changed time zone; found in the external code review (not reproduced
  on a real server)
- **Found on:** 2026-10-28
- **Severity:** low
- **Status:** won't fix (documented in the README and in `CLAUDE.md`)

`sysjobhistory` and `MSdistribution_history`/`MSlogreader_history` store
local time without an offset. The collectors apply
`DATEPART(TZOFFSET, SYSDATETIMEOFFSET())`, the current offset, to an event
that may be from before the change. The error is the size of the change
(1 h for daylight saving time) and only applies to the last exposed event:
it goes away on the next job run or the next agent record. A continuous
agent records every few minutes; a weekly job can carry the error for up to
a week.

**Why not fix it:** the correct per-event conversion needs the server's
time zone (`AT TIME ZONE`), and `CURRENT_TIMEZONE_ID()` only exists from SQL
Server 2022 and on Azure SQL. On 2016 and 2019 the only option left would be
reading the registry with `xp_regread`, which needs more permission, and the
version would need a branch in the code. The gain does not pay for that
complexity. `backups` does not suffer from this: it uses the `time_zone`
stored in each backup (HML-010). `log_shipping` reads UTC columns.

---

## HML-059 — Coverage dashboard without the metrics of the newest collectors

- **Environment:** `dashboards/mssql-exporter-homolog-coverage.json`
- **Found on:** 2026-10-28, in the external reviews; count checked on
  2026-10-30 (descriptor names in the code against the JSON)
- **Severity:** low (affects only the homologation, not the dashboards for
  everyday use)
- **Status:** open

The "Metrics with data (of 142)" panel has a fixed denominator and claims to
cover every metric, but 29 of the 170 descriptors were left out, so a broken
new collector does not lower the coverage:

- `transaction_log`: `sqlserver_database_log_{vlf_count,active_vlf_count,active_bytes,since_last_log_backup_bytes,since_last_checkpoint_bytes}`
- `databases`: `sqlserver_database_log_reuse_wait`
- `volumes` on Azure SQL MI: `sqlserver_instance_storage_{reserved,used}_bytes`
- `missing_indexes`: the 6 `sqlserver_missing_index_*` series
- `query_store`: the 11 `sqlserver_query_store_*` series
- `replication`: `sqlserver_replication_agent_{run_status,last_history_timestamp_seconds,delivery_latency_seconds}` and `sqlserver_replication_pending_commands`

**Next step:** include those series and the right denominator in the next
homologation round that uses the dashboard. It is worth generating the list
from the descriptors so it does not age again.
