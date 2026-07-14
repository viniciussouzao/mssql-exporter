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
		var role, syncHealth float64
		if err := rows.Scan(&agName, &replicaServer, &role, &syncHealth); err != nil {
			scanErrs = append(scanErrs, fmt.Errorf("failed to scan replica state row: %w", err))
			continue
		}

		ch <- prometheus.MustNewConstMetric(c.replicaRole, prometheus.GaugeValue, role, agName, replicaServer)
		ch <- prometheus.MustNewConstMetric(c.replicaSyncHlth, prometheus.GaugeValue, syncHealth, agName, replicaServer)
	}

	return errors.Join(append(scanErrs, rows.Err())...)
}

func (c *SQLServerAvailabilityGroups) collectDatabaseStates(ctx context.Context, ch chan<- prometheus.Metric) error {
	rows, err := c.db.QueryContext(ctx, `
		SELECT
			ag.name,
			ar.replica_server_name,
			DB_NAME(drs.database_id),
			drs.synchronization_state,
			ISNULL(drs.log_send_queue_size, 0),
			ISNULL(drs.redo_queue_size, 0)
		FROM sys.dm_hadr_database_replica_states drs
		INNER JOIN sys.availability_replicas ar
			ON drs.replica_id = ar.replica_id
		INNER JOIN sys.availability_groups ag
			ON drs.group_id = ag.group_id
	`)
	if err != nil {
		return fmt.Errorf("failed to query dm_hadr_database_replica_states: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var scanErrs []error
	for rows.Next() {
		var agName, replicaServer string
		var dbName sql.NullString
		var syncState, logSendQueueKB, redoQueueKB float64
		if err := rows.Scan(&agName, &replicaServer, &dbName, &syncState, &logSendQueueKB, &redoQueueKB); err != nil {
			scanErrs = append(scanErrs, fmt.Errorf("failed to scan database replica state row: %w", err))
			continue
		}
		if !dbName.Valid {
			continue
		}

		labels := []string{agName, replicaServer, dbName.String}
		ch <- prometheus.MustNewConstMetric(c.dbSyncState, prometheus.GaugeValue, syncState, labels...)
		ch <- prometheus.MustNewConstMetric(c.dbLogSendQueueKB, prometheus.GaugeValue, logSendQueueKB, labels...)
		ch <- prometheus.MustNewConstMetric(c.dbRedoQueueKB, prometheus.GaugeValue, redoQueueKB, labels...)
	}

	return errors.Join(append(scanErrs, rows.Err())...)
}
