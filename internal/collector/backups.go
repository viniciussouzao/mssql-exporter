package collector

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/viniciussouzao/mssql-exporter/internal/db"
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

func (c *SQLServerBackups) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.lastCompletedAt
	ch <- c.lastDurationSecs
	ch <- c.lastSizeBytes
}

func (c *SQLServerBackups) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), db.CurrentOptions().ConnTimeout)
	defer cancel()

	rows, err := c.db.QueryContext(ctx, `
		SELECT
			bs.database_name,
			CASE bs.type
				WHEN 'D' THEN 'full'
				WHEN 'I' THEN 'differential'
				WHEN 'L' THEN 'log'
				ELSE bs.type
			END AS backup_type,
			bs.backup_finish_date,
			bs.backup_start_date,
			COALESCE(bs.compressed_backup_size, bs.backup_size) AS size_bytes
		FROM msdb.dbo.backupset bs
		INNER JOIN (
			SELECT database_name, type, MAX(backup_finish_date) AS max_finish
			FROM msdb.dbo.backupset
			WHERE backup_finish_date IS NOT NULL
			GROUP BY database_name, type
		) latest
			ON bs.database_name = latest.database_name
			AND bs.type        = latest.type
			AND bs.backup_finish_date = latest.max_finish
	`)
	if err != nil {
		slog.Error("failed to query backupset", slog.Any("error", err))
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			dbName      string
			backupType  string
			finishDate  time.Time
			startDate   time.Time
			sizeBytes   float64
		)

		if err := rows.Scan(&dbName, &backupType, &finishDate, &startDate, &sizeBytes); err != nil {
			slog.Error("failed to scan backupset row", slog.Any("error", err))
			continue
		}

		labels := []string{dbName, backupType}
		duration := finishDate.Sub(startDate).Seconds()

		ch <- prometheus.MustNewConstMetric(c.lastCompletedAt, prometheus.GaugeValue, float64(finishDate.Unix()), labels...)
		ch <- prometheus.MustNewConstMetric(c.lastDurationSecs, prometheus.GaugeValue, duration, labels...)
		ch <- prometheus.MustNewConstMetric(c.lastSizeBytes, prometheus.GaugeValue, sizeBytes, labels...)
	}
}
