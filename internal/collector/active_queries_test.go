package collector

import (
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

// activeRequestsQueryRegex pins the user-only filter, the exporter's own
// request exclusion, the overflow-safe elapsed time and the bucket order,
// which must match longRunningThresholds.
var activeRequestsQueryRegex = sqlContainsInOrder(
	"MAX(DATEDIFF(SECOND, r.start_time, GETDATE()))",
	"GETDATE()) >= 10 THEN",
	"GETDATE()) >= 60 THEN",
	"GETDATE()) >= 300 THEN",
	"GETDATE()) >= 1800 THEN",
	"FROM sys.dm_exec_requests r",
	"WHERE s.is_user_process = 1",
	"AND r.session_id <> @@SPID",
	"GROUP BY r.command",
)

var openTransactionsQueryRegex = sqlContainsInOrder(
	"COUNT(DISTINCT CASE WHEN s.status = 'sleeping' THEN s.session_id END)",
	"FROM sys.dm_tran_session_transactions st",
	"sys.dm_tran_active_transactions at",
	"WHERE s.is_user_process = 1",
	"AND st.session_id <> @@SPID",
)

var activeRequestsColumns = []string{"command", "max_elapsed", "over_10", "over_60", "over_300", "over_1800"}
var openTransactionsColumns = []string{"open", "oldest", "sleeping"}

func expectNoOpenTransactions(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(openTransactionsQueryRegex).WillReturnRows(
		sqlmock.NewRows(openTransactionsColumns).AddRow(0, 0, 0))
}

func TestActiveQueriesScrapeHappyPath(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery(activeRequestsQueryRegex).WillReturnRows(
		sqlmock.NewRows(activeRequestsColumns).
			AddRow("SELECT", 75.0, 3, 1, 0, 0).
			AddRow("WAITFOR", 1900.0, 2, 2, 2, 1))
	mock.ExpectQuery(openTransactionsQueryRegex).WillReturnRows(
		sqlmock.NewRows(openTransactionsColumns).AddRow(4, 620, 1))

	c := NewActiveQueriesCollector(db)
	families := gatherScraper(t, c)

	for command, want := range map[string]float64{"SELECT": 75, "WAITFOR": 1900} {
		if v, ok := metricValue(families, "sqlserver_request_max_elapsed_seconds", map[string]string{"command": command}); !ok || v != want {
			t.Errorf("request_max_elapsed_seconds{%s} = %v (found=%v), want %v", command, v, ok, want)
		}
	}
	// Buckets are summed across commands.
	for over, want := range map[string]float64{"10": 5, "60": 3, "300": 2, "1800": 1} {
		if v, ok := metricValue(families, "sqlserver_requests_long_running", map[string]string{"over_seconds": over}); !ok || v != want {
			t.Errorf("requests_long_running{over_seconds=%s} = %v (found=%v), want %v", over, v, ok, want)
		}
	}
	for name, want := range map[string]float64{
		"sqlserver_open_transactions":                       4,
		"sqlserver_oldest_open_transaction_seconds":         620,
		"sqlserver_sleeping_sessions_with_open_transaction": 1,
	} {
		if v, ok := metricValue(families, name, nil); !ok || v != want {
			t.Errorf("%s = %v (found=%v), want %v", name, v, ok, want)
		}
	}
}

func TestActiveQueriesScrapeIdleEmitsZeroBuckets(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery(activeRequestsQueryRegex).WillReturnRows(sqlmock.NewRows(activeRequestsColumns))
	expectNoOpenTransactions(mock)

	c := NewActiveQueriesCollector(db)
	families := gatherScraper(t, c)

	for _, over := range []string{"10", "60", "300", "1800"} {
		if v, ok := metricValue(families, "sqlserver_requests_long_running", map[string]string{"over_seconds": over}); !ok || v != 0 {
			t.Errorf("requests_long_running{over_seconds=%s} = %v (found=%v), want 0 with no user request", over, v, ok)
		}
	}
	if _, ok := metricValue(families, "sqlserver_request_max_elapsed_seconds", nil); ok {
		t.Error("request_max_elapsed_seconds present, want it absent with no user request")
	}
	if v, ok := metricValue(families, "sqlserver_oldest_open_transaction_seconds", nil); !ok || v != 0 {
		t.Errorf("oldest_open_transaction_seconds = %v (found=%v), want 0", v, ok)
	}
}

func TestActiveQueriesScrapeNegativeElapsedClampedToZero(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery(activeRequestsQueryRegex).WillReturnRows(
		sqlmock.NewRows(activeRequestsColumns).AddRow("SELECT", -3600.0, 0, 0, 0, 0))
	mock.ExpectQuery(openTransactionsQueryRegex).WillReturnRows(
		sqlmock.NewRows(openTransactionsColumns).AddRow(1, -3600, 0))

	c := NewActiveQueriesCollector(db)
	families := gatherScraper(t, c)

	if v, ok := metricValue(families, "sqlserver_request_max_elapsed_seconds", map[string]string{"command": "SELECT"}); !ok || v != 0 {
		t.Errorf("request_max_elapsed_seconds{SELECT} = %v (found=%v), want 0 after a backwards clock change", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_oldest_open_transaction_seconds", nil); !ok || v != 0 {
		t.Errorf("oldest_open_transaction_seconds = %v (found=%v), want 0 after a backwards clock change", v, ok)
	}
}

func TestActiveQueriesScrapeRetriesOnceOnAGNotAccessible(t *testing.T) {
	for _, number := range []int32{976, 983} {
		t.Run(sqlServerError{number: number}.Error(), func(t *testing.T) {
			db, mock := newSQLMock(t)

			// The failed attempt read one row before the error: it must not
			// be emitted, or the retry would produce duplicate series.
			mock.ExpectQuery(activeRequestsQueryRegex).WillReturnRows(
				sqlmock.NewRows(activeRequestsColumns).
					AddRow("SELECT", 99.0, 9, 9, 9, 9).
					AddRow("WAITFOR", 99.0, 9, 9, 9, 9).
					RowError(1, sqlServerError{number: number}))
			mock.ExpectQuery(activeRequestsQueryRegex).WillReturnRows(
				sqlmock.NewRows(activeRequestsColumns).AddRow("SELECT", 12.0, 1, 0, 0, 0))
			expectNoOpenTransactions(mock)

			c := NewActiveQueriesCollector(db)
			families := gatherScraper(t, c)

			if v, ok := metricValue(families, "sqlserver_request_max_elapsed_seconds", map[string]string{"command": "SELECT"}); !ok || v != 12 {
				t.Errorf("request_max_elapsed_seconds{SELECT} = %v (found=%v), want 12 from the retry", v, ok)
			}
			if v, ok := metricValue(families, "sqlserver_requests_long_running", map[string]string{"over_seconds": "10"}); !ok || v != 1 {
				t.Errorf("requests_long_running{10} = %v (found=%v), want 1 from the retry", v, ok)
			}
		})
	}
}

func TestActiveQueriesScrapeRequestsErrorKeepsTransactions(t *testing.T) {
	db, mock := newSQLMock(t)

	queryErr := errors.New("requests failed")
	mock.ExpectQuery(activeRequestsQueryRegex).WillReturnError(queryErr)
	mock.ExpectQuery(openTransactionsQueryRegex).WillReturnRows(
		sqlmock.NewRows(openTransactionsColumns).AddRow(2, 30, 0))

	c := NewActiveQueriesCollector(db)
	families, err := gatherScraperErr(t, c)

	if !errors.Is(err, queryErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v", err, queryErr)
	}
	if _, ok := metricValue(families, "sqlserver_requests_long_running", nil); ok {
		t.Error("requests_long_running present, want it absent when the requests query fails (not a misleading 0)")
	}
	if v, ok := metricValue(families, "sqlserver_open_transactions", nil); !ok || v != 2 {
		t.Errorf("open_transactions = %v (found=%v), want 2 despite the requests query failing", v, ok)
	}
}

func TestActiveQueriesScrapeTransactionsError(t *testing.T) {
	db, mock := newSQLMock(t)

	queryErr := errors.New("transactions failed")
	mock.ExpectQuery(activeRequestsQueryRegex).WillReturnRows(sqlmock.NewRows(activeRequestsColumns))
	mock.ExpectQuery(openTransactionsQueryRegex).WillReturnError(queryErr)

	c := NewActiveQueriesCollector(db)
	families, err := gatherScraperErr(t, c)

	if !errors.Is(err, queryErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v", err, queryErr)
	}
	if _, ok := metricValue(families, "sqlserver_open_transactions", nil); ok {
		t.Error("open_transactions present, want it absent when its query fails")
	}
	if v, ok := metricValue(families, "sqlserver_requests_long_running", map[string]string{"over_seconds": "10"}); !ok || v != 0 {
		t.Errorf("requests_long_running{10} = %v (found=%v), want 0 despite the transactions query failing", v, ok)
	}
}

func TestActiveQueriesScrapeScanErrorKeepsOtherRows(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery(activeRequestsQueryRegex).WillReturnRows(
		sqlmock.NewRows(activeRequestsColumns).
			AddRow("BAD", "not-a-number", 0, 0, 0, 0).
			AddRow("UPDATE", 15.0, 1, 0, 0, 0))
	expectNoOpenTransactions(mock)

	c := NewActiveQueriesCollector(db)
	families, err := gatherScraperErr(t, c)

	if err == nil {
		t.Error("Scrape() error = nil, want the scan error reported")
	}
	if v, ok := metricValue(families, "sqlserver_request_max_elapsed_seconds", map[string]string{"command": "UPDATE"}); !ok || v != 15 {
		t.Errorf("request_max_elapsed_seconds{UPDATE} = %v (found=%v), want 15 for the row after the bad one", v, ok)
	}
}
