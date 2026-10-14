package collector

import (
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

// The *_utc columns must be the ones read: the plain date columns are local
// server time (see HML-010 in BACKLOG-HOMOLOG.md).
var logShippingPrimaryQueryRegex = sqlContainsInOrder(
	"last_backup_date_utc",
	"FROM msdb.dbo.log_shipping_monitor_primary",
)

var logShippingSecondaryQueryRegex = sqlContainsInOrder(
	"last_copied_date_utc",
	"last_restored_date_utc",
	"FROM msdb.dbo.log_shipping_monitor_secondary",
)

var logShippingPrimaryColumns = []string{"primary_server", "primary_database", "last_backup_date_utc", "backup_threshold"}
var logShippingSecondaryColumns = []string{"secondary_server", "secondary_database", "primary_server", "primary_database",
	"last_copied_date_utc", "last_restored_date_utc", "last_restored_latency", "restore_threshold"}

func TestLogShippingScrapeHappyPath(t *testing.T) {
	db, mock := newSQLMock(t)

	lastBackup := time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)
	lastCopied := time.Date(2026, 10, 13, 12, 1, 0, 0, time.UTC)
	lastRestored := time.Date(2026, 10, 13, 12, 2, 0, 0, time.UTC)
	mock.ExpectQuery(logShippingPrimaryQueryRegex).WillReturnRows(
		sqlmock.NewRows(logShippingPrimaryColumns).
			AddRow("SQL1", "LSDB", lastBackup, 60))
	mock.ExpectQuery(logShippingSecondaryQueryRegex).WillReturnRows(
		sqlmock.NewRows(logShippingSecondaryColumns).
			AddRow("SQL2", "LSDB", "SQL1", "LSDB", lastCopied, lastRestored, 2, 45))

	c := NewLogShippingCollector(db)
	families := gatherScraper(t, c)

	primary := map[string]string{"primary_server": "SQL1", "primary_database": "LSDB"}
	if v, ok := metricValue(families, "sqlserver_log_shipping_primary_last_backup_timestamp_seconds", primary); !ok || v != float64(lastBackup.Unix()) {
		t.Errorf("primary_last_backup_timestamp_seconds = %v (found=%v), want %v", v, ok, lastBackup.Unix())
	}
	if v, ok := metricValue(families, "sqlserver_log_shipping_primary_backup_threshold_seconds", primary); !ok || v != 3600 {
		t.Errorf("primary_backup_threshold_seconds = %v (found=%v), want 3600", v, ok)
	}

	secondary := map[string]string{"secondary_server": "SQL2", "secondary_database": "LSDB", "primary_server": "SQL1", "primary_database": "LSDB"}
	for name, want := range map[string]float64{
		"sqlserver_log_shipping_secondary_last_copied_timestamp_seconds":   float64(lastCopied.Unix()),
		"sqlserver_log_shipping_secondary_last_restored_timestamp_seconds": float64(lastRestored.Unix()),
		"sqlserver_log_shipping_secondary_last_restored_latency_seconds":   120,
		"sqlserver_log_shipping_secondary_restore_threshold_seconds":       2700,
	} {
		if v, ok := metricValue(families, name, secondary); !ok || v != want {
			t.Errorf("%s = %v (found=%v), want %v", name, v, ok, want)
		}
	}
}

func TestLogShippingScrapeNotConfiguredEmitsNothingNoError(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery(logShippingPrimaryQueryRegex).WillReturnRows(sqlmock.NewRows(logShippingPrimaryColumns))
	mock.ExpectQuery(logShippingSecondaryQueryRegex).WillReturnRows(sqlmock.NewRows(logShippingSecondaryColumns))

	c := NewLogShippingCollector(db)
	families := gatherScraper(t, c)

	if len(families) != 0 {
		t.Errorf("got %d metric families, want none when log shipping is not configured", len(families))
	}
}

func TestLogShippingScrapeNeverRunOmitsTimestamps(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery(logShippingPrimaryQueryRegex).WillReturnRows(
		sqlmock.NewRows(logShippingPrimaryColumns).
			AddRow("SQL1", "LSDB", nil, 60))
	mock.ExpectQuery(logShippingSecondaryQueryRegex).WillReturnRows(
		sqlmock.NewRows(logShippingSecondaryColumns).
			AddRow("SQL2", "LSDB", "SQL1", "LSDB", nil, nil, nil, 45))

	c := NewLogShippingCollector(db)
	families := gatherScraper(t, c)

	if _, ok := metricValue(families, "sqlserver_log_shipping_primary_last_backup_timestamp_seconds", nil); ok {
		t.Error("primary_last_backup_timestamp_seconds present, want it omitted before the first backup")
	}
	if _, ok := metricValue(families, "sqlserver_log_shipping_secondary_last_restored_timestamp_seconds", nil); ok {
		t.Error("secondary_last_restored_timestamp_seconds present, want it omitted before the first restore")
	}
	if v, ok := metricValue(families, "sqlserver_log_shipping_secondary_restore_threshold_seconds", nil); !ok || v != 2700 {
		t.Errorf("secondary_restore_threshold_seconds = %v (found=%v), want 2700", v, ok)
	}
}

func TestLogShippingScrapeCombinesQueryErrors(t *testing.T) {
	db, mock := newSQLMock(t)

	primaryErr := errors.New("SELECT permission was denied on log_shipping_monitor_primary")
	secondaryErr := errors.New("SELECT permission was denied on log_shipping_monitor_secondary")
	mock.ExpectQuery(logShippingPrimaryQueryRegex).WillReturnError(primaryErr)
	mock.ExpectQuery(logShippingSecondaryQueryRegex).WillReturnError(secondaryErr)

	c := NewLogShippingCollector(db)
	err := scrapeErr(t, c)

	if !errors.Is(err, primaryErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v", err, primaryErr)
	}
	if !errors.Is(err, secondaryErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v", err, secondaryErr)
	}
}
