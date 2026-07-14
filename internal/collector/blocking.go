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
			"Number of requests currently blocked by another session",
			nil, nil,
		),
		blockedMaxWait: prometheus.NewDesc(
			"sqlserver_blocked_max_wait_ms",
			"Longest current wait time in milliseconds among blocked requests",
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
	// blocking_session_id > 0 only: negative values are documented sentinels
	// (orphaned/deferred transactions, untracked latch owners), not sessions.
	row := c.db.QueryRowContext(ctx, `
		SELECT COUNT(*), ISNULL(MAX(wait_time), 0)
		FROM sys.dm_exec_requests
		WHERE blocking_session_id > 0
	`)

	var blocked, maxWaitMs float64
	if err := row.Scan(&blocked, &maxWaitMs); err != nil {
		return fmt.Errorf("failed to query blocked requests: %w", err)
	}

	ch <- prometheus.MustNewConstMetric(c.blockedSessions, prometheus.GaugeValue, blocked)
	ch <- prometheus.MustNewConstMetric(c.blockedMaxWait, prometheus.GaugeValue, maxWaitMs)

	return nil
}

func (c *SQLServerBlocking) collectRequests(ctx context.Context, ch chan<- prometheus.Metric) error {
	rows, err := c.db.QueryContext(ctx, `
		SELECT status, COUNT(*)
		FROM sys.dm_exec_requests
		GROUP BY status
	`)
	if err != nil {
		return fmt.Errorf("failed to query dm_exec_requests: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var scanErrs []error
	for rows.Next() {
		var status string
		var count float64
		if err := rows.Scan(&status, &count); err != nil {
			scanErrs = append(scanErrs, fmt.Errorf("failed to scan requests row: %w", err))
			continue
		}

		ch <- prometheus.MustNewConstMetric(c.activeRequests, prometheus.GaugeValue, count, status)
	}

	return errors.Join(append(scanErrs, rows.Err())...)
}
