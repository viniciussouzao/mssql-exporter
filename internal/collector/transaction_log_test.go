package collector

import (
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

// transactionLogQueryRegex pins the filters: ONLINE databases the login can
// access, and no local AG secondary replicas.
var transactionLogQueryRegex = sqlContainsInOrder(
	"CROSS APPLY sys.dm_db_log_stats(d.database_id)",
	"d.state = 0",
	"HAS_DBACCESS(d.name) = 1",
	"sys.dm_hadr_database_replica_states",
	"rs.is_primary_replica = 0",
)

var transactionLogColumns = []string{
	"name", "total_vlf_count", "active_vlf_count", "active_log_size_mb",
	"log_since_last_log_backup_mb", "log_since_last_checkpoint_mb",
}

func TestTransactionLogScrapeHappyPath(t *testing.T) {
	db, mock := newSQLMock(t)

	rows := sqlmock.NewRows(transactionLogColumns).
		AddRow("AppDB", 12.0, 8.0, 87.5, 87.25, 3.0).
		AddRow("master", 8.0, 1.0, 0.25, nil, 0.25)
	mock.ExpectQuery(transactionLogQueryRegex).WillReturnRows(rows)

	c := NewTransactionLogCollector(db)
	families := gatherScraper(t, c)

	app := map[string]string{"db_name": "AppDB"}
	cases := map[string]float64{
		"sqlserver_database_log_vlf_count":                   12,
		"sqlserver_database_log_active_vlf_count":            8,
		"sqlserver_database_log_active_bytes":                87.5 * 1024 * 1024,
		"sqlserver_database_log_since_last_log_backup_bytes": 87.25 * 1024 * 1024,
		"sqlserver_database_log_since_last_checkpoint_bytes": 3 * 1024 * 1024,
	}
	for name, want := range cases {
		if v, ok := metricValue(families, name, app); !ok || v != want {
			t.Errorf("%s{AppDB} = %v (found=%v), want %v", name, v, ok, want)
		}
	}

	// No log backup chain (SIMPLE recovery): the series is absent, not 0.
	master := map[string]string{"db_name": "master"}
	if _, ok := metricValue(families, "sqlserver_database_log_since_last_log_backup_bytes", master); ok {
		t.Error("log_since_last_log_backup_bytes{master} present, want it absent for a NULL value")
	}
	if v, ok := metricValue(families, "sqlserver_database_log_vlf_count", master); !ok || v != 8 {
		t.Errorf("log_vlf_count{master} = %v (found=%v), want 8", v, ok)
	}
}

func TestTransactionLogScrapeQueryError(t *testing.T) {
	db, mock := newSQLMock(t)

	queryErr := errors.New("Invalid object name 'sys.dm_db_log_stats'")
	mock.ExpectQuery(transactionLogQueryRegex).WillReturnError(queryErr)

	c := NewTransactionLogCollector(db)
	if err := scrapeErr(t, c); !errors.Is(err, queryErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v", err, queryErr)
	}
}

func TestTransactionLogScrapeScanError(t *testing.T) {
	db, mock := newSQLMock(t)

	rows := sqlmock.NewRows(transactionLogColumns).
		AddRow("BadDB", "not-a-number", 1.0, 1.0, nil, 1.0).
		AddRow("AppDB", 12.0, 8.0, 87.5, 87.25, 3.0)
	mock.ExpectQuery(transactionLogQueryRegex).WillReturnRows(rows)

	c := NewTransactionLogCollector(db)
	families, err := gatherScraperErr(t, c)
	if err == nil {
		t.Error("Scrape() error = nil, want error from bad row")
	}
	if v, ok := metricValue(families, "sqlserver_database_log_vlf_count", map[string]string{"db_name": "AppDB"}); !ok || v != 12 {
		t.Errorf("log_vlf_count{AppDB} = %v (found=%v), want 12 despite the earlier scan error", v, ok)
	}
}

func TestTransactionLogScrapeEmpty(t *testing.T) {
	db, mock := newSQLMock(t)
	mock.ExpectQuery(transactionLogQueryRegex).WillReturnRows(sqlmock.NewRows(transactionLogColumns))

	c := NewTransactionLogCollector(db)
	families := gatherScraper(t, c)
	if _, ok := metricValue(families, "sqlserver_database_log_vlf_count", nil); ok {
		t.Error("log_vlf_count present, want no metrics for an empty result")
	}
}
