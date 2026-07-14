package collector

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/viniciussouzao/mssql-exporter/internal/db"
)

type SQLServerDatabases struct {
	db             *sql.DB
	state          *prometheus.Desc
	recoveryModel  *prometheus.Desc
	dataSizeBytes  *prometheus.Desc
	logSizeBytes   *prometheus.Desc
	logUsedPercent *prometheus.Desc
}

func NewDatabasesCollector(db *sql.DB) *SQLServerDatabases {
	return &SQLServerDatabases{
		db: db,
		state: prometheus.NewDesc(
			"sqlserver_database_state",
			"Current state of the database (1 for the current state, labeled by state)",
			[]string{"db_name", "state"}, nil,
		),
		recoveryModel: prometheus.NewDesc(
			"sqlserver_database_recovery_model",
			"Recovery model of the database (1 for the current model, labeled by recovery_model)",
			[]string{"db_name", "recovery_model"}, nil,
		),
		dataSizeBytes: prometheus.NewDesc(
			"sqlserver_database_data_size_bytes",
			"Total allocated size of data files per database in bytes",
			[]string{"db_name"}, nil,
		),
		logSizeBytes: prometheus.NewDesc(
			"sqlserver_database_log_size_bytes",
			"Total allocated size of log files per database in bytes",
			[]string{"db_name"}, nil,
		),
		logUsedPercent: prometheus.NewDesc(
			"sqlserver_database_log_used_percent",
			"Percentage of the transaction log currently in use per database",
			[]string{"db_name"}, nil,
		),
	}
}

func (c *SQLServerDatabases) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.state
	ch <- c.recoveryModel
	ch <- c.dataSizeBytes
	ch <- c.logSizeBytes
	ch <- c.logUsedPercent
}

func (c *SQLServerDatabases) Collect(ch chan<- prometheus.Metric) {
	c.collectStates(ch)
	c.collectFileSizes(ch)
	c.collectLogUsage(ch)
}

func (c *SQLServerDatabases) collectStates(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), db.CurrentOptions().ConnTimeout)
	defer cancel()

	rows, err := c.db.QueryContext(ctx, `
		SELECT name, state_desc, recovery_model_desc
		FROM sys.databases
	`)
	if err != nil {
		slog.Error("failed to query sys.databases", slog.Any("error", err))
		return
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var name, stateDesc, recoveryModelDesc string
		if err := rows.Scan(&name, &stateDesc, &recoveryModelDesc); err != nil {
			slog.Error("failed to scan sys.databases row", slog.Any("error", err))
			continue
		}

		ch <- prometheus.MustNewConstMetric(c.state, prometheus.GaugeValue, 1, name, stateDesc)
		ch <- prometheus.MustNewConstMetric(c.recoveryModel, prometheus.GaugeValue, 1, name, recoveryModelDesc)
	}

	if err := rows.Err(); err != nil {
		slog.Error("error iterating sys.databases rows", slog.Any("error", err))
	}
}

func (c *SQLServerDatabases) collectFileSizes(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), db.CurrentOptions().ConnTimeout)
	defer cancel()

	rows, err := c.db.QueryContext(ctx, `
		SELECT
			DB_NAME(database_id) AS db_name,
			type,
			SUM(CAST(size AS BIGINT)) * 8192 AS size_bytes
		FROM sys.master_files
		WHERE type IN (0, 1)
		GROUP BY database_id, type
	`)
	if err != nil {
		slog.Error("failed to query sys.master_files for database sizes", slog.Any("error", err))
		return
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var dbName sql.NullString
		var fileType int
		var sizeBytes float64
		if err := rows.Scan(&dbName, &fileType, &sizeBytes); err != nil {
			slog.Error("failed to scan database file size row", slog.Any("error", err))
			continue
		}
		// DB_NAME can return NULL for databases in transient states (drop/restore in progress).
		if !dbName.Valid {
			continue
		}

		switch fileType {
		case 0:
			ch <- prometheus.MustNewConstMetric(c.dataSizeBytes, prometheus.GaugeValue, sizeBytes, dbName.String)
		case 1:
			ch <- prometheus.MustNewConstMetric(c.logSizeBytes, prometheus.GaugeValue, sizeBytes, dbName.String)
		}
	}

	if err := rows.Err(); err != nil {
		slog.Error("error iterating database file size rows", slog.Any("error", err))
	}
}

func (c *SQLServerDatabases) collectLogUsage(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), db.CurrentOptions().ConnTimeout)
	defer cancel()

	rows, err := c.db.QueryContext(ctx, `
		SELECT RTRIM(instance_name), cntr_value
		FROM sys.dm_os_performance_counters
		WHERE object_name LIKE '%Databases%'
			AND counter_name = 'Percent Log Used'
			AND instance_name <> '_Total'
	`)
	if err != nil {
		slog.Error("failed to query log usage performance counters", slog.Any("error", err))
		return
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var dbName string
		var value float64
		if err := rows.Scan(&dbName, &value); err != nil {
			slog.Error("failed to scan log usage row", slog.Any("error", err))
			continue
		}

		ch <- prometheus.MustNewConstMetric(c.logUsedPercent, prometheus.GaugeValue, value, dbName)
	}

	if err := rows.Err(); err != nil {
		slog.Error("error iterating log usage rows", slog.Any("error", err))
	}
}
