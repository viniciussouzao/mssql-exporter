package collector

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

type SQLServerBlocking struct {
	db              *sql.DB
	blockedSessions *prometheus.Desc
	blockedMaxWait  *prometheus.Desc
	activeRequests  *prometheus.Desc
}

func NewBlockingCollector(db *sql.DB) *SQLServerBlocking {
	return &SQLServerBlocking{
		db: db,
		blockedSessions: prometheus.NewDesc(
			"sqlserver_blocked_sessions",
			"Number of sessions currently blocked by another session",
			nil, nil,
		),
		blockedMaxWait: prometheus.NewDesc(
			"sqlserver_blocked_max_wait_ms",
			"Longest current wait time in milliseconds among blocked sessions",
			nil, nil,
		),
		activeRequests: prometheus.NewDesc(
			"sqlserver_active_requests",
			"Number of requests currently executing, by status",
			[]string{"status"}, nil,
		),
	}
}

func (c *SQLServerBlocking) Name() string {
	return "blocking"
}

func (c *SQLServerBlocking) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.blockedSessions
	ch <- c.blockedMaxWait
	ch <- c.activeRequests
}

func (c *SQLServerBlocking) Scrape(ctx context.Context, ch chan<- prometheus.Metric) error {
	return errors.Join(
		c.collectBlocked(ctx, ch),
		c.collectRequests(ctx, ch),
	)
}

func (c *SQLServerBlocking) collectBlocked(ctx context.Context, ch chan<- prometheus.Metric) error {
	// dm_os_waiting_tasks has no database context, unlike dm_exec_requests,
	// which intermittently fails with error 976 on an AG secondary (HML-015).
	// It has one row per waiting task, so a blocked parallel request is
	// counted once per session; a task waiting on its own session (parallel
	// exchange) is not blocking. blocking_session_id > 0 only: negative
	// values are documented sentinels (orphaned/deferred transactions,
	// untracked latch owners), not sessions.
	row := c.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT session_id), ISNULL(MAX(wait_duration_ms), 0)
		FROM sys.dm_os_waiting_tasks
		WHERE blocking_session_id > 0
			AND blocking_session_id <> session_id
	`)

	var blocked, maxWaitMs float64
	if err := row.Scan(&blocked, &maxWaitMs); err != nil {
		return fmt.Errorf("failed to query blocked sessions: %w", err)
	}

	ch <- prometheus.MustNewConstMetric(c.blockedSessions, prometheus.GaugeValue, blocked)
	ch <- prometheus.MustNewConstMetric(c.blockedMaxWait, prometheus.GaugeValue, maxWaitMs)

	return nil
}

// isAGDatabaseNotAccessible reports SQL Server errors raised while an
// availability database on this replica is momentarily not accessible:
// 976 (not readable, HML-015) and 983 (replica neither PRIMARY nor
// SECONDARY, e.g. RESOLVING right after SQL Server starts, HML-020).
func isAGDatabaseNotAccessible(err error) bool {
	switch sqlErrorNumber(err) {
	case 976, 983:
		return true
	}
	return false
}

func (c *SQLServerBlocking) collectRequests(ctx context.Context, ch chan<- prometheus.Metric) error {
	// On an AG replica, reading dm_exec_requests intermittently fails while
	// an availability database is not accessible (HML-015, HML-020); it is
	// transient, so one retry avoids reporting the whole collector as
	// failed. Counts are buffered so a failed attempt never emits
	// duplicates of the retry's metrics.
	counts, err := c.queryRequestsByStatus(ctx)
	if isAGDatabaseNotAccessible(err) {
		counts, err = c.queryRequestsByStatus(ctx)
	}

	for status, count := range counts {
		ch <- prometheus.MustNewConstMetric(c.activeRequests, prometheus.GaugeValue, count, status)
	}

	return err
}

func (c *SQLServerBlocking) queryRequestsByStatus(ctx context.Context) (map[string]float64, error) {
	rows, err := c.db.QueryContext(ctx, `
		SELECT status, COUNT(*)
		FROM sys.dm_exec_requests
		GROUP BY status
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query dm_exec_requests: %w", err)
	}
	defer func() { _ = rows.Close() }()

	counts := make(map[string]float64)
	var scanErrs []error
	for rows.Next() {
		var status string
		var count float64
		if err := rows.Scan(&status, &count); err != nil {
			scanErrs = append(scanErrs, fmt.Errorf("failed to scan requests row: %w", err))
			continue
		}
		counts[status] = count
	}

	return counts, errors.Join(append(scanErrs, rows.Err())...)
}

// sqlErrorNumber returns the SQL Server error number carried by err (0 if
// none), matching the driver's error through its SQLErrorNumber method.
func sqlErrorNumber(err error) int32 {
	var sqlErr interface{ SQLErrorNumber() int32 }
	if errors.As(err, &sqlErr) {
		return sqlErr.SQLErrorNumber()
	}
	return 0
}
