package collector

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

type SQLServerDatabaseMirroring struct {
	db                 *sql.DB
	info               *prometheus.Desc
	role               *prometheus.Desc
	state              *prometheus.Desc
	safetyLevel        *prometheus.Desc
	witnessState       *prometheus.Desc
	logSendQueueKB     *prometheus.Desc
	redoQueueKB        *prometheus.Desc
	logSentBytes       *prometheus.Desc
	logReceivedBytes   *prometheus.Desc
	redoBytes          *prometheus.Desc
	transactionDelayMS *prometheus.Desc
	writeTransactions  *prometheus.Desc
}

func NewDatabaseMirroringCollector(db *sql.DB) *SQLServerDatabaseMirroring {
	dbLabels := []string{"db_name"}
	return &SQLServerDatabaseMirroring{
		db: db,
		info: prometheus.NewDesc(
			"sqlserver_mirroring_info",
			"Database mirroring session of the database, always 1; partner_instance is the server instance of the other partner",
			[]string{"db_name", "partner_instance"}, nil,
		),
		role: prometheus.NewDesc(
			"sqlserver_mirroring_role",
			"Current mirroring role of the database (1 = principal, 2 = mirror)",
			dbLabels, nil,
		),
		state: prometheus.NewDesc(
			"sqlserver_mirroring_state",
			"Mirroring state of the database (0 = suspended, 1 = disconnected from the other partner, 2 = synchronizing, 3 = pending failover, 4 = synchronized, 5 = not synchronized/failover not possible, 6 = synchronized/failover possible)",
			dbLabels, nil,
		),
		safetyLevel: prometheus.NewDesc(
			"sqlserver_mirroring_safety_level",
			"Transaction safety of the mirroring session (0 = unknown, 1 = off/asynchronous, 2 = full/synchronous)",
			dbLabels, nil,
		),
		witnessState: prometheus.NewDesc(
			"sqlserver_mirroring_witness_state",
			"State of the witness, only present when one is configured (0 = unknown, 1 = connected, 2 = disconnected)",
			dbLabels, nil,
		),
		logSendQueueKB: prometheus.NewDesc(
			"sqlserver_mirroring_log_send_queue_kb",
			"Amount of log in kilobytes not yet sent to the mirror, as seen by this partner (both partners report it)",
			dbLabels, nil,
		),
		redoQueueKB: prometheus.NewDesc(
			"sqlserver_mirroring_redo_queue_kb",
			"Amount of log in kilobytes not yet redone on the mirror, as seen by this partner (both partners report it)",
			dbLabels, nil,
		),
		logSentBytes: prometheus.NewDesc(
			"sqlserver_mirroring_log_sent_bytes_total",
			"Total bytes of log sent to the mirror since SQL Server started",
			dbLabels, nil,
		),
		logReceivedBytes: prometheus.NewDesc(
			"sqlserver_mirroring_log_received_bytes_total",
			"Total bytes of log received from the principal since SQL Server started",
			dbLabels, nil,
		),
		redoBytes: prometheus.NewDesc(
			"sqlserver_mirroring_redo_bytes_total",
			"Total bytes of log redone on the mirror since SQL Server started",
			dbLabels, nil,
		),
		transactionDelayMS: prometheus.NewDesc(
			"sqlserver_mirroring_transaction_delay_ms_total",
			"Total milliseconds transactions waited for the mirror to acknowledge the commit (high-safety mode); divide its rate by the write transactions rate for the average delay",
			dbLabels, nil,
		),
		writeTransactions: prometheus.NewDesc(
			"sqlserver_mirroring_write_transactions_total",
			"Total transactions that wrote to the mirrored database and waited for the log to be sent to the mirror",
			dbLabels, nil,
		),
	}
}

func (c *SQLServerDatabaseMirroring) Name() string {
	return "database_mirroring"
}

func (c *SQLServerDatabaseMirroring) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.info
	ch <- c.role
	ch <- c.state
	ch <- c.safetyLevel
	ch <- c.witnessState
	ch <- c.logSendQueueKB
	ch <- c.redoQueueKB
	ch <- c.logSentBytes
	ch <- c.logReceivedBytes
	ch <- c.redoBytes
	ch <- c.transactionDelayMS
	ch <- c.writeTransactions
}

// Scrape emits nothing (and no error) when no database is mirrored:
// sys.database_mirroring only has rows with a mirroring_guid for mirrored
// databases and the Database Mirroring counters have no instances.
func (c *SQLServerDatabaseMirroring) Scrape(ctx context.Context, ch chan<- prometheus.Metric) error {
	return errors.Join(
		c.collectStates(ctx, ch),
		c.collectCounters(ctx, ch),
	)
}

// collectStates only sees the principal for non-sysadmin logins: on the
// mirror, sys.database_mirroring returns NULL in every mirroring_* column
// unless the caller is sysadmin (granting SELECT on the view doesn't help,
// validated in HML-011). The session state reported by the principal is
// enough for alerting, and the mirror still exposes its counters.
func (c *SQLServerDatabaseMirroring) collectStates(ctx context.Context, ch chan<- prometheus.Metric) error {
	rows, err := c.db.QueryContext(ctx, `
		SELECT
			d.name,
			COALESCE(dm.mirroring_partner_instance, ''),
			dm.mirroring_role,
			dm.mirroring_state,
			dm.mirroring_safety_level,
			CASE WHEN dm.mirroring_witness_name <> '' THEN dm.mirroring_witness_state END
		FROM sys.database_mirroring dm
		INNER JOIN sys.databases d
			ON d.database_id = dm.database_id
		WHERE dm.mirroring_guid IS NOT NULL
	`)
	if err != nil {
		return fmt.Errorf("failed to query database_mirroring: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var scanErrs []error
	for rows.Next() {
		var dbName, partner string
		// Without a witness SQL Server reports witness state 0 (unknown) and
		// an empty witness name; the query turns that into NULL so the series
		// is omitted instead of looking like a witness in trouble.
		var role, state, safetyLevel, witnessState sql.NullFloat64
		if err := rows.Scan(&dbName, &partner, &role, &state, &safetyLevel, &witnessState); err != nil {
			scanErrs = append(scanErrs, fmt.Errorf("failed to scan database_mirroring row: %w", err))
			continue
		}

		ch <- prometheus.MustNewConstMetric(c.info, prometheus.GaugeValue, 1, dbName, partner)
		if role.Valid {
			ch <- prometheus.MustNewConstMetric(c.role, prometheus.GaugeValue, role.Float64, dbName)
		}
		if state.Valid {
			ch <- prometheus.MustNewConstMetric(c.state, prometheus.GaugeValue, state.Float64, dbName)
		}
		if safetyLevel.Valid {
			ch <- prometheus.MustNewConstMetric(c.safetyLevel, prometheus.GaugeValue, safetyLevel.Float64, dbName)
		}
		if witnessState.Valid {
			ch <- prometheus.MustNewConstMetric(c.witnessState, prometheus.GaugeValue, witnessState.Float64, dbName)
		}
	}

	return errors.Join(append(scanErrs, rows.Err())...)
}

func (c *SQLServerDatabaseMirroring) collectCounters(ctx context.Context, ch chan<- prometheus.Metric) error {
	// The "/sec" counters and Transaction Delay are cumulative raw values in
	// dm_os_performance_counters, hence exposed as counters.
	rows, err := c.db.QueryContext(ctx, `
		SELECT RTRIM(instance_name), RTRIM(counter_name), cntr_value
		FROM sys.dm_os_performance_counters
		WHERE object_name LIKE '%Database Mirroring%'
			AND instance_name <> '_Total'
			AND counter_name IN ('Log Send Queue KB', 'Redo Queue KB', 'Log Bytes Sent/sec',
				'Log Bytes Received/sec', 'Redo Bytes/sec', 'Transaction Delay', 'Mirrored Write Transactions/sec')
	`)
	if err != nil {
		return fmt.Errorf("failed to query database mirroring performance counters: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var scanErrs []error
	for rows.Next() {
		var dbName, counterName string
		var value float64
		if err := rows.Scan(&dbName, &counterName, &value); err != nil {
			scanErrs = append(scanErrs, fmt.Errorf("failed to scan database mirroring counter row: %w", err))
			continue
		}

		switch counterName {
		case "Log Send Queue KB":
			ch <- prometheus.MustNewConstMetric(c.logSendQueueKB, prometheus.GaugeValue, value, dbName)
		case "Redo Queue KB":
			ch <- prometheus.MustNewConstMetric(c.redoQueueKB, prometheus.GaugeValue, value, dbName)
		case "Log Bytes Sent/sec":
			ch <- prometheus.MustNewConstMetric(c.logSentBytes, prometheus.CounterValue, value, dbName)
		case "Log Bytes Received/sec":
			ch <- prometheus.MustNewConstMetric(c.logReceivedBytes, prometheus.CounterValue, value, dbName)
		case "Redo Bytes/sec":
			ch <- prometheus.MustNewConstMetric(c.redoBytes, prometheus.CounterValue, value, dbName)
		case "Transaction Delay":
			ch <- prometheus.MustNewConstMetric(c.transactionDelayMS, prometheus.CounterValue, value, dbName)
		case "Mirrored Write Transactions/sec":
			ch <- prometheus.MustNewConstMetric(c.writeTransactions, prometheus.CounterValue, value, dbName)
		}
	}

	return errors.Join(append(scanErrs, rows.Err())...)
}
