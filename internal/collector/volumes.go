package collector

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

type SQLServerVolumes struct {
	db             *sql.DB
	totalBytes     *prometheus.Desc
	availableBytes *prometheus.Desc
}

func NewVolumesCollector(db *sql.DB) *SQLServerVolumes {
	labels := []string{"volume_mount_point"}
	return &SQLServerVolumes{
		db: db,
		totalBytes: prometheus.NewDesc(
			"sqlserver_volume_total_bytes",
			"Total size in bytes of the volume hosting database files",
			labels, nil,
		),
		availableBytes: prometheus.NewDesc(
			"sqlserver_volume_available_bytes",
			"Available free space in bytes on the volume hosting database files",
			labels, nil,
		),
	}
}

func (c *SQLServerVolumes) Name() string {
	return "volumes"
}

func (c *SQLServerVolumes) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.totalBytes
	ch <- c.availableBytes
}

// Scrape reports disk space per volume as seen by SQL Server itself, which
// also works when the exporter runs on a different machine than the server.
func (c *SQLServerVolumes) Scrape(ctx context.Context, ch chan<- prometheus.Metric) error {
	// volume_mount_point can be NULL on SQL Server on Linux; fall back to the
	// logical volume name so the series still has a usable label.
	rows, err := c.db.QueryContext(ctx, `
		SELECT DISTINCT
			COALESCE(NULLIF(vs.volume_mount_point, ''), vs.logical_volume_name, ''),
			vs.total_bytes,
			vs.available_bytes
		FROM sys.master_files mf
		CROSS APPLY sys.dm_os_volume_stats(mf.database_id, mf.file_id) vs
	`)
	if err != nil {
		return fmt.Errorf("failed to query dm_os_volume_stats: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var scanErrs []error
	for rows.Next() {
		var mountPoint string
		var totalBytes, availableBytes float64
		if err := rows.Scan(&mountPoint, &totalBytes, &availableBytes); err != nil {
			scanErrs = append(scanErrs, fmt.Errorf("failed to scan volume stats row: %w", err))
			continue
		}

		ch <- prometheus.MustNewConstMetric(c.totalBytes, prometheus.GaugeValue, totalBytes, mountPoint)
		ch <- prometheus.MustNewConstMetric(c.availableBytes, prometheus.GaugeValue, availableBytes, mountPoint)
	}

	return errors.Join(append(scanErrs, rows.Err())...)
}
