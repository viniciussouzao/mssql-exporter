package collector

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/viniciussouzao/mssql-exporter/internal/db"
)

// Scraper is implemented by every SQL Server collector. Scrape returns an
// error instead of logging it so the Exporter can expose per-collector
// success/duration metrics and log failures uniformly. The context carries
// the shared scrape deadline; scrapers must pass it to every query.
type Scraper interface {
	Name() string
	Describe(ch chan<- *prometheus.Desc)
	Scrape(ctx context.Context, ch chan<- prometheus.Metric) error
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
	"blocking",
	"active_queries",
	"availability_groups",
	"agent_jobs",
	"volumes",
	"schedulers",
	"database_mirroring",
	"log_shipping",
	"top_queries",
	"index_fragmentation",
	"query_store",
}

// scrapersDisabledByDefault lists collectors that only run when explicitly
// enabled: log_shipping needs extra grants and covers a niche setup, so
// enabling it by default would report scrape failures on most instances;
// top_queries scans the plan cache and its labels churn with the workload;
// index_fragmentation reads index pages (heavy I/O on large objects);
// query_store reads every database and needs access to each of them.
var scrapersDisabledByDefault = map[string]bool{
	"log_shipping":        true,
	"top_queries":         true,
	"index_fragmentation": true,
	"query_store":         true,
}

var scraperFactories = map[string]func(*sql.DB, Options) Scraper{
	"wait_stats":          func(db *sql.DB, _ Options) Scraper { return NewWaitStatsCollector(db) },
	"io_file_stats":       func(db *sql.DB, _ Options) Scraper { return NewIOFileStatsCollector(db) },
	"sessions":            func(db *sql.DB, _ Options) Scraper { return NewSessionsCollector(db) },
	"server_info":         func(db *sql.DB, _ Options) Scraper { return NewServerInfoCollector(db) },
	"memory_clerks":       func(db *sql.DB, _ Options) Scraper { return NewMemoryClerksCollector(db) },
	"backups":             func(db *sql.DB, _ Options) Scraper { return NewBackupsCollector(db) },
	"tempdb":              func(db *sql.DB, _ Options) Scraper { return NewTempDBCollector(db) },
	"perf_counters":       func(db *sql.DB, _ Options) Scraper { return NewPerfCountersCollector(db) },
	"databases":           func(db *sql.DB, _ Options) Scraper { return NewDatabasesCollector(db) },
	"blocking":            func(db *sql.DB, _ Options) Scraper { return NewBlockingCollector(db) },
	"active_queries":      func(db *sql.DB, _ Options) Scraper { return NewActiveQueriesCollector(db) },
	"availability_groups": func(db *sql.DB, _ Options) Scraper { return NewAvailabilityGroupsCollector(db) },
	"agent_jobs":          func(db *sql.DB, _ Options) Scraper { return NewAgentJobsCollector(db) },
	"volumes":             func(db *sql.DB, _ Options) Scraper { return NewVolumesCollector(db) },
	"schedulers":          func(db *sql.DB, _ Options) Scraper { return NewSchedulersCollector(db) },
	"database_mirroring":  func(db *sql.DB, _ Options) Scraper { return NewDatabaseMirroringCollector(db) },
	"log_shipping":        func(db *sql.DB, _ Options) Scraper { return NewLogShippingCollector(db) },
	"top_queries":         func(db *sql.DB, o Options) Scraper { return NewTopQueriesCollector(db, o.TopQueries) },
	"index_fragmentation": func(db *sql.DB, o Options) Scraper {
		return NewIndexFragmentationCollector(db, o.IndexFragmentation)
	},
	"query_store": func(db *sql.DB, o Options) Scraper { return NewQueryStoreCollector(db, o.QueryStore) },
}

// ScraperNames returns the names of all available collectors in
// registration order.
func ScraperNames() []string {
	names := make([]string, len(scraperNames))
	copy(names, scraperNames)
	return names
}

// ScraperEnabledByDefault reports whether the named collector runs when no
// flag or config file entry toggles it.
func ScraperEnabledByDefault(name string) bool {
	return !scrapersDisabledByDefault[name]
}

// pinger is the subset of *sql.DB the Exporter needs for the availability
// check; it exists so tests can fake reachability.
type pinger interface {
	PingContext(ctx context.Context) error
}

// deadlineGrace is how long Collect waits past the scrape deadline for the
// ping or a scraper to return before giving up on it. The driver can block
// well past a cancelled context (e.g. waiting for the server to confirm the
// cancellation while an instance fails over), and without this bound the
// whole scrape outlives Prometheus' scrape_timeout and is lost (HML-037).
const deadlineGrace = 500 * time.Millisecond

// Exporter is the top-level prometheus.Collector. It pings SQL Server on
// every scrape (exposed as sqlserver_up), runs the enabled scrapers and
// exposes per-scraper duration and success metrics.
type Exporter struct {
	db       pinger
	scrapers []Scraper
	// timeout overrides db.CurrentOptions().ConnTimeout when set (tests).
	timeout time.Duration

	up             *prometheus.Desc
	scrapeDuration *prometheus.Desc
	scrapeSuccess  *prometheus.Desc
}

// NewExporter builds an Exporter running the scrapers enabled in the given
// map (missing entries fall back to ScraperEnabledByDefault).
func NewExporter(db *sql.DB, enabled map[string]bool) *Exporter {
	return NewExporterWithOptions(db, enabled, DefaultOptions())
}

// NewExporterWithOptions is NewExporter with collector options.
func NewExporterWithOptions(db *sql.DB, enabled map[string]bool, opts Options) *Exporter {
	var scrapers []Scraper
	for _, name := range scraperNames {
		on, ok := enabled[name]
		if !ok {
			on = ScraperEnabledByDefault(name)
		}
		if !on {
			continue
		}
		scrapers = append(scrapers, scraperFactories[name](db, opts))
	}
	return newExporter(db, scrapers)
}

func newExporter(db pinger, scrapers []Scraper) *Exporter {
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

// scrapeResult is what a scraper goroutine reports once Scrape returns.
type scrapeResult struct {
	name     string
	duration float64
	err      error
}

// Collect pings SQL Server and, when reachable, runs every enabled scraper
// concurrently under a single shared deadline (ConnTimeout), so the total
// scrape time is bounded by one timeout instead of the sum of all scrapers.
// Collect returns by the deadline plus deadlineGrace even when the driver
// ignores the cancelled context: a ping that has not returned reports
// sqlserver_up 0, and a scraper that has not returned reports success 0.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	timeout := e.timeout
	if timeout <= 0 {
		timeout = db.CurrentOptions().ConnTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	giveUp := time.NewTimer(timeout + deadlineGrace)
	defer giveUp.Stop()

	// Buffered so an abandoned ping can still deliver its result and exit.
	pingErr := make(chan error, 1)
	go func() { pingErr <- e.db.PingContext(ctx) }()
	select {
	case err := <-pingErr:
		if err != nil {
			slog.Error("sql server is unreachable", slog.Any("error", err))
			ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0)
			return
		}
	case <-giveUp.C:
		slog.Error("sql server is unreachable", slog.Any("error", errors.New("ping did not return before the scrape deadline")))
		ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0)
		return
	}
	ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 1)

	// Scrapers send to metrics instead of ch: once Collect returns, the
	// registry stops reading ch, so a scraper still running past the deadline
	// must not write to it. results is buffered so no scraper ever blocks on it.
	metrics := make(chan prometheus.Metric)
	results := make(chan scrapeResult, len(e.scrapers))
	start := time.Now()
	for _, s := range e.scrapers {
		go func(s Scraper) {
			begin := time.Now()
			err := s.Scrape(ctx, metrics)
			results <- scrapeResult{name: s.Name(), duration: time.Since(begin).Seconds(), err: err}
		}(s)
	}

	pending := make(map[string]bool, len(e.scrapers))
	for _, s := range e.scrapers {
		pending[s.Name()] = true
	}
	for len(pending) > 0 {
		select {
		case m := <-metrics:
			ch <- m
		case r := <-results:
			delete(pending, r.name)
			e.reportScrape(ch, r)
		case <-giveUp.C:
			for name := range pending {
				e.reportScrape(ch, scrapeResult{
					name:     name,
					duration: time.Since(start).Seconds(),
					err:      errors.New("collector did not return before the scrape deadline"),
				})
			}
			go drain(metrics, results, len(pending))
			return
		}
	}
}

// reportScrape emits the duration and success metrics of one scraper.
func (e *Exporter) reportScrape(ch chan<- prometheus.Metric, r scrapeResult) {
	success := 1.0
	if r.err != nil {
		success = 0
		slog.Error("collector scrape failed", slog.String("collector", r.name), slog.Any("error", r.err))
	}
	ch <- prometheus.MustNewConstMetric(e.scrapeDuration, prometheus.GaugeValue, r.duration, r.name)
	ch <- prometheus.MustNewConstMetric(e.scrapeSuccess, prometheus.GaugeValue, success, r.name)
}

// drain discards what abandoned scrapers still send, until all of them return.
func drain(metrics <-chan prometheus.Metric, results <-chan scrapeResult, remaining int) {
	for remaining > 0 {
		select {
		case <-metrics:
		case <-results:
			remaining--
		}
	}
}
