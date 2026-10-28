//go:build integration

// Package integration runs the exporter against a real SQL Server.
//
// MSSQL_EXPORTER_TEST_DSN must point to a disposable instance as a sysadmin
// (e.g. sa on the Docker image): the tests create a database, a backup and a
// monitoring login with the grants documented in the README, then collect
// as that login. Run with `make test-integration` (starts a container) or
// `go test -tags integration ./integration/...` against your own instance.
package integration

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	_ "github.com/microsoft/go-mssqldb"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/viniciussouzao/mssql-exporter/internal/collector"
	"github.com/viniciussouzao/mssql-exporter/internal/db"
)

const (
	testDatabase = "exporter_it"
	testLogin    = "exporter_it_login"
)

// monitoringDSN is the DSN of the least-privilege login, set by TestMain.
var monitoringDSN string

func TestMain(m *testing.M) {
	adminDSN := os.Getenv("MSSQL_EXPORTER_TEST_DSN")
	if adminDSN == "" {
		fmt.Println("MSSQL_EXPORTER_TEST_DSN not set, skipping integration tests")
		os.Exit(0)
	}

	dsn, err := setUp(adminDSN)
	if err != nil {
		fmt.Fprintln(os.Stderr, "integration setup failed:", err)
		os.Exit(1)
	}
	monitoringDSN = dsn
	os.Exit(m.Run())
}

// setUp waits for SQL Server, creates the fixtures as the admin login and
// returns the DSN of the monitoring login.
func setUp(adminDSN string) (string, error) {
	admin, err := sql.Open("sqlserver", adminDSN)
	if err != nil {
		return "", err
	}
	defer func() { _ = admin.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	for {
		if err = admin.PingContext(ctx); err == nil {
			break
		}
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("sql server not reachable: %w", err)
		case <-time.After(2 * time.Second):
		}
	}

	password, err := randomPassword()
	if err != nil {
		return "", err
	}

	// The grants are the "Permissions" section of the README: if they stop
	// being enough for a collector, TestAllCollectorsSucceed fails.
	statements := []string{
		`IF DB_ID(N'` + testDatabase + `') IS NULL CREATE DATABASE [` + testDatabase + `]`,
		`ALTER DATABASE [` + testDatabase + `] SET RECOVERY FULL`,
		`ALTER DATABASE [` + testDatabase + `] SET QUERY_STORE = ON`,
		`USE [` + testDatabase + `];
		IF OBJECT_ID(N'dbo.items') IS NULL
		BEGIN
			CREATE TABLE dbo.items (id int IDENTITY PRIMARY KEY, grp int NOT NULL, payload char(100) NOT NULL);
			INSERT dbo.items (grp, payload) SELECT TOP (20000) ABS(CHECKSUM(NEWID())) % 100, 'x' FROM sys.all_objects a CROSS JOIN sys.all_objects b;
		END`,
		`BACKUP DATABASE [` + testDatabase + `] TO DISK = N'` + testDatabase + `.bak' WITH INIT`,
		`IF SUSER_ID(N'` + testLogin + `') IS NULL
			CREATE LOGIN [` + testLogin + `] WITH PASSWORD = N'` + password + `'
		ELSE
			ALTER LOGIN [` + testLogin + `] WITH PASSWORD = N'` + password + `'`,
		`GRANT VIEW SERVER STATE TO [` + testLogin + `]`,
		`GRANT VIEW ANY DEFINITION TO [` + testLogin + `]`,
		`GRANT CONNECT ANY DATABASE TO [` + testLogin + `]`,
		`USE msdb;
		IF USER_ID(N'` + testLogin + `') IS NULL CREATE USER [` + testLogin + `] FOR LOGIN [` + testLogin + `];
		GRANT SELECT ON dbo.sysjobs TO [` + testLogin + `];
		GRANT SELECT ON dbo.sysjobhistory TO [` + testLogin + `];
		GRANT SELECT ON dbo.log_shipping_monitor_primary TO [` + testLogin + `];
		GRANT SELECT ON dbo.log_shipping_monitor_secondary TO [` + testLogin + `];`,
	}
	for _, stmt := range statements {
		if _, err := admin.ExecContext(ctx, stmt); err != nil {
			return "", fmt.Errorf("fixture statement failed: %w\n%s", err, stmt)
		}
	}

	u, err := url.Parse(adminDSN)
	if err != nil {
		return "", fmt.Errorf("MSSQL_EXPORTER_TEST_DSN must be a sqlserver:// URL: %w", err)
	}
	u.User = url.UserPassword(testLogin, password)
	return u.String(), nil
}

// randomPassword satisfies the SQL Server password policy (upper, lower,
// digit and symbol).
func randomPassword() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "It-" + hex.EncodeToString(b) + "-Aa1", nil
}

// newExporter connects as the monitoring login with every collector
// enabled.
func newExporter(t *testing.T, dsn string) *collector.Exporter {
	t.Helper()
	conn, err := sql.Open("sqlserver", dsn)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	opts := db.DefaultOptions()
	conn.SetMaxOpenConns(opts.MaxOpenConns)

	enabled := map[string]bool{}
	for _, name := range collector.ScraperNames() {
		enabled[name] = true
	}
	return collector.NewExporterWithOptions(conn, enabled, collector.DefaultOptions())
}

func gather(t *testing.T, c prometheus.Collector) []*dto.MetricFamily {
	t.Helper()
	reg := prometheus.NewPedanticRegistry()
	if err := reg.Register(c); err != nil {
		t.Fatalf("register: %v", err)
	}
	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	return families
}

// series returns the samples of a metric whose labels include the given
// ones.
func series(families []*dto.MetricFamily, name string, labels map[string]string) []*dto.Metric {
	var out []*dto.Metric
	for _, f := range families {
		if f.GetName() != name {
			continue
		}
	metric:
		for _, m := range f.GetMetric() {
			got := map[string]string{}
			for _, l := range m.GetLabel() {
				got[l.GetName()] = l.GetValue()
			}
			for k, v := range labels {
				if got[k] != v {
					continue metric
				}
			}
			out = append(out, m)
		}
	}
	return out
}

func value(m *dto.Metric) float64 {
	if m.GetCounter() != nil {
		return m.GetCounter().GetValue()
	}
	return m.GetGauge().GetValue()
}

// backgroundCollectors serve the result of a background collection that
// starts on the first scrape.
var backgroundCollectors = []string{"index_fragmentation", "query_store"}

// TestAllCollectorsSucceed enables every collector as the least-privilege
// login and requires each one to succeed, which also checks that the
// README grants are enough.
func TestAllCollectorsSucceed(t *testing.T) {
	e := newExporter(t, monitoringDSN)

	var families []*dto.MetricFamily
	deadline := time.Now().Add(90 * time.Second)
	for {
		families = gather(t, e)
		done := true
		for _, name := range backgroundCollectors {
			s := series(families, "sqlserver_"+name+"_last_collection_success", nil)
			if len(s) == 0 {
				done = false
			}
		}
		if done || time.Now().After(deadline) {
			break
		}
		time.Sleep(2 * time.Second)
	}

	if s := series(families, "sqlserver_up", nil); len(s) != 1 || value(s[0]) != 1 {
		t.Fatalf("sqlserver_up = %v, want 1", s)
	}
	for _, name := range collector.ScraperNames() {
		s := series(families, "sqlserver_scrape_collector_success", map[string]string{"collector": name})
		if len(s) != 1 || value(s[0]) != 1 {
			t.Errorf("collector %s did not succeed (scrape_collector_success = %v)", name, s)
		}
	}
	for _, name := range backgroundCollectors {
		s := series(families, "sqlserver_"+name+"_last_collection_success", nil)
		if len(s) != 1 || value(s[0]) != 1 {
			t.Errorf("%s background collection did not succeed (last_collection_success = %v)", name, s)
		}
	}
}

// TestCollectedValues checks a few values against the fixtures.
func TestCollectedValues(t *testing.T) {
	families := gather(t, newExporter(t, monitoringDSN))
	dbLabel := map[string]string{"db_name": testDatabase}

	if s := series(families, "sqlserver_database_state", map[string]string{"db_name": testDatabase, "state": "ONLINE"}); len(s) != 1 || value(s[0]) != 1 {
		t.Errorf("database_state{%s,ONLINE} = %v, want 1", testDatabase, s)
	}
	if s := series(families, "sqlserver_database_recovery_model", map[string]string{"db_name": testDatabase, "recovery_model": "FULL"}); len(s) != 1 {
		t.Errorf("database_recovery_model{%s,FULL} missing", testDatabase)
	}
	if s := series(families, "sqlserver_database_data_size_bytes", dbLabel); len(s) != 1 || value(s[0]) <= 0 {
		t.Errorf("database_data_size_bytes{%s} = %v, want > 0", testDatabase, s)
	}
	s := series(families, "sqlserver_backup_last_completed_timestamp_seconds", map[string]string{"database_name": testDatabase, "backup_type": "full"})
	if len(s) == 0 {
		t.Errorf("backup_last_completed_timestamp_seconds{%s} missing, want the fixture backup", testDatabase)
	} else if age := time.Since(time.Unix(int64(value(s[0])), 0)); age < -time.Minute || age > time.Hour {
		t.Errorf("last backup of %s is %v old, want within the last hour (UTC conversion)", testDatabase, age)
	}
	for _, m := range series(families, "sqlserver_volume_total_bytes", nil) {
		for _, l := range m.GetLabel() {
			if l.GetName() == "volume_mount_point" && l.GetValue() == "" {
				t.Error("volume_total_bytes with an empty volume_mount_point")
			}
		}
	}
	if s := series(families, "sqlserver_engine_edition", nil); len(s) != 1 {
		t.Errorf("sqlserver_engine_edition = %v, want one series", s)
	}
}

// TestUnreachableServerReportsDown checks that a server that cannot be
// reached yields sqlserver_up 0 within the query timeout instead of an error.
func TestUnreachableServerReportsDown(t *testing.T) {
	e := newExporter(t, "sqlserver://nobody:nothing@127.0.0.1:1?connection+timeout=2")

	start := time.Now()
	families := gather(t, e)
	if elapsed := time.Since(start); elapsed > db.DefaultOptions().ConnTimeout+2*time.Second {
		t.Errorf("scrape of an unreachable server took %v", elapsed)
	}
	if s := series(families, "sqlserver_up", nil); len(s) != 1 || value(s[0]) != 0 {
		t.Errorf("sqlserver_up = %v, want 0", s)
	}
	if s := series(families, "sqlserver_scrape_collector_success", nil); len(s) != 0 {
		t.Errorf("collector metrics present while the server is down: %d series", len(s))
	}
}
