package collector

import (
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

// mirroringStatesQueryRegex pins the filter that keeps databases which are
// not mirrored (mirroring_guid NULL, including AG databases) out.
var mirroringStatesQueryRegex = sqlContainsInOrder(
	// No witness: state 0 with an empty name, must not be emitted.
	"CASE WHEN dm.mirroring_witness_name <> '' THEN dm.mirroring_witness_state END",
	"FROM sys.database_mirroring dm",
	"WHERE dm.mirroring_guid IS NOT NULL",
)

// mirroringCountersQueryRegex pins the per-database instances: the _Total
// instance would otherwise be reported as a database named "_Total".
var mirroringCountersQueryRegex = sqlContainsInOrder(
	"FROM sys.dm_os_performance_counters",
	"object_name LIKE '%Database Mirroring%'",
	"instance_name <> '_Total'",
)

var mirroringStatesColumns = []string{"name", "partner_instance", "role", "state", "safety_level", "witness_state"}
var mirroringCountersColumns = []string{"instance_name", "counter_name", "cntr_value"}

func TestDatabaseMirroringScrapeHappyPath(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery(mirroringStatesQueryRegex).WillReturnRows(
		sqlmock.NewRows(mirroringStatesColumns).
			AddRow("MirrorDB", "SQL2", 1.0, 4.0, 2.0, 1.0))
	mock.ExpectQuery(mirroringCountersQueryRegex).WillReturnRows(
		sqlmock.NewRows(mirroringCountersColumns).
			AddRow("MirrorDB", "Log Send Queue KB", 128.0).
			AddRow("MirrorDB", "Redo Queue KB", 0.0).
			AddRow("MirrorDB", "Log Bytes Sent/sec", 4096.0).
			AddRow("MirrorDB", "Log Bytes Received/sec", 0.0).
			AddRow("MirrorDB", "Redo Bytes/sec", 0.0).
			AddRow("MirrorDB", "Transaction Delay", 350.0).
			AddRow("MirrorDB", "Mirrored Write Transactions/sec", 70.0))

	c := NewDatabaseMirroringCollector(db)
	families := gatherScraper(t, c)

	labels := map[string]string{"db_name": "MirrorDB"}
	if v, ok := metricValue(families, "sqlserver_mirroring_info", map[string]string{"db_name": "MirrorDB", "partner_instance": "SQL2"}); !ok || v != 1 {
		t.Errorf("mirroring_info = %v (found=%v), want 1", v, ok)
	}
	for name, want := range map[string]float64{
		"sqlserver_mirroring_role":              1,
		"sqlserver_mirroring_state":             4,
		"sqlserver_mirroring_safety_level":      2,
		"sqlserver_mirroring_witness_state":     1,
		"sqlserver_mirroring_log_send_queue_kb": 128,
		"sqlserver_mirroring_redo_queue_kb":     0,
	} {
		if v, ok := metricValue(families, name, labels); !ok || v != want {
			t.Errorf("%s = %v (found=%v), want %v", name, v, ok, want)
		}
	}
	for name, want := range map[string]float64{
		"sqlserver_mirroring_log_sent_bytes_total":       4096,
		"sqlserver_mirroring_log_received_bytes_total":   0,
		"sqlserver_mirroring_redo_bytes_total":           0,
		"sqlserver_mirroring_transaction_delay_ms_total": 350,
		"sqlserver_mirroring_write_transactions_total":   70,
	} {
		if v, ok := counterValue(families, name, labels); !ok || v != want {
			t.Errorf("%s = %v (found=%v), want %v", name, v, ok, want)
		}
	}
}

func TestDatabaseMirroringScrapeNotConfiguredEmitsNothingNoError(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery(mirroringStatesQueryRegex).WillReturnRows(sqlmock.NewRows(mirroringStatesColumns))
	mock.ExpectQuery(mirroringCountersQueryRegex).WillReturnRows(sqlmock.NewRows(mirroringCountersColumns))

	c := NewDatabaseMirroringCollector(db)
	families := gatherScraper(t, c)

	if len(families) != 0 {
		t.Errorf("got %d metric families, want none when no database is mirrored", len(families))
	}
}

func TestDatabaseMirroringScrapeNullColumnsOmitted(t *testing.T) {
	db, mock := newSQLMock(t)

	// No witness configured and a disconnected session: NULL witness state
	// and partner instance must not fail the scrape.
	mock.ExpectQuery(mirroringStatesQueryRegex).WillReturnRows(
		sqlmock.NewRows(mirroringStatesColumns).
			AddRow("MirrorDB", "", 2.0, 1.0, 1.0, nil))
	mock.ExpectQuery(mirroringCountersQueryRegex).WillReturnRows(sqlmock.NewRows(mirroringCountersColumns))

	c := NewDatabaseMirroringCollector(db)
	families := gatherScraper(t, c)

	labels := map[string]string{"db_name": "MirrorDB"}
	if v, ok := metricValue(families, "sqlserver_mirroring_state", labels); !ok || v != 1 {
		t.Errorf("mirroring_state = %v (found=%v), want 1", v, ok)
	}
	if _, ok := metricValue(families, "sqlserver_mirroring_witness_state", labels); ok {
		t.Error("mirroring_witness_state present, want it omitted when there is no witness")
	}
}

func TestDatabaseMirroringScrapeCombinesQueryErrors(t *testing.T) {
	db, mock := newSQLMock(t)

	statesErr := errors.New("states failed")
	countersErr := errors.New("counters failed")
	mock.ExpectQuery(mirroringStatesQueryRegex).WillReturnError(statesErr)
	mock.ExpectQuery(mirroringCountersQueryRegex).WillReturnError(countersErr)

	c := NewDatabaseMirroringCollector(db)
	err := scrapeErr(t, c)

	if !errors.Is(err, statesErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v", err, statesErr)
	}
	if !errors.Is(err, countersErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v", err, countersErr)
	}
}

func TestDatabaseMirroringScrapeScanErrorKeepsOtherRows(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery(mirroringStatesQueryRegex).WillReturnRows(
		sqlmock.NewRows(mirroringStatesColumns).
			AddRow("BadDB", "SQL2", "not-a-number", 4.0, 2.0, nil).
			AddRow("MirrorDB", "SQL2", 2.0, 4.0, 2.0, nil))
	mock.ExpectQuery(mirroringCountersQueryRegex).WillReturnRows(sqlmock.NewRows(mirroringCountersColumns))

	c := NewDatabaseMirroringCollector(db)
	families, err := gatherScraperErr(t, c)

	if err == nil {
		t.Error("Scrape() error = nil, want the scan error reported")
	}
	if v, ok := metricValue(families, "sqlserver_mirroring_role", map[string]string{"db_name": "MirrorDB"}); !ok || v != 2 {
		t.Errorf("mirroring_role = %v (found=%v), want 2 for the row after the bad one", v, ok)
	}
}
