package collector

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// SQLServerReplication reports transactional replication agents from the
// distribution databases of this instance (run it on the Distributor): the
// Log Reader and Distribution agents' last status, latency and, for the
// Distribution agents, the commands still waiting in the distribution
// database. Merge and snapshot agents are not covered.
type SQLServerReplication struct {
	db              *sql.DB
	runStatus       *prometheus.Desc
	lastHistory     *prometheus.Desc
	deliveryLatency *prometheus.Desc
	pendingCommands *prometheus.Desc
}

func NewReplicationCollector(db *sql.DB) *SQLServerReplication {
	labels := []string{"distribution_db", "agent_type", "agent_name", "publisher", "publisher_db", "publication", "subscriber", "subscriber_db"}
	return &SQLServerReplication{
		db: db,
		runStatus: prometheus.NewDesc(
			"sqlserver_replication_agent_run_status",
			"Status in the agent's last history entry (1 = started, 2 = succeeded, 3 = in progress, 4 = idle, 5 = retrying, 6 = failed)",
			labels, nil,
		),
		lastHistory: prometheus.NewDesc(
			"sqlserver_replication_agent_last_history_timestamp_seconds",
			"Unix timestamp (UTC) of the agent's last history entry",
			labels, nil,
		),
		deliveryLatency: prometheus.NewDesc(
			"sqlserver_replication_agent_delivery_latency_seconds",
			"Delivery latency reported in the agent's last history entry, in seconds",
			labels, nil,
		),
		pendingCommands: prometheus.NewDesc(
			"sqlserver_replication_pending_commands",
			"Commands in the distribution database not yet delivered to the subscription (Distribution agents only)",
			labels, nil,
		),
	}
}

func (c *SQLServerReplication) Name() string {
	return "replication"
}

func (c *SQLServerReplication) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.runStatus
	ch <- c.lastHistory
	ch <- c.deliveryLatency
	ch <- c.pendingCommands
}

func (c *SQLServerReplication) Scrape(ctx context.Context, ch chan<- prometheus.Metric) error {
	dbs, err := listDistributionDatabases(ctx, c.db)
	if err != nil {
		return err
	}

	errs := make([]error, 0, len(dbs))
	for _, name := range dbs {
		errs = append(errs, c.scrapeDistributionDB(ctx, ch, name))
	}
	return errors.Join(errs...)
}

// listDistributionDatabases returns the distribution databases the login can
// read. None means the instance is not a Distributor: nothing is emitted.
func listDistributionDatabases(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT name
		FROM sys.databases
		WHERE is_distributor = 1
			AND state = 0
			AND HAS_DBACCESS(name) = 1
		ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list distribution databases: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var names []string
	var scanErrs []error
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			scanErrs = append(scanErrs, fmt.Errorf("failed to scan distribution database row: %w", err))
			continue
		}
		names = append(names, name)
	}
	return names, errors.Join(append(scanErrs, rows.Err())...)
}

// replicationAgentsQuery reads the agents of one distribution database. The
// publisher and subscriber ids map to MSreplservers (SQL Server 2016 SP2 and
// later; the ids in master.sys.servers do not match there). Virtual
// subscriptions (subscriber_id < 0, created for immediate_sync publications)
// are skipped. History times are the Distributor's local time and are
// converted to UTC with its current offset.
func replicationAgentsQuery(distributionDB string) string {
	q := strings.NewReplacer("{db}", quoteIdentifier(distributionDB))
	return q.Replace(`
		SELECT
			N'distribution',
			a.name,
			ISNULL(ps.srvname, N''),
			a.publisher_db,
			ISNULL(a.publication, N''),
			ISNULL(ss.srvname, N''),
			a.subscriber_db,
			h.runstatus,
			DATEADD(MINUTE, -DATEPART(TZOFFSET, SYSDATETIMEOFFSET()), h.[time]),
			h.delivery_latency,
			st.pending
		FROM {db}.dbo.MSdistribution_agents a
		LEFT JOIN {db}.dbo.MSreplservers ps ON ps.srvid = a.publisher_id
		LEFT JOIN {db}.dbo.MSreplservers ss ON ss.srvid = a.subscriber_id
		OUTER APPLY (
			SELECT TOP (1) runstatus, [time], delivery_latency
			FROM {db}.dbo.MSdistribution_history
			WHERE agent_id = a.id
			ORDER BY [time] DESC
		) h
		OUTER APPLY (
			SELECT SUM(CAST(UndelivCmdsInDistDB AS bigint)) AS pending
			FROM {db}.dbo.MSdistribution_status
			WHERE agent_id = a.id
		) st
		WHERE a.subscriber_id >= 0
		UNION ALL
		SELECT
			N'logreader',
			a.name,
			ISNULL(ps.srvname, N''),
			a.publisher_db,
			N'',
			N'',
			N'',
			h.runstatus,
			DATEADD(MINUTE, -DATEPART(TZOFFSET, SYSDATETIMEOFFSET()), h.[time]),
			h.delivery_latency,
			NULL
		FROM {db}.dbo.MSlogreader_agents a
		LEFT JOIN {db}.dbo.MSreplservers ps ON ps.srvid = a.publisher_id
		OUTER APPLY (
			SELECT TOP (1) runstatus, [time], delivery_latency
			FROM {db}.dbo.MSlogreader_history
			WHERE agent_id = a.id
			ORDER BY [time] DESC
		) h
	`)
}

func (c *SQLServerReplication) scrapeDistributionDB(ctx context.Context, ch chan<- prometheus.Metric, distributionDB string) error {
	rows, err := c.db.QueryContext(ctx, replicationAgentsQuery(distributionDB))
	if err != nil {
		return fmt.Errorf("failed to query replication agents in %s: %w", distributionDB, err)
	}
	defer func() { _ = rows.Close() }()

	var scanErrs []error
	for rows.Next() {
		var agentType, agentName, publisher, publisherDB, publication, subscriber, subscriberDB string
		var runStatus, latencyMs, pending sql.NullFloat64
		var lastHistory sql.NullTime
		if err := rows.Scan(&agentType, &agentName, &publisher, &publisherDB, &publication, &subscriber, &subscriberDB,
			&runStatus, &lastHistory, &latencyMs, &pending); err != nil {
			scanErrs = append(scanErrs, fmt.Errorf("failed to scan replication agent row in %s: %w", distributionDB, err))
			continue
		}

		labels := []string{distributionDB, agentType, agentName, publisher, publisherDB, publication, subscriber, subscriberDB}
		// An agent that never ran has no history: only the pending commands.
		if runStatus.Valid {
			ch <- prometheus.MustNewConstMetric(c.runStatus, prometheus.GaugeValue, runStatus.Float64, labels...)
		}
		if lastHistory.Valid {
			ch <- prometheus.MustNewConstMetric(c.lastHistory, prometheus.GaugeValue, float64(lastHistory.Time.Unix()), labels...)
		}
		if latencyMs.Valid {
			ch <- prometheus.MustNewConstMetric(c.deliveryLatency, prometheus.GaugeValue, latencyMs.Float64/1000, labels...)
		}
		if pending.Valid {
			ch <- prometheus.MustNewConstMetric(c.pendingCommands, prometheus.GaugeValue, pending.Float64, labels...)
		}
	}

	return errors.Join(append(scanErrs, rows.Err())...)
}
