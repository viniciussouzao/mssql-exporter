# Alerting rules

[`contrib/prometheus/alerts.yml`](../contrib/prometheus/alerts.yml) (also in
the release archives) has example alerting rules built from the
expressions of [Exposed Metrics](metrics.md#exposed-metrics): exporter and SQL Server
reachability, failing collectors, databases in `SUSPECT`/`RECOVERY_PENDING`,
a log that cannot be truncated or whose log backups stopped, stale full and
log backups, many VLFs, full volumes, blocking, long requests and
transactions, CPU pressure, Availability Groups, mirroring, log shipping,
failed Agent jobs, replication agents (failed, stale history, growing
queue, latency) and the Query Store and background collections. Load it
with `rule_files:` and tune thresholds and `for:` windows to your workload;
rules over opt-in collectors never fire while they are disabled, and the log
backup age rule also needs the `databases` collector (it skips databases in
`SIMPLE` recovery). The
`MSSQLExporterDown` rule selects `job="sqlserver_exporter"`, the job name of
the example above.

```yaml
rule_files:
  - alerts.yml
```

`make test-alerts` validates the file and runs its unit tests
([`alerts_test.yml`](../contrib/prometheus/alerts_test.yml)) with `promtool`
from the Prometheus image; CI runs it on every pull request.
