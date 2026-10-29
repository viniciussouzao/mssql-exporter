package collector

import (
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestIOFileStatsScrapeHappyPath(t *testing.T) {
	db, mock := newSQLMock(t)

	rows := sqlmock.NewRows([]string{
		"db_name", "file_name", "file_type", "physical_name",
		"num_of_reads", "num_of_writes", "num_of_bytes_read", "num_of_bytes_written",
		"io_stall_read_ms", "io_stall_write_ms",
	}).AddRow("AppDB", "AppDB_Data", "ROWS", "C:\\data\\AppDB.mdf", 100.0, 50.0, 8192.0, 4096.0, 10.0, 5.0)
	mock.ExpectQuery("FROM sys.dm_io_virtual_file_stats").WillReturnRows(rows)

	c := NewIOFileStatsCollector(db)
	families := gatherScraper(t, c)

	labels := map[string]string{
		"db_name": "AppDB", "file_name": "AppDB_Data", "file_type": "ROWS", "physical_name": "C:\\data\\AppDB.mdf",
	}
	cases := map[string]float64{
		"sqlserver_io_file_reads_total":          100,
		"sqlserver_io_file_writes_total":         50,
		"sqlserver_io_file_read_bytes_total":     8192,
		"sqlserver_io_file_write_bytes_total":    4096,
		"sqlserver_io_file_stall_read_ms_total":  10,
		"sqlserver_io_file_stall_write_ms_total": 5,
	}
	for name, want := range cases {
		if v, ok := counterValue(families, name, labels); !ok || v != want {
			t.Errorf("%s = %v (found=%v), want %v", name, v, ok, want)
		}
	}
}

func TestIOFileStatsScrapeQueryError(t *testing.T) {
	db, mock := newSQLMock(t)

	queryErr := errors.New("boom")
	mock.ExpectQuery("FROM sys.dm_io_virtual_file_stats").WillReturnError(queryErr)

	c := NewIOFileStatsCollector(db)
	if err := scrapeErr(t, c); !errors.Is(err, queryErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v", err, queryErr)
	}
}

func TestIOFileStatsScrapeScanError(t *testing.T) {
	db, mock := newSQLMock(t)

	rows := sqlmock.NewRows([]string{
		"db_name", "file_name", "file_type", "physical_name",
		"num_of_reads", "num_of_writes", "num_of_bytes_read", "num_of_bytes_written",
		"io_stall_read_ms", "io_stall_write_ms",
	}).
		AddRow("BadDB", "Bad", "ROWS", "C:\\bad.mdf", "not-a-number", 1.0, 1.0, 1.0, 1.0, 1.0).
		AddRow("AppDB", "AppDB_Data", "ROWS", "C:\\data\\AppDB.mdf", 100.0, 50.0, 8192.0, 4096.0, 10.0, 5.0)
	mock.ExpectQuery("FROM sys.dm_io_virtual_file_stats").WillReturnRows(rows)

	c := NewIOFileStatsCollector(db)
	families, err := gatherScraperErr(t, c)

	if err == nil {
		t.Error("Scrape() error = nil, want error from bad row")
	}
	labels := map[string]string{
		"db_name": "AppDB", "file_name": "AppDB_Data", "file_type": "ROWS", "physical_name": "C:\\data\\AppDB.mdf",
	}
	if v, ok := counterValue(families, "sqlserver_io_file_reads_total", labels); !ok || v != 100 {
		t.Errorf("io_file_reads_total = %v (found=%v), want 100 despite earlier scan error", v, ok)
	}
}

func TestIOFileStatsScrapeRowIterationError(t *testing.T) {
	db, mock := newSQLMock(t)

	rowErr := errors.New("row iteration failed")
	rows := sqlmock.NewRows([]string{
		"db_name", "file_name", "file_type", "physical_name",
		"num_of_reads", "num_of_writes", "num_of_bytes_read", "num_of_bytes_written",
		"io_stall_read_ms", "io_stall_write_ms",
	}).
		AddRow("AppDB", "AppDB_Data", "ROWS", "C:\\data\\AppDB.mdf", 100.0, 50.0, 8192.0, 4096.0, 10.0, 5.0).
		AddRow("AppDB", "AppDB_Log", "LOG", "C:\\data\\AppDB.ldf", 20.0, 10.0, 1024.0, 512.0, 1.0, 1.0).
		RowError(1, rowErr)
	mock.ExpectQuery("FROM sys.dm_io_virtual_file_stats").WillReturnRows(rows)

	c := NewIOFileStatsCollector(db)
	families, err := gatherScraperErr(t, c)

	if !errors.Is(err, rowErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v (rows.Err())", err, rowErr)
	}
	labels := map[string]string{
		"db_name": "AppDB", "file_name": "AppDB_Data", "file_type": "ROWS", "physical_name": "C:\\data\\AppDB.mdf",
	}
	if v, ok := counterValue(families, "sqlserver_io_file_reads_total", labels); !ok || v != 100 {
		t.Errorf("io_file_reads_total = %v (found=%v), want 100 despite the later iteration error", v, ok)
	}
}

func TestIOFileStatsScrapeEmpty(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery("FROM sys.dm_io_virtual_file_stats").WillReturnRows(sqlmock.NewRows([]string{
		"db_name", "file_name", "file_type", "physical_name",
		"num_of_reads", "num_of_writes", "num_of_bytes_read", "num_of_bytes_written",
		"io_stall_read_ms", "io_stall_write_ms",
	}))

	c := NewIOFileStatsCollector(db)
	families := gatherScraper(t, c)
	if _, ok := counterValue(families, "sqlserver_io_file_reads_total", nil); ok {
		t.Error("io_file_reads_total present, want no metrics for empty result set (e.g. missing VIEW ANY DEFINITION)")
	}
}

func TestIOFileStatsScrapeSkipsNullDatabaseName(t *testing.T) {
	db, mock := newSQLMock(t)

	// DB_NAME is NULL for a database being dropped or restored: the row is
	// skipped without failing the collector.
	rows := sqlmock.NewRows([]string{
		"db_name", "file_name", "file_type", "physical_name",
		"num_of_reads", "num_of_writes", "num_of_bytes_read", "num_of_bytes_written",
		"io_stall_read_ms", "io_stall_write_ms",
	}).
		AddRow(nil, "Gone_Data", "ROWS", "C:\\data\\Gone.mdf", 1.0, 1.0, 1.0, 1.0, 1.0, 1.0).
		AddRow("AppDB", "AppDB_Data", "ROWS", "C:\\data\\AppDB.mdf", 100.0, 50.0, 8192.0, 4096.0, 10.0, 5.0)
	mock.ExpectQuery("FROM sys.dm_io_virtual_file_stats").WillReturnRows(rows)

	c := NewIOFileStatsCollector(db)
	families := gatherScraper(t, c)

	if v, ok := counterValue(families, "sqlserver_io_file_reads_total", map[string]string{"db_name": "AppDB"}); !ok || v != 100 {
		t.Errorf("io_file_reads_total{AppDB} = %v (found=%v), want 100", v, ok)
	}
	if _, ok := counterValue(families, "sqlserver_io_file_reads_total", map[string]string{"file_name": "Gone_Data"}); ok {
		t.Error("io_file_reads_total{Gone_Data} present, want the row with a NULL database name skipped")
	}
}
