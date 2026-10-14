package collector

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

// blockedQueryRegex pins the source of the blocking metrics: dm_os_waiting_tasks
// has no database context (dm_exec_requests fails with error 976 on an AG
// secondary, HML-015), sessions are counted once, and a parallel request
// waiting on its own session is not blocking.
var blockedQueryRegex = sqlContainsInOrder(
	"COUNT(DISTINCT session_id)",
	"FROM sys.dm_os_waiting_tasks",
	"blocking_session_id > 0",
	"blocking_session_id <> session_id",
)

const requestsQueryRegex = "FROM sys.dm_exec_requests(.|\n)*GROUP BY status"

// sqlServerError mimics the driver's error, which exposes the SQL Server
// error number through SQLErrorNumber.
type sqlServerError struct{ number int32 }

func (e sqlServerError) Error() string         { return fmt.Sprintf("mssql: error %d", e.number) }
func (e sqlServerError) SQLErrorNumber() int32 { return e.number }

func TestBlockingScrapeHappyPath(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery(blockedQueryRegex).WillReturnRows(
		sqlmock.NewRows([]string{"blocked", "max_wait"}).AddRow(3.0, 1500.0))
	mock.ExpectQuery(requestsQueryRegex).WillReturnRows(
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
	mock.ExpectQuery(blockedQueryRegex).WillReturnError(blockedErr)
	mock.ExpectQuery(requestsQueryRegex).WillReturnRows(
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

	mock.ExpectQuery(blockedQueryRegex).WillReturnRows(
		sqlmock.NewRows([]string{"blocked", "max_wait"}).AddRow(0.0, 0.0))
	mock.ExpectQuery(requestsQueryRegex).WillReturnRows(
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
	mock.ExpectQuery(blockedQueryRegex).WillReturnRows(
		sqlmock.NewRows([]string{"blocked", "max_wait"}).AddRow(0.0, 0.0))
	mock.ExpectQuery(requestsQueryRegex).WillReturnRows(
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

func TestBlockingScrapeRequestsRetriesOnceOnAGNotAccessible(t *testing.T) {
	// 976: database not readable (HML-015); 983: replica RESOLVING right
	// after SQL Server starts (HML-020).
	for _, number := range []int32{976, 983} {
		t.Run(fmt.Sprint(number), func(t *testing.T) {
			db, mock := newSQLMock(t)

			mock.ExpectQuery(blockedQueryRegex).WillReturnRows(
				sqlmock.NewRows([]string{"blocked", "max_wait"}).AddRow(0.0, 0.0))
			// The first attempt fails mid-iteration after one row was read: none of
			// its counts may be emitted, or the retry would produce duplicates.
			mock.ExpectQuery(requestsQueryRegex).WillReturnRows(
				sqlmock.NewRows([]string{"status", "count"}).
					AddRow("running", 9.0).
					AddRow("suspended", 9.0).
					RowError(1, sqlServerError{number: number}))
			mock.ExpectQuery(requestsQueryRegex).WillReturnRows(
				sqlmock.NewRows([]string{"status", "count"}).
					AddRow("running", 1.0).
					AddRow("suspended", 2.0))

			c := NewBlockingCollector(db)
			families := gatherScraper(t, c)

			if v, ok := metricValue(families, "sqlserver_active_requests", map[string]string{"status": "running"}); !ok || v != 1 {
				t.Errorf("active_requests{running} = %v (found=%v), want 1 from the retry", v, ok)
			}
			if v, ok := metricValue(families, "sqlserver_active_requests", map[string]string{"status": "suspended"}); !ok || v != 2 {
				t.Errorf("active_requests{suspended} = %v (found=%v), want 2 from the retry", v, ok)
			}
		})
	}
}

func TestBlockingScrapeRequestsAGNotAccessibleTwiceReportsError(t *testing.T) {
	db, mock := newSQLMock(t)

	agErr := sqlServerError{number: 976}
	mock.ExpectQuery(blockedQueryRegex).WillReturnRows(
		sqlmock.NewRows([]string{"blocked", "max_wait"}).AddRow(0.0, 0.0))
	mock.ExpectQuery(requestsQueryRegex).WillReturnRows(
		sqlmock.NewRows([]string{"status", "count"}).AddRow("running", 1.0).RowError(0, agErr))
	mock.ExpectQuery(requestsQueryRegex).WillReturnRows(
		sqlmock.NewRows([]string{"status", "count"}).AddRow("running", 1.0).RowError(0, agErr))

	c := NewBlockingCollector(db)
	families, err := gatherScraperErr(t, c)

	if !errors.Is(err, agErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v after the single retry", err, agErr)
	}
	if v, ok := metricValue(families, "sqlserver_blocked_sessions", nil); !ok || v != 0 {
		t.Errorf("blocked_sessions = %v (found=%v), want 0 despite the requests query failing", v, ok)
	}
}

func TestBlockingScrapeRequestsOtherErrorsAreNotRetried(t *testing.T) {
	db, mock := newSQLMock(t)

	rowErr := sqlServerError{number: 1205}
	mock.ExpectQuery(blockedQueryRegex).WillReturnRows(
		sqlmock.NewRows([]string{"blocked", "max_wait"}).AddRow(0.0, 0.0))
	mock.ExpectQuery(requestsQueryRegex).WillReturnRows(
		sqlmock.NewRows([]string{"status", "count"}).AddRow("running", 1.0).RowError(0, rowErr))

	c := NewBlockingCollector(db)
	err := scrapeErr(t, c)

	if !errors.Is(err, rowErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v", err, rowErr)
	}
	// A retry would hit sqlmock with no expectation left.
	if err != nil && strings.Contains(err.Error(), "all expectations were already fulfilled") {
		t.Errorf("Scrape() retried an error other than 976/983: %v", err)
	}
}
