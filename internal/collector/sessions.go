package collector

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/viniciussouzao/mssql-exporter/internal/db"
)

type SQLServerSessions struct {
	db           *sql.DB
	sessionState *prometheus.Desc
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
	}
}

func (c *SQLServerSessions) Name() string {
	return "sessions"
}

func (c *SQLServerSessions) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.sessionState
}

func (c *SQLServerSessions) Scrape(ch chan<- prometheus.Metric) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.CurrentOptions().ConnTimeout)
	defer cancel()

	rows, err := c.db.QueryContext(ctx, `
	SELECT status, COUNT(*) AS session_count
	FROM sys.dm_exec_sessions
	GROUP BY status
	`)

	if err != nil {
		return fmt.Errorf("failed to query dm_exec_sessions: %w", err)
	}

	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var sessionState string
		var sessionCount int64

		if err := rows.Scan(&sessionState, &sessionCount); err != nil {
			slog.Error("failed to scan sessions row", slog.Any("error", err))
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.sessionState,
			prometheus.GaugeValue,
			float64(sessionCount),
			sessionState,
		)
	}

	return rows.Err()
}
