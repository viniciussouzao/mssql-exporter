package collector

import (
	"errors"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

// statesQueryRegex pins the column order scanned by collectStates.
var statesQueryRegex = sqlContainsInOrder(
	"SELECT name, state_desc, recovery_model_desc, log_reuse_wait_desc",
	"FROM sys.databases",
)

// logUsageQueryRegex pins the join with sys.databases: the counter also has
// instances for hidden databases (mssqlsystemresource, model_msdb,
// model_replicatedmaster) and _Total, which must not become series.
var logUsageQueryRegex = sqlContainsInOrder(
	"SELECT d.name, pc.cntr_value",
	"FROM sys.dm_os_performance_counters pc",
	"INNER JOIN sys.databases d",
	"ON d.name = RTRIM(pc.instance_name) COLLATE DATABASE_DEFAULT",
	"AND pc.counter_name = 'Percent Log Used'",
)

// logUsageMIQueryRegex pins the Managed Instance match on the physical name,
// with the database name preferred.
var logUsageMIQueryRegex = sqlContainsInOrder(
	"SELECT d.name, pc.cntr_value",
	"FROM sys.databases d",
	"CROSS APPLY",
	"SELECT TOP (1) p.cntr_value",
	"AND p.counter_name = 'Percent Log Used'",
	"IN (d.name, d.physical_database_name)",
	"ORDER BY CASE WHEN RTRIM(p.instance_name) COLLATE DATABASE_DEFAULT = d.name THEN 0 ELSE 1 END",
)

func TestDatabasesScrapeHappyPath(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery(statesQueryRegex).WillReturnRows(
		sqlmock.NewRows([]string{"name", "state_desc", "recovery_model_desc", "log_reuse_wait_desc"}).
			AddRow("AppDB", "ONLINE", "FULL", "LOG_BACKUP"))
	mock.ExpectQuery("FROM sys.master_files").WillReturnRows(
		sqlmock.NewRows([]string{"db_name", "type", "size_bytes"}).
			AddRow("AppDB", 0, 1_000_000.0).
			AddRow("AppDB", 1, 200_000.0))
	expectEngineEdition(mock, 3)
	mock.ExpectQuery(logUsageQueryRegex).WillReturnRows(
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
	if v, ok := metricValue(families, "sqlserver_database_log_reuse_wait", map[string]string{"db_name": "AppDB", "reason": "LOG_BACKUP"}); !ok || v != 1 {
		t.Errorf("database_log_reuse_wait = %v (found=%v), want 1", v, ok)
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
		sqlmock.NewRows([]string{"name", "state_desc", "recovery_model_desc", "log_reuse_wait_desc"}))
	mock.ExpectQuery("FROM sys.master_files").WillReturnRows(
		sqlmock.NewRows([]string{"db_name", "type", "size_bytes"}).
			AddRow(nil, 0, 1_000_000.0))
	expectEngineEdition(mock, 3)
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
	expectEngineEdition(mock, 3)
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
		sqlmock.NewRows([]string{"name", "state_desc", "recovery_model_desc", "log_reuse_wait_desc"}).
			AddRow("AppDB", "ONLINE", "FULL", "LOG_BACKUP").
			AddRow("OtherDB", "ONLINE", "FULL", "NOTHING").
			RowError(1, rowErr))
	mock.ExpectQuery("FROM sys.master_files").WillReturnRows(
		sqlmock.NewRows([]string{"db_name", "type", "size_bytes"}))
	expectEngineEdition(mock, 3)
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

// physical_database_name only exists on SQL Server 2019+: the query sent to
// every server other than Managed Instance must not reference it.
func TestDatabasesLogUsageQueryAvoidsPhysicalName(t *testing.T) {
	if strings.Contains(logUsageQuery, "physical_database_name") {
		t.Error("logUsageQuery references physical_database_name, which fails before SQL Server 2019")
	}
}

func TestDatabasesScrapeLogUsageManagedInstance(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery("FROM sys.databases").WillReturnRows(
		sqlmock.NewRows([]string{"name", "state_desc", "recovery_model_desc", "log_reuse_wait_desc"}))
	mock.ExpectQuery("FROM sys.master_files").WillReturnRows(
		sqlmock.NewRows([]string{"db_name", "type", "size_bytes"}))
	expectEngineEdition(mock, engineEditionManagedInstance)
	mock.ExpectQuery(logUsageMIQueryRegex).WillReturnRows(
		sqlmock.NewRows([]string{"name", "cntr_value"}).
			AddRow("master", 44.0).
			AddRow("LoadTestDB", 7.0))

	c := NewDatabasesCollector(db)
	families := gatherScraper(t, c)

	if v, ok := metricValue(families, "sqlserver_database_log_used_percent", map[string]string{"db_name": "LoadTestDB"}); !ok || v != 7 {
		t.Errorf("database_log_used_percent{db_name=LoadTestDB} = %v (found=%v), want 7", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_database_log_used_percent", map[string]string{"db_name": "master"}); !ok || v != 44 {
		t.Errorf("database_log_used_percent{db_name=master} = %v (found=%v), want 44", v, ok)
	}
}

func TestDatabasesScrapeLogUsageEngineEditionError(t *testing.T) {
	db, mock := newSQLMock(t)

	editionErr := errors.New("edition failed")
	mock.ExpectQuery("FROM sys.databases").WillReturnRows(
		sqlmock.NewRows([]string{"name", "state_desc", "recovery_model_desc", "log_reuse_wait_desc"}).
			AddRow("AppDB", "ONLINE", "FULL", "NOTHING"))
	mock.ExpectQuery("FROM sys.master_files").WillReturnRows(
		sqlmock.NewRows([]string{"db_name", "type", "size_bytes"}))
	mock.ExpectQuery("SERVERPROPERTY\\('EngineEdition'\\)").WillReturnError(editionErr)

	c := NewDatabasesCollector(db)
	families, err := gatherScraperErr(t, c)

	if !errors.Is(err, editionErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v", err, editionErr)
	}
	if v, ok := metricValue(families, "sqlserver_database_state", map[string]string{"db_name": "AppDB", "state": "ONLINE"}); !ok || v != 1 {
		t.Errorf("database_state = %v (found=%v), want 1 despite the engine edition error", v, ok)
	}
	if _, ok := metricValue(families, "sqlserver_database_log_used_percent", nil); ok {
		t.Error("database_log_used_percent present, want none when the engine edition is unknown")
	}
}
