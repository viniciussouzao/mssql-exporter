package collector

import (
	"errors"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

var replicationAgentColumns = []string{
	"agent_type", "agent_name", "publisher", "publisher_db", "publication", "subscriber", "subscriber_db",
	"runstatus", "last_history", "delivery_latency", "pending",
}

func expectDistributionDBs(mock sqlmock.Sqlmock, names ...string) {
	rows := sqlmock.NewRows([]string{"name"})
	for _, n := range names {
		rows.AddRow(n)
	}
	mock.ExpectQuery(sqlContainsInOrder("FROM sys.databases", "is_distributor = 1", "HAS_DBACCESS(name) = 1")).WillReturnRows(rows)
}

func TestReplicationScrapeHappyPath(t *testing.T) {
	db, mock := newSQLMock(t)
	expectDistributionDBs(mock, "distribution")

	last := time.Date(2026, 10, 28, 13, 54, 41, 0, time.UTC)
	rows := sqlmock.NewRows(replicationAgentColumns).
		AddRow("distribution", "SRV-ReplPub-ItemsPub-SRV-3", "SRV", "ReplPub", "ItemsPub", "SRV", "ReplSub", 3.0, last, 1703.0, 12.0).
		AddRow("logreader", "SRV-ReplPub-1", "SRV", "ReplPub", "", "", "", 6.0, last, 4000.0, nil)
	mock.ExpectQuery(sqlContainsInOrder(
		"FROM [distribution].dbo.MSdistribution_agents a",
		"[distribution].dbo.MSreplservers",
		"[distribution].dbo.MSdistribution_status",
		"WHERE a.subscriber_id >= 0",
		"FROM [distribution].dbo.MSlogreader_agents a",
	)).WillReturnRows(rows)

	c := NewReplicationCollector(db)
	families := gatherScraper(t, c)

	dist := map[string]string{"distribution_db": "distribution", "agent_type": "distribution", "publication": "ItemsPub", "subscriber_db": "ReplSub"}
	cases := map[string]float64{
		"sqlserver_replication_agent_run_status":                     3,
		"sqlserver_replication_agent_last_history_timestamp_seconds": float64(last.Unix()),
		"sqlserver_replication_agent_delivery_latency_seconds":       1.703,
		"sqlserver_replication_pending_commands":                     12,
	}
	for name, want := range cases {
		if v, ok := metricValue(families, name, dist); !ok || v != want {
			t.Errorf("%s{distribution agent} = %v (found=%v), want %v", name, v, ok, want)
		}
	}

	lr := map[string]string{"agent_type": "logreader", "publisher_db": "ReplPub"}
	if v, ok := metricValue(families, "sqlserver_replication_agent_run_status", lr); !ok || v != 6 {
		t.Errorf("run_status{logreader} = %v (found=%v), want 6 (failed)", v, ok)
	}
	if _, ok := metricValue(families, "sqlserver_replication_pending_commands", lr); ok {
		t.Error("pending_commands present for the log reader, want it only for distribution agents")
	}
}

// An agent that never ran has no history row: only what is known is exposed.
func TestReplicationScrapeAgentWithoutHistory(t *testing.T) {
	db, mock := newSQLMock(t)
	expectDistributionDBs(mock, "distribution")
	mock.ExpectQuery("MSdistribution_agents").WillReturnRows(sqlmock.NewRows(replicationAgentColumns).
		AddRow("distribution", "new-agent", "SRV", "Pub", "P", "SRV", "Sub", nil, nil, nil, 0.0))

	c := NewReplicationCollector(db)
	families := gatherScraper(t, c)
	if _, ok := metricValue(families, "sqlserver_replication_agent_run_status", nil); ok {
		t.Error("run_status present, want none without history")
	}
	if v, ok := metricValue(families, "sqlserver_replication_pending_commands", map[string]string{"agent_name": "new-agent"}); !ok || v != 0 {
		t.Errorf("pending_commands = %v (found=%v), want 0", v, ok)
	}
}

// Not a Distributor: nothing is queried beyond the database list.
func TestReplicationScrapeNotDistributor(t *testing.T) {
	db, mock := newSQLMock(t)
	expectDistributionDBs(mock)

	c := NewReplicationCollector(db)
	families := gatherScraper(t, c)
	if _, ok := metricValue(families, "sqlserver_replication_agent_run_status", nil); ok {
		t.Error("replication metrics present on an instance without distribution databases")
	}
}

func TestReplicationScrapeListError(t *testing.T) {
	db, mock := newSQLMock(t)
	listErr := errors.New("boom")
	mock.ExpectQuery("is_distributor = 1").WillReturnError(listErr)

	c := NewReplicationCollector(db)
	if err := scrapeErr(t, c); !errors.Is(err, listErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v", err, listErr)
	}
}

// A failing distribution database does not hide the others.
func TestReplicationScrapeOneDistributionDBFails(t *testing.T) {
	db, mock := newSQLMock(t)
	expectDistributionDBs(mock, "distA", "distB")
	queryErr := errors.New("permission denied")
	mock.ExpectQuery(`\[distA\]\.dbo\.MSdistribution_agents`).WillReturnError(queryErr)
	mock.ExpectQuery(`\[distB\]\.dbo\.MSdistribution_agents`).WillReturnRows(sqlmock.NewRows(replicationAgentColumns).
		AddRow("logreader", "lr", "SRV", "Pub", "", "", "", 4.0, time.Now(), 0.0, nil))

	c := NewReplicationCollector(db)
	families, err := gatherScraperErr(t, c)
	if !errors.Is(err, queryErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v", err, queryErr)
	}
	if v, ok := metricValue(families, "sqlserver_replication_agent_run_status", map[string]string{"distribution_db": "distB"}); !ok || v != 4 {
		t.Errorf("run_status{distB} = %v (found=%v), want 4", v, ok)
	}
}

func TestReplicationAgentsQueryQuotesDatabase(t *testing.T) {
	q := replicationAgentsQuery("dist]x")
	if !strings.Contains(q, "[dist]]x].dbo.MSdistribution_agents") {
		t.Errorf("distribution database not quoted as an identifier:\n%s", q)
	}
}
