package collector

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

type SQLServerAvailabilityGroups struct {
	db               *sql.DB
	replicaRole      *prometheus.Desc
	replicaSyncHlth  *prometheus.Desc
	dbSyncState      *prometheus.Desc
	dbLogSendQueueKB *prometheus.Desc
	dbRedoQueueKB    *prometheus.Desc
	dbIsSuspended    *prometheus.Desc
	dbSuspendReason  *prometheus.Desc
	dbSecondaryLag   *prometheus.Desc
}

func NewAvailabilityGroupsCollector(db *sql.DB) *SQLServerAvailabilityGroups {
	replicaLabels := []string{"ag_name", "replica_server"}
	dbLabels := []string{"ag_name", "replica_server", "db_name"}
	return &SQLServerAvailabilityGroups{
		db: db,
		replicaRole: prometheus.NewDesc(
			"sqlserver_hadr_replica_role",
			"Current role of the availability replica (1 = primary, 2 = secondary, 0 = resolving)",
			replicaLabels, nil,
		),
		replicaSyncHlth: prometheus.NewDesc(
			"sqlserver_hadr_replica_synchronization_health",
			"Synchronization health of the availability replica (0 = not healthy, 1 = partially healthy, 2 = healthy)",
			replicaLabels, nil,
		),
		dbSyncState: prometheus.NewDesc(
			"sqlserver_hadr_database_synchronization_state",
			"Synchronization state of the database replica (0 = not synchronizing, 1 = synchronizing, 2 = synchronized, 3 = reverting, 4 = initializing)",
			dbLabels, nil,
		),
		dbLogSendQueueKB: prometheus.NewDesc(
			"sqlserver_hadr_database_log_send_queue_kb",
			"Amount of log records in kilobytes not yet sent to the secondary replica",
			dbLabels, nil,
		),
		dbRedoQueueKB: prometheus.NewDesc(
			"sqlserver_hadr_database_redo_queue_kb",
			"Amount of log records in kilobytes not yet redone on the secondary replica",
			dbLabels, nil,
		),
		dbIsSuspended: prometheus.NewDesc(
			"sqlserver_hadr_database_is_suspended",
			"Whether data movement for the database replica is suspended (1 = suspended, 0 = running). While suspended SQL Server reports no log send queue, so that series is absent",
			dbLabels, nil,
		),
		dbSuspendReason: prometheus.NewDesc(
			"sqlserver_hadr_database_suspend_reason",
			"Reason data movement is suspended, only present while suspended (0 = user, 1 = partner/forced failover, 2 = redo, 3 = capture, 4 = apply, 5 = restart, 6 = undo, 7 = revalidation, 8 = synchronization point calculation error)",
			dbLabels, nil,
		),
		dbSecondaryLag: prometheus.NewDesc(
			"sqlserver_hadr_database_secondary_lag_seconds",
			"Seconds the secondary database replica is behind the primary, as reported on the primary; 0 while data movement is suspended (requires SQL Server 2016+)",
			dbLabels, nil,
		),
	}
}

func (c *SQLServerAvailabilityGroups) Name() string {
	return "availability_groups"
}

func (c *SQLServerAvailabilityGroups) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.replicaRole
	ch <- c.replicaSyncHlth
	ch <- c.dbSyncState
	ch <- c.dbLogSendQueueKB
	ch <- c.dbRedoQueueKB
	ch <- c.dbIsSuspended
	ch <- c.dbSuspendReason
	ch <- c.dbSecondaryLag
}

// Scrape emits nothing (and no error) when Always On availability groups is
// disabled: the dm_hadr_* DMVs exist on every edition and simply return no
// rows in that case.
func (c *SQLServerAvailabilityGroups) Scrape(ctx context.Context, ch chan<- prometheus.Metric) error {
	return errors.Join(
		c.collectReplicaStates(ctx, ch),
		c.collectDatabaseStates(ctx, ch),
	)
}

func (c *SQLServerAvailabilityGroups) collectReplicaStates(ctx context.Context, ch chan<- prometheus.Metric) error {
	rows, err := c.db.QueryContext(ctx, `
		SELECT
			ag.name,
			ar.replica_server_name,
			ars.role,
			ars.synchronization_health
		FROM sys.dm_hadr_availability_replica_states ars
		INNER JOIN sys.availability_replicas ar
			ON ars.replica_id = ar.replica_id
		INNER JOIN sys.availability_groups ag
			ON ars.group_id = ag.group_id
	`)
	if err != nil {
		return fmt.Errorf("failed to query dm_hadr_availability_replica_states: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var scanErrs []error
	for rows.Next() {
		var agName, replicaServer string
		// State columns are NULL for disconnected/transitioning remote
		// replicas; emit only what SQL Server actually knows instead of
		// failing the scrape during the exact incident being monitored.
		var role, syncHealth sql.NullFloat64
		if err := rows.Scan(&agName, &replicaServer, &role, &syncHealth); err != nil {
			scanErrs = append(scanErrs, fmt.Errorf("failed to scan replica state row: %w", err))
			continue
		}

		if role.Valid {
			ch <- prometheus.MustNewConstMetric(c.replicaRole, prometheus.GaugeValue, role.Float64, agName, replicaServer)
		}
		if syncHealth.Valid {
			ch <- prometheus.MustNewConstMetric(c.replicaSyncHlth, prometheus.GaugeValue, syncHealth.Float64, agName, replicaServer)
		}
	}

	return errors.Join(append(scanErrs, rows.Err())...)
}

func (c *SQLServerAvailabilityGroups) collectDatabaseStates(ctx context.Context, ch chan<- prometheus.Metric) error {
	// database_id is instance-scoped and unsafe for rows describing remote
	// replicas; group_database_id is the identifier shared by all replicas,
	// so the name is resolved through sys.availability_databases_cluster.
	rows, err := c.db.QueryContext(ctx, `
		SELECT
			ag.name,
			ar.replica_server_name,
			adc.database_name,
			drs.synchronization_state,
			drs.log_send_queue_size,
			drs.redo_queue_size,
			CAST(drs.is_suspended AS INT),
			drs.suspend_reason,
			drs.secondary_lag_seconds
		FROM sys.dm_hadr_database_replica_states drs
		INNER JOIN sys.availability_replicas ar
			ON drs.replica_id = ar.replica_id
		INNER JOIN sys.availability_groups ag
			ON drs.group_id = ag.group_id
		INNER JOIN sys.availability_databases_cluster adc
			ON adc.group_database_id = drs.group_database_id
	`)
	if err != nil {
		return fmt.Errorf("failed to query dm_hadr_database_replica_states: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var scanErrs []error
	for rows.Next() {
		var agName, replicaServer, dbName string
		// NULL queue/state values mean SQL Server has no usable measurement
		// (e.g. disconnected replica) — omitting the sample is correct;
		// converting to zero would make an outage look healthy.
		var syncState, logSendQueueKB, redoQueueKB, isSuspended, suspendReason, secondaryLag sql.NullFloat64
		if err := rows.Scan(&agName, &replicaServer, &dbName, &syncState, &logSendQueueKB, &redoQueueKB,
			&isSuspended, &suspendReason, &secondaryLag); err != nil {
			scanErrs = append(scanErrs, fmt.Errorf("failed to scan database replica state row: %w", err))
			continue
		}

		labels := []string{agName, replicaServer, dbName}
		if syncState.Valid {
			ch <- prometheus.MustNewConstMetric(c.dbSyncState, prometheus.GaugeValue, syncState.Float64, labels...)
		}
		if logSendQueueKB.Valid {
			ch <- prometheus.MustNewConstMetric(c.dbLogSendQueueKB, prometheus.GaugeValue, logSendQueueKB.Float64, labels...)
		}
		if redoQueueKB.Valid {
			ch <- prometheus.MustNewConstMetric(c.dbRedoQueueKB, prometheus.GaugeValue, redoQueueKB.Float64, labels...)
		}
		// is_suspended stays available while the queues go NULL, so a
		// suspended replica is visible instead of just a gap in the queues.
		if isSuspended.Valid {
			ch <- prometheus.MustNewConstMetric(c.dbIsSuspended, prometheus.GaugeValue, isSuspended.Float64, labels...)
			if isSuspended.Float64 == 1 && suspendReason.Valid {
				ch <- prometheus.MustNewConstMetric(c.dbSuspendReason, prometheus.GaugeValue, suspendReason.Float64, labels...)
			}
		}
		if secondaryLag.Valid {
			ch <- prometheus.MustNewConstMetric(c.dbSecondaryLag, prometheus.GaugeValue, secondaryLag.Float64, labels...)
		}
	}

	return errors.Join(append(scanErrs, rows.Err())...)
}
