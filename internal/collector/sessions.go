package collector

import (
	"context"
	"database/sql"
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

func (c *SQLServerSessions) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.sessionState
}

func (c *SQLServerSessions) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), db.CurrentOptions().ConnTimeout)
	defer cancel()

	rows, err := c.db.QueryContext(ctx, `
	SELECT status, COUNT(*) AS session_count
	FROM sys.dm_exec_sessions
	GROUP BY status
	`)

	if err != nil {
		slog.Error("failed to query sql server for sessions stats:", slog.Any("error", err))
		return
	}

	defer rows.Close()

	for rows.Next() {
		var sessionState string
		var sessionCount int64

		if err := rows.Scan(&sessionState, &sessionCount); err != nil {
			slog.Error("failed to scan sql server sessions stats:", slog.Any("error", err))
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.sessionState,
			prometheus.GaugeValue,
			float64(sessionCount),
			sessionState,
		)
	}

}
