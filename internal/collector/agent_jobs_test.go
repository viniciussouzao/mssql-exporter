package collector

import (
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

// agentJobsQueryRegex pins down the run-history selection rules: only
// step_id 0 (the job outcome row, not an individual step), only completed
// outcomes, and the single most recent one. Losing any of these could
// surface an in-progress run, an individual step, or a stale run.
var agentJobsQueryRegex = sqlContainsInOrder(
	// run_date/run_time are local time: the UTC conversion must stay.
	"DATEADD(MINUTE, -DATEPART(TZOFFSET, SYSDATETIMEOFFSET()), DATETIMEFROMPARTS(",
	"SELECT TOP 1 run_status, run_date, run_time, run_duration",
	"FROM msdb.dbo.sysjobhistory h",
	"h.step_id = 0",
	"h.run_status IN (0, 1, 2, 3)",
	"ORDER BY h.instance_id DESC",
)

func TestAgentJobsScrapeHappyPath(t *testing.T) {
	db, mock := newSQLMock(t)

	lastRun := time.Date(2026, 7, 15, 3, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{"name", "enabled", "run_status", "last_run", "duration_seconds"}).
		AddRow("NightlyBackup", 1, 1.0, lastRun, 125.0)
	mock.ExpectQuery(agentJobsQueryRegex).WillReturnRows(rows)

	c := NewAgentJobsCollector(db)
	families := gatherScraper(t, c)

	labels := map[string]string{"job_name": "NightlyBackup"}
	if v, ok := metricValue(families, "sqlserver_agent_job_enabled", labels); !ok || v != 1 {
		t.Errorf("agent_job_enabled = %v (found=%v), want 1", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_agent_job_last_run_status", labels); !ok || v != 1 {
		t.Errorf("agent_job_last_run_status = %v (found=%v), want 1", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_agent_job_last_run_timestamp_seconds", labels); !ok || v != float64(lastRun.Unix()) {
		t.Errorf("agent_job_last_run_timestamp_seconds = %v (found=%v), want %v", v, ok, lastRun.Unix())
	}
	if v, ok := metricValue(families, "sqlserver_agent_job_last_run_duration_seconds", labels); !ok || v != 125 {
		t.Errorf("agent_job_last_run_duration_seconds = %v (found=%v), want 125", v, ok)
	}
}

func TestAgentJobsScrapeNeverRunOmitsHistoryMetrics(t *testing.T) {
	db, mock := newSQLMock(t)

	rows := sqlmock.NewRows([]string{"name", "enabled", "run_status", "last_run", "duration_seconds"}).
		AddRow("NeverRunJob", 0, nil, nil, nil)
	mock.ExpectQuery(agentJobsQueryRegex).WillReturnRows(rows)

	c := NewAgentJobsCollector(db)
	families := gatherScraper(t, c)

	labels := map[string]string{"job_name": "NeverRunJob"}
	if v, ok := metricValue(families, "sqlserver_agent_job_enabled", labels); !ok || v != 0 {
		t.Errorf("agent_job_enabled = %v (found=%v), want 0", v, ok)
	}
	if _, ok := metricValue(families, "sqlserver_agent_job_last_run_status", labels); ok {
		t.Error("agent_job_last_run_status present, want omitted for a job with no run history")
	}
}

func TestAgentJobsScrapeQueryError(t *testing.T) {
	db, mock := newSQLMock(t)

	queryErr := errors.New("boom")
	mock.ExpectQuery(agentJobsQueryRegex).WillReturnError(queryErr)

	c := NewAgentJobsCollector(db)
	if err := scrapeErr(t, c); !errors.Is(err, queryErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v", err, queryErr)
	}
}

func TestAgentJobsScrapeScanError(t *testing.T) {
	db, mock := newSQLMock(t)

	rows := sqlmock.NewRows([]string{"name", "enabled", "run_status", "last_run", "duration_seconds"}).
		AddRow("BadJob", "not-an-int", nil, nil, nil).
		AddRow("GoodJob", 1, nil, nil, nil)
	mock.ExpectQuery(agentJobsQueryRegex).WillReturnRows(rows)

	c := NewAgentJobsCollector(db)
	families, err := gatherScraperErr(t, c)

	if err == nil {
		t.Error("Scrape() error = nil, want error from bad row")
	}
	if v, ok := metricValue(families, "sqlserver_agent_job_enabled", map[string]string{"job_name": "GoodJob"}); !ok || v != 1 {
		t.Errorf("agent_job_enabled{GoodJob} = %v (found=%v), want 1 despite earlier scan error", v, ok)
	}
}

func TestAgentJobsScrapeRowIterationError(t *testing.T) {
	db, mock := newSQLMock(t)

	rowErr := errors.New("row iteration failed")
	rows := sqlmock.NewRows([]string{"name", "enabled", "run_status", "last_run", "duration_seconds"}).
		AddRow("GoodJob", 1, nil, nil, nil).
		AddRow("OtherJob", 1, nil, nil, nil).
		RowError(1, rowErr)
	mock.ExpectQuery(agentJobsQueryRegex).WillReturnRows(rows)

	c := NewAgentJobsCollector(db)
	families, err := gatherScraperErr(t, c)

	if !errors.Is(err, rowErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v (rows.Err())", err, rowErr)
	}
	if v, ok := metricValue(families, "sqlserver_agent_job_enabled", map[string]string{"job_name": "GoodJob"}); !ok || v != 1 {
		t.Errorf("agent_job_enabled{GoodJob} = %v (found=%v), want 1 despite the later iteration error", v, ok)
	}
}
