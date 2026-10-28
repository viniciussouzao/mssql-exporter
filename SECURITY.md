# Security Policy

## Supported versions

Only the latest release gets fixes. Upgrade to it before reporting, and
check whether the problem is still there.

## Reporting a vulnerability

Please do not open a public issue. Report it privately through GitHub:
**Security → Report a vulnerability** on this repository (private
vulnerability reporting). Include:

- the version (`mssql-exporter --version`) and the platform (OS, SQL Server
  edition, managed service if any);
- what an attacker can do and what they need (network access to the
  metrics endpoint, a SQL Server login, local access to the host...);
- steps or a proof of concept to reproduce it.

You should get an answer within 7 days. Once a fix is released, the advisory
is published with credit to the reporter, unless you prefer otherwise.

## Scope

In scope: the exporter binary, the Docker image, the Windows service and the
systemd unit shipped in the release archives, and the example configuration.

Out of scope: SQL Server itself, Prometheus and Grafana, and deployments that
expose the metrics endpoint to untrusted networks without TLS and basic
authentication (see Security in the README).

Never include real credentials, DSNs or metric dumps with private data in a
report.
