package collector

import (
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

var missingIndexesQueryRegex = sqlContainsInOrder(
	"SELECT TOP (@p1)",
	"FROM sys.dm_db_missing_index_details d",
	"JOIN sys.dm_db_missing_index_groups g",
	"JOIN sys.dm_db_missing_index_group_stats s",
	"ORDER BY improvement DESC",
)

var missingIndexesColumns = []string{
	"db_name", "schema", "table", "index_handle", "equality_columns", "inequality_columns", "included_columns",
	"user_seeks", "user_scans", "avg_user_impact", "avg_total_user_cost", "improvement",
}

func TestMissingIndexesScrapeHappyPath(t *testing.T) {
	db, mock := newSQLMock(t)

	rows := sqlmock.NewRows(missingIndexesColumns).
		AddRow("AppDB", "dbo", "Orders", "7", "[Status]", "[OrderDate]", "[CustomerID], [Amount]", 20.0, 1.0, 95.86, 4.83, 97.2)
	mock.ExpectQuery(missingIndexesQueryRegex).WithArgs(20).WillReturnRows(rows)

	c := NewMissingIndexesCollector(db, MissingIndexesOptions{Limit: 20})
	families := gatherScraper(t, c)

	labels := map[string]string{"db_name": "AppDB", "schema": "dbo", "table": "Orders", "index_handle": "7"}
	cases := map[string]float64{
		"sqlserver_missing_index_improvement_measure":     97.2,
		"sqlserver_missing_index_avg_user_impact_percent": 95.86,
		"sqlserver_missing_index_avg_total_user_cost":     4.83,
	}
	for name, want := range cases {
		if v, ok := metricValue(families, name, labels); !ok || v != want {
			t.Errorf("%s = %v (found=%v), want %v", name, v, ok, want)
		}
	}
	for name, want := range map[string]float64{
		"sqlserver_missing_index_user_seeks_total": 20,
		"sqlserver_missing_index_user_scans_total": 1,
	} {
		if v, ok := counterValue(families, name, labels); !ok || v != want {
			t.Errorf("%s = %v (found=%v), want %v", name, v, ok, want)
		}
	}

	info := map[string]string{
		"table": "Orders", "equality_columns": "[Status]", "inequality_columns": "[OrderDate]",
		"included_columns": "[CustomerID], [Amount]",
	}
	if v, ok := metricValue(families, "sqlserver_missing_index_info", info); !ok || v != 1 {
		t.Errorf("missing_index_info = %v (found=%v), want 1 with the index columns", v, ok)
	}
}

func TestMissingIndexesScrapeQueryError(t *testing.T) {
	db, mock := newSQLMock(t)

	queryErr := errors.New("boom")
	mock.ExpectQuery(missingIndexesQueryRegex).WillReturnError(queryErr)

	c := NewMissingIndexesCollector(db, DefaultOptions().MissingIndexes)
	if err := scrapeErr(t, c); !errors.Is(err, queryErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v", err, queryErr)
	}
}

func TestMissingIndexesScrapeScanError(t *testing.T) {
	db, mock := newSQLMock(t)

	rows := sqlmock.NewRows(missingIndexesColumns).
		AddRow("AppDB", "dbo", "Bad", "1", "", "", "", "not-a-number", 0.0, 1.0, 1.0, 1.0).
		AddRow("AppDB", "dbo", "Orders", "7", "[CustomerID]", "", "", 20.0, 0.0, 97.26, 4.8, 93.4)
	mock.ExpectQuery(missingIndexesQueryRegex).WillReturnRows(rows)

	c := NewMissingIndexesCollector(db, DefaultOptions().MissingIndexes)
	families, err := gatherScraperErr(t, c)
	if err == nil {
		t.Error("Scrape() error = nil, want error from bad row")
	}
	if _, ok := metricValue(families, "sqlserver_missing_index_improvement_measure", map[string]string{"table": "Orders"}); !ok {
		t.Error("improvement_measure{Orders} missing, want rows after a bad one still exposed")
	}
}

func TestMissingIndexesScrapeEmpty(t *testing.T) {
	db, mock := newSQLMock(t)
	mock.ExpectQuery(missingIndexesQueryRegex).WillReturnRows(sqlmock.NewRows(missingIndexesColumns))

	c := NewMissingIndexesCollector(db, DefaultOptions().MissingIndexes)
	families := gatherScraper(t, c)
	if _, ok := metricValue(families, "sqlserver_missing_index_info", nil); ok {
		t.Error("missing_index_info present, want nothing without suggestions")
	}
}
