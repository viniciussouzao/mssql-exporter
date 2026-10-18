package collector

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
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

func (c *SQLServerDatabases) Name() string {
	return "databases"
}

func (c *SQLServerDatabases) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.state
	ch <- c.recoveryModel
	ch <- c.dataSizeBytes
	ch <- c.logSizeBytes
	ch <- c.logUsedPercent
}

func (c *SQLServerDatabases) Scrape(ctx context.Context, ch chan<- prometheus.Metric) error {
	return errors.Join(
		c.collectStates(ctx, ch),
		c.collectFileSizes(ctx, ch),
		c.collectLogUsage(ctx, ch),
	)
}

func (c *SQLServerDatabases) collectStates(ctx context.Context, ch chan<- prometheus.Metric) error {
	rows, err := c.db.QueryContext(ctx, `
		SELECT name, state_desc, recovery_model_desc
		FROM sys.databases
	`)
	if err != nil {
		return fmt.Errorf("failed to query sys.databases: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var scanErrs []error
	for rows.Next() {
		var name, stateDesc, recoveryModelDesc string
		if err := rows.Scan(&name, &stateDesc, &recoveryModelDesc); err != nil {
			scanErrs = append(scanErrs, fmt.Errorf("failed to scan sys.databases row: %w", err))
			continue
		}

		ch <- prometheus.MustNewConstMetric(c.state, prometheus.GaugeValue, 1, name, stateDesc)
		ch <- prometheus.MustNewConstMetric(c.recoveryModel, prometheus.GaugeValue, 1, name, recoveryModelDesc)
	}

	return errors.Join(append(scanErrs, rows.Err())...)
}

func (c *SQLServerDatabases) collectFileSizes(ctx context.Context, ch chan<- prometheus.Metric) error {
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
		return fmt.Errorf("failed to query sys.master_files for database sizes: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var scanErrs []error
	for rows.Next() {
		var dbName sql.NullString
		var fileType int
		var sizeBytes float64
		if err := rows.Scan(&dbName, &fileType, &sizeBytes); err != nil {
			scanErrs = append(scanErrs, fmt.Errorf("failed to scan database file size row: %w", err))
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

	return errors.Join(append(scanErrs, rows.Err())...)
}

// collectLogUsage reads the "Percent Log Used" counter of the databases listed
// in sys.databases. The counter also has instances for hidden databases
// (mssqlsystemresource, and model_msdb/model_replicatedmaster on SQL Server
// 2025) that no other metric reports, plus _Total; the join drops them. On
// Amazon RDS the counter has no instance for master, model, msdb and rdsadmin,
// so those databases get no series there.
func (c *SQLServerDatabases) collectLogUsage(ctx context.Context, ch chan<- prometheus.Metric) error {
	rows, err := c.db.QueryContext(ctx, `
		SELECT d.name, pc.cntr_value
		FROM sys.dm_os_performance_counters pc
		INNER JOIN sys.databases d
			ON d.name = RTRIM(pc.instance_name) COLLATE DATABASE_DEFAULT
		WHERE pc.object_name LIKE '%Databases%'
			AND pc.counter_name = 'Percent Log Used'
	`)
	if err != nil {
		return fmt.Errorf("failed to query log usage performance counters: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var scanErrs []error
	for rows.Next() {
		var dbName string
		var value float64
		if err := rows.Scan(&dbName, &value); err != nil {
			scanErrs = append(scanErrs, fmt.Errorf("failed to scan log usage row: %w", err))
			continue
		}

		ch <- prometheus.MustNewConstMetric(c.logUsedPercent, prometheus.GaugeValue, value, dbName)
	}

	return errors.Join(append(scanErrs, rows.Err())...)
}
