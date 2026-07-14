package collector

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/viniciussouzao/mssql-exporter/internal/db"
)

type SQLServerMemoryClerks struct {
	db           *sql.DB
	memoryUsedKB *prometheus.Desc
}

func NewMemoryClerksCollector(db *sql.DB) *SQLServerMemoryClerks {
	return &SQLServerMemoryClerks{
		db: db,
		memoryUsedKB: prometheus.NewDesc(
			"sqlserver_memory_clerk_used_kb",
			"Memory used in KB by each memory clerk",
			[]string{"clerk_type", "clerk_name", "memory_node_id"},
			nil,
		),
	}
}

func (c *SQLServerMemoryClerks) Name() string {
	return "memory_clerks"
}

func (c *SQLServerMemoryClerks) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.memoryUsedKB
}

func (c *SQLServerMemoryClerks) Scrape(ch chan<- prometheus.Metric) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.CurrentOptions().ConnTimeout)
	defer cancel()

	rows, err := c.db.QueryContext(ctx, `
		SELECT type, name, memory_node_id, SUM(virtual_memory_committed_kb) AS memory_used_kb
		FROM sys.dm_os_memory_clerks
		GROUP BY type, name, memory_node_id
	`)
	if err != nil {
		return fmt.Errorf("failed to query dm_os_memory_clerks: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var (
			clerkType    string
			clerkName    string
			memoryNodeID int
			memoryUsedKB int64
		)

		if err := rows.Scan(&clerkType, &clerkName, &memoryNodeID, &memoryUsedKB); err != nil {
			slog.Error("failed to scan memory_clerks row", slog.Any("error", err))
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.memoryUsedKB,
			prometheus.GaugeValue,
			float64(memoryUsedKB),
			clerkType,
			clerkName,
			strconv.Itoa(memoryNodeID),
		)
	}

	return rows.Err()
}
