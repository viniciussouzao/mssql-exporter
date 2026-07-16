package collector

import (
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestSchedulersScrapeHappyPath(t *testing.T) {
	db, mock := newSQLMock(t)

	row := sqlmock.NewRows([]string{"runnable", "current", "workers", "work_queue", "online"}).
		AddRow(2.0, 10.0, 8.0, 0.0, 4.0)
	mock.ExpectQuery("SELECT(.|\n)*FROM sys.dm_os_schedulers").WillReturnRows(row)

	c := NewSchedulersCollector(db)
	families := gatherScraper(t, c)

	cases := map[string]float64{
		"sqlserver_schedulers_runnable_tasks": 2,
		"sqlserver_schedulers_current_tasks":  10,
		"sqlserver_schedulers_active_workers": 8,
		"sqlserver_schedulers_work_queue":     0,
		"sqlserver_schedulers_online":         4,
	}
	for name, want := range cases {
		if v, ok := metricValue(families, name, nil); !ok || v != want {
			t.Errorf("%s = %v (found=%v), want %v", name, v, ok, want)
		}
	}
}

func TestSchedulersScrapeQueryError(t *testing.T) {
	db, mock := newSQLMock(t)

	queryErr := errors.New("boom")
	mock.ExpectQuery("SELECT(.|\n)*FROM sys.dm_os_schedulers").WillReturnError(queryErr)

	c := NewSchedulersCollector(db)
	if err := scrapeErr(t, c); !errors.Is(err, queryErr) {
		t.Errorf("Scrape() error = %v, want it to wrap %v", err, queryErr)
	}
}

func TestSchedulersScrapeScanError(t *testing.T) {
	db, mock := newSQLMock(t)

	row := sqlmock.NewRows([]string{"runnable", "current", "workers", "work_queue", "online"}).
		AddRow("not-a-number", 10.0, 8.0, 0.0, 4.0)
	mock.ExpectQuery("SELECT(.|\n)*FROM sys.dm_os_schedulers").WillReturnRows(row)

	c := NewSchedulersCollector(db)
	if err := scrapeErr(t, c); err == nil {
		t.Error("Scrape() error = nil, want error from bad row scan")
	}
}
