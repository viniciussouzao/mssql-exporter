package collector

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/viniciussouzao/mssql-exporter/internal/db"
)

type SQLServerPerfCounters struct {
	db                  *sql.DB
	batchRequests       *prometheus.Desc
	sqlCompilations     *prometheus.Desc
	sqlRecompilations   *prometheus.Desc
	pageLifeExpectancy  *prometheus.Desc
	bufferCacheHitRatio *prometheus.Desc
	lazyWrites          *prometheus.Desc
	deadlocks           *prometheus.Desc
	lockWaits           *prometheus.Desc
	userConnections     *prometheus.Desc
	pageSplits          *prometheus.Desc
}

func NewPerfCountersCollector(db *sql.DB) *SQLServerPerfCounters {
	return &SQLServerPerfCounters{
		db: db,
		batchRequests: prometheus.NewDesc(
			"sqlserver_batch_requests_total",
			"Total number of Transact-SQL command batches received since SQL Server started",
			nil, nil,
		),
		sqlCompilations: prometheus.NewDesc(
			"sqlserver_sql_compilations_total",
			"Total number of SQL compilations since SQL Server started",
			nil, nil,
		),
		sqlRecompilations: prometheus.NewDesc(
			"sqlserver_sql_recompilations_total",
			"Total number of SQL statement recompiles since SQL Server started",
			nil, nil,
		),
		pageLifeExpectancy: prometheus.NewDesc(
			"sqlserver_page_life_expectancy_seconds",
			"Number of seconds a page will stay in the buffer pool without references",
			nil, nil,
		),
		bufferCacheHitRatio: prometheus.NewDesc(
			"sqlserver_buffer_cache_hit_ratio",
			"Ratio of pages found in the buffer cache without having to read from disk (0-1)",
			nil, nil,
		),
		lazyWrites: prometheus.NewDesc(
			"sqlserver_lazy_writes_total",
			"Total number of buffers written by the buffer manager's lazy writer since SQL Server started",
			nil, nil,
		),
		deadlocks: prometheus.NewDesc(
			"sqlserver_deadlocks_total",
			"Total number of lock requests that resulted in a deadlock since SQL Server started",
			nil, nil,
		),
		lockWaits: prometheus.NewDesc(
			"sqlserver_lock_waits_total",
			"Total number of lock requests that required the caller to wait since SQL Server started",
			nil, nil,
		),
		userConnections: prometheus.NewDesc(
			"sqlserver_user_connections",
			"Number of users currently connected to SQL Server",
			nil, nil,
		),
		pageSplits: prometheus.NewDesc(
			"sqlserver_page_splits_total",
			"Total number of page splits that occurred as the result of overflowing index pages since SQL Server started",
			nil, nil,
		),
	}
}

func (c *SQLServerPerfCounters) Name() string {
	return "perf_counters"
}

func (c *SQLServerPerfCounters) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.batchRequests
	ch <- c.sqlCompilations
	ch <- c.sqlRecompilations
	ch <- c.pageLifeExpectancy
	ch <- c.bufferCacheHitRatio
	ch <- c.lazyWrites
	ch <- c.deadlocks
	ch <- c.lockWaits
	ch <- c.userConnections
	ch <- c.pageSplits
}

func (c *SQLServerPerfCounters) Scrape(ch chan<- prometheus.Metric) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.CurrentOptions().ConnTimeout)
	defer cancel()

	rows, err := c.db.QueryContext(ctx, `
		SELECT RTRIM(counter_name), cntr_value
		FROM sys.dm_os_performance_counters
		WHERE (object_name LIKE '%SQL Statistics%'
				AND counter_name IN ('Batch Requests/sec', 'SQL Compilations/sec', 'SQL Re-Compilations/sec'))
			OR (object_name LIKE '%Buffer Manager%'
				AND counter_name IN ('Page life expectancy', 'Buffer cache hit ratio', 'Buffer cache hit ratio base', 'Lazy writes/sec'))
			OR (object_name LIKE '%Locks%'
				AND counter_name IN ('Number of Deadlocks/sec', 'Lock Waits/sec')
				AND instance_name = '_Total')
			OR (object_name LIKE '%General Statistics%'
				AND counter_name = 'User Connections')
			OR (object_name LIKE '%Access Methods%'
				AND counter_name = 'Page Splits/sec')
	`)
	if err != nil {
		return fmt.Errorf("failed to query dm_os_performance_counters: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var hitRatio, hitRatioBase float64

	for rows.Next() {
		var counterName string
		var value float64
		if err := rows.Scan(&counterName, &value); err != nil {
			slog.Error("failed to scan performance counter row", slog.Any("error", err))
			continue
		}

		switch counterName {
		case "Batch Requests/sec":
			ch <- prometheus.MustNewConstMetric(c.batchRequests, prometheus.CounterValue, value)
		case "SQL Compilations/sec":
			ch <- prometheus.MustNewConstMetric(c.sqlCompilations, prometheus.CounterValue, value)
		case "SQL Re-Compilations/sec":
			ch <- prometheus.MustNewConstMetric(c.sqlRecompilations, prometheus.CounterValue, value)
		case "Page life expectancy":
			ch <- prometheus.MustNewConstMetric(c.pageLifeExpectancy, prometheus.GaugeValue, value)
		case "Buffer cache hit ratio":
			hitRatio = value
		case "Buffer cache hit ratio base":
			hitRatioBase = value
		case "Lazy writes/sec":
			ch <- prometheus.MustNewConstMetric(c.lazyWrites, prometheus.CounterValue, value)
		case "Number of Deadlocks/sec":
			ch <- prometheus.MustNewConstMetric(c.deadlocks, prometheus.CounterValue, value)
		case "Lock Waits/sec":
			ch <- prometheus.MustNewConstMetric(c.lockWaits, prometheus.CounterValue, value)
		case "User Connections":
			ch <- prometheus.MustNewConstMetric(c.userConnections, prometheus.GaugeValue, value)
		case "Page Splits/sec":
			ch <- prometheus.MustNewConstMetric(c.pageSplits, prometheus.CounterValue, value)
		}
	}

	// "Buffer cache hit ratio" is a fraction counter: the raw value only makes
	// sense divided by its "base" counter, so it is emitted after the scan.
	if hitRatioBase > 0 {
		ch <- prometheus.MustNewConstMetric(c.bufferCacheHitRatio, prometheus.GaugeValue, hitRatio/hitRatioBase)
	}

	return rows.Err()
}
