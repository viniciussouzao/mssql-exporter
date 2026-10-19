package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

// accessibleDatabasesQueryRegex pins the database filter: user databases
// only, online, and accessible to the login (no 916 on databases without a
// user, mirrors or non-readable AG secondaries).
var accessibleDatabasesQueryRegex = sqlContainsInOrder(
	"FROM sys.databases",
	"WHERE database_id > 4",
	"AND state = 0",
	"AND HAS_DBACCESS(name) = 1",
)

// fragmentationQueryRegex pins LIMITED mode, the three-part catalog names of
// the database being read, the partition aggregation and the size filter.
func fragmentationQueryRegex(quotedDB string) string {
	return sqlContainsInOrder(
		// Heaps get an empty index_name: a placeholder like "HEAP" could
		// collide with a real index of that name (duplicate series).
		"ISNULL(i.name, '')",
		"SUM(ps.avg_fragmentation_in_percent * ps.page_count) / NULLIF(SUM(ps.page_count), 0)",
		"FROM sys.dm_db_index_physical_stats(DB_ID(@p1), NULL, NULL, NULL, 'LIMITED') ps",
		"INNER JOIN "+quotedDB+".sys.indexes i",
		"INNER JOIN "+quotedDB+".sys.objects o",
		"INNER JOIN "+quotedDB+".sys.schemas s",
		"WHERE ps.alloc_unit_type_desc = 'IN_ROW_DATA'",
		"AND o.is_ms_shipped = 0",
		"GROUP BY s.name, o.name, i.name",
		"HAVING SUM(ps.page_count) >= @p2",
	)
}

var fragmentationColumns = []string{"schema", "table", "index", "fragmentation", "pages"}

// newTestFragmentationCollector returns a collector whose background loop
// never starts, so tests drive refresh directly, with a fake clock.
func newTestFragmentationCollector(t *testing.T, opts IndexFragmentationOptions) (*SQLServerIndexFragmentation, sqlmock.Sqlmock) {
	t.Helper()
	db, mock := newSQLMock(t)
	c := NewIndexFragmentationCollector(db, opts)
	c.startOnce.Do(func() {})
	clock := time.Date(2026, 10, 15, 18, 0, 0, 0, time.UTC)
	c.now = func() time.Time {
		clock = clock.Add(90 * time.Second)
		return clock
	}
	return c, mock
}

func TestIndexFragmentationBeforeFirstCollectionEmitsNothing(t *testing.T) {
	c, _ := newTestFragmentationCollector(t, DefaultOptions().IndexFragmentation)

	families := gatherScraper(t, c)
	if len(families) != 0 {
		t.Errorf("got %d metric families before the first collection, want none", len(families))
	}
}

func TestIndexFragmentationRefreshAndScrape(t *testing.T) {
	c, mock := newTestFragmentationCollector(t, DefaultOptions().IndexFragmentation)

	mock.ExpectQuery(accessibleDatabasesQueryRegex).WillReturnRows(
		sqlmock.NewRows([]string{"name"}).AddRow("AppDB").AddRow("Odd]Name"))
	mock.ExpectQuery(fragmentationQueryRegex("[AppDB]")).WithArgs("AppDB", 1000).WillReturnRows(
		sqlmock.NewRows(fragmentationColumns).
			AddRow("dbo", "Orders", "PK_Orders", 37.5, 2048).
			AddRow("dbo", "Logs", "", 80.0, 1500))
	mock.ExpectQuery(fragmentationQueryRegex("[Odd]]Name]")).WithArgs("Odd]Name", 1000).WillReturnRows(
		sqlmock.NewRows(fragmentationColumns).
			AddRow("sales", "Items", "IX_Items_sku", nil, 1000))

	c.refresh(context.Background())
	families := gatherScraper(t, c)

	orders := map[string]string{"db_name": "AppDB", "schema_name": "dbo", "table_name": "Orders", "index_name": "PK_Orders"}
	if v, ok := metricValue(families, "sqlserver_index_fragmentation_percent", orders); !ok || v != 37.5 {
		t.Errorf("fragmentation{PK_Orders} = %v (found=%v), want 37.5", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_index_page_count", orders); !ok || v != 2048 {
		t.Errorf("page_count{PK_Orders} = %v (found=%v), want 2048", v, ok)
	}
	heap := map[string]string{"table_name": "Logs", "index_name": ""}
	if v, ok := metricValue(families, "sqlserver_index_fragmentation_percent", heap); !ok || v != 80 {
		t.Errorf("fragmentation{heap Logs} = %v (found=%v), want 80 with an empty index_name", v, ok)
	}
	// NULL fragmentation (no pages counted) is reported as 0.
	items := map[string]string{"db_name": "Odd]Name", "index_name": "IX_Items_sku"}
	if v, ok := metricValue(families, "sqlserver_index_fragmentation_percent", items); !ok || v != 0 {
		t.Errorf("fragmentation{IX_Items_sku} = %v (found=%v), want 0", v, ok)
	}

	for name, want := range map[string]float64{
		"sqlserver_index_fragmentation_last_collection_success":          1,
		"sqlserver_index_fragmentation_last_collection_databases":        2,
		"sqlserver_index_fragmentation_last_collection_duration_seconds": 90,
		"sqlserver_index_fragmentation_last_collection_timestamp_seconds": float64(
			time.Date(2026, 10, 15, 18, 3, 0, 0, time.UTC).Unix()),
	} {
		if v, ok := metricValue(families, name, nil); !ok || v != want {
			t.Errorf("%s = %v (found=%v), want %v", name, v, ok, want)
		}
	}
}

func TestIndexFragmentationDatabasesOption(t *testing.T) {
	opts := DefaultOptions().IndexFragmentation
	opts.Databases = []string{"APPDB"}
	opts.MinPageCount = 50
	c, mock := newTestFragmentationCollector(t, opts)

	mock.ExpectQuery(accessibleDatabasesQueryRegex).WillReturnRows(
		sqlmock.NewRows([]string{"name"}).AddRow("AppDB").AddRow("OtherDB"))
	// Only AppDB (matched case-insensitively against "APPDB"), with the
	// configured page filter.
	mock.ExpectQuery(fragmentationQueryRegex("[AppDB]")).WithArgs("AppDB", 50).WillReturnRows(
		sqlmock.NewRows(fragmentationColumns).AddRow("dbo", "T", "IX", 5.0, 60))

	c.refresh(context.Background())
	families := gatherScraper(t, c)

	if v, ok := metricValue(families, "sqlserver_index_fragmentation_last_collection_databases", nil); !ok || v != 1 {
		t.Errorf("last_collection_databases = %v (found=%v), want 1", v, ok)
	}
}

func TestIndexFragmentationDatabaseErrorKeepsOthers(t *testing.T) {
	c, mock := newTestFragmentationCollector(t, DefaultOptions().IndexFragmentation)

	mock.ExpectQuery(accessibleDatabasesQueryRegex).WillReturnRows(
		sqlmock.NewRows([]string{"name"}).AddRow("BadDB").AddRow("GoodDB"))
	mock.ExpectQuery(fragmentationQueryRegex("[BadDB]")).WillReturnError(errors.New("lock request time out period exceeded"))
	mock.ExpectQuery(fragmentationQueryRegex("[GoodDB]")).WillReturnRows(
		sqlmock.NewRows(fragmentationColumns).AddRow("dbo", "T", "IX", 12.0, 3000))

	c.refresh(context.Background())
	families, err := gatherScraperErr(t, c)

	if err != nil {
		t.Errorf("Scrape() error = %v, want nil: background failures are reported by last_collection_success", err)
	}
	if v, ok := metricValue(families, "sqlserver_index_fragmentation_last_collection_success", nil); !ok || v != 0 {
		t.Errorf("last_collection_success = %v (found=%v), want 0", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_index_fragmentation_last_collection_databases", nil); !ok || v != 1 {
		t.Errorf("last_collection_databases = %v (found=%v), want 1", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_index_fragmentation_percent", map[string]string{"db_name": "GoodDB"}); !ok || v != 12 {
		t.Errorf("fragmentation{GoodDB} = %v (found=%v), want 12 despite BadDB failing", v, ok)
	}
}

func TestIndexFragmentationListErrorReportsFailure(t *testing.T) {
	c, mock := newTestFragmentationCollector(t, DefaultOptions().IndexFragmentation)

	mock.ExpectQuery(accessibleDatabasesQueryRegex).WillReturnError(errors.New("boom"))

	c.refresh(context.Background())
	families := gatherScraper(t, c)

	if v, ok := metricValue(families, "sqlserver_index_fragmentation_last_collection_success", nil); !ok || v != 0 {
		t.Errorf("last_collection_success = %v (found=%v), want 0", v, ok)
	}
	if _, ok := metricValue(families, "sqlserver_index_fragmentation_percent", nil); ok {
		t.Error("fragmentation_percent present, want none when the database list failed")
	}
}

func TestIndexFragmentationLatestResultReplacesPrevious(t *testing.T) {
	c, mock := newTestFragmentationCollector(t, DefaultOptions().IndexFragmentation)

	mock.ExpectQuery(accessibleDatabasesQueryRegex).WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("AppDB"))
	mock.ExpectQuery(fragmentationQueryRegex("[AppDB]")).WillReturnRows(
		sqlmock.NewRows(fragmentationColumns).AddRow("dbo", "Dropped", "IX", 50.0, 2000))
	c.refresh(context.Background())

	mock.ExpectQuery(accessibleDatabasesQueryRegex).WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("AppDB"))
	mock.ExpectQuery(fragmentationQueryRegex("[AppDB]")).WillReturnRows(
		sqlmock.NewRows(fragmentationColumns).AddRow("dbo", "Kept", "IX", 1.0, 2000))
	c.refresh(context.Background())

	families := gatherScraper(t, c)
	if _, ok := metricValue(families, "sqlserver_index_page_count", map[string]string{"table_name": "Dropped"}); ok {
		t.Error("page_count{Dropped} present, want only the latest collection's indexes")
	}
	if _, ok := metricValue(families, "sqlserver_index_page_count", map[string]string{"table_name": "Kept"}); !ok {
		t.Error("page_count{Kept} missing from the latest collection")
	}
}

func TestIndexFragmentationScrapeStartsBackgroundCollectionOnce(t *testing.T) {
	db, mock := newSQLMock(t)
	opts := DefaultOptions().IndexFragmentation
	opts.Interval = time.Hour
	c := NewIndexFragmentationCollector(db, opts)

	mock.ExpectQuery(accessibleDatabasesQueryRegex).WillReturnRows(sqlmock.NewRows([]string{"name"}))

	// The first scrape only starts the collection; a second one must not
	// start another (sqlmock would fail on an unexpected second query).
	_ = gatherScraper(t, c)
	_ = gatherScraper(t, c)

	deadline := time.Now().Add(5 * time.Second)
	for {
		c.mu.Lock()
		done := c.last != nil
		c.mu.Unlock()
		if done {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("background collection did not finish")
		}
		time.Sleep(10 * time.Millisecond)
	}
	families := gatherScraper(t, c)
	if v, ok := metricValue(families, "sqlserver_index_fragmentation_last_collection_success", nil); !ok || v != 1 {
		t.Errorf("last_collection_success = %v (found=%v), want 1 after the background collection", v, ok)
	}
}

func TestQuoteIdentifier(t *testing.T) {
	for in, want := range map[string]string{
		"AppDB":      "[AppDB]",
		"Odd]Name":   "[Odd]]Name]",
		"a]]b":       "[a]]]]b]",
		"with space": "[with space]",
	} {
		if got := quoteIdentifier(in); got != want {
			t.Errorf("quoteIdentifier(%q) = %q, want %q", in, got, want)
		}
	}
}
