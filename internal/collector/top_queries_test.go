package collector

import (
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

// topQueriesQueryRegex pins the single plan cache scan aggregated per
// query_hash, the representative plan, the TOP applied before the per-row
// APPLYs, and the database taken from the plan attributes.
var topQueriesQueryRegex = sqlContainsInOrder(
	"SUM(total_worker_time) OVER (PARTITION BY query_hash) AS worker_time_us",
	"ROW_NUMBER() OVER (PARTITION BY query_hash ORDER BY total_worker_time DESC) AS rn",
	"FROM sys.dm_exec_query_stats",
	"CASE WHEN @p2 = 1 THEN",
	"SELECT TOP (@p1) *",
	"WHERE rn = 1",
	"ORDER BY worker_time_us DESC",
	"sys.dm_exec_plan_attributes(top_n.plan_handle)",
	"WHERE attribute = N'dbid'",
	"sys.dm_exec_sql_text(top_n.sql_handle)",
)

var topQueriesColumns = []string{"query_hash", "db_name", "executions", "worker_time_us", "elapsed_time_us",
	"logical_reads", "physical_reads", "logical_writes", "text"}

func TestTopQueriesScrapeHappyPath(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery(topQueriesQueryRegex).WithArgs(50, 0).WillReturnRows(
		sqlmock.NewRows(topQueriesColumns).
			// Text present despite include_text off: the collector must still
			// not expose it (the query also returns NULL then).
			AddRow("0x1A2B3C4D5E6F7081", "LoadTestDB", 1000, 2500000, 4000000, 80000, 120, 30, "SELECT 1"))

	c := NewTopQueriesCollector(db, DefaultOptions().TopQueries)
	families := gatherScraper(t, c)

	labels := map[string]string{"db_name": "LoadTestDB", "query_hash": "0x1A2B3C4D5E6F7081"}
	for name, want := range map[string]float64{
		"sqlserver_top_query_executions_total":      1000,
		"sqlserver_top_query_cpu_seconds_total":     2.5,
		"sqlserver_top_query_elapsed_seconds_total": 4,
		"sqlserver_top_query_logical_reads_total":   80000,
		"sqlserver_top_query_physical_reads_total":  120,
		"sqlserver_top_query_logical_writes_total":  30,
	} {
		if v, ok := counterValue(families, name, labels); !ok || v != want {
			t.Errorf("%s = %v (found=%v), want %v", name, v, ok, want)
		}
	}
	if _, ok := metricValue(families, "sqlserver_top_query_info", nil); ok {
		t.Error("top_query_info present, want it absent unless include_text is enabled")
	}
}

func TestTopQueriesScrapeIncludeTextObfuscatesAndTruncates(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery(topQueriesQueryRegex).WithArgs(10, 1).WillReturnRows(
		sqlmock.NewRows(topQueriesColumns).
			AddRow("0x01", "LoadTestDB", 5, 100, 100, 1, 0, 0, "SELECT id FROM dbo.Orders WHERE cpf = '123.456.789-00'").
			AddRow("0x02", "", 1, 50, 50, 1, 0, 0, nil))

	opts := TopQueriesOptions{Limit: 10, IncludeText: true, TextLength: 30}
	c := NewTopQueriesCollector(db, opts)
	families := gatherScraper(t, c)

	want := map[string]string{"db_name": "LoadTestDB", "query_hash": "0x01", "query_text": "SELECT id FROM dbo.Orders W..."}
	if v, ok := metricValue(families, "sqlserver_top_query_info", want); !ok || v != 1 {
		t.Errorf("top_query_info%v = %v (found=%v), want 1", want, v, ok)
	}
	// A plan evicted between the scan and the text lookup has no text: the
	// counters are still emitted, the info series is not.
	if _, ok := metricValue(families, "sqlserver_top_query_info", map[string]string{"query_hash": "0x02"}); ok {
		t.Error("top_query_info{0x02} present, want it absent for a query without text")
	}
	if v, ok := counterValue(families, "sqlserver_top_query_executions_total", map[string]string{"query_hash": "0x02"}); !ok || v != 1 {
		t.Errorf("top_query_executions_total{0x02} = %v (found=%v), want 1", v, ok)
	}
}

func TestTopQueriesScrapeTextNeverLeaksLiterals(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery(topQueriesQueryRegex).WithArgs(10, 1).WillReturnRows(
		sqlmock.NewRows(topQueriesColumns).
			AddRow("0x03", "LoadTestDB", 1, 1, 1, 1, 0, 0, "UPDATE t SET card = N'4111111111111111' WHERE id = 42"))

	c := NewTopQueriesCollector(db, TopQueriesOptions{Limit: 10, IncludeText: true, TextLength: 120})
	families := gatherScraper(t, c)

	want := map[string]string{"query_text": "UPDATE t SET card = ? WHERE id = ?"}
	if v, ok := metricValue(families, "sqlserver_top_query_info", want); !ok || v != 1 {
		t.Errorf("top_query_info%v = %v (found=%v), want 1 with literals obfuscated", want, v, ok)
	}
}

func TestTopQueriesScrapeQueryError(t *testing.T) {
	db, mock := newSQLMock(t)

	queryErr := errors.New("VIEW SERVER STATE permission was denied")
	mock.ExpectQuery(topQueriesQueryRegex).WillReturnError(queryErr)

	c := NewTopQueriesCollector(db, DefaultOptions().TopQueries)
	if err := scrapeErr(t, c); !errors.Is(err, queryErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v", err, queryErr)
	}
}

func TestTopQueriesScrapeScanErrorKeepsOtherRows(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery(topQueriesQueryRegex).WillReturnRows(
		sqlmock.NewRows(topQueriesColumns).
			AddRow("0xBAD", "db", "not-a-number", 1, 1, 1, 0, 0, nil).
			AddRow("0x04", "db", 7, 1, 1, 1, 0, 0, nil))

	c := NewTopQueriesCollector(db, DefaultOptions().TopQueries)
	families, err := gatherScraperErr(t, c)

	if err == nil {
		t.Error("Scrape() error = nil, want the scan error reported")
	}
	if v, ok := counterValue(families, "sqlserver_top_query_executions_total", map[string]string{"query_hash": "0x04"}); !ok || v != 7 {
		t.Errorf("top_query_executions_total{0x04} = %v (found=%v), want 7 for the row after the bad one", v, ok)
	}
}
