# Contributing to mssql-exporter

Thanks for helping. Bug reports, platform reports ("it works on X") and pull
requests are all welcome.

- **Bugs and feature requests:** open an issue with the matching form. Do not
  paste DSNs, passwords or host names.
- **Security vulnerabilities:** report them privately, as described in
  [SECURITY.md](SECURITY.md), not in a public issue.
- **Larger changes** (a new collector, a new platform, a behaviour change):
  open an issue first, so we can agree on the metrics and their cost before
  you write the code. [ROADMAP.md](ROADMAP.md) lists what is planned and
  what was left out on purpose.

## Development setup

You need Go 1.27.3 or later, and Docker for the integration and alert tests.

```bash
make build             # bin/mssql-exporter
make check             # gofmt, go vet and unit tests (go-sqlmock, no SQL Server needed)
make test-integration  # integration tests against a disposable SQL Server container
make test-alerts       # promtool check + unit tests of contrib/prometheus/alerts.yml
make vulncheck         # govulncheck
```

`make test-integration` creates a monitoring login with exactly the grants of
the README's [Permissions](README.md#permissions) section and requires every
collector to succeed as that login. If your collector needs another grant,
add it to the fixture in `integration/` **and** to the README.

The architecture (Exporter, `Scraper` interface, shared scrape deadline),
the coding conventions and the platform notes are described in
[CLAUDE.md](CLAUDE.md); it is written for AI assistants, but it is the most
complete description of the code.

## Pull requests

- One topic per PR, with a [conventional commit](https://www.conventionalcommits.org/)
  title (`feat:`, `fix:`, `docs:`, `ci:`, ...).
- `make check` must pass; CI also runs golangci-lint, govulncheck, gitleaks,
  the alert tests and the integration tests on SQL Server 2022 and 2025.
- New or changed collectors come with `go-sqlmock` tests (happy path, query
  error, scan error, empty result) and the README's metric documentation.
- If a metric used in `contrib/prometheus/alerts.yml` changes, update the
  rule and its tests in `alerts_test.yml` in the same PR.
- Say how you tested it: SQL Server version and platform, and a sample of
  the new metrics (without host names or other private data).
- Releases are cut by the maintainer by adding a `release:patch`,
  `release:minor` or `release:major` label before merging; you don't need to
  bump versions.

## Adding a New Collector

Follow these steps to add a new collector:

### 1. Create Collector File

Create `internal/collector/your_collector.go`:

```go
package collector

import (
    "context"
    "database/sql"
    "errors"
    "fmt"

    "github.com/prometheus/client_golang/prometheus"
)

type YourCollector struct {
    db         *sql.DB
    metricDesc *prometheus.Desc
}

func NewYourCollector(db *sql.DB) *YourCollector {
    return &YourCollector{
        db: db,
        metricDesc: prometheus.NewDesc(
            "sqlserver_your_metric",
            "Description of your metric",
            []string{"label1"},  // Label names
            nil,
        ),
    }
}

func (c *YourCollector) Name() string { return "your_collector" }

func (c *YourCollector) Describe(ch chan<- *prometheus.Desc) {
    ch <- c.metricDesc
}

func (c *YourCollector) Scrape(ctx context.Context, ch chan<- prometheus.Metric) error {
    rows, err := c.db.QueryContext(ctx, `
        SELECT column1, column2
        FROM sys.dm_your_dmv
    `)

    if err != nil {
        return fmt.Errorf("failed to query your_dmv: %w", err)
    }
    defer func() { _ = rows.Close() }()

    var scanErrs []error
    for rows.Next() {
        var label string
        var value float64
        if err := rows.Scan(&label, &value); err != nil {
            scanErrs = append(scanErrs, fmt.Errorf("failed to scan your_dmv row: %w", err))
            continue
        }

        ch <- prometheus.MustNewConstMetric(
            c.metricDesc,
            prometheus.GaugeValue,
            value,
            label,
        )
    }

    return errors.Join(append(scanErrs, rows.Err())...)
}
```

### 2. Register Collector

Edit `internal/collector/exporter.go` and add the collector to `scraperNames`
(this defines the registration order AND the `--collector.<name>` flag /
`collectors:` YAML key) and to `scraperFactories`:

```go
var scraperNames = []string{
    // ...
    "your_collector",
}

var scraperFactories = map[string]func(*sql.DB, Options) Scraper{
    // ...
    "your_collector": func(db *sql.DB, _ Options) Scraper { return NewYourCollector(db) },
}
```

No change in `main.go` is needed — the Exporter picks it up automatically,
including the enable/disable toggle.

- **Opt-in collectors** (extra grants, expensive queries or high cardinality)
  also go in `scrapersDisabledByDefault`.
- **Options** beyond on/off: add a struct to `Options` in `options.go` with
  its defaults in `DefaultOptions()`, and the flag/env/YAML plumbing and
  validation in `internal/config/config.go` (see `top_queries`).
- **Expensive or per-database collectors** run in the background on their
  own interval and `Scrape` serves the last result plus
  `_last_collection_{timestamp_seconds,duration_seconds,success,databases}`
  (see `index_fragmentation.go` and `query_store.go`). Per-database queries
  use `listAccessibleDatabases` (`HAS_DBACCESS`) and three-part names quoted
  with `quoteIdentifier`.

### 3. Test

1. Add `internal/collector/your_collector_test.go` using `newSQLMock` from
   `testutil_test.go` (covers happy path, query error, scan error, empty result)
2. Verify SQL query works against your SQL Server instance
3. Run `make build && make run`
4. Check metrics: `curl http://localhost:10060/metrics | grep sqlserver_your_metric`
5. Check self-metrics: `curl http://localhost:10060/metrics | grep 'scrape_collector.*your_collector'`
6. Run `make check` to ensure code quality
7. Document the metrics in README.md and, if useful, the Grafana dashboard
