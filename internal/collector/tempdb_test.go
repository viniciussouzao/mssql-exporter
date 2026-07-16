package collector

import (
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestTempDBScrapeHappyPath(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery("FROM tempdb.sys.dm_db_file_space_usage").WillReturnRows(
		sqlmock.NewRows([]string{"free", "version_store", "internal", "user"}).
			AddRow(1000.0, 10.0, 20.0, 30.0))
	mock.ExpectQuery("FROM sys.dm_os_performance_counters").WillReturnRows(
		sqlmock.NewRows([]string{"counter_name", "cntr_value"}).
			AddRow("Active Temp Tables", 5.0).
			AddRow("Temp Tables Creation Rate", 42.0))

	c := NewTempDBCollector(db)
	families := gatherScraper(t, c)

	gauges := map[string]float64{
		"sqlserver_tempdb_free_space_kb":       1000,
		"sqlserver_tempdb_version_store_kb":    10,
		"sqlserver_tempdb_internal_objects_kb": 20,
		"sqlserver_tempdb_user_objects_kb":     30,
		"sqlserver_tempdb_active_temp_tables":  5,
	}
	for name, want := range gauges {
		if v, ok := metricValue(families, name, nil); !ok || v != want {
			t.Errorf("%s = %v (found=%v), want %v", name, v, ok, want)
		}
	}
	if v, ok := counterValue(families, "sqlserver_tempdb_temp_tables_created_total", nil); !ok || v != 42 {
		t.Errorf("tempdb_temp_tables_created_total = %v (found=%v), want 42", v, ok)
	}
}

func TestTempDBScrapeFileSpaceQueryError(t *testing.T) {
	db, mock := newSQLMock(t)

	fileSpaceErr := errors.New("boom")
	mock.ExpectQuery("FROM tempdb.sys.dm_db_file_space_usage").WillReturnError(fileSpaceErr)
	mock.ExpectQuery("FROM sys.dm_os_performance_counters").WillReturnRows(
		sqlmock.NewRows([]string{"counter_name", "cntr_value"}).
			AddRow("Active Temp Tables", 5.0))

	c := NewTempDBCollector(db)
	families, err := gatherScraperErr(t, c)

	if !errors.Is(err, fileSpaceErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v", err, fileSpaceErr)
	}
	if v, ok := metricValue(families, "sqlserver_tempdb_active_temp_tables", nil); !ok || v != 5 {
		t.Errorf("active_temp_tables = %v (found=%v), want 5 despite the other query failing", v, ok)
	}
}

func TestTempDBScrapePerfCountersScanError(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery("FROM tempdb.sys.dm_db_file_space_usage").WillReturnRows(
		sqlmock.NewRows([]string{"free", "version_store", "internal", "user"}).
			AddRow(1000.0, 10.0, 20.0, 30.0))
	mock.ExpectQuery("FROM sys.dm_os_performance_counters").WillReturnRows(
		sqlmock.NewRows([]string{"counter_name", "cntr_value"}).
			AddRow("Active Temp Tables", "not-a-number").
			AddRow("Temp Tables Creation Rate", 42.0))

	c := NewTempDBCollector(db)
	families, err := gatherScraperErr(t, c)

	if err == nil {
		t.Error("Scrape() error = nil, want error from bad perf counter row")
	}
	if v, ok := counterValue(families, "sqlserver_tempdb_temp_tables_created_total", nil); !ok || v != 42 {
		t.Errorf("temp_tables_created_total = %v (found=%v), want 42 despite earlier scan error", v, ok)
	}
}

func TestTempDBScrapePerfCountersRowIterationError(t *testing.T) {
	db, mock := newSQLMock(t)

	rowErr := errors.New("row iteration failed")
	mock.ExpectQuery("FROM tempdb.sys.dm_db_file_space_usage").WillReturnRows(
		sqlmock.NewRows([]string{"free", "version_store", "internal", "user"}).
			AddRow(1000.0, 10.0, 20.0, 30.0))
	mock.ExpectQuery("FROM sys.dm_os_performance_counters").WillReturnRows(
		sqlmock.NewRows([]string{"counter_name", "cntr_value"}).
			AddRow("Active Temp Tables", 5.0).
			AddRow("Temp Tables Creation Rate", 42.0).
			RowError(1, rowErr))

	c := NewTempDBCollector(db)
	families, err := gatherScraperErr(t, c)

	if !errors.Is(err, rowErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v (rows.Err())", err, rowErr)
	}
	if v, ok := metricValue(families, "sqlserver_tempdb_active_temp_tables", nil); !ok || v != 5 {
		t.Errorf("active_temp_tables = %v (found=%v), want 5 despite the later iteration error", v, ok)
	}
}
