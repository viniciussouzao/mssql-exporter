# Documentation

The documentation of mssql-exporter, grouped by what you are trying to do.
For a first run, start with the [Quick Start](../README.md#quick-start) in
the main README.

## Install and run

| Page | What it covers |
|---|---|
| [Installation](install.md) | Windows service, systemd unit, container image, a local Prometheus + Grafana stack, and importing the dashboards and alerting rules |
| [Configuration](configuration.md) | Flags, environment variables, the YAML config file and their precedence, connection pool, collector options, TLS and basic auth on the metrics endpoint |
| [Permissions](permissions.md) | The monitoring login: grants per collector and per platform (Amazon RDS, Google Cloud SQL, Azure SQL Managed Instance), and the DMVs read |
| [Encrypted connections](tls.md) | `encrypt=true` and `encrypt=strict` (TDS 8.0), certificates per platform |
| [Upgrading](upgrading.md) | Changes to check per version, newest first |

## Reference

| Page | What it covers |
|---|---|
| [Metrics](metrics.md) | Collectors (default and opt-in), self-metrics, and every exposed metric with its labels |
| [Alerting rules](alerting.md) | The rules in `contrib/prometheus/alerts.yml`, how to load and tune them |
| [Dashboards](../dashboards/README.md) | The Grafana dashboards (general, Amazon RDS, Google Cloud SQL, Azure SQL Managed Instance), how to import them, screenshots |

## How it was tested

| Page | What it covers |
|---|---|
| [Validation](validation.md) | Summary of the manual homologation: environments, scenarios, findings that changed the exporter, what is not covered |
| [Homologation backlog](validation/BACKLOG-HOMOLOG.en.md) | The full log of every finding (HML-001 to HML-059), with the commands, results and fixes. Original in Portuguese: [BACKLOG-HOMOLOG.md](validation/BACKLOG-HOMOLOG.md) |

## Project

| Page | What it covers |
|---|---|
| [Contributing](../CONTRIBUTING.md) | Development setup, tests, the pull request checklist, releases and adding a collector |
| [Roadmap](../ROADMAP.md) | What is done and the next candidates |
| [Security](../SECURITY.md) | How to report a vulnerability, supported versions |
