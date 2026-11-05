## What changes

<!-- What the PR does and why. Link the issue, if there is one. -->

## How it was tested

<!-- make check, make test-integration, make test-alerts, and against which
SQL Server (version, platform). For a new or changed metric, paste a sample
of the output (no host names or private data). -->

## Checklist

- [ ] `make check` passes
- [ ] Tests for new or changed collectors (`go-sqlmock`, see CONTRIBUTING.md)
- [ ] README updated for new or changed metrics, flags or grants
- [ ] `contrib/prometheus/alerts.yml` and its tests updated if a metric they use changed
