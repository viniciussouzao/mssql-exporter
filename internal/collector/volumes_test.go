package collector

import (
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestVolumesScrapeHappyPath(t *testing.T) {
	db, mock := newSQLMock(t)
	expectEngineEdition(mock, 3)

	rows := sqlmock.NewRows([]string{"mount_point", "total_bytes", "available_bytes"}).
		AddRow("C:\\", 1000.0, 400.0)
	mock.ExpectQuery("GROUP BY mount_point").WillReturnRows(rows)

	c := NewVolumesCollector(db)
	families := gatherScraper(t, c)

	labels := map[string]string{"volume_mount_point": "C:\\"}
	if v, ok := metricValue(families, "sqlserver_volume_total_bytes", labels); !ok || v != 1000 {
		t.Errorf("volume_total_bytes = %v (found=%v), want 1000", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_volume_available_bytes", labels); !ok || v != 400 {
		t.Errorf("volume_available_bytes = %v (found=%v), want 400", v, ok)
	}
}

// SQL Server on Linux returns no mount point nor logical volume name: the
// label falls back to the directory of the database file. Databases changing
// state are skipped (HML-019).
func TestVolumesScrapeLinuxFallbackAndStateFilter(t *testing.T) {
	db, mock := newSQLMock(t)
	expectEngineEdition(mock, 3)

	rows := sqlmock.NewRows([]string{"mount_point", "total_bytes", "available_bytes"}).
		AddRow("/var/opt/mssql/data/", 1000.0, 400.0)
	mock.ExpectQuery(`(?s)NULLIF\(vs\.logical_volume_name, ''\),\s*LEFT\(mf\.physical_name, .*CHARINDEX\('/', REVERSE\(mf\.physical_name\)\).*WHERE d\.state IN \(0, 1\).*GROUP BY mount_point`).
		WillReturnRows(rows)

	c := NewVolumesCollector(db)
	families := gatherScraper(t, c)

	labels := map[string]string{"volume_mount_point": "/var/opt/mssql/data/"}
	if v, ok := metricValue(families, "sqlserver_volume_total_bytes", labels); !ok || v != 1000 {
		t.Errorf("volume_total_bytes = %v (found=%v), want 1000", v, ok)
	}
}

func TestVolumesScrapeQueryError(t *testing.T) {
	db, mock := newSQLMock(t)
	expectEngineEdition(mock, 3)

	queryErr := errors.New("boom")
	mock.ExpectQuery("GROUP BY mount_point").WillReturnError(queryErr)

	c := NewVolumesCollector(db)
	if err := scrapeErr(t, c); !errors.Is(err, queryErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v", err, queryErr)
	}
}

func TestVolumesScrapeScanError(t *testing.T) {
	db, mock := newSQLMock(t)
	expectEngineEdition(mock, 3)

	rows := sqlmock.NewRows([]string{"mount_point", "total_bytes", "available_bytes"}).
		AddRow("C:\\", "not-a-number", 400.0).
		AddRow("D:\\", 2000.0, 900.0)
	mock.ExpectQuery("GROUP BY mount_point").WillReturnRows(rows)

	c := NewVolumesCollector(db)
	families, err := gatherScraperErr(t, c)

	if err == nil {
		t.Error("Scrape() error = nil, want error from bad row")
	}
	if v, ok := metricValue(families, "sqlserver_volume_total_bytes", map[string]string{"volume_mount_point": "D:\\"}); !ok || v != 2000 {
		t.Errorf("volume_total_bytes{D:\\\\} = %v (found=%v), want 2000 despite earlier scan error", v, ok)
	}
}

func TestVolumesScrapeRowIterationError(t *testing.T) {
	db, mock := newSQLMock(t)
	expectEngineEdition(mock, 3)

	rowErr := errors.New("row iteration failed")
	rows := sqlmock.NewRows([]string{"mount_point", "total_bytes", "available_bytes"}).
		AddRow("C:\\", 1000.0, 400.0).
		AddRow("D:\\", 2000.0, 900.0).
		RowError(1, rowErr)
	mock.ExpectQuery("GROUP BY mount_point").WillReturnRows(rows)

	c := NewVolumesCollector(db)
	families, err := gatherScraperErr(t, c)

	if !errors.Is(err, rowErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v (rows.Err())", err, rowErr)
	}
	if v, ok := metricValue(families, "sqlserver_volume_total_bytes", map[string]string{"volume_mount_point": "C:\\"}); !ok || v != 1000 {
		t.Errorf("volume_total_bytes{C:\\\\} = %v (found=%v), want 1000 despite the later iteration error", v, ok)
	}
}

func TestVolumesScrapeEmpty(t *testing.T) {
	db, mock := newSQLMock(t)
	expectEngineEdition(mock, 3)

	mock.ExpectQuery("GROUP BY mount_point").WillReturnRows(
		sqlmock.NewRows([]string{"mount_point", "total_bytes", "available_bytes"}))

	c := NewVolumesCollector(db)
	families := gatherScraper(t, c)
	if _, ok := metricValue(families, "sqlserver_volume_total_bytes", nil); ok {
		t.Error("volume_total_bytes present, want no metrics for empty result set (e.g. missing VIEW ANY DEFINITION)")
	}
}

func expectEngineEdition(mock sqlmock.Sqlmock, edition int) {
	mock.ExpectQuery("SERVERPROPERTY\\('EngineEdition'\\)").WillReturnRows(
		sqlmock.NewRows([]string{"engine_edition"}).AddRow(edition))
}

func TestVolumesScrapeEngineEditionError(t *testing.T) {
	db, mock := newSQLMock(t)

	editionErr := errors.New("edition failed")
	mock.ExpectQuery("SERVERPROPERTY").WillReturnError(editionErr)

	c := NewVolumesCollector(db)
	if err := scrapeErr(t, c); !errors.Is(err, editionErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v", err, editionErr)
	}
}

// On Azure SQL Managed Instance dm_os_volume_stats is not queried: the
// instance storage comes from server_resource_stats.
func TestVolumesScrapeManagedInstanceStorage(t *testing.T) {
	db, mock := newSQLMock(t)
	expectEngineEdition(mock, 8)
	mock.ExpectQuery("FROM master.sys.server_resource_stats").WillReturnRows(
		sqlmock.NewRows([]string{"reserved_storage_mb", "storage_space_used_mb"}).AddRow(65536, 120.0))

	c := NewVolumesCollector(db)
	families := gatherScraper(t, c)

	if v, ok := metricValue(families, "sqlserver_instance_storage_reserved_bytes", nil); !ok || v != 65536*1024*1024 {
		t.Errorf("instance_storage_reserved_bytes = %v (found=%v), want %v", v, ok, 65536*1024*1024)
	}
	if v, ok := metricValue(families, "sqlserver_instance_storage_used_bytes", nil); !ok || v != 120*1024*1024 {
		t.Errorf("instance_storage_used_bytes = %v (found=%v), want %v", v, ok, 120*1024*1024)
	}
	if _, ok := metricValue(families, "sqlserver_volume_total_bytes", nil); ok {
		t.Error("volume_total_bytes present, want no volume series on Managed Instance")
	}
}

func TestVolumesScrapeManagedInstanceStorageQueryError(t *testing.T) {
	db, mock := newSQLMock(t)
	expectEngineEdition(mock, 8)
	queryErr := errors.New("boom")
	mock.ExpectQuery("FROM master.sys.server_resource_stats").WillReturnError(queryErr)

	c := NewVolumesCollector(db)
	if err := scrapeErr(t, c); !errors.Is(err, queryErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v", err, queryErr)
	}
}

// server_resource_stats is empty for the first minutes after the instance is
// created.
func TestVolumesScrapeManagedInstanceStorageEmpty(t *testing.T) {
	db, mock := newSQLMock(t)
	expectEngineEdition(mock, 8)
	mock.ExpectQuery("FROM master.sys.server_resource_stats").WillReturnRows(
		sqlmock.NewRows([]string{"reserved_storage_mb", "storage_space_used_mb"}))

	c := NewVolumesCollector(db)
	families := gatherScraper(t, c)
	if _, ok := metricValue(families, "sqlserver_instance_storage_used_bytes", nil); ok {
		t.Error("instance_storage_used_bytes present, want none for an empty server_resource_stats")
	}
}
