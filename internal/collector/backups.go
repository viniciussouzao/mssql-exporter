package collector

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type SQLServerBackups struct {
	db               *sql.DB
	lastCompletedAt  *prometheus.Desc
	lastDurationSecs *prometheus.Desc
	lastSizeBytes    *prometheus.Desc
}

func NewBackupsCollector(db *sql.DB) *SQLServerBackups {
	labels := []string{"database_name", "backup_type"}
	return &SQLServerBackups{
		db: db,
		lastCompletedAt: prometheus.NewDesc(
			"sqlserver_backup_last_completed_timestamp_seconds",
			"Unix timestamp of the last completed backup",
			labels, nil,
		),
		lastDurationSecs: prometheus.NewDesc(
			"sqlserver_backup_last_duration_seconds",
			"Duration in seconds of the last completed backup",
			labels, nil,
		),
		lastSizeBytes: prometheus.NewDesc(
			"sqlserver_backup_last_size_bytes",
			"Size in bytes of the last completed backup (compressed when available)",
			labels, nil,
		),
	}
}

func (c *SQLServerBackups) Name() string {
	return "backups"
}

func (c *SQLServerBackups) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.lastCompletedAt
	ch <- c.lastDurationSecs
	ch <- c.lastSizeBytes
}

func (c *SQLServerBackups) Scrape(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Only backups taken by this instance count: RESTORE also writes the
	// restored backup into the local backupset (with the original
	// server_name), so a log shipping secondary or a mirror would otherwise
	// report the primary's backups as its own.
	//
	// The latest backup per database/type is picked by backup_set_id (an
	// identity), not by date: backupset dates are the server's local time,
	// so after a time zone change or a DST fall-back an older backup can
	// carry a later local date and hide newer ones.
	//
	// Backups written to the NUL device are discarded by SQL Server, so they
	// cannot be restored and must not count as the latest backup: Amazon RDS
	// takes them of model, msdb and rdsadmin at instance creation, and
	// BACKUP ... TO DISK = 'NUL' is a common way to cut the log chain. A
	// striped backup with one NUL stripe is unusable too. The filter runs
	// before the ROW_NUMBER, so the previous real backup is reported instead.
	//
	// The dates are converted to UTC because the metric is a Unix timestamp.
	// time_zone is the server's UTC offset (15-minute intervals) when the
	// backup ran, so it stays correct across DST changes; 127 means unknown,
	// in which case the current offset is used.
	rows, err := c.db.QueryContext(ctx, `
		SELECT
			bs.database_name,
			CASE bs.type
				WHEN 'D' THEN 'full'
				WHEN 'I' THEN 'differential'
				WHEN 'L' THEN 'log'
				ELSE bs.type
			END AS backup_type,
			DATEADD(MINUTE, -tz.utc_offset_minutes, bs.backup_finish_date) AS backup_finish_utc,
			DATEADD(MINUTE, -tz.utc_offset_minutes, bs.backup_start_date) AS backup_start_utc,
			COALESCE(bs.compressed_backup_size, bs.backup_size) AS size_bytes
		FROM (
			SELECT
				database_name, type, backup_finish_date, backup_start_date,
				compressed_backup_size, backup_size, time_zone,
				ROW_NUMBER() OVER (
					PARTITION BY database_name, type
					ORDER BY backup_set_id DESC
				) AS rn
			FROM msdb.dbo.backupset b
			WHERE backup_finish_date IS NOT NULL
				AND server_name = @@SERVERNAME
				AND NOT EXISTS (
					SELECT 1
					FROM msdb.dbo.backupmediafamily mf
					WHERE mf.media_set_id = b.media_set_id
						AND UPPER(RTRIM(mf.physical_device_name)) IN ('NUL', 'NUL:')
				)
		) bs
		CROSS APPLY (
			SELECT CASE
				WHEN bs.time_zone BETWEEN -48 AND 48 THEN bs.time_zone * 15
				ELSE DATEPART(TZOFFSET, SYSDATETIMEOFFSET())
			END AS utc_offset_minutes
		) tz
		WHERE bs.rn = 1
	`)
	if err != nil {
		return fmt.Errorf("failed to query backupset: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var scanErrs []error
	for rows.Next() {
		var (
			dbName     string
			backupType string
			finishDate time.Time
			startDate  time.Time
			sizeBytes  float64
		)

		if err := rows.Scan(&dbName, &backupType, &finishDate, &startDate, &sizeBytes); err != nil {
			scanErrs = append(scanErrs, fmt.Errorf("failed to scan backupset row: %w", err))
			continue
		}

		labels := []string{dbName, backupType}
		duration := finishDate.Sub(startDate).Seconds()

		ch <- prometheus.MustNewConstMetric(c.lastCompletedAt, prometheus.GaugeValue, float64(finishDate.Unix()), labels...)
		ch <- prometheus.MustNewConstMetric(c.lastDurationSecs, prometheus.GaugeValue, duration, labels...)
		ch <- prometheus.MustNewConstMetric(c.lastSizeBytes, prometheus.GaugeValue, sizeBytes, labels...)
	}

	return errors.Join(append(scanErrs, rows.Err())...)
}
