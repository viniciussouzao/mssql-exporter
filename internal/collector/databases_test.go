package collector

import (
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestDatabasesScrapeHappyPath(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery("FROM sys.databases").WillReturnRows(
		sqlmock.NewRows([]string{"name", "state_desc", "recovery_model_desc"}).
			AddRow("AppDB", "ONLINE", "FULL"))
	mock.ExpectQuery("FROM sys.master_files").WillReturnRows(
		sqlmock.NewRows([]string{"db_name", "type", "size_bytes"}).
			AddRow("AppDB", 0, 1_000_000.0).
			AddRow("AppDB", 1, 200_000.0))
	mock.ExpectQuery("FROM sys.dm_os_performance_counters").WillReturnRows(
		sqlmock.NewRows([]string{"instance_name", "cntr_value"}).
			AddRow("AppDB", 12.5))

	c := NewDatabasesCollector(db)
	families := gatherScraper(t, c)

	if v, ok := metricValue(families, "sqlserver_database_state", map[string]string{"db_name": "AppDB", "state": "ONLINE"}); !ok || v != 1 {
		t.Errorf("database_state = %v (found=%v), want 1", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_database_recovery_model", map[string]string{"db_name": "AppDB", "recovery_model": "FULL"}); !ok || v != 1 {
		t.Errorf("database_recovery_model = %v (found=%v), want 1", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_database_data_size_bytes", map[string]string{"db_name": "AppDB"}); !ok || v != 1_000_000 {
		t.Errorf("database_data_size_bytes = %v (found=%v), want 1000000", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_database_log_size_bytes", map[string]string{"db_name": "AppDB"}); !ok || v != 200_000 {
		t.Errorf("database_log_size_bytes = %v (found=%v), want 200000", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_database_log_used_percent", map[string]string{"db_name": "AppDB"}); !ok || v != 12.5 {
		t.Errorf("database_log_used_percent = %v (found=%v), want 12.5", v, ok)
	}
}

func TestDatabasesScrapeFileSizeSkipsNullDBName(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery("FROM sys.databases").WillReturnRows(
		sqlmock.NewRows([]string{"name", "state_desc", "recovery_model_desc"}))
	mock.ExpectQuery("FROM sys.master_files").WillReturnRows(
		sqlmock.NewRows([]string{"db_name", "type", "size_bytes"}).
			AddRow(nil, 0, 1_000_000.0))
	mock.ExpectQuery("FROM sys.dm_os_performance_counters").WillReturnRows(
		sqlmock.NewRows([]string{"instance_name", "cntr_value"}))

	c := NewDatabasesCollector(db)
	families := gatherScraper(t, c)

	if _, ok := metricValue(families, "sqlserver_database_data_size_bytes", nil); ok {
		t.Error("database_data_size_bytes present, want row with NULL db_name to be skipped")
	}
}

func TestDatabasesScrapeCombinesErrorsAcrossQueries(t *testing.T) {
	db, mock := newSQLMock(t)

	statesErr := errors.New("states failed")
	logUsageErr := errors.New("log usage failed")
	mock.ExpectQuery("FROM sys.databases").WillReturnError(statesErr)
	mock.ExpectQuery("FROM sys.master_files").WillReturnRows(
		sqlmock.NewRows([]string{"db_name", "type", "size_bytes"}).
			AddRow("AppDB", 0, 1_000_000.0))
	mock.ExpectQuery("FROM sys.dm_os_performance_counters").WillReturnError(logUsageErr)

	c := NewDatabasesCollector(db)
	families, err := gatherScraperErr(t, c)

	if !errors.Is(err, statesErr) {
		t.Errorf("Scrape() error = %v, want it to wrap the states error %v", err, statesErr)
	}
	if !errors.Is(err, logUsageErr) {
		t.Errorf("Scrape() error = %v, want it to wrap the log usage error %v", err, logUsageErr)
	}
	if v, ok := metricValue(families, "sqlserver_database_data_size_bytes", map[string]string{"db_name": "AppDB"}); !ok || v != 1_000_000 {
		t.Errorf("database_data_size_bytes = %v (found=%v), want 1000000 despite the other queries failing", v, ok)
	}
}

func TestDatabasesScrapeStatesRowIterationError(t *testing.T) {
	db, mock := newSQLMock(t)

	rowErr := errors.New("row iteration failed")
	mock.ExpectQuery("FROM sys.databases").WillReturnRows(
		sqlmock.NewRows([]string{"name", "state_desc", "recovery_model_desc"}).
			AddRow("AppDB", "ONLINE", "FULL").
			AddRow("OtherDB", "ONLINE", "FULL").
			RowError(1, rowErr))
	mock.ExpectQuery("FROM sys.master_files").WillReturnRows(
		sqlmock.NewRows([]string{"db_name", "type", "size_bytes"}))
	mock.ExpectQuery("FROM sys.dm_os_performance_counters").WillReturnRows(
		sqlmock.NewRows([]string{"instance_name", "cntr_value"}))

	c := NewDatabasesCollector(db)
	families, err := gatherScraperErr(t, c)

	if !errors.Is(err, rowErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v (rows.Err())", err, rowErr)
	}
	if v, ok := metricValue(families, "sqlserver_database_state", map[string]string{"db_name": "AppDB", "state": "ONLINE"}); !ok || v != 1 {
		t.Errorf("database_state = %v (found=%v), want 1 despite the later iteration error", v, ok)
	}
}
