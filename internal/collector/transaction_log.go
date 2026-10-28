package collector

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

const bytesPerMB = 1024 * 1024

// SQLServerTransactionLog reports transaction log health per database from
// sys.dm_db_log_stats (SQL Server 2016 SP2 and later): VLF count, active log
// and log generated since the last log backup and the last checkpoint. It
// works where the "Percent Log Used" counter misses databases (the system
// databases on Amazon RDS, HML-026). The truncation reason is already
// exposed as sqlserver_database_log_reuse_wait by the databases collector.
type SQLServerTransactionLog struct {
	db                  *sql.DB
	vlfCount            *prometheus.Desc
	activeVLFCount      *prometheus.Desc
	activeBytes         *prometheus.Desc
	sinceLastLogBackup  *prometheus.Desc
	sinceLastCheckpoint *prometheus.Desc
}

func NewTransactionLogCollector(db *sql.DB) *SQLServerTransactionLog {
	labels := []string{"db_name"}
	return &SQLServerTransactionLog{
		db: db,
		vlfCount: prometheus.NewDesc(
			"sqlserver_database_log_vlf_count",
			"Number of virtual log files (VLFs) in the transaction log",
			labels, nil,
		),
		activeVLFCount: prometheus.NewDesc(
			"sqlserver_database_log_active_vlf_count",
			"Number of active virtual log files (VLFs), which cannot be reused yet",
			labels, nil,
		),
		activeBytes: prometheus.NewDesc(
			"sqlserver_database_log_active_bytes",
			"Size of the active part of the transaction log in bytes",
			labels, nil,
		),
		sinceLastLogBackup: prometheus.NewDesc(
			"sqlserver_database_log_since_last_log_backup_bytes",
			"Log generated since the last transaction log backup in bytes (absent when the database has no log backup chain, e.g. SIMPLE recovery)",
			labels, nil,
		),
		sinceLastCheckpoint: prometheus.NewDesc(
			"sqlserver_database_log_since_last_checkpoint_bytes",
			"Log generated since the last checkpoint in bytes",
			labels, nil,
		),
	}
}

func (c *SQLServerTransactionLog) Name() string {
	return "transaction_log"
}

func (c *SQLServerTransactionLog) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.vlfCount
	ch <- c.activeVLFCount
	ch <- c.activeBytes
	ch <- c.sinceLastLogBackup
	ch <- c.sinceLastCheckpoint
}

// Scrape reads every ONLINE database the login can access. Local secondary
// replicas of an Availability Group are skipped: their log mirrors the
// primary's and an unreadable secondary rejects the query.
func (c *SQLServerTransactionLog) Scrape(ctx context.Context, ch chan<- prometheus.Metric) error {
	rows, err := c.db.QueryContext(ctx, `
		SELECT
			d.name,
			ls.total_vlf_count,
			ls.active_vlf_count,
			ls.active_log_size_mb,
			ls.log_since_last_log_backup_mb,
			ls.log_since_last_checkpoint_mb
		FROM sys.databases d
		CROSS APPLY sys.dm_db_log_stats(d.database_id) ls
		WHERE d.state = 0
			AND HAS_DBACCESS(d.name) = 1
			AND NOT EXISTS (
				SELECT 1 FROM sys.dm_hadr_database_replica_states rs
				WHERE rs.database_id = d.database_id
					AND rs.is_local = 1
					AND rs.is_primary_replica = 0
			)
	`)
	if err != nil {
		return fmt.Errorf("failed to query dm_db_log_stats: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var scanErrs []error
	for rows.Next() {
		var (
			name                        string
			vlfs, activeVLFs            float64
			activeMB, sinceCheckpointMB float64
			sinceLogBackupMB            sql.NullFloat64
		)
		if err := rows.Scan(&name, &vlfs, &activeVLFs, &activeMB, &sinceLogBackupMB, &sinceCheckpointMB); err != nil {
			scanErrs = append(scanErrs, fmt.Errorf("failed to scan dm_db_log_stats row: %w", err))
			continue
		}

		ch <- prometheus.MustNewConstMetric(c.vlfCount, prometheus.GaugeValue, vlfs, name)
		ch <- prometheus.MustNewConstMetric(c.activeVLFCount, prometheus.GaugeValue, activeVLFs, name)
		ch <- prometheus.MustNewConstMetric(c.activeBytes, prometheus.GaugeValue, activeMB*bytesPerMB, name)
		ch <- prometheus.MustNewConstMetric(c.sinceLastCheckpoint, prometheus.GaugeValue, sinceCheckpointMB*bytesPerMB, name)
		if sinceLogBackupMB.Valid {
			ch <- prometheus.MustNewConstMetric(c.sinceLastLogBackup, prometheus.GaugeValue, sinceLogBackupMB.Float64*bytesPerMB, name)
		}
	}

	return errors.Join(append(scanErrs, rows.Err())...)
}
