package collector

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/viniciussouzao/mssql-exporter/internal/db"
)

// Scraper is implemented by every SQL Server collector. Scrape returns an
// error instead of logging it so the Exporter can expose per-collector
// success/duration metrics and log failures uniformly.
type Scraper interface {
	Name() string
	Describe(ch chan<- *prometheus.Desc)
	Scrape(ch chan<- prometheus.Metric) error
}

// scraperNames holds the registration order; each name is also the flag
// (--collector.<name>) and YAML (collectors: <name>) key for the collector.
var scraperNames = []string{
	"wait_stats",
	"io_file_stats",
	"sessions",
	"server_info",
	"memory_clerks",
	"backups",
	"tempdb",
	"perf_counters",
	"databases",
}

var scraperFactories = map[string]func(*sql.DB) Scraper{
	"wait_stats":    func(db *sql.DB) Scraper { return NewWaitStatsCollector(db) },
	"io_file_stats": func(db *sql.DB) Scraper { return NewIOFileStatsCollector(db) },
	"sessions":      func(db *sql.DB) Scraper { return NewSessionsCollector(db) },
	"server_info":   func(db *sql.DB) Scraper { return NewServerInfoCollector(db) },
	"memory_clerks": func(db *sql.DB) Scraper { return NewMemoryClerksCollector(db) },
	"backups":       func(db *sql.DB) Scraper { return NewBackupsCollector(db) },
	"tempdb":        func(db *sql.DB) Scraper { return NewTempDBCollector(db) },
	"perf_counters": func(db *sql.DB) Scraper { return NewPerfCountersCollector(db) },
	"databases":     func(db *sql.DB) Scraper { return NewDatabasesCollector(db) },
}

// ScraperNames returns the names of all available collectors in
// registration order.
func ScraperNames() []string {
	names := make([]string, len(scraperNames))
	copy(names, scraperNames)
	return names
}

// Exporter is the top-level prometheus.Collector. It pings SQL Server on
// every scrape (exposed as sqlserver_up), runs the enabled scrapers and
// exposes per-scraper duration and success metrics.
type Exporter struct {
	db       *sql.DB
	scrapers []Scraper

	up             *prometheus.Desc
	scrapeDuration *prometheus.Desc
	scrapeSuccess  *prometheus.Desc
}

// NewExporter builds an Exporter running the scrapers enabled in the given
// map (missing entries default to enabled).
func NewExporter(db *sql.DB, enabled map[string]bool) *Exporter {
	var scrapers []Scraper
	for _, name := range scraperNames {
		if on, ok := enabled[name]; ok && !on {
			continue
		}
		scrapers = append(scrapers, scraperFactories[name](db))
	}

	return &Exporter{
		db:       db,
		scrapers: scrapers,
		up: prometheus.NewDesc(
			"sqlserver_up",
			"Whether the SQL Server instance is reachable (1) or not (0)",
			nil, nil,
		),
		scrapeDuration: prometheus.NewDesc(
			"sqlserver_scrape_collector_duration_seconds",
			"Duration of a collector scrape",
			[]string{"collector"}, nil,
		),
		scrapeSuccess: prometheus.NewDesc(
			"sqlserver_scrape_collector_success",
			"Whether a collector scrape succeeded (1) or not (0)",
			[]string{"collector"}, nil,
		),
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.up
	ch <- e.scrapeDuration
	ch <- e.scrapeSuccess
	for _, s := range e.scrapers {
		s.Describe(ch)
	}
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), db.CurrentOptions().ConnTimeout)
	defer cancel()

	if err := e.db.PingContext(ctx); err != nil {
		slog.Error("sql server is unreachable", slog.Any("error", err))
		ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0)
		return
	}
	ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 1)

	for _, s := range e.scrapers {
		start := time.Now()
		err := s.Scrape(ch)
		duration := time.Since(start).Seconds()

		success := 1.0
		if err != nil {
			success = 0
			slog.Error("collector scrape failed", slog.String("collector", s.Name()), slog.Any("error", err))
		}

		ch <- prometheus.MustNewConstMetric(e.scrapeDuration, prometheus.GaugeValue, duration, s.Name())
		ch <- prometheus.MustNewConstMetric(e.scrapeSuccess, prometheus.GaugeValue, success, s.Name())
	}
}
