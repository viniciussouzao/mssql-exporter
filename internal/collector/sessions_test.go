package collector

import (
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

const sessionsByStateQueryRegex = "SELECT status, COUNT"

// sessionsByLoginQueryRegex pins the user-process filter: system sessions
// run under sa/the service account and would swamp the per-login counts.
var sessionsByLoginQueryRegex = sqlContainsInOrder(
	"SELECT login_name, COUNT(*)",
	"FROM sys.dm_exec_sessions",
	"WHERE is_user_process = 1",
	"GROUP BY login_name",
)

var sessionsByLoginColumns = []string{"login_name", "session_count"}

// expectNoLogins satisfies the by-login query for tests about the state query.
func expectNoLogins(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(sessionsByLoginQueryRegex).WillReturnRows(sqlmock.NewRows(sessionsByLoginColumns))
}

func TestSessionsScrapeHappyPath(t *testing.T) {
	db, mock := newSQLMock(t)

	rows := sqlmock.NewRows([]string{"status", "session_count"}).
		AddRow("running", 3).
		AddRow("sleeping", 10)
	mock.ExpectQuery(sessionsByStateQueryRegex).WillReturnRows(rows)
	expectNoLogins(mock)

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
	mock.ExpectQuery(sessionsByStateQueryRegex).WillReturnError(queryErr)
	expectNoLogins(mock)

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
	mock.ExpectQuery(sessionsByStateQueryRegex).WillReturnRows(rows)
	expectNoLogins(mock)

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
	mock.ExpectQuery(sessionsByStateQueryRegex).WillReturnRows(rows)
	expectNoLogins(mock)

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

	mock.ExpectQuery(sessionsByStateQueryRegex).WillReturnRows(sqlmock.NewRows([]string{"status", "session_count"}))
	expectNoLogins(mock)

	c := NewSessionsCollector(db)
	families := gatherScraper(t, c)
	if _, ok := metricValue(families, "sqlserver_sessions_total", nil); ok {
		t.Error("sessions_total present, want no metrics for empty result set")
	}
	if _, ok := metricValue(families, "sqlserver_sessions_by_login", nil); ok {
		t.Error("sessions_by_login present, want no metrics for empty result set")
	}
}

func TestSessionsScrapeByLogin(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery(sessionsByStateQueryRegex).WillReturnRows(
		sqlmock.NewRows([]string{"status", "session_count"}).AddRow("sleeping", 12))
	mock.ExpectQuery(sessionsByLoginQueryRegex).WillReturnRows(
		sqlmock.NewRows(sessionsByLoginColumns).
			AddRow("app_user", 9).
			AddRow(`DOMAIN\jdoe`, 2))

	c := NewSessionsCollector(db)
	families := gatherScraper(t, c)

	for login, want := range map[string]float64{"app_user": 9, `DOMAIN\jdoe`: 2} {
		if v, ok := metricValue(families, "sqlserver_sessions_by_login", map[string]string{"login_name": login}); !ok || v != want {
			t.Errorf("sessions_by_login{%s} = %v (found=%v), want %v", login, v, ok, want)
		}
	}
}

func TestSessionsScrapeByLoginErrorKeepsStateMetrics(t *testing.T) {
	db, mock := newSQLMock(t)

	loginErr := errors.New("by login failed")
	mock.ExpectQuery(sessionsByStateQueryRegex).WillReturnRows(
		sqlmock.NewRows([]string{"status", "session_count"}).AddRow("running", 3))
	mock.ExpectQuery(sessionsByLoginQueryRegex).WillReturnError(loginErr)

	c := NewSessionsCollector(db)
	families, err := gatherScraperErr(t, c)

	if !errors.Is(err, loginErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v", err, loginErr)
	}
	if v, ok := metricValue(families, "sqlserver_sessions_total", map[string]string{"session_state": "running"}); !ok || v != 3 {
		t.Errorf("sessions_total{running} = %v (found=%v), want 3 despite the by-login query failing", v, ok)
	}
}

func TestSessionsScrapeByLoginScanErrorKeepsOtherRows(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery(sessionsByStateQueryRegex).WillReturnRows(sqlmock.NewRows([]string{"status", "session_count"}))
	mock.ExpectQuery(sessionsByLoginQueryRegex).WillReturnRows(
		sqlmock.NewRows(sessionsByLoginColumns).
			AddRow("bad_login", "not-a-number").
			AddRow("app_user", 4))

	c := NewSessionsCollector(db)
	families, err := gatherScraperErr(t, c)

	if err == nil {
		t.Error("Scrape() error = nil, want the scan error reported")
	}
	if v, ok := metricValue(families, "sqlserver_sessions_by_login", map[string]string{"login_name": "app_user"}); !ok || v != 4 {
		t.Errorf("sessions_by_login{app_user} = %v (found=%v), want 4 for the row after the bad one", v, ok)
	}
}
