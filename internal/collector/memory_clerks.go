package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

type SQLServerMemoryClerks struct {
	db           *sql.DB
	clerkType    *prometheus.Desc
	clerkName    *prometheus.Desc
	memoryNodeId *prometheus.Desc
	memoryUsedKB *prometheus.Desc
}

func NewMemoryClerksCollector(db *sql.DB) *SQLServerMemoryClerks {
	return &SQLServerMemoryClerks{
		db: db,
		clerkType: prometheus.NewDesc(
			"sqlserver_memory_clerks",
			"Type of memory clerk",
			[]string{"clerk_type"},
			nil,
		),

		clerkName: prometheus.NewDesc(
			"sqlserver_memory_clerk_name",
			"Name of memory clerk",
			[]string{"clerk_name"},
			nil,
		),

		memoryNodeId: prometheus.NewDesc(
			"sqlserver_memory_node_id",
			"Memory node ID used by the clerk",
			[]string{"memory_node_id"},
			nil,
		),

		memoryUsedKB: prometheus.NewDesc(
			"sqlserver_memory_used_kb",
			"Memory used in KB by the clerk",
			[]string{"clerk_type", "clerk_name", "memory_node_id"},
			nil,
		),
	}
}

func (c *SQLServerMemoryClerks) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.clerkType
	ch <- c.clerkName
	ch <- c.memoryNodeId
	ch <- c.memoryUsedKB
}

func (c *SQLServerMemoryClerks) Collect(ch chan<- prometheus.Metric) {
	rows, err := c.db.Query(`
	SELECT clerk_type, clerk_name, memory_node_id, SUM(virtual_memory_committed_kb) AS memory_used_kb
	FROM sys.dm_os_memory_clerks
	GROUP BY clerk_type, clerk_name, memory_node_id
	`)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			clerkType    string
			clerkName    string
			memoryNodeId int
			memoryUsedKB int64
		)

		if err := rows.Scan(&clerkType, &clerkName, &memoryNodeId, &memoryUsedKB); err != nil {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.memoryUsedKB,
			prometheus.GaugeValue,
			float64(memoryUsedKB),
			clerkType,
			clerkName,
			string(memoryNodeId),
		)
	}
}
