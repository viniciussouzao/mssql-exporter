package collector

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

// longRunningThresholds are the elapsed-time buckets (seconds) of
// sqlserver_requests_long_running. The SQL query has one COUNT per entry, in
// this order.
var longRunningThresholds = []int{10, 60, 300, 1800}

// SQLServerActiveQueries reports long-running user requests and open user
// transactions without per-query labels, so cardinality stays fixed.
type SQLServerActiveQueries struct {
	db                   *sql.DB
	longRunning          *prometheus.Desc
	maxElapsed           *prometheus.Desc
	openTransactions     *prometheus.Desc
	oldestTransaction    *prometheus.Desc
	sleepingWithOpenTran *prometheus.Desc
}

func NewActiveQueriesCollector(db *sql.DB) *SQLServerActiveQueries {
	return &SQLServerActiveQueries{
		db: db,
		longRunning: prometheus.NewDesc(
			"sqlserver_requests_long_running",
			"Number of user requests running for at least over_seconds",
			[]string{"over_seconds"}, nil,
		),
		maxElapsed: prometheus.NewDesc(
			"sqlserver_request_max_elapsed_seconds",
			"Elapsed time of the longest running user request, per command (only commands with active user requests)",
			[]string{"command"}, nil,
		),
		openTransactions: prometheus.NewDesc(
			"sqlserver_open_transactions",
			"Number of open transactions of user sessions",
			nil, nil,
		),
		oldestTransaction: prometheus.NewDesc(
			"sqlserver_oldest_open_transaction_seconds",
			"Age of the oldest open transaction of a user session (0 when there is none)",
			nil, nil,
		),
		sleepingWithOpenTran: prometheus.NewDesc(
			"sqlserver_sleeping_sessions_with_open_transaction",
			"Number of idle (sleeping) user sessions holding an open transaction, e.g. a BEGIN TRAN never committed",
			nil, nil,
		),
	}
}

func (c *SQLServerActiveQueries) Name() string {
	return "active_queries"
}

func (c *SQLServerActiveQueries) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.longRunning
	ch <- c.maxElapsed
	ch <- c.openTransactions
	ch <- c.oldestTransaction
	ch <- c.sleepingWithOpenTran
}

func (c *SQLServerActiveQueries) Scrape(ctx context.Context, ch chan<- prometheus.Metric) error {
	return errors.Join(
		c.collectRequests(ctx, ch),
		c.collectTransactions(ctx, ch),
	)
}

// requestsByCommand is the per-command result of the requests query.
type requestsByCommand struct {
	command    string
	maxElapsed float64
	overCounts []float64 // one per longRunningThresholds entry
}

func (c *SQLServerActiveQueries) collectRequests(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Same transient AG errors as the blocking collector's requests query
	// (HML-015, HML-020): retry once, emitting only a complete result.
	results, err := c.queryRequests(ctx)
	if isAGDatabaseNotAccessible(err) {
		results, err = c.queryRequests(ctx)
	}

	totals := make([]float64, len(longRunningThresholds))
	for _, r := range results {
		ch <- prometheus.MustNewConstMetric(c.maxElapsed, prometheus.GaugeValue, r.maxElapsed, r.command)
		for i, n := range r.overCounts {
			totals[i] += n
		}
	}
	// The buckets are emitted even with no user request running, so alerts
	// and graphs see 0 rather than an absent series; a failed query emits
	// nothing.
	if err == nil {
		for i, seconds := range longRunningThresholds {
			ch <- prometheus.MustNewConstMetric(c.longRunning, prometheus.GaugeValue, totals[i], strconv.Itoa(seconds))
		}
	}

	return err
}

func (c *SQLServerActiveQueries) queryRequests(ctx context.Context) ([]requestsByCommand, error) {
	// start_time and GETDATE() are both server local time. DATEDIFF in
	// seconds instead of total_elapsed_time (int milliseconds), which
	// overflows after ~24 days. The exporter's own request (@@SPID) and
	// system sessions are excluded, and so is the telemetry service of Azure
	// SQL Managed Instance (EngineEdition 8): a user session that keeps a
	// request reading an Extended Events live target open from instance
	// start, so it would count in every bucket.
	rows, err := c.db.QueryContext(ctx, `
		SELECT
			RTRIM(r.command),
			MAX(DATEDIFF(SECOND, r.start_time, GETDATE())),
			COUNT(CASE WHEN DATEDIFF(SECOND, r.start_time, GETDATE()) >= 10 THEN 1 END),
			COUNT(CASE WHEN DATEDIFF(SECOND, r.start_time, GETDATE()) >= 60 THEN 1 END),
			COUNT(CASE WHEN DATEDIFF(SECOND, r.start_time, GETDATE()) >= 300 THEN 1 END),
			COUNT(CASE WHEN DATEDIFF(SECOND, r.start_time, GETDATE()) >= 1800 THEN 1 END)
		FROM sys.dm_exec_requests r
		INNER JOIN sys.dm_exec_sessions s
			ON s.session_id = r.session_id
		WHERE s.is_user_process = 1
			AND r.session_id <> @@SPID
			AND NOT (
				CAST(SERVERPROPERTY('EngineEdition') AS int) = 8
				AND ISNULL(s.program_name, N'') = N'TdService'
				AND s.login_name = N'NT AUTHORITY\SYSTEM'
			)
		GROUP BY r.command
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query long-running requests: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var results []requestsByCommand
	var scanErrs []error
	for rows.Next() {
		r := requestsByCommand{overCounts: make([]float64, len(longRunningThresholds))}
		dest := []any{&r.command, &r.maxElapsed}
		for i := range r.overCounts {
			dest = append(dest, &r.overCounts[i])
		}
		if err := rows.Scan(dest...); err != nil {
			scanErrs = append(scanErrs, fmt.Errorf("failed to scan long-running requests row: %w", err))
			continue
		}
		// A request started before a backwards clock change would show a
		// negative elapsed time.
		r.maxElapsed = max(r.maxElapsed, 0)
		results = append(results, r)
	}

	return results, errors.Join(append(scanErrs, rows.Err())...)
}

func (c *SQLServerActiveQueries) collectTransactions(ctx context.Context, ch chan<- prometheus.Metric) error {
	// A sleeping session can only hold an open transaction if it was opened
	// explicitly and never committed: the classic forgotten BEGIN TRAN that
	// keeps locks and prevents log truncation.
	row := c.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*),
			ISNULL(MAX(DATEDIFF(SECOND, at.transaction_begin_time, GETDATE())), 0),
			COUNT(DISTINCT CASE WHEN s.status = 'sleeping' THEN s.session_id END)
		FROM sys.dm_tran_session_transactions st
		INNER JOIN sys.dm_tran_active_transactions at
			ON at.transaction_id = st.transaction_id
		INNER JOIN sys.dm_exec_sessions s
			ON s.session_id = st.session_id
		WHERE s.is_user_process = 1
			AND st.session_id <> @@SPID
	`)

	var open, oldest, sleeping float64
	if err := row.Scan(&open, &oldest, &sleeping); err != nil {
		return fmt.Errorf("failed to query open transactions: %w", err)
	}

	ch <- prometheus.MustNewConstMetric(c.openTransactions, prometheus.GaugeValue, open)
	ch <- prometheus.MustNewConstMetric(c.oldestTransaction, prometheus.GaugeValue, max(oldest, 0))
	ch <- prometheus.MustNewConstMetric(c.sleepingWithOpenTran, prometheus.GaugeValue, sleeping)

	return nil
}
