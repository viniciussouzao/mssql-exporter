package collector

import (
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestMemoryClerksScrapeHappyPath(t *testing.T) {
	db, mock := newSQLMock(t)

	rows := sqlmock.NewRows([]string{"type", "name", "memory_node_id", "memory_used_kb"}).
		AddRow("MEMORYCLERK_SQLBUFFERPOOL", "Buffer Pool", 0, 102400)
	mock.ExpectQuery("SELECT type, name, memory_node_id").WillReturnRows(rows)

	c := NewMemoryClerksCollector(db)
	families := gatherScraper(t, c)

	labels := map[string]string{"clerk_type": "MEMORYCLERK_SQLBUFFERPOOL", "clerk_name": "Buffer Pool", "memory_node_id": "0"}
	if v, ok := metricValue(families, "sqlserver_memory_clerk_used_kb", labels); !ok || v != 102400 {
		t.Errorf("memory_clerk_used_kb = %v (found=%v), want 102400", v, ok)
	}
}

func TestMemoryClerksScrapeQueryError(t *testing.T) {
	db, mock := newSQLMock(t)

	queryErr := errors.New("boom")
	mock.ExpectQuery("SELECT type, name, memory_node_id").WillReturnError(queryErr)

	c := NewMemoryClerksCollector(db)
	if err := scrapeErr(t, c); !errors.Is(err, queryErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v", err, queryErr)
	}
}

func TestMemoryClerksScrapeScanError(t *testing.T) {
	db, mock := newSQLMock(t)

	rows := sqlmock.NewRows([]string{"type", "name", "memory_node_id", "memory_used_kb"}).
		AddRow("MEMORYCLERK_BAD", "Bad", "not-an-int", 1).
		AddRow("MEMORYCLERK_SQLBUFFERPOOL", "Buffer Pool", 0, 100)
	mock.ExpectQuery("SELECT type, name, memory_node_id").WillReturnRows(rows)

	c := NewMemoryClerksCollector(db)
	families, err := gatherScraperErr(t, c)

	if err == nil {
		t.Error("Scrape() error = nil, want error from bad row")
	}
	labels := map[string]string{"clerk_type": "MEMORYCLERK_SQLBUFFERPOOL", "clerk_name": "Buffer Pool", "memory_node_id": "0"}
	if v, ok := metricValue(families, "sqlserver_memory_clerk_used_kb", labels); !ok || v != 100 {
		t.Errorf("memory_clerk_used_kb = %v (found=%v), want 100 despite earlier scan error", v, ok)
	}
}

func TestMemoryClerksScrapeRowIterationError(t *testing.T) {
	db, mock := newSQLMock(t)

	rowErr := errors.New("row iteration failed")
	rows := sqlmock.NewRows([]string{"type", "name", "memory_node_id", "memory_used_kb"}).
		AddRow("MEMORYCLERK_SQLBUFFERPOOL", "Buffer Pool", 0, 100).
		AddRow("MEMORYCLERK_SQLGENERAL", "General", 0, 50).
		RowError(1, rowErr)
	mock.ExpectQuery("SELECT type, name, memory_node_id").WillReturnRows(rows)

	c := NewMemoryClerksCollector(db)
	families, err := gatherScraperErr(t, c)

	if !errors.Is(err, rowErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v (rows.Err())", err, rowErr)
	}
	labels := map[string]string{"clerk_type": "MEMORYCLERK_SQLBUFFERPOOL", "clerk_name": "Buffer Pool", "memory_node_id": "0"}
	if v, ok := metricValue(families, "sqlserver_memory_clerk_used_kb", labels); !ok || v != 100 {
		t.Errorf("memory_clerk_used_kb = %v (found=%v), want 100 despite the later iteration error", v, ok)
	}
}

func TestMemoryClerksScrapeEmpty(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery("SELECT type, name, memory_node_id").WillReturnRows(
		sqlmock.NewRows([]string{"type", "name", "memory_node_id", "memory_used_kb"}))

	c := NewMemoryClerksCollector(db)
	families := gatherScraper(t, c)
	if _, ok := metricValue(families, "sqlserver_memory_clerk_used_kb", nil); ok {
		t.Error("memory_clerk_used_kb present, want no metrics for empty result set")
	}
}
