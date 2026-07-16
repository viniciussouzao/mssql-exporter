package collector

import (
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestSessionsScrapeHappyPath(t *testing.T) {
	db, mock := newSQLMock(t)

	rows := sqlmock.NewRows([]string{"status", "session_count"}).
		AddRow("running", 3).
		AddRow("sleeping", 10)
	mock.ExpectQuery("SELECT status, COUNT").WillReturnRows(rows)

	c := NewSessionsCollector(db)
	families := gatherScraper(t, c)

	if v, ok := metricValue(families, "sqlserver_sessions_total", map[string]string{"session_state": "running"}); !ok || v != 3 {
		t.Errorf("sessions_total{running} = %v (found=%v), want 3", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_sessions_total", map[string]string{"session_state": "sleeping"}); !ok || v != 10 {
		t.Errorf("sessions_total{sleeping} = %v (found=%v), want 10", v, ok)
	}
}

func TestSessionsScrapeQueryError(t *testing.T) {
	db, mock := newSQLMock(t)

	queryErr := errors.New("boom")
	mock.ExpectQuery("SELECT status, COUNT").WillReturnError(queryErr)

	c := NewSessionsCollector(db)
	if err := scrapeErr(t, c); !errors.Is(err, queryErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v", err, queryErr)
	}
}

func TestSessionsScrapeScanError(t *testing.T) {
	db, mock := newSQLMock(t)

	rows := sqlmock.NewRows([]string{"status", "session_count"}).
		AddRow("running", "not-a-number").
		AddRow("sleeping", 10)
	mock.ExpectQuery("SELECT status, COUNT").WillReturnRows(rows)

	c := NewSessionsCollector(db)
	families, err := gatherScraperErr(t, c)

	if err == nil {
		t.Error("Scrape() error = nil, want error from bad row")
	}
	if v, ok := metricValue(families, "sqlserver_sessions_total", map[string]string{"session_state": "sleeping"}); !ok || v != 10 {
		t.Errorf("sessions_total{sleeping} = %v (found=%v), want 10 despite earlier scan error", v, ok)
	}
}

func TestSessionsScrapeRowIterationError(t *testing.T) {
	db, mock := newSQLMock(t)

	rowErr := errors.New("row iteration failed")
	rows := sqlmock.NewRows([]string{"status", "session_count"}).
		AddRow("running", 3).
		AddRow("sleeping", 10).
		RowError(1, rowErr)
	mock.ExpectQuery("SELECT status, COUNT").WillReturnRows(rows)

	c := NewSessionsCollector(db)
	families, err := gatherScraperErr(t, c)

	if !errors.Is(err, rowErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v (rows.Err())", err, rowErr)
	}
	if v, ok := metricValue(families, "sqlserver_sessions_total", map[string]string{"session_state": "running"}); !ok || v != 3 {
		t.Errorf("sessions_total{running} = %v (found=%v), want 3 despite the later iteration error", v, ok)
	}
}

func TestSessionsScrapeEmpty(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery("SELECT status, COUNT").WillReturnRows(sqlmock.NewRows([]string{"status", "session_count"}))

	c := NewSessionsCollector(db)
	families := gatherScraper(t, c)
	if _, ok := metricValue(families, "sqlserver_sessions_total", nil); ok {
		t.Error("sessions_total present, want no metrics for empty result set")
	}
}
