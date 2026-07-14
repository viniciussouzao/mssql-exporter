package collector

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/viniciussouzao/mssql-exporter/internal/db"
)

type SQLServerIOFileStats struct {
	db *sql.DB

	reads        *prometheus.Desc
	writes       *prometheus.Desc
	readBytes    *prometheus.Desc
	writeBytes   *prometheus.Desc
	stallReadMs  *prometheus.Desc
	stallWriteMs *prometheus.Desc
}

func NewIOFileStatsCollector(db *sql.DB) *SQLServerIOFileStats {
	return &SQLServerIOFileStats{
		db: db,

		reads: prometheus.NewDesc(
			"sqlserver_io_file_reads_total",
			"Total number of reads from the file",
			[]string{"db_name", "file_name", "file_type", "physical_name"},
			nil,
		),

		writes: prometheus.NewDesc(
			"sqlserver_io_file_writes_total",
			"Total number of writes to the file",
			[]string{"db_name", "file_name", "file_type", "physical_name"},
			nil,
		),

		readBytes: prometheus.NewDesc(
			"sqlserver_io_file_read_bytes_total",
			"Total number of bytes read from the file",
			[]string{"db_name", "file_name", "file_type", "physical_name"},
			nil,
		),

		writeBytes: prometheus.NewDesc(
			"sqlserver_io_file_write_bytes_total",
			"Total number of bytes written to the file",
			[]string{"db_name", "file_name", "file_type", "physical_name"},
			nil,
		),

		stallReadMs: prometheus.NewDesc(
			"sqlserver_io_file_stall_read_ms_total",
			"Total stall time in milliseconds for reads from the file",
			[]string{"db_name", "file_name", "file_type", "physical_name"},
			nil,
		),

		stallWriteMs: prometheus.NewDesc(
			"sqlserver_io_file_stall_write_ms_total",
			"Total stall time in milliseconds for writes to the file",
			[]string{"db_name", "file_name", "file_type", "physical_name"},
			nil,
		),
	}
}

func (c *SQLServerIOFileStats) Name() string {
	return "io_file_stats"
}

func (c *SQLServerIOFileStats) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.reads
	ch <- c.writes
	ch <- c.readBytes
	ch <- c.writeBytes
	ch <- c.stallReadMs
	ch <- c.stallWriteMs
}

func (c *SQLServerIOFileStats) Scrape(ch chan<- prometheus.Metric) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.CurrentOptions().ConnTimeout)
	defer cancel()

	rows, err := c.db.QueryContext(ctx, `
	SELECT
            DB_NAME(vfs.database_id) AS db_name,
            mf.name AS file_name,
            mf.type_desc AS file_type,
            mf.physical_name,

            vfs.num_of_reads,
            vfs.num_of_writes,
            vfs.num_of_bytes_read,
            vfs.num_of_bytes_written,
            vfs.io_stall_read_ms,
            vfs.io_stall_write_ms

        FROM sys.dm_io_virtual_file_stats(NULL, NULL) AS vfs
        INNER JOIN sys.master_files AS mf
            ON vfs.database_id = mf.database_id
            AND vfs.file_id = mf.file_id`)

	if err != nil {
		return fmt.Errorf("failed to query dm_io_virtual_file_stats: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var (
			dbName, fileName, fileType, physicalName string
			reads, writes                            float64
			readBytes, writeBytes                    float64
			stallRead, stallWrite                    float64
		)

		if err := rows.Scan(
			&dbName,
			&fileName,
			&fileType,
			&physicalName,
			&reads,
			&writes,
			&readBytes,
			&writeBytes,
			&stallRead,
			&stallWrite,
		); err != nil {
			slog.Error("failed to scan io file stats row", slog.Any("error", err))
			continue
		}

		labels := []string{dbName, fileName, fileType, physicalName}

		ch <- prometheus.MustNewConstMetric(c.reads, prometheus.CounterValue, reads, labels...)
		ch <- prometheus.MustNewConstMetric(c.writes, prometheus.CounterValue, writes, labels...)
		ch <- prometheus.MustNewConstMetric(c.readBytes, prometheus.CounterValue, readBytes, labels...)
		ch <- prometheus.MustNewConstMetric(c.writeBytes, prometheus.CounterValue, writeBytes, labels...)
		ch <- prometheus.MustNewConstMetric(c.stallReadMs, prometheus.CounterValue, stallRead, labels...)
		ch <- prometheus.MustNewConstMetric(c.stallWriteMs, prometheus.CounterValue, stallWrite, labels...)
	}

	return rows.Err()
}
