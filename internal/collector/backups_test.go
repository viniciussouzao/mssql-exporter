package collector

import (
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

// backupsQueryRegex pins down how the latest backup per database/type is
// picked: by backup_set_id (an identity), never by the local-time
// backup_finish_date, which goes backwards after a time zone change or a DST
// fall-back. It also pins the local-to-UTC conversion: backupset dates carry
// no offset, so emitting them as-is skews the timestamp by the UTC offset.
// The server_name filter keeps restored backups (log shipping secondary,
// mirror) from being reported as local ones (HML-012).
var backupsQueryRegex = sqlContainsInOrder(
	"DATEADD(MINUTE, -tz.utc_offset_minutes, bs.backup_finish_date)",
	"DATEADD(MINUTE, -tz.utc_offset_minutes, bs.backup_start_date)",
	"PARTITION BY database_name, type",
	"ORDER BY backup_set_id DESC",
	"FROM msdb.dbo.backupset",
	"WHERE backup_finish_date IS NOT NULL",
	"AND server_name = @@SERVERNAME",
	"WHEN bs.time_zone BETWEEN -48 AND 48 THEN bs.time_zone * 15",
	"ELSE DATEPART(TZOFFSET, SYSDATETIMEOFFSET())",
	"WHERE bs.rn = 1",
)

func TestBackupsScrapeHappyPath(t *testing.T) {
	db, mock := newSQLMock(t)

	start := time.Date(2026, 7, 15, 1, 0, 0, 0, time.UTC)
	finish := start.Add(90 * time.Second)

	rows := sqlmock.NewRows([]string{"database_name", "backup_type", "backup_finish_date", "backup_start_date", "size_bytes"}).
		AddRow("AppDB", "full", finish, start, 5_000_000.0)
	mock.ExpectQuery(backupsQueryRegex).WillReturnRows(rows)

	c := NewBackupsCollector(db)
	families := gatherScraper(t, c)

	labels := map[string]string{"database_name": "AppDB", "backup_type": "full"}
	if v, ok := metricValue(families, "sqlserver_backup_last_completed_timestamp_seconds", labels); !ok || v != float64(finish.Unix()) {
		t.Errorf("last_completed_timestamp_seconds = %v (found=%v), want %v", v, ok, finish.Unix())
	}
	if v, ok := metricValue(families, "sqlserver_backup_last_duration_seconds", labels); !ok || v != 90 {
		t.Errorf("last_duration_seconds = %v (found=%v), want 90", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_backup_last_size_bytes", labels); !ok || v != 5_000_000 {
		t.Errorf("last_size_bytes = %v (found=%v), want 5000000", v, ok)
	}
}

func TestBackupsScrapeQueryError(t *testing.T) {
	db, mock := newSQLMock(t)

	queryErr := errors.New("boom")
	mock.ExpectQuery(backupsQueryRegex).WillReturnError(queryErr)

	c := NewBackupsCollector(db)
	if err := scrapeErr(t, c); !errors.Is(err, queryErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v", err, queryErr)
	}
}

func TestBackupsScrapeScanError(t *testing.T) {
	db, mock := newSQLMock(t)

	start := time.Date(2026, 7, 15, 1, 0, 0, 0, time.UTC)
	finish := start.Add(60 * time.Second)

	rows := sqlmock.NewRows([]string{"database_name", "backup_type", "backup_finish_date", "backup_start_date", "size_bytes"}).
		AddRow("BadDB", "full", "not-a-time", start, 1.0).
		AddRow("AppDB", "log", finish, start, 1024.0)
	mock.ExpectQuery(backupsQueryRegex).WillReturnRows(rows)

	c := NewBackupsCollector(db)
	families, err := gatherScraperErr(t, c)

	if err == nil {
		t.Error("Scrape() error = nil, want error from bad row")
	}
	labels := map[string]string{"database_name": "AppDB", "backup_type": "log"}
	if v, ok := metricValue(families, "sqlserver_backup_last_size_bytes", labels); !ok || v != 1024 {
		t.Errorf("last_size_bytes = %v (found=%v), want 1024 despite earlier scan error", v, ok)
	}
}

func TestBackupsScrapeRowIterationError(t *testing.T) {
	db, mock := newSQLMock(t)

	start := time.Date(2026, 7, 15, 1, 0, 0, 0, time.UTC)
	finish := start.Add(60 * time.Second)
	rowErr := errors.New("row iteration failed")

	rows := sqlmock.NewRows([]string{"database_name", "backup_type", "backup_finish_date", "backup_start_date", "size_bytes"}).
		AddRow("AppDB", "full", finish, start, 1024.0).
		AddRow("OtherDB", "full", finish, start, 2048.0).
		RowError(1, rowErr)
	mock.ExpectQuery(backupsQueryRegex).WillReturnRows(rows)

	c := NewBackupsCollector(db)
	families, err := gatherScraperErr(t, c)

	if !errors.Is(err, rowErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v (rows.Err())", err, rowErr)
	}
	if v, ok := metricValue(families, "sqlserver_backup_last_size_bytes", map[string]string{"database_name": "AppDB", "backup_type": "full"}); !ok || v != 1024 {
		t.Errorf("last_size_bytes = %v (found=%v), want 1024 despite the later iteration error", v, ok)
	}
}

func TestBackupsScrapeEmpty(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery(backupsQueryRegex).WillReturnRows(
		sqlmock.NewRows([]string{"database_name", "backup_type", "backup_finish_date", "backup_start_date", "size_bytes"}))

	c := NewBackupsCollector(db)
	families := gatherScraper(t, c)
	if _, ok := metricValue(families, "sqlserver_backup_last_size_bytes", nil); ok {
		t.Error("last_size_bytes present, want no metrics for empty result set")
	}
}
