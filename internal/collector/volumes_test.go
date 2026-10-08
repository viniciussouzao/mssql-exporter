package collector

import (
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestVolumesScrapeHappyPath(t *testing.T) {
	db, mock := newSQLMock(t)

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

func TestVolumesScrapeQueryError(t *testing.T) {
	db, mock := newSQLMock(t)

	queryErr := errors.New("boom")
	mock.ExpectQuery("GROUP BY mount_point").WillReturnError(queryErr)

	c := NewVolumesCollector(db)
	if err := scrapeErr(t, c); !errors.Is(err, queryErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v", err, queryErr)
	}
}

func TestVolumesScrapeScanError(t *testing.T) {
	db, mock := newSQLMock(t)

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

	mock.ExpectQuery("GROUP BY mount_point").WillReturnRows(
		sqlmock.NewRows([]string{"mount_point", "total_bytes", "available_bytes"}))

	c := NewVolumesCollector(db)
	families := gatherScraper(t, c)
	if _, ok := metricValue(families, "sqlserver_volume_total_bytes", nil); ok {
		t.Error("volume_total_bytes present, want no metrics for empty result set (e.g. missing VIEW ANY DEFINITION)")
	}
}
