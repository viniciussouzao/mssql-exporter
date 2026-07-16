package collector

import (
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

// perfCountersQueryRegex pins the '_Total' instance filter on the Locks
// counters: without it, per-database instances would be summed alongside
// (or instead of) the instance-level total, duplicating or skewing the
// deadlocks/lock-waits series.
var perfCountersQueryRegex = sqlContainsInOrder(
	"Number of Deadlocks/sec", "Lock Waits/sec",
	"instance_name = '_Total'",
)

func TestPerfCountersScrapeHappyPath(t *testing.T) {
	db, mock := newSQLMock(t)

	rows := sqlmock.NewRows([]string{"counter_name", "cntr_value"}).
		AddRow("Batch Requests/sec", 1000.0).
		AddRow("SQL Compilations/sec", 50.0).
		AddRow("SQL Re-Compilations/sec", 5.0).
		AddRow("Page life expectancy", 3600.0).
		AddRow("Buffer cache hit ratio", 950.0).
		AddRow("Buffer cache hit ratio base", 1000.0).
		AddRow("Lazy writes/sec", 2.0).
		AddRow("Number of Deadlocks/sec", 0.0).
		AddRow("Lock Waits/sec", 3.0).
		AddRow("User Connections", 42.0).
		AddRow("Page Splits/sec", 7.0)
	mock.ExpectQuery(perfCountersQueryRegex).WillReturnRows(rows)

	c := NewPerfCountersCollector(db)
	families := gatherScraper(t, c)

	counters := map[string]float64{
		"sqlserver_batch_requests_total":     1000,
		"sqlserver_sql_compilations_total":   50,
		"sqlserver_sql_recompilations_total": 5,
		"sqlserver_lazy_writes_total":        2,
		"sqlserver_deadlocks_total":          0,
		"sqlserver_lock_waits_total":         3,
		"sqlserver_page_splits_total":        7,
	}
	for name, want := range counters {
		if v, ok := counterValue(families, name, nil); !ok || v != want {
			t.Errorf("%s = %v (found=%v), want %v", name, v, ok, want)
		}
	}
	gauges := map[string]float64{
		"sqlserver_page_life_expectancy_seconds": 3600,
		"sqlserver_user_connections":             42,
		"sqlserver_buffer_cache_hit_ratio":       0.95,
	}
	for name, want := range gauges {
		if v, ok := metricValue(families, name, nil); !ok || v != want {
			t.Errorf("%s = %v (found=%v), want %v", name, v, ok, want)
		}
	}
}

func TestPerfCountersScrapeHitRatioOmittedWithoutBase(t *testing.T) {
	db, mock := newSQLMock(t)

	rows := sqlmock.NewRows([]string{"counter_name", "cntr_value"}).
		AddRow("Batch Requests/sec", 1.0)
	mock.ExpectQuery(perfCountersQueryRegex).WillReturnRows(rows)

	c := NewPerfCountersCollector(db)
	families := gatherScraper(t, c)

	if _, ok := metricValue(families, "sqlserver_buffer_cache_hit_ratio", nil); ok {
		t.Error("buffer_cache_hit_ratio present, want it omitted when the base counter row is absent")
	}
}

func TestPerfCountersScrapeQueryError(t *testing.T) {
	db, mock := newSQLMock(t)

	queryErr := errors.New("boom")
	mock.ExpectQuery(perfCountersQueryRegex).WillReturnError(queryErr)

	c := NewPerfCountersCollector(db)
	if err := scrapeErr(t, c); !errors.Is(err, queryErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v", err, queryErr)
	}
}

func TestPerfCountersScrapeScanError(t *testing.T) {
	db, mock := newSQLMock(t)

	rows := sqlmock.NewRows([]string{"counter_name", "cntr_value"}).
		AddRow("Batch Requests/sec", "not-a-number").
		AddRow("User Connections", 10.0)
	mock.ExpectQuery(perfCountersQueryRegex).WillReturnRows(rows)

	c := NewPerfCountersCollector(db)
	families, err := gatherScraperErr(t, c)

	if err == nil {
		t.Error("Scrape() error = nil, want error from bad row")
	}
	if v, ok := metricValue(families, "sqlserver_user_connections", nil); !ok || v != 10 {
		t.Errorf("user_connections = %v (found=%v), want 10 despite earlier scan error", v, ok)
	}
}

func TestPerfCountersScrapeRowIterationError(t *testing.T) {
	db, mock := newSQLMock(t)

	rowErr := errors.New("row iteration failed")
	rows := sqlmock.NewRows([]string{"counter_name", "cntr_value"}).
		AddRow("User Connections", 10.0).
		AddRow("Page Splits/sec", 3.0).
		RowError(1, rowErr)
	mock.ExpectQuery(perfCountersQueryRegex).WillReturnRows(rows)

	c := NewPerfCountersCollector(db)
	families, err := gatherScraperErr(t, c)

	if !errors.Is(err, rowErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v (rows.Err())", err, rowErr)
	}
	if v, ok := metricValue(families, "sqlserver_user_connections", nil); !ok || v != 10 {
		t.Errorf("user_connections = %v (found=%v), want 10 despite the later iteration error", v, ok)
	}
}
