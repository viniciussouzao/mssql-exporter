package collector

import (
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestBlockingScrapeHappyPath(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery("FROM sys.dm_exec_requests(.|\n)*WHERE blocking_session_id").WillReturnRows(
		sqlmock.NewRows([]string{"blocked", "max_wait"}).AddRow(3.0, 1500.0))
	mock.ExpectQuery("FROM sys.dm_exec_requests(.|\n)*GROUP BY status").WillReturnRows(
		sqlmock.NewRows([]string{"status", "count"}).
			AddRow("running", 2.0).
			AddRow("suspended", 3.0))

	c := NewBlockingCollector(db)
	families := gatherScraper(t, c)

	if v, ok := metricValue(families, "sqlserver_blocked_sessions", nil); !ok || v != 3 {
		t.Errorf("blocked_sessions = %v (found=%v), want 3", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_blocked_max_wait_ms", nil); !ok || v != 1500 {
		t.Errorf("blocked_max_wait_ms = %v (found=%v), want 1500", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_active_requests", map[string]string{"status": "suspended"}); !ok || v != 3 {
		t.Errorf("active_requests{suspended} = %v (found=%v), want 3", v, ok)
	}
}

func TestBlockingScrapeBlockedQueryError(t *testing.T) {
	db, mock := newSQLMock(t)

	blockedErr := errors.New("boom")
	mock.ExpectQuery("FROM sys.dm_exec_requests(.|\n)*WHERE blocking_session_id").WillReturnError(blockedErr)
	mock.ExpectQuery("FROM sys.dm_exec_requests(.|\n)*GROUP BY status").WillReturnRows(
		sqlmock.NewRows([]string{"status", "count"}).AddRow("running", 1.0))

	c := NewBlockingCollector(db)
	families, err := gatherScraperErr(t, c)

	if !errors.Is(err, blockedErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v", err, blockedErr)
	}
	if v, ok := metricValue(families, "sqlserver_active_requests", map[string]string{"status": "running"}); !ok || v != 1 {
		t.Errorf("active_requests{running} = %v (found=%v), want 1 despite the other query failing", v, ok)
	}
}

func TestBlockingScrapeRequestsScanError(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery("FROM sys.dm_exec_requests(.|\n)*WHERE blocking_session_id").WillReturnRows(
		sqlmock.NewRows([]string{"blocked", "max_wait"}).AddRow(0.0, 0.0))
	mock.ExpectQuery("FROM sys.dm_exec_requests(.|\n)*GROUP BY status").WillReturnRows(
		sqlmock.NewRows([]string{"status", "count"}).
			AddRow("running", "not-a-number").
			AddRow("suspended", 2.0))

	c := NewBlockingCollector(db)
	families, err := gatherScraperErr(t, c)

	if err == nil {
		t.Error("Scrape() error = nil, want error from bad row")
	}
	if v, ok := metricValue(families, "sqlserver_active_requests", map[string]string{"status": "suspended"}); !ok || v != 2 {
		t.Errorf("active_requests{suspended} = %v (found=%v), want 2 despite earlier scan error", v, ok)
	}
}

func TestBlockingScrapeRequestsRowIterationError(t *testing.T) {
	db, mock := newSQLMock(t)

	rowErr := errors.New("row iteration failed")
	mock.ExpectQuery("FROM sys.dm_exec_requests(.|\n)*WHERE blocking_session_id").WillReturnRows(
		sqlmock.NewRows([]string{"blocked", "max_wait"}).AddRow(0.0, 0.0))
	mock.ExpectQuery("FROM sys.dm_exec_requests(.|\n)*GROUP BY status").WillReturnRows(
		sqlmock.NewRows([]string{"status", "count"}).
			AddRow("running", 1.0).
			AddRow("suspended", 2.0).
			RowError(1, rowErr))

	c := NewBlockingCollector(db)
	families, err := gatherScraperErr(t, c)

	if !errors.Is(err, rowErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v (rows.Err())", err, rowErr)
	}
	if v, ok := metricValue(families, "sqlserver_active_requests", map[string]string{"status": "running"}); !ok || v != 1 {
		t.Errorf("active_requests{running} = %v (found=%v), want 1 despite the later iteration error", v, ok)
	}
}
