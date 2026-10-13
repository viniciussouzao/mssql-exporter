package collector

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

type SQLServerAgentJobs struct {
	db              *sql.DB
	jobEnabled      *prometheus.Desc
	lastRunStatus   *prometheus.Desc
	lastRunAt       *prometheus.Desc
	lastRunDuration *prometheus.Desc
}

func NewAgentJobsCollector(db *sql.DB) *SQLServerAgentJobs {
	labels := []string{"job_name"}
	return &SQLServerAgentJobs{
		db: db,
		jobEnabled: prometheus.NewDesc(
			"sqlserver_agent_job_enabled",
			"Whether the SQL Agent job is enabled (1) or disabled (0)",
			labels, nil,
		),
		lastRunStatus: prometheus.NewDesc(
			"sqlserver_agent_job_last_run_status",
			"Outcome of the last completed run of the SQL Agent job (0 = failed, 1 = succeeded, 2 = retry, 3 = canceled)",
			labels, nil,
		),
		lastRunAt: prometheus.NewDesc(
			"sqlserver_agent_job_last_run_timestamp_seconds",
			"Unix timestamp of the start of the last completed run of the SQL Agent job",
			labels, nil,
		),
		lastRunDuration: prometheus.NewDesc(
			"sqlserver_agent_job_last_run_duration_seconds",
			"Duration in seconds of the last completed run of the SQL Agent job",
			labels, nil,
		),
	}
}

func (c *SQLServerAgentJobs) Name() string {
	return "agent_jobs"
}

func (c *SQLServerAgentJobs) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.jobEnabled
	ch <- c.lastRunStatus
	ch <- c.lastRunAt
	ch <- c.lastRunDuration
}

// Scrape reads SQL Agent job outcomes from msdb. The login needs SELECT on
// msdb.dbo.sysjobs and msdb.dbo.sysjobhistory (see README); the SQL Agent
// roles grant access through procedures and views, not these base tables.
func (c *SQLServerAgentJobs) Scrape(ctx context.Context, ch chan<- prometheus.Metric) error {
	rows, err := c.db.QueryContext(ctx, `
		SELECT
			j.name,
			CAST(j.enabled AS INT),
			h.run_status,
			-- run_date/run_time are the server's local time; convert to UTC
			-- with the current offset (sysjobhistory keeps no offset).
			DATEADD(MINUTE, -DATEPART(TZOFFSET, SYSDATETIMEOFFSET()), DATETIMEFROMPARTS(
				h.run_date / 10000, (h.run_date / 100) % 100, h.run_date % 100,
				h.run_time / 10000, (h.run_time / 100) % 100, h.run_time % 100, 0
			)) AS last_run_utc,
			(h.run_duration / 10000) * 3600
				+ ((h.run_duration / 100) % 100) * 60
				+ (h.run_duration % 100) AS duration_seconds
		FROM msdb.dbo.sysjobs j
		OUTER APPLY (
			SELECT TOP 1 run_status, run_date, run_time, run_duration
			FROM msdb.dbo.sysjobhistory h
			WHERE h.job_id = j.job_id
				AND h.step_id = 0
				AND h.run_status IN (0, 1, 2, 3) -- completed outcomes only (4 = in progress, 5 = unknown)
			-- instance_id is an identity: ordering by the local run_date/
			-- run_time would pick an older run after a time zone change or
			-- a DST fall-back.
			ORDER BY h.instance_id DESC
		) h
	`)
	if err != nil {
		return fmt.Errorf("failed to query msdb.dbo.sysjobs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var scanErrs []error
	for rows.Next() {
		var (
			jobName         string
			enabled         float64
			runStatus       sql.NullFloat64
			lastRun         sql.NullTime
			durationSeconds sql.NullFloat64
		)

		if err := rows.Scan(&jobName, &enabled, &runStatus, &lastRun, &durationSeconds); err != nil {
			scanErrs = append(scanErrs, fmt.Errorf("failed to scan agent job row: %w", err))
			continue
		}

		ch <- prometheus.MustNewConstMetric(c.jobEnabled, prometheus.GaugeValue, enabled, jobName)

		// Jobs that never completed a run have no history row.
		if runStatus.Valid {
			ch <- prometheus.MustNewConstMetric(c.lastRunStatus, prometheus.GaugeValue, runStatus.Float64, jobName)
		}
		if lastRun.Valid {
			ch <- prometheus.MustNewConstMetric(c.lastRunAt, prometheus.GaugeValue, float64(lastRun.Time.Unix()), jobName)
		}
		if durationSeconds.Valid {
			ch <- prometheus.MustNewConstMetric(c.lastRunDuration, prometheus.GaugeValue, durationSeconds.Float64, jobName)
		}
	}

	return errors.Join(append(scanErrs, rows.Err())...)
}
