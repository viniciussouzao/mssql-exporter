package collector

import (
	"errors"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

// waitStatsQueryRegex pins down the exclusion filters that keep noisy wait
// types (SLEEP/XE/BROKER) out of the metric: removing any of them should
// break the mock match and fail the test.
var waitStatsQueryRegex = sqlContainsInOrder(
	"FROM sys.dm_os_wait_stats",
	"wait_type NOT LIKE 'SLEEP%'",
	"wait_type NOT LIKE 'XE%'",
	"wait_type NOT LIKE 'BROKER_%'",
)

func TestWaitStatsScrapeHappyPath(t *testing.T) {
	db, mock := newSQLMock(t)

	rows := sqlmock.NewRows([]string{"wait_type", "wait_time_ms", "waiting_tasks_count", "max_wait_time_ms"}).
		AddRow("PAGEIOLATCH_SH", 1500.0, 42.0, 200.0)
	mock.ExpectQuery(waitStatsQueryRegex).WillReturnRows(rows)

	c := NewWaitStatsCollector(db, WaitStatsOptions{})
	families := gatherScraper(t, c)

	labels := map[string]string{"wait_type": "PAGEIOLATCH_SH"}
	if v, ok := counterValue(families, "sqlserver_wait_time_ms_total", labels); !ok || v != 1500 {
		t.Errorf("wait_time_ms = %v (found=%v), want 1500", v, ok)
	}
	if v, ok := counterValue(families, "sqlserver_waiting_tasks_total", labels); !ok || v != 42 {
		t.Errorf("waiting_tasks = %v (found=%v), want 42", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_max_wait_time_ms", labels); !ok || v != 200 {
		t.Errorf("max_wait_time_ms = %v (found=%v), want 200", v, ok)
	}
}

func TestWaitStatsScrapeQueryError(t *testing.T) {
	db, mock := newSQLMock(t)

	queryErr := errors.New("boom")
	mock.ExpectQuery(waitStatsQueryRegex).WillReturnError(queryErr)

	c := NewWaitStatsCollector(db, WaitStatsOptions{})
	if err := scrapeErr(t, c); !errors.Is(err, queryErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v", err, queryErr)
	}
}

func TestWaitStatsScrapeScanError(t *testing.T) {
	db, mock := newSQLMock(t)

	rows := sqlmock.NewRows([]string{"wait_type", "wait_time_ms", "waiting_tasks_count", "max_wait_time_ms"}).
		AddRow("BAD_WAIT", "not-a-number", 1.0, 1.0).
		AddRow("PAGEIOLATCH_SH", 1500.0, 42.0, 200.0)
	mock.ExpectQuery(waitStatsQueryRegex).WillReturnRows(rows)

	c := NewWaitStatsCollector(db, WaitStatsOptions{})
	families, err := gatherScraperErr(t, c)

	if err == nil {
		t.Error("Scrape() error = nil, want error from bad row")
	}
	if v, ok := counterValue(families, "sqlserver_wait_time_ms_total", map[string]string{"wait_type": "PAGEIOLATCH_SH"}); !ok || v != 1500 {
		t.Errorf("wait_time_ms = %v (found=%v), want 1500 despite earlier scan error", v, ok)
	}
}

func TestWaitStatsScrapeRowIterationError(t *testing.T) {
	db, mock := newSQLMock(t)

	rowErr := errors.New("row iteration failed")
	rows := sqlmock.NewRows([]string{"wait_type", "wait_time_ms", "waiting_tasks_count", "max_wait_time_ms"}).
		AddRow("PAGEIOLATCH_SH", 1500.0, 42.0, 200.0).
		AddRow("LCK_M_X", 300.0, 3.0, 50.0).
		RowError(1, rowErr)
	mock.ExpectQuery(waitStatsQueryRegex).WillReturnRows(rows)

	c := NewWaitStatsCollector(db, WaitStatsOptions{})
	families, err := gatherScraperErr(t, c)

	if !errors.Is(err, rowErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v (rows.Err())", err, rowErr)
	}
	if v, ok := counterValue(families, "sqlserver_wait_time_ms_total", map[string]string{"wait_type": "PAGEIOLATCH_SH"}); !ok || v != 1500 {
		t.Errorf("wait_time_ms = %v (found=%v), want 1500 despite the later iteration error", v, ok)
	}
}

func TestWaitStatsScrapeEmpty(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery(waitStatsQueryRegex).WillReturnRows(
		sqlmock.NewRows([]string{"wait_type", "wait_time_ms", "waiting_tasks_count", "max_wait_time_ms"}))

	c := NewWaitStatsCollector(db, WaitStatsOptions{})
	families := gatherScraper(t, c)
	if _, ok := counterValue(families, "sqlserver_wait_time_ms_total", nil); ok {
		t.Error("wait_time_ms present, want no metrics for empty result set")
	}
}

// By default idle/background waits and the wait types that never happened
// are skipped; include_idle restores the previous query (only the
// SLEEP/XE/BROKER prefixes are excluded).
func TestWaitStatsQueryIdleFilter(t *testing.T) {
	def := waitStatsQuery(WaitStatsOptions{})
	for _, want := range []string{"waiting_tasks_count > 0", "N'LAZYWRITER_SLEEP'", "N'SOS_WORK_DISPATCHER'", "N'HADR_WORK_QUEUE'", "wait_type NOT LIKE 'SLEEP%'"} {
		if !strings.Contains(def, want) {
			t.Errorf("default query lacks %q:\n%s", want, def)
		}
	}
	if strings.Contains(def, "HADR_SYNC_COMMIT") {
		t.Error("default query filters HADR_SYNC_COMMIT, a wait that slows commits")
	}

	all := waitStatsQuery(WaitStatsOptions{IncludeIdle: true})
	for _, unwanted := range []string{"waiting_tasks_count > 0", "NOT IN"} {
		if strings.Contains(all, unwanted) {
			t.Errorf("include_idle query contains %q:\n%s", unwanted, all)
		}
	}
	if !strings.Contains(all, "wait_type NOT LIKE 'XE%'") {
		t.Errorf("include_idle query lost the prefix filters:\n%s", all)
	}
}

func TestWaitStatsScrapeDefaultQueryFiltersIdle(t *testing.T) {
	db, mock := newSQLMock(t)

	rows := sqlmock.NewRows([]string{"wait_type", "wait_time_ms", "waiting_tasks_count", "max_wait_time_ms"}).
		AddRow("PAGEIOLATCH_SH", 1500.0, 42.0, 200.0)
	mock.ExpectQuery(sqlContainsInOrder("FROM sys.dm_os_wait_stats", "waiting_tasks_count > 0", "NOT IN")).WillReturnRows(rows)

	c := NewWaitStatsCollector(db, DefaultOptions().WaitStats)
	families := gatherScraper(t, c)
	if _, ok := counterValue(families, "sqlserver_wait_time_ms_total", map[string]string{"wait_type": "PAGEIOLATCH_SH"}); !ok {
		t.Error("wait_time_ms{PAGEIOLATCH_SH} missing")
	}
}
