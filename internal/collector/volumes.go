package collector

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

type SQLServerVolumes struct {
	db                   *sql.DB
	totalBytes           *prometheus.Desc
	availableBytes       *prometheus.Desc
	storageReservedBytes *prometheus.Desc
	storageUsedBytes     *prometheus.Desc
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
		storageReservedBytes: prometheus.NewDesc(
			"sqlserver_instance_storage_reserved_bytes",
			"Storage reserved for the instance in bytes (Azure SQL Managed Instance only)",
			nil, nil,
		),
		storageUsedBytes: prometheus.NewDesc(
			"sqlserver_instance_storage_used_bytes",
			"Instance storage used by database files in bytes (Azure SQL Managed Instance only)",
			nil, nil,
		),
	}
}

func (c *SQLServerVolumes) Name() string {
	return "volumes"
}

func (c *SQLServerVolumes) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.totalBytes
	ch <- c.availableBytes
	ch <- c.storageReservedBytes
	ch <- c.storageUsedBytes
}

// Scrape reports disk space per volume as seen by SQL Server itself, which
// also works when the exporter runs on a different machine than the server.
// On Azure SQL Managed Instance it reports the instance storage instead.
func (c *SQLServerVolumes) Scrape(ctx context.Context, ch chan<- prometheus.Metric) error {
	engineEdition, err := queryEngineEdition(ctx, c.db)
	if err != nil {
		return err
	}
	if engineEdition == engineEditionManagedInstance {
		return c.scrapeInstanceStorage(ctx, ch)
	}
	return c.scrapeVolumes(ctx, ch)
}

// scrapeInstanceStorage reads the latest 15-second sample of
// server_resource_stats (published a few minutes late). dm_os_volume_stats is
// not usable on Managed Instance: total_bytes there is the size of the files
// on the volume, smaller than available_bytes, not the volume capacity.
func (c *SQLServerVolumes) scrapeInstanceStorage(ctx context.Context, ch chan<- prometheus.Metric) error {
	rows, err := c.db.QueryContext(ctx, `
		SELECT TOP (1) reserved_storage_mb, storage_space_used_mb
		FROM master.sys.server_resource_stats
		ORDER BY end_time DESC
	`)
	if err != nil {
		return fmt.Errorf("failed to query server_resource_stats: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var scanErrs []error
	for rows.Next() {
		var reservedMB, usedMB float64
		if err := rows.Scan(&reservedMB, &usedMB); err != nil {
			scanErrs = append(scanErrs, fmt.Errorf("failed to scan server_resource_stats row: %w", err))
			continue
		}

		ch <- prometheus.MustNewConstMetric(c.storageReservedBytes, prometheus.GaugeValue, reservedMB*1024*1024)
		ch <- prometheus.MustNewConstMetric(c.storageUsedBytes, prometheus.GaugeValue, usedMB*1024*1024)
	}

	return errors.Join(append(scanErrs, rows.Err())...)
}

func (c *SQLServerVolumes) scrapeVolumes(ctx context.Context, ch chan<- prometheus.Metric) error {
	// SQL Server on Linux (Cloud SQL, the Docker images) returns NULL mount
	// point, volume id and logical volume name; fall back to the directory of
	// the database file so the series still has a usable label. Files in
	// different directories of the same file system then report the same
	// bytes under separate labels.
	// dm_os_volume_stats is evaluated once per database file, so under write
	// load files on the same volume can report different available_bytes in a
	// single execution. Group by mount point (DISTINCT is not enough) so each
	// volume yields exactly one row; otherwise the duplicate series make the
	// whole /metrics response fail.
	// Only ONLINE and RESTORING (AG and log shipping secondaries) databases:
	// dm_os_volume_stats waits on a database that is changing state and the
	// query runs into the scrape timeout (HML-019).
	rows, err := c.db.QueryContext(ctx, `
		SELECT mount_point, MAX(total_bytes), MIN(available_bytes)
		FROM (
			SELECT
				COALESCE(
					NULLIF(vs.volume_mount_point, ''),
					NULLIF(vs.logical_volume_name, ''),
					LEFT(mf.physical_name, LEN(mf.physical_name) - CHARINDEX('/', REVERSE(mf.physical_name)) + 1)
				) AS mount_point,
				vs.total_bytes,
				vs.available_bytes
			FROM sys.master_files mf
			JOIN sys.databases d ON d.database_id = mf.database_id
			CROSS APPLY sys.dm_os_volume_stats(mf.database_id, mf.file_id) vs
			WHERE d.state IN (0, 1)
		) v
		GROUP BY mount_point
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
