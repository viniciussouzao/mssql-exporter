package collector

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

type SQLServerSessions struct {
	db             *sql.DB
	sessionState   *prometheus.Desc
	sessionsByUser *prometheus.Desc
}

func NewSessionsCollector(db *sql.DB) *SQLServerSessions {
	return &SQLServerSessions{
		db: db,
		sessionState: prometheus.NewDesc(
			"sqlserver_sessions_total",
			"Number of sessions in SQL Server by state",
			[]string{"session_state"},
			nil,
		),
		sessionsByUser: prometheus.NewDesc(
			"sqlserver_sessions_by_login",
			"Number of user sessions in SQL Server by login",
			[]string{"login_name"},
			nil,
		),
	}
}

func (c *SQLServerSessions) Name() string {
	return "sessions"
}

func (c *SQLServerSessions) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.sessionState
	ch <- c.sessionsByUser
}

func (c *SQLServerSessions) Scrape(ctx context.Context, ch chan<- prometheus.Metric) error {
	return errors.Join(
		c.collectByState(ctx, ch),
		c.collectByLogin(ctx, ch),
	)
}

func (c *SQLServerSessions) collectByState(ctx context.Context, ch chan<- prometheus.Metric) error {
	rows, err := c.db.QueryContext(ctx, `
	SELECT status, COUNT(*) AS session_count
	FROM sys.dm_exec_sessions
	GROUP BY status
	`)

	if err != nil {
		return fmt.Errorf("failed to query dm_exec_sessions: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var scanErrs []error
	for rows.Next() {
		var sessionState string
		var sessionCount int64

		if err := rows.Scan(&sessionState, &sessionCount); err != nil {
			scanErrs = append(scanErrs, fmt.Errorf("failed to scan sessions row: %w", err))
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.sessionState,
			prometheus.GaugeValue,
			float64(sessionCount),
			sessionState,
		)
	}

	return errors.Join(append(scanErrs, rows.Err())...)
}

// collectByLogin counts user sessions only: system sessions run under the
// sa/service login and would hide the application logins' load. The label
// cardinality is the number of logins with open sessions.
func (c *SQLServerSessions) collectByLogin(ctx context.Context, ch chan<- prometheus.Metric) error {
	rows, err := c.db.QueryContext(ctx, `
	SELECT login_name, COUNT(*) AS session_count
	FROM sys.dm_exec_sessions
	WHERE is_user_process = 1
	GROUP BY login_name
	`)

	if err != nil {
		return fmt.Errorf("failed to query dm_exec_sessions by login: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var scanErrs []error
	for rows.Next() {
		var loginName string
		var sessionCount int64

		if err := rows.Scan(&loginName, &sessionCount); err != nil {
			scanErrs = append(scanErrs, fmt.Errorf("failed to scan sessions by login row: %w", err))
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.sessionsByUser,
			prometheus.GaugeValue,
			float64(sessionCount),
			loginName,
		)
	}

	return errors.Join(append(scanErrs, rows.Err())...)
}
