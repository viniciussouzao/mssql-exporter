package collector

import (
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestServerInfoScrapeHappyPath(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery("SERVERPROPERTY").WillReturnRows(
		sqlmock.NewRows([]string{"hostname", "servername", "edition", "product_version", "engine_edition", "is_hadr_enabled"}).
			AddRow("host1", "host1\\SQLEXPRESS", "Developer Edition", "16.0.1000.6", 3, 1))
	mock.ExpectQuery("FROM sys.dm_os_sys_info").WillReturnRows(
		sqlmock.NewRows([]string{"cpu_count"}).AddRow(8))
	mock.ExpectQuery("FROM sys.dm_os_sys_memory").WillReturnRows(
		sqlmock.NewRows([]string{"total_physical_memory_mb", "available_physical_memory_mb"}).
			AddRow(16_000_000, 4_000_000))

	c := NewServerInfoCollector(db)
	families := gatherScraper(t, c)

	if v, ok := metricValue(families, "sqlserver_hostname", map[string]string{"hostname": "host1"}); !ok || v != 1 {
		t.Errorf("hostname = %v (found=%v), want 1", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_engine_edition", nil); !ok || v != 3 {
		t.Errorf("engine_edition = %v (found=%v), want 3", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_visible_cpus", nil); !ok || v != 8 {
		t.Errorf("visible_cpus = %v (found=%v), want 8", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_total_physical_memory_kb", nil); !ok || v != 16_000_000 {
		t.Errorf("total_physical_memory_kb = %v (found=%v), want 16000000", v, ok)
	}
}

// Azure SQL Managed Instance returns NULL for SERVERPROPERTY('MachineName');
// the other properties must still be exposed.
func TestServerInfoScrapeNullPropertyOmitsOnlyThatMetric(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery("SERVERPROPERTY").WillReturnRows(
		sqlmock.NewRows([]string{"hostname", "servername", "edition", "product_version", "engine_edition", "is_hadr_enabled"}).
			AddRow(nil, "mi1.abc.database.windows.net", "SQL Azure", "12.0.2000.8", 8, 0))
	mock.ExpectQuery("FROM sys.dm_os_sys_info").WillReturnRows(
		sqlmock.NewRows([]string{"cpu_count"}).AddRow(4))
	mock.ExpectQuery("FROM sys.dm_os_sys_memory").WillReturnRows(
		sqlmock.NewRows([]string{"total_physical_memory_mb", "available_physical_memory_mb"}).
			AddRow(23_000_000, 9_000_000))

	c := NewServerInfoCollector(db)
	families := gatherScraper(t, c)

	if _, ok := metricValue(families, "sqlserver_hostname", nil); ok {
		t.Error("hostname present, want omitted since MachineName is NULL")
	}
	if v, ok := metricValue(families, "sqlserver_servername", map[string]string{"servername": "mi1.abc.database.windows.net"}); !ok || v != 1 {
		t.Errorf("servername = %v (found=%v), want 1", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_edition", map[string]string{"edition": "SQL Azure"}); !ok || v != 1 {
		t.Errorf("edition = %v (found=%v), want 1", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_engine_edition", nil); !ok || v != 8 {
		t.Errorf("engine_edition = %v (found=%v), want 8", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_is_hadr_enabled", nil); !ok || v != 0 {
		t.Errorf("is_hadr_enabled = %v (found=%v), want 0", v, ok)
	}
}

func TestServerInfoScrapePartialFailureStillEmitsOtherMetrics(t *testing.T) {
	db, mock := newSQLMock(t)

	propertiesErr := errors.New("properties failed")
	memoryErr := errors.New("memory failed")
	mock.ExpectQuery("SERVERPROPERTY").WillReturnError(propertiesErr)
	mock.ExpectQuery("FROM sys.dm_os_sys_info").WillReturnRows(
		sqlmock.NewRows([]string{"cpu_count"}).AddRow(4))
	mock.ExpectQuery("FROM sys.dm_os_sys_memory").WillReturnError(memoryErr)

	c := NewServerInfoCollector(db)
	families, err := gatherScraperErr(t, c)

	if !errors.Is(err, propertiesErr) {
		t.Errorf("Scrape() error = %v, want it to wrap the server-properties error %v", err, propertiesErr)
	}
	if !errors.Is(err, memoryErr) {
		t.Errorf("Scrape() error = %v, want it to wrap the memory error %v", err, memoryErr)
	}
	if v, ok := metricValue(families, "sqlserver_visible_cpus", nil); !ok || v != 4 {
		t.Errorf("visible_cpus = %v (found=%v), want 4 despite the other queries failing", v, ok)
	}
	if _, ok := metricValue(families, "sqlserver_hostname", nil); ok {
		t.Error("hostname present, want omitted since the server properties query failed")
	}
	if _, ok := metricValue(families, "sqlserver_total_physical_memory_kb", nil); ok {
		t.Error("total_physical_memory_kb present, want omitted since the memory query failed")
	}
}
