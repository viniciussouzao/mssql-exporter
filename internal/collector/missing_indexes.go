package collector

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

// SQLServerMissingIndexes exposes the indexes the query optimizer reported
// as missing (sys.dm_db_missing_index_*), the top N by estimated improvement.
// The DMVs keep at most 600 suggestions, are reset when SQL Server restarts
// and drop a table's suggestions when an index is created on it or the
// table is modified (DDL), so the series come and go.
type SQLServerMissingIndexes struct {
	db               *sql.DB
	opts             MissingIndexesOptions
	improvement      *prometheus.Desc
	userSeeks        *prometheus.Desc
	userScans        *prometheus.Desc
	avgUserImpact    *prometheus.Desc
	avgTotalUserCost *prometheus.Desc
	info             *prometheus.Desc
}

func NewMissingIndexesCollector(db *sql.DB, opts MissingIndexesOptions) *SQLServerMissingIndexes {
	labels := []string{"db_name", "schema", "table", "index_handle"}
	return &SQLServerMissingIndexes{
		db:   db,
		opts: opts,
		improvement: prometheus.NewDesc(
			"sqlserver_missing_index_improvement_measure",
			"Estimated benefit of the missing index: avg_total_user_cost * avg_user_impact / 100 * (user_seeks + user_scans)",
			labels, nil,
		),
		userSeeks: prometheus.NewDesc(
			"sqlserver_missing_index_user_seeks_total",
			"Seeks by user queries that the missing index would have served",
			labels, nil,
		),
		userScans: prometheus.NewDesc(
			"sqlserver_missing_index_user_scans_total",
			"Scans by user queries that the missing index would have served",
			labels, nil,
		),
		avgUserImpact: prometheus.NewDesc(
			"sqlserver_missing_index_avg_user_impact_percent",
			"Average estimated reduction of the query cost with the missing index, in percent",
			labels, nil,
		),
		avgTotalUserCost: prometheus.NewDesc(
			"sqlserver_missing_index_avg_total_user_cost",
			"Average cost of the user queries that would benefit from the missing index",
			labels, nil,
		),
		info: prometheus.NewDesc(
			"sqlserver_missing_index_info",
			"Columns of the missing index suggested by the optimizer (value is always 1)",
			append(labels, "equality_columns", "inequality_columns", "included_columns"), nil,
		),
	}
}

func (c *SQLServerMissingIndexes) Name() string {
	return "missing_indexes"
}

func (c *SQLServerMissingIndexes) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.improvement
	ch <- c.userSeeks
	ch <- c.userScans
	ch <- c.avgUserImpact
	ch <- c.avgTotalUserCost
	ch <- c.info
}

func (c *SQLServerMissingIndexes) Scrape(ctx context.Context, ch chan<- prometheus.Metric) error {
	rows, err := c.db.QueryContext(ctx, `
		SELECT TOP (@p1)
			ISNULL(DB_NAME(d.database_id), ''),
			ISNULL(OBJECT_SCHEMA_NAME(d.object_id, d.database_id), ''),
			ISNULL(OBJECT_NAME(d.object_id, d.database_id), ''),
			CONVERT(varchar(20), d.index_handle),
			ISNULL(d.equality_columns, ''),
			ISNULL(d.inequality_columns, ''),
			ISNULL(d.included_columns, ''),
			s.user_seeks,
			s.user_scans,
			s.avg_user_impact,
			s.avg_total_user_cost,
			s.avg_total_user_cost * (s.avg_user_impact / 100.0) * (s.user_seeks + s.user_scans) AS improvement
		FROM sys.dm_db_missing_index_details d
		JOIN sys.dm_db_missing_index_groups g ON g.index_handle = d.index_handle
		JOIN sys.dm_db_missing_index_group_stats s ON s.group_handle = g.index_group_handle
		ORDER BY improvement DESC
	`, c.opts.Limit)
	if err != nil {
		return fmt.Errorf("failed to query dm_db_missing_index_details: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var scanErrs []error
	for rows.Next() {
		var dbName, schema, table, handle, equality, inequality, included string
		var seeks, scans, impact, cost, improvement float64
		if err := rows.Scan(&dbName, &schema, &table, &handle, &equality, &inequality, &included,
			&seeks, &scans, &impact, &cost, &improvement); err != nil {
			scanErrs = append(scanErrs, fmt.Errorf("failed to scan missing index row: %w", err))
			continue
		}

		labels := []string{dbName, schema, table, handle}
		ch <- prometheus.MustNewConstMetric(c.improvement, prometheus.GaugeValue, improvement, labels...)
		ch <- prometheus.MustNewConstMetric(c.userSeeks, prometheus.CounterValue, seeks, labels...)
		ch <- prometheus.MustNewConstMetric(c.userScans, prometheus.CounterValue, scans, labels...)
		ch <- prometheus.MustNewConstMetric(c.avgUserImpact, prometheus.GaugeValue, impact, labels...)
		ch <- prometheus.MustNewConstMetric(c.avgTotalUserCost, prometheus.GaugeValue, cost, labels...)
		ch <- prometheus.MustNewConstMetric(c.info, prometheus.GaugeValue, 1, append(labels, equality, inequality, included)...)
	}

	return errors.Join(append(scanErrs, rows.Err())...)
}
