package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

// queryStoreQueryRegex pins the three-part names of the database being read
// and the forced plan counts.
func queryStoreQueryRegex(quotedDB string) string {
	return sqlContainsInOrder(
		"o.actual_state_desc",
		"o.desired_state_desc",
		"o.readonly_reason",
		"o.current_storage_size_mb",
		"o.max_storage_size_mb",
		"FROM "+quotedDB+".sys.database_query_store_options o",
		"COUNT(CASE WHEN force_failure_count > 0 THEN 1 END)",
		"FROM "+quotedDB+".sys.query_store_plan",
		"WHERE is_forced_plan = 1",
	)
}

var queryStoreColumns = []string{"actual", "desired", "readonly_reason", "used_mb", "max_mb", "forced", "failing"}

// newTestQueryStoreCollector returns a collector whose background loop never
// starts, so tests drive refresh directly, with a fake clock.
func newTestQueryStoreCollector(t *testing.T, opts QueryStoreOptions) (*SQLServerQueryStore, sqlmock.Sqlmock) {
	t.Helper()
	db, mock := newSQLMock(t)
	c := NewQueryStoreCollector(db, opts)
	c.startOnce.Do(func() {})
	clock := time.Date(2026, 10, 19, 18, 0, 0, 0, time.UTC)
	c.now = func() time.Time {
		clock = clock.Add(2 * time.Second)
		return clock
	}
	return c, mock
}

func TestQueryStoreBeforeFirstCollectionEmitsNothing(t *testing.T) {
	c, _ := newTestQueryStoreCollector(t, DefaultOptions().QueryStore)

	families := gatherScraper(t, c)
	if len(families) != 0 {
		t.Errorf("got %d metric families before the first collection, want none", len(families))
	}
}

func TestQueryStoreRefreshAndScrape(t *testing.T) {
	c, mock := newTestQueryStoreCollector(t, DefaultOptions().QueryStore)

	mock.ExpectQuery(accessibleDatabasesQueryRegex).WillReturnRows(
		sqlmock.NewRows([]string{"name"}).AddRow("AppDB").AddRow("Odd]Name"))
	mock.ExpectQuery(queryStoreQueryRegex("[AppDB]")).WillReturnRows(
		sqlmock.NewRows(queryStoreColumns).AddRow("READ_ONLY", "READ_WRITE", 65536, 1000, 1000, 3, 1))
	mock.ExpectQuery(queryStoreQueryRegex("[Odd]]Name]")).WillReturnRows(
		sqlmock.NewRows(queryStoreColumns).AddRow("READ_WRITE", "READ_WRITE", 0, 12, 1000, 0, 0))

	c.refresh(context.Background())
	families := gatherScraper(t, c)

	app := map[string]string{"db_name": "AppDB"}
	for name, want := range map[string]float64{
		"sqlserver_query_store_readonly_reason":      65536,
		"sqlserver_query_store_storage_used_bytes":   1000 * 1024 * 1024,
		"sqlserver_query_store_storage_max_bytes":    1000 * 1024 * 1024,
		"sqlserver_query_store_forced_plans":         3,
		"sqlserver_query_store_forced_plans_failing": 1,
	} {
		if v, ok := metricValue(families, name, app); !ok || v != want {
			t.Errorf("%s{AppDB} = %v (found=%v), want %v", name, v, ok, want)
		}
	}
	if v, ok := metricValue(families, "sqlserver_query_store_actual_state", map[string]string{"db_name": "AppDB", "state": "READ_ONLY"}); !ok || v != 1 {
		t.Errorf("actual_state{AppDB,READ_ONLY} = %v (found=%v), want 1", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_query_store_desired_state", map[string]string{"db_name": "AppDB", "state": "READ_WRITE"}); !ok || v != 1 {
		t.Errorf("desired_state{AppDB,READ_WRITE} = %v (found=%v), want 1", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_query_store_storage_used_bytes", map[string]string{"db_name": "Odd]Name"}); !ok || v != 12*1024*1024 {
		t.Errorf("storage_used_bytes{Odd]Name} = %v (found=%v), want 12 MB", v, ok)
	}

	for name, want := range map[string]float64{
		"sqlserver_query_store_last_collection_success":          1,
		"sqlserver_query_store_last_collection_databases":        2,
		"sqlserver_query_store_last_collection_duration_seconds": 2,
		"sqlserver_query_store_last_collection_timestamp_seconds": float64(
			time.Date(2026, 10, 19, 18, 0, 4, 0, time.UTC).Unix()),
	} {
		if v, ok := metricValue(families, name, nil); !ok || v != want {
			t.Errorf("%s = %v (found=%v), want %v", name, v, ok, want)
		}
	}
}

func TestQueryStoreNullOptionsAreOmitted(t *testing.T) {
	c, mock := newTestQueryStoreCollector(t, DefaultOptions().QueryStore)

	mock.ExpectQuery(accessibleDatabasesQueryRegex).WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("AppDB"))
	mock.ExpectQuery(queryStoreQueryRegex("[AppDB]")).WillReturnRows(
		sqlmock.NewRows(queryStoreColumns).AddRow("OFF", "OFF", nil, nil, nil, 0, 0))

	c.refresh(context.Background())
	families := gatherScraper(t, c)

	for _, name := range []string{
		"sqlserver_query_store_readonly_reason",
		"sqlserver_query_store_storage_used_bytes",
		"sqlserver_query_store_storage_max_bytes",
	} {
		if _, ok := metricValue(families, name, nil); ok {
			t.Errorf("%s present, want NULL to be omitted instead of reported as 0", name)
		}
	}
	if v, ok := metricValue(families, "sqlserver_query_store_actual_state", map[string]string{"state": "OFF"}); !ok || v != 1 {
		t.Errorf("actual_state{OFF} = %v (found=%v), want 1", v, ok)
	}
}

func TestQueryStoreDatabasesOption(t *testing.T) {
	opts := DefaultOptions().QueryStore
	opts.Databases = []string{"appdb"}
	c, mock := newTestQueryStoreCollector(t, opts)

	mock.ExpectQuery(accessibleDatabasesQueryRegex).WillReturnRows(
		sqlmock.NewRows([]string{"name"}).AddRow("AppDB").AddRow("OtherDB"))
	mock.ExpectQuery(queryStoreQueryRegex("[AppDB]")).WillReturnRows(
		sqlmock.NewRows(queryStoreColumns).AddRow("READ_WRITE", "READ_WRITE", 0, 1, 100, 0, 0))

	c.refresh(context.Background())
	families := gatherScraper(t, c)

	if v, ok := metricValue(families, "sqlserver_query_store_last_collection_databases", nil); !ok || v != 1 {
		t.Errorf("last_collection_databases = %v (found=%v), want 1", v, ok)
	}
}

func TestQueryStoreDatabaseErrorKeepsOthers(t *testing.T) {
	c, mock := newTestQueryStoreCollector(t, DefaultOptions().QueryStore)

	mock.ExpectQuery(accessibleDatabasesQueryRegex).WillReturnRows(
		sqlmock.NewRows([]string{"name"}).AddRow("BadDB").AddRow("GoodDB"))
	mock.ExpectQuery(queryStoreQueryRegex("[BadDB]")).WillReturnError(errors.New("Invalid object name 'sys.database_query_store_options'"))
	mock.ExpectQuery(queryStoreQueryRegex("[GoodDB]")).WillReturnRows(
		sqlmock.NewRows(queryStoreColumns).AddRow("READ_WRITE", "READ_WRITE", 0, 5, 100, 1, 0))

	c.refresh(context.Background())
	families, err := gatherScraperErr(t, c)

	if err != nil {
		t.Errorf("Scrape() error = %v, want nil: background failures are reported by last_collection_success", err)
	}
	if v, ok := metricValue(families, "sqlserver_query_store_last_collection_success", nil); !ok || v != 0 {
		t.Errorf("last_collection_success = %v (found=%v), want 0", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_query_store_forced_plans", map[string]string{"db_name": "GoodDB"}); !ok || v != 1 {
		t.Errorf("forced_plans{GoodDB} = %v (found=%v), want 1 despite BadDB failing", v, ok)
	}
	if _, ok := metricValue(families, "sqlserver_query_store_forced_plans", map[string]string{"db_name": "BadDB"}); ok {
		t.Error("forced_plans{BadDB} present, want the failed database skipped")
	}
}

func TestQueryStoreListErrorReportsFailure(t *testing.T) {
	c, mock := newTestQueryStoreCollector(t, DefaultOptions().QueryStore)

	mock.ExpectQuery(accessibleDatabasesQueryRegex).WillReturnError(errors.New("boom"))

	c.refresh(context.Background())
	families := gatherScraper(t, c)

	if v, ok := metricValue(families, "sqlserver_query_store_last_collection_success", nil); !ok || v != 0 {
		t.Errorf("last_collection_success = %v (found=%v), want 0", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_query_store_last_collection_databases", nil); !ok || v != 0 {
		t.Errorf("last_collection_databases = %v (found=%v), want 0", v, ok)
	}
}

func TestQueryStoreScrapeStartsBackgroundCollectionOnce(t *testing.T) {
	db, mock := newSQLMock(t)
	opts := DefaultOptions().QueryStore
	opts.Interval = time.Hour
	c := NewQueryStoreCollector(db, opts)

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
	if v, ok := metricValue(families, "sqlserver_query_store_last_collection_success", nil); !ok || v != 1 {
		t.Errorf("last_collection_success = %v (found=%v), want 1 after the background collection", v, ok)
	}
}
