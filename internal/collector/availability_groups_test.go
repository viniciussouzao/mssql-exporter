package collector

import (
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestAvailabilityGroupsScrapeHappyPath(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery("FROM sys.dm_hadr_availability_replica_states").WillReturnRows(
		sqlmock.NewRows([]string{"ag_name", "replica_server", "role", "sync_health"}).
			AddRow("AG1", "SQL1", 1.0, 2.0))
	mock.ExpectQuery("FROM sys.dm_hadr_database_replica_states").WillReturnRows(
		sqlmock.NewRows([]string{"ag_name", "replica_server", "db_name", "sync_state", "log_send_queue", "redo_queue", "is_suspended", "suspend_reason", "secondary_lag"}).
			AddRow("AG1", "SQL1", "AppDB", 2.0, 0.0, 0.0, 0.0, nil, 0.0))

	c := NewAvailabilityGroupsCollector(db)
	families := gatherScraper(t, c)

	replicaLabels := map[string]string{"ag_name": "AG1", "replica_server": "SQL1"}
	if v, ok := metricValue(families, "sqlserver_hadr_replica_role", replicaLabels); !ok || v != 1 {
		t.Errorf("hadr_replica_role = %v (found=%v), want 1", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_hadr_replica_synchronization_health", replicaLabels); !ok || v != 2 {
		t.Errorf("hadr_replica_synchronization_health = %v (found=%v), want 2", v, ok)
	}
	dbLabels := map[string]string{"ag_name": "AG1", "replica_server": "SQL1", "db_name": "AppDB"}
	if v, ok := metricValue(families, "sqlserver_hadr_database_synchronization_state", dbLabels); !ok || v != 2 {
		t.Errorf("hadr_database_synchronization_state = %v (found=%v), want 2", v, ok)
	}
}

func TestAvailabilityGroupsScrapeDisabledEmitsNothingNoError(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery("FROM sys.dm_hadr_availability_replica_states").WillReturnRows(
		sqlmock.NewRows([]string{"ag_name", "replica_server", "role", "sync_health"}))
	mock.ExpectQuery("FROM sys.dm_hadr_database_replica_states").WillReturnRows(
		sqlmock.NewRows([]string{"ag_name", "replica_server", "db_name", "sync_state", "log_send_queue", "redo_queue", "is_suspended", "suspend_reason", "secondary_lag"}))

	c := NewAvailabilityGroupsCollector(db)
	families := gatherScraper(t, c)

	if _, ok := metricValue(families, "sqlserver_hadr_replica_role", nil); ok {
		t.Error("hadr_replica_role present, want no metrics when Always On is disabled")
	}
}

func TestAvailabilityGroupsScrapeNullStateColumnsOmitted(t *testing.T) {
	db, mock := newSQLMock(t)

	mock.ExpectQuery("FROM sys.dm_hadr_availability_replica_states").WillReturnRows(
		sqlmock.NewRows([]string{"ag_name", "replica_server", "role", "sync_health"}).
			AddRow("AG1", "SQL2", nil, nil))
	mock.ExpectQuery("FROM sys.dm_hadr_database_replica_states").WillReturnRows(
		sqlmock.NewRows([]string{"ag_name", "replica_server", "db_name", "sync_state", "log_send_queue", "redo_queue", "is_suspended", "suspend_reason", "secondary_lag"}).
			AddRow("AG1", "SQL2", "AppDB", nil, nil, nil, nil, nil, nil))

	c := NewAvailabilityGroupsCollector(db)
	families := gatherScraper(t, c)

	replicaLabels := map[string]string{"ag_name": "AG1", "replica_server": "SQL2"}
	if _, ok := metricValue(families, "sqlserver_hadr_replica_role", replicaLabels); ok {
		t.Error("hadr_replica_role present, want it omitted for NULL role")
	}
	dbLabels := map[string]string{"ag_name": "AG1", "replica_server": "SQL2", "db_name": "AppDB"}
	if _, ok := metricValue(families, "sqlserver_hadr_database_redo_queue_kb", dbLabels); ok {
		t.Error("hadr_database_redo_queue_kb present, want it omitted for NULL redo queue")
	}
}

func TestAvailabilityGroupsScrapeCombinesQueryErrors(t *testing.T) {
	db, mock := newSQLMock(t)

	replicaErr := errors.New("replica states failed")
	databaseErr := errors.New("database states failed")
	mock.ExpectQuery("FROM sys.dm_hadr_availability_replica_states").WillReturnError(replicaErr)
	mock.ExpectQuery("FROM sys.dm_hadr_database_replica_states").WillReturnError(databaseErr)

	c := NewAvailabilityGroupsCollector(db)
	err := scrapeErr(t, c)

	if !errors.Is(err, replicaErr) {
		t.Errorf("Scrape() error = %v, want it to wrap the replica states error %v", err, replicaErr)
	}
	if !errors.Is(err, databaseErr) {
		t.Errorf("Scrape() error = %v, want it to wrap the database states error %v", err, databaseErr)
	}
}

func TestAvailabilityGroupsScrapeReplicaStatesRowIterationError(t *testing.T) {
	db, mock := newSQLMock(t)

	rowErr := errors.New("row iteration failed")
	mock.ExpectQuery("FROM sys.dm_hadr_availability_replica_states").WillReturnRows(
		sqlmock.NewRows([]string{"ag_name", "replica_server", "role", "sync_health"}).
			AddRow("AG1", "SQL1", 1.0, 2.0).
			AddRow("AG1", "SQL2", 2.0, 2.0).
			RowError(1, rowErr))
	mock.ExpectQuery("FROM sys.dm_hadr_database_replica_states").WillReturnRows(
		sqlmock.NewRows([]string{"ag_name", "replica_server", "db_name", "sync_state", "log_send_queue", "redo_queue", "is_suspended", "suspend_reason", "secondary_lag"}))

	c := NewAvailabilityGroupsCollector(db)
	families, err := gatherScraperErr(t, c)

	if !errors.Is(err, rowErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v (rows.Err())", err, rowErr)
	}
	if v, ok := metricValue(families, "sqlserver_hadr_replica_role", map[string]string{"ag_name": "AG1", "replica_server": "SQL1"}); !ok || v != 1 {
		t.Errorf("hadr_replica_role = %v (found=%v), want 1 despite the later iteration error", v, ok)
	}
}

func TestAvailabilityGroupsScrapeDatabaseStatesRowIterationError(t *testing.T) {
	db, mock := newSQLMock(t)

	rowErr := errors.New("row iteration failed")
	mock.ExpectQuery("FROM sys.dm_hadr_availability_replica_states").WillReturnRows(
		sqlmock.NewRows([]string{"ag_name", "replica_server", "role", "sync_health"}))
	mock.ExpectQuery("FROM sys.dm_hadr_database_replica_states").WillReturnRows(
		sqlmock.NewRows([]string{"ag_name", "replica_server", "db_name", "sync_state", "log_send_queue", "redo_queue", "is_suspended", "suspend_reason", "secondary_lag"}).
			AddRow("AG1", "SQL1", "AppDB", 2.0, 0.0, 0.0, 0.0, nil, 0.0).
			AddRow("AG1", "SQL1", "OtherDB", 2.0, 0.0, 0.0, 0.0, nil, 0.0).
			RowError(1, rowErr))

	c := NewAvailabilityGroupsCollector(db)
	families, err := gatherScraperErr(t, c)

	if !errors.Is(err, rowErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v (rows.Err())", err, rowErr)
	}
	dbLabels := map[string]string{"ag_name": "AG1", "replica_server": "SQL1", "db_name": "AppDB"}
	if v, ok := metricValue(families, "sqlserver_hadr_database_synchronization_state", dbLabels); !ok || v != 2 {
		t.Errorf("hadr_database_synchronization_state = %v (found=%v), want 2 despite the later iteration error", v, ok)
	}
}

func TestAvailabilityGroupsScrapeSuspendedReplica(t *testing.T) {
	db, mock := newSQLMock(t)

	// Shape observed on a real AG with data movement suspended: the primary
	// still lists the secondary, but SQL Server reports no log send queue.
	mock.ExpectQuery("FROM sys.dm_hadr_availability_replica_states").WillReturnRows(
		sqlmock.NewRows([]string{"ag_name", "replica_server", "role", "sync_health"}))
	mock.ExpectQuery("FROM sys.dm_hadr_database_replica_states").WillReturnRows(
		sqlmock.NewRows([]string{"ag_name", "replica_server", "db_name", "sync_state", "log_send_queue", "redo_queue", "is_suspended", "suspend_reason", "secondary_lag"}).
			AddRow("AG1", "SQL1", "AppDB", 2.0, nil, nil, 0.0, nil, nil).
			AddRow("AG1", "SQL2", "AppDB", 0.0, nil, 0.0, 1.0, 0.0, 312.0))

	c := NewAvailabilityGroupsCollector(db)
	families := gatherScraper(t, c)

	secondary := map[string]string{"ag_name": "AG1", "replica_server": "SQL2", "db_name": "AppDB"}
	if v, ok := metricValue(families, "sqlserver_hadr_database_is_suspended", secondary); !ok || v != 1 {
		t.Errorf("hadr_database_is_suspended = %v (found=%v), want 1", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_hadr_database_suspend_reason", secondary); !ok || v != 0 {
		t.Errorf("hadr_database_suspend_reason = %v (found=%v), want 0 (user)", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_hadr_database_secondary_lag_seconds", secondary); !ok || v != 312 {
		t.Errorf("hadr_database_secondary_lag_seconds = %v (found=%v), want 312", v, ok)
	}
	if _, ok := metricValue(families, "sqlserver_hadr_database_log_send_queue_kb", secondary); ok {
		t.Error("hadr_database_log_send_queue_kb present, want it omitted for NULL while suspended")
	}

	primary := map[string]string{"ag_name": "AG1", "replica_server": "SQL1", "db_name": "AppDB"}
	if v, ok := metricValue(families, "sqlserver_hadr_database_is_suspended", primary); !ok || v != 0 {
		t.Errorf("hadr_database_is_suspended{primary} = %v (found=%v), want 0", v, ok)
	}
	if _, ok := metricValue(families, "sqlserver_hadr_database_suspend_reason", primary); ok {
		t.Error("hadr_database_suspend_reason present for a running replica, want it only while suspended")
	}
	if _, ok := metricValue(families, "sqlserver_hadr_database_secondary_lag_seconds", primary); ok {
		t.Error("hadr_database_secondary_lag_seconds present for NULL lag, want it omitted")
	}
}
