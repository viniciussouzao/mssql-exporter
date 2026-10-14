package collector

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

// SQLServerLogShipping reads the log shipping monitor tables in msdb, which
// SQL Server keeps locally on the primary and the secondary (and on the
// monitor server, when one is configured). It needs SELECT on both tables,
// so it is disabled by default (see scrapersDisabledByDefault).
type SQLServerLogShipping struct {
	db                   *sql.DB
	primaryLastBackup    *prometheus.Desc
	primaryThreshold     *prometheus.Desc
	secondaryLastCopied  *prometheus.Desc
	secondaryLastRestore *prometheus.Desc
	secondaryLatency     *prometheus.Desc
	secondaryThreshold   *prometheus.Desc
}

func NewLogShippingCollector(db *sql.DB) *SQLServerLogShipping {
	primaryLabels := []string{"primary_server", "primary_database"}
	secondaryLabels := []string{"secondary_server", "secondary_database", "primary_server", "primary_database"}
	return &SQLServerLogShipping{
		db: db,
		primaryLastBackup: prometheus.NewDesc(
			"sqlserver_log_shipping_primary_last_backup_timestamp_seconds",
			"Unix timestamp (UTC) of the last log backup taken by the log shipping backup job on the primary",
			primaryLabels, nil,
		),
		primaryThreshold: prometheus.NewDesc(
			"sqlserver_log_shipping_primary_backup_threshold_seconds",
			"Configured maximum time between log shipping backups before an alert is raised",
			primaryLabels, nil,
		),
		secondaryLastCopied: prometheus.NewDesc(
			"sqlserver_log_shipping_secondary_last_copied_timestamp_seconds",
			"Unix timestamp (UTC) of the last log backup file copied to the secondary",
			secondaryLabels, nil,
		),
		secondaryLastRestore: prometheus.NewDesc(
			"sqlserver_log_shipping_secondary_last_restored_timestamp_seconds",
			"Unix timestamp (UTC) of the last log backup restored on the secondary",
			secondaryLabels, nil,
		),
		secondaryLatency: prometheus.NewDesc(
			"sqlserver_log_shipping_secondary_last_restored_latency_seconds",
			"Time between the creation of the last restored log backup on the primary and its restore on the secondary (minute precision)",
			secondaryLabels, nil,
		),
		secondaryThreshold: prometheus.NewDesc(
			"sqlserver_log_shipping_secondary_restore_threshold_seconds",
			"Configured maximum time between log shipping restores before an alert is raised",
			secondaryLabels, nil,
		),
	}
}

func (c *SQLServerLogShipping) Name() string {
	return "log_shipping"
}

func (c *SQLServerLogShipping) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.primaryLastBackup
	ch <- c.primaryThreshold
	ch <- c.secondaryLastCopied
	ch <- c.secondaryLastRestore
	ch <- c.secondaryLatency
	ch <- c.secondaryThreshold
}

// Scrape emits nothing (and no error) when log shipping is not configured:
// the monitor tables exist in every msdb and are simply empty.
func (c *SQLServerLogShipping) Scrape(ctx context.Context, ch chan<- prometheus.Metric) error {
	return errors.Join(
		c.collectPrimaries(ctx, ch),
		c.collectSecondaries(ctx, ch),
	)
}

func (c *SQLServerLogShipping) collectPrimaries(ctx context.Context, ch chan<- prometheus.Metric) error {
	// The *_utc columns are already UTC, so no time zone conversion is
	// needed (unlike backupset/sysjobhistory, see backups and agent_jobs).
	rows, err := c.db.QueryContext(ctx, `
		SELECT
			primary_server,
			primary_database,
			last_backup_date_utc,
			backup_threshold
		FROM msdb.dbo.log_shipping_monitor_primary
	`)
	if err != nil {
		return fmt.Errorf("failed to query log_shipping_monitor_primary: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var scanErrs []error
	for rows.Next() {
		var server, database string
		// last_backup_date_utc is NULL until the first backup job run.
		var lastBackup sql.NullTime
		var thresholdMinutes sql.NullFloat64
		if err := rows.Scan(&server, &database, &lastBackup, &thresholdMinutes); err != nil {
			scanErrs = append(scanErrs, fmt.Errorf("failed to scan log_shipping_monitor_primary row: %w", err))
			continue
		}

		if lastBackup.Valid {
			ch <- prometheus.MustNewConstMetric(c.primaryLastBackup, prometheus.GaugeValue, float64(lastBackup.Time.Unix()), server, database)
		}
		if thresholdMinutes.Valid {
			ch <- prometheus.MustNewConstMetric(c.primaryThreshold, prometheus.GaugeValue, thresholdMinutes.Float64*60, server, database)
		}
	}

	return errors.Join(append(scanErrs, rows.Err())...)
}

func (c *SQLServerLogShipping) collectSecondaries(ctx context.Context, ch chan<- prometheus.Metric) error {
	rows, err := c.db.QueryContext(ctx, `
		SELECT
			secondary_server,
			secondary_database,
			primary_server,
			primary_database,
			last_copied_date_utc,
			last_restored_date_utc,
			last_restored_latency,
			restore_threshold
		FROM msdb.dbo.log_shipping_monitor_secondary
	`)
	if err != nil {
		return fmt.Errorf("failed to query log_shipping_monitor_secondary: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var scanErrs []error
	for rows.Next() {
		var secondaryServer, secondaryDatabase, primaryServer, primaryDatabase string
		// Dates and latency are NULL until the first copy/restore job run.
		var lastCopied, lastRestored sql.NullTime
		var latencyMinutes, thresholdMinutes sql.NullFloat64
		if err := rows.Scan(&secondaryServer, &secondaryDatabase, &primaryServer, &primaryDatabase,
			&lastCopied, &lastRestored, &latencyMinutes, &thresholdMinutes); err != nil {
			scanErrs = append(scanErrs, fmt.Errorf("failed to scan log_shipping_monitor_secondary row: %w", err))
			continue
		}

		labels := []string{secondaryServer, secondaryDatabase, primaryServer, primaryDatabase}
		if lastCopied.Valid {
			ch <- prometheus.MustNewConstMetric(c.secondaryLastCopied, prometheus.GaugeValue, float64(lastCopied.Time.Unix()), labels...)
		}
		if lastRestored.Valid {
			ch <- prometheus.MustNewConstMetric(c.secondaryLastRestore, prometheus.GaugeValue, float64(lastRestored.Time.Unix()), labels...)
		}
		if latencyMinutes.Valid {
			ch <- prometheus.MustNewConstMetric(c.secondaryLatency, prometheus.GaugeValue, latencyMinutes.Float64*60, labels...)
		}
		if thresholdMinutes.Valid {
			ch <- prometheus.MustNewConstMetric(c.secondaryThreshold, prometheus.GaugeValue, thresholdMinutes.Float64*60, labels...)
		}
	}

	return errors.Join(append(scanErrs, rows.Err())...)
}
