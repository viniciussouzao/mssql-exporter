# Installation

How to run the exporter as a service or a container, and a local Prometheus
and Grafana stack to try it. For a first run see the
[Quick Start](../README.md#quick-start).

## Running as a Windows service

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

## Running as a Linux service (systemd)

The Linux release archives include a hardened unit file,
`contrib/systemd/mssql-exporter.service`. It runs the exporter as a dedicated
system user with no login shell and no home directory, and reads
`/etc/mssql-exporter/config.yml`. As root, or with `sudo` on each command:

```bash
# Latest release (archives include the unit since v0.9.1); or set VERSION by hand
VERSION=$(curl -fsSL https://api.github.com/repos/viniciussouzao/mssql-exporter/releases/latest | sed -n 's/.*"tag_name": *"v\([^"]*\)".*/\1/p')
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

## Running in Docker

Each release from v0.19.1 publishes a multi-arch image (`linux/amd64`,
`linux/arm64`) to GitHub Container Registry, tagged `vX.Y.Z`, `X.Y` and
`latest`:

```bash
docker run -d --name mssql-exporter -p 10060:10060 \
  --env-file mssql-exporter.env \
  ghcr.io/viniciussouzao/mssql-exporter:latest
```

where `mssql-exporter.env` holds the DSN (an `-e` flag would leave it in the
shell history; both remain visible to `docker inspect`):

```
SQL_SERVER_DSN=sqlserver://mssql_exporter:<password>@sqlhost:1433?database=master&app+name=mssql-exporter
```

- Pin a version tag (`:vX.Y.Z`) instead of `latest` in production.
- To use a config file, mount it and pass `--config.file`, with absolute
  paths for the files it references (web config, certificates); the
  container user `exporter` must be able to read them:
  `-v /etc/mssql-exporter:/etc/mssql-exporter:ro ghcr.io/viniciussouzao/mssql-exporter:latest --config.file /etc/mssql-exporter/config.yml`
- The image runs as the non-root user `exporter` and has a `HEALTHCHECK` on
  `/healthz`; with TLS, basic auth or another listen address override it (see
  [Security](configuration.md#security)).
- Leave `os_collectors` disabled: they would describe the container, not the
  database server.
- `make docker-build` builds an image from source with the `Dockerfile`; the
  published one uses `Dockerfile.release`, which copies the binary GoReleaser
  built.

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
pool (since v0.14.0, see [Upgrading](upgrading.md#v0140)).

Ready-to-use dashboards are included in [`dashboards/`](../dashboards/): a
general one, one for **Amazon RDS** that shows the exporter's engine
metrics next to the host metrics from CloudWatch, one for **Google Cloud
SQL** that does the same with Cloud Monitoring, and one for **Azure SQL
Managed Instance** with Azure Monitor. See
[`dashboards/README.md`](../dashboards/README.md) for import instructions.
