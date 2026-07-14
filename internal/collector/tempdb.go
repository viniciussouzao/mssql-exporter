package collector

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

type SQLServerTempDB struct {
	db                 *sql.DB
	freeSpaceKB        *prometheus.Desc
	versionStoreKB     *prometheus.Desc
	internalObjectsKB  *prometheus.Desc
	userObjectsKB      *prometheus.Desc
	activeTempTables   *prometheus.Desc
	tempTablesCreated  *prometheus.Desc
}

func NewTempDBCollector(db *sql.DB) *SQLServerTempDB {
	return &SQLServerTempDB{
		db: db,
		freeSpaceKB: prometheus.NewDesc(
			"sqlserver_tempdb_free_space_kb",
			"Free space available in tempdb in kilobytes",
			nil, nil,
		),
		versionStoreKB: prometheus.NewDesc(
			"sqlserver_tempdb_version_store_kb",
			"Space used by the version store in tempdb in kilobytes",
			nil, nil,
		),
		internalObjectsKB: prometheus.NewDesc(
			"sqlserver_tempdb_internal_objects_kb",
			"Space used by internal objects in tempdb in kilobytes",
			nil, nil,
		),
		userObjectsKB: prometheus.NewDesc(
			"sqlserver_tempdb_user_objects_kb",
			"Space used by user objects (temp tables, table variables) in tempdb in kilobytes",
			nil, nil,
		),
		activeTempTables: prometheus.NewDesc(
			"sqlserver_tempdb_active_temp_tables",
			"Number of active temporary tables in tempdb",
			nil, nil,
		),
		tempTablesCreated: prometheus.NewDesc(
			"sqlserver_tempdb_temp_tables_created_total",
			"Total number of temporary tables created since SQL Server started",
			nil, nil,
		),
	}
}

func (c *SQLServerTempDB) Name() string {
	return "tempdb"
}

func (c *SQLServerTempDB) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.freeSpaceKB
	ch <- c.versionStoreKB
	ch <- c.internalObjectsKB
	ch <- c.userObjectsKB
	ch <- c.activeTempTables
	ch <- c.tempTablesCreated
}

func (c *SQLServerTempDB) Scrape(ctx context.Context, ch chan<- prometheus.Metric) error {
	return errors.Join(
		c.collectFileSpaceUsage(ctx, ch),
		c.collectPerfCounters(ctx, ch),
	)
}

func (c *SQLServerTempDB) collectFileSpaceUsage(ctx context.Context, ch chan<- prometheus.Metric) error {
	row := c.db.QueryRowContext(ctx, `
		SELECT
			SUM(unallocated_extent_page_count)       * 8,
			SUM(version_store_reserved_page_count)   * 8,
			SUM(internal_object_reserved_page_count) * 8,
			SUM(user_object_reserved_page_count)     * 8
		FROM tempdb.sys.dm_db_file_space_usage
	`)

	var freeKB, versionStoreKB, internalKB, userKB float64
	if err := row.Scan(&freeKB, &versionStoreKB, &internalKB, &userKB); err != nil {
		return fmt.Errorf("failed to query tempdb file space usage: %w", err)
	}

	ch <- prometheus.MustNewConstMetric(c.freeSpaceKB, prometheus.GaugeValue, freeKB)
	ch <- prometheus.MustNewConstMetric(c.versionStoreKB, prometheus.GaugeValue, versionStoreKB)
	ch <- prometheus.MustNewConstMetric(c.internalObjectsKB, prometheus.GaugeValue, internalKB)
	ch <- prometheus.MustNewConstMetric(c.userObjectsKB, prometheus.GaugeValue, userKB)

	return nil
}

func (c *SQLServerTempDB) collectPerfCounters(ctx context.Context, ch chan<- prometheus.Metric) error {
	rows, err := c.db.QueryContext(ctx, `
		SELECT RTRIM(counter_name), cntr_value
		FROM sys.dm_os_performance_counters
		WHERE object_name LIKE '%General Statistics%'
		AND RTRIM(counter_name) IN ('Active Temp Tables', 'Temp Tables Creation Rate')
	`)
	if err != nil {
		return fmt.Errorf("failed to query tempdb performance counters: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var counterName string
		var value float64
		if err := rows.Scan(&counterName, &value); err != nil {
			slog.Error("failed to scan tempdb perf counter row", slog.Any("error", err))
			continue
		}

		switch counterName {
		case "Active Temp Tables":
			ch <- prometheus.MustNewConstMetric(c.activeTempTables, prometheus.GaugeValue, value)
		case "Temp Tables Creation Rate":
			ch <- prometheus.MustNewConstMetric(c.tempTablesCreated, prometheus.CounterValue, value)
		}
	}

	return rows.Err()
}
