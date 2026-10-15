package collector

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

// SQLServerTopQueries exposes the top N queries by total CPU time from the
// plan cache (sys.dm_exec_query_stats), aggregated by query_hash, modeled on
// postgres_exporter's stat_statements and mysqld_exporter's
// perf_schema.eventsstatements collectors. Disabled by default: it scans the
// plan cache on every scrape and its series come and go as queries enter and
// leave the top N.
type SQLServerTopQueries struct {
	db             *sql.DB
	opts           TopQueriesOptions
	executions     *prometheus.Desc
	cpuSeconds     *prometheus.Desc
	elapsedSeconds *prometheus.Desc
	logicalReads   *prometheus.Desc
	physicalReads  *prometheus.Desc
	logicalWrites  *prometheus.Desc
	info           *prometheus.Desc
}

func NewTopQueriesCollector(db *sql.DB, opts TopQueriesOptions) *SQLServerTopQueries {
	labels := []string{"db_name", "query_hash"}
	return &SQLServerTopQueries{
		db:   db,
		opts: opts,
		executions: prometheus.NewDesc(
			"sqlserver_top_query_executions_total",
			"Executions of the query since its plans were cached (top queries by CPU)",
			labels, nil,
		),
		cpuSeconds: prometheus.NewDesc(
			"sqlserver_top_query_cpu_seconds_total",
			"CPU time used by the query since its plans were cached",
			labels, nil,
		),
		elapsedSeconds: prometheus.NewDesc(
			"sqlserver_top_query_elapsed_seconds_total",
			"Elapsed time of the query's executions since its plans were cached; divide its rate by the executions rate for the average duration",
			labels, nil,
		),
		logicalReads: prometheus.NewDesc(
			"sqlserver_top_query_logical_reads_total",
			"Logical reads (pages) of the query since its plans were cached",
			labels, nil,
		),
		physicalReads: prometheus.NewDesc(
			"sqlserver_top_query_physical_reads_total",
			"Physical reads (pages) of the query since its plans were cached",
			labels, nil,
		),
		logicalWrites: prometheus.NewDesc(
			"sqlserver_top_query_logical_writes_total",
			"Logical writes (pages) of the query since its plans were cached",
			labels, nil,
		),
		info: prometheus.NewDesc(
			"sqlserver_top_query_info",
			"Statement text of a top query, literals replaced with ? and truncated; always 1 (join on query_hash). Only with include_text",
			[]string{"db_name", "query_hash", "query_text"}, nil,
		),
	}
}

func (c *SQLServerTopQueries) Name() string {
	return "top_queries"
}

func (c *SQLServerTopQueries) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.executions
	ch <- c.cpuSeconds
	ch <- c.elapsedSeconds
	ch <- c.logicalReads
	ch <- c.physicalReads
	ch <- c.logicalWrites
	ch <- c.info
}

// Scrape reads the plan cache once: the window functions aggregate every
// plan of a query_hash and pick its most expensive plan as the
// representative for the database and the statement text, and the TOP is
// applied before the per-row APPLYs so they only run N times. dbid comes
// from the plan attributes (the database the statement ran in), since
// dm_exec_sql_text only has it for module code.
func (c *SQLServerTopQueries) Scrape(ctx context.Context, ch chan<- prometheus.Metric) error {
	includeText := 0
	if c.opts.IncludeText {
		includeText = 1
	}

	rows, err := c.db.QueryContext(ctx, `
		WITH per_hash AS (
			SELECT
				query_hash,
				plan_handle,
				sql_handle,
				statement_start_offset,
				statement_end_offset,
				SUM(execution_count) OVER (PARTITION BY query_hash) AS executions,
				SUM(total_worker_time) OVER (PARTITION BY query_hash) AS worker_time_us,
				SUM(total_elapsed_time) OVER (PARTITION BY query_hash) AS elapsed_time_us,
				SUM(total_logical_reads) OVER (PARTITION BY query_hash) AS logical_reads,
				SUM(total_physical_reads) OVER (PARTITION BY query_hash) AS physical_reads,
				SUM(total_logical_writes) OVER (PARTITION BY query_hash) AS logical_writes,
				ROW_NUMBER() OVER (PARTITION BY query_hash ORDER BY total_worker_time DESC) AS rn
			FROM sys.dm_exec_query_stats
		)
		SELECT
			CONVERT(varchar(18), top_n.query_hash, 1),
			ISNULL(DB_NAME(pa.dbid), ''),
			top_n.executions,
			top_n.worker_time_us,
			top_n.elapsed_time_us,
			top_n.logical_reads,
			top_n.physical_reads,
			top_n.logical_writes,
			CASE WHEN @p2 = 1 THEN LEFT(SUBSTRING(st.text, top_n.statement_start_offset / 2 + 1,
				(CASE top_n.statement_end_offset WHEN -1 THEN DATALENGTH(st.text) ELSE top_n.statement_end_offset END
					- top_n.statement_start_offset) / 2 + 1), 4000) END
		FROM (
			SELECT TOP (@p1) *
			FROM per_hash
			WHERE rn = 1
			ORDER BY worker_time_us DESC
		) top_n
		OUTER APPLY (
			SELECT CONVERT(int, value) AS dbid
			FROM sys.dm_exec_plan_attributes(top_n.plan_handle)
			WHERE attribute = N'dbid'
		) pa
		OUTER APPLY sys.dm_exec_sql_text(top_n.sql_handle) st
	`, c.opts.Limit, includeText)
	if err != nil {
		return fmt.Errorf("failed to query dm_exec_query_stats: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var scanErrs []error
	for rows.Next() {
		var queryHash, dbName string
		var executions, workerTimeUs, elapsedTimeUs, logicalReads, physicalReads, logicalWrites float64
		var text sql.NullString
		if err := rows.Scan(&queryHash, &dbName, &executions, &workerTimeUs, &elapsedTimeUs,
			&logicalReads, &physicalReads, &logicalWrites, &text); err != nil {
			scanErrs = append(scanErrs, fmt.Errorf("failed to scan top queries row: %w", err))
			continue
		}

		ch <- prometheus.MustNewConstMetric(c.executions, prometheus.CounterValue, executions, dbName, queryHash)
		ch <- prometheus.MustNewConstMetric(c.cpuSeconds, prometheus.CounterValue, workerTimeUs/1e6, dbName, queryHash)
		ch <- prometheus.MustNewConstMetric(c.elapsedSeconds, prometheus.CounterValue, elapsedTimeUs/1e6, dbName, queryHash)
		ch <- prometheus.MustNewConstMetric(c.logicalReads, prometheus.CounterValue, logicalReads, dbName, queryHash)
		ch <- prometheus.MustNewConstMetric(c.physicalReads, prometheus.CounterValue, physicalReads, dbName, queryHash)
		ch <- prometheus.MustNewConstMetric(c.logicalWrites, prometheus.CounterValue, logicalWrites, dbName, queryHash)
		if c.opts.IncludeText && text.Valid {
			ch <- prometheus.MustNewConstMetric(c.info, prometheus.GaugeValue, 1, dbName, queryHash, obfuscateSQL(text.String, c.opts.TextLength))
		}
	}

	return errors.Join(append(scanErrs, rows.Err())...)
}
