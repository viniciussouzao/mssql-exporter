package collector

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/viniciussouzao/mssql-exporter/internal/db"
)

type SQLServerWaitStats struct {
	db            *sql.DB
	waitTimeMs    *prometheus.Desc
	waitingTasks  *prometheus.Desc
	maxWaitTimeMs *prometheus.Desc
}

func NewWaitStatsCollector(db *sql.DB) *SQLServerWaitStats {
	return &SQLServerWaitStats{
		db: db,
		waitTimeMs: prometheus.NewDesc(
			"sqlserver_wait_time_ms",
			"wait time in milliseconds for SQL Server",
			[]string{"wait_type"},
			nil,
		),

		waitingTasks: prometheus.NewDesc(
			"sqlserver_waiting_tasks",
			"number of waiting tasks in SQL Server",
			[]string{"wait_type"},
			nil,
		),

		maxWaitTimeMs: prometheus.NewDesc(
			"sqlserver_max_wait_time_ms",
			"maximum wait time in milliseconds for SQL Server",
			[]string{"wait_type"},
			nil,
		),
	}
}

func (c *SQLServerWaitStats) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.waitTimeMs
	ch <- c.waitingTasks
	ch <- c.maxWaitTimeMs
}

func (c *SQLServerWaitStats) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), db.CurrentOptions().ConnTimeout)
	defer cancel()

	rows, err := c.db.QueryContext(ctx, `
	SELECT wait_type, wait_time_ms, waiting_tasks_count, max_wait_time_ms
	FROM sys.dm_os_wait_stats
	WHERE wait_type NOT LIKE 'SLEEP%' 
 	AND wait_type NOT LIKE 'XE%' 
  	AND wait_type NOT LIKE 'BROKER_%'
	`)

	if err != nil {
		slog.Error("failed to query sql server:", slog.Any("error", err))
		return
	}

	defer rows.Close()

	for rows.Next() {
		var waitType string
		var waitTimeMs float64
		var waitingTasks float64
		var maxWaitTimeMs float64

		err := rows.Scan(&waitType, &waitTimeMs, &waitingTasks, &maxWaitTimeMs)
		if err != nil {
			slog.Error("failed to scan row:", slog.Any("error", err))
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.waitTimeMs,
			prometheus.GaugeValue,
			waitTimeMs,
			waitType,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitingTasks,
			prometheus.GaugeValue,
			waitingTasks,
			waitType,
		)

		ch <- prometheus.MustNewConstMetric(
			c.maxWaitTimeMs,
			prometheus.GaugeValue,
			maxWaitTimeMs,
			waitType,
		)
	}
}
