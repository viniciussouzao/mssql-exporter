package collector

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

type fakePinger struct {
	err   error
	delay time.Duration // ignores the context, like a driver stuck on a dead connection
}

func (p fakePinger) PingContext(_ context.Context) error {
	if p.delay > 0 {
		time.Sleep(p.delay)
	}
	return p.err
}

type fakeScraper struct {
	name  string
	delay time.Duration
	err   error
	desc  *prometheus.Desc
	calls atomic.Int32
}

func newFakeScraper(name string, delay time.Duration, err error) *fakeScraper {
	return &fakeScraper{
		name:  name,
		delay: delay,
		err:   err,
		desc:  prometheus.NewDesc("sqlserver_fake_"+name, "fake metric", nil, nil),
	}
}

func (s *fakeScraper) Name() string { return s.name }

func (s *fakeScraper) Describe(ch chan<- *prometheus.Desc) { ch <- s.desc }

func (s *fakeScraper) Scrape(_ context.Context, ch chan<- prometheus.Metric) error {
	s.calls.Add(1)
	if s.delay > 0 {
		time.Sleep(s.delay)
	}
	if s.err != nil {
		return s.err
	}
	ch <- prometheus.MustNewConstMetric(s.desc, prometheus.GaugeValue, 1)
	return nil
}

func gather(t *testing.T, e *Exporter) []*dto.MetricFamily {
	t.Helper()
	reg := prometheus.NewPedanticRegistry()
	if err := reg.Register(e); err != nil {
		t.Fatalf("failed to register exporter: %v", err)
	}
	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("failed to gather: %v", err)
	}
	return families
}

func metricValue(families []*dto.MetricFamily, name string, labels map[string]string) (float64, bool) {
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
			return m.GetGauge().GetValue(), true
		}
	}
	return 0, false
}

func TestExporterUpWhenReachable(t *testing.T) {
	e := newExporter(fakePinger{}, []Scraper{newFakeScraper("ok", 0, nil)})
	families := gather(t, e)

	if v, ok := metricValue(families, "sqlserver_up", nil); !ok || v != 1 {
		t.Errorf("sqlserver_up = %v (found=%v), want 1", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_scrape_collector_success", map[string]string{"collector": "ok"}); !ok || v != 1 {
		t.Errorf("scrape_collector_success{ok} = %v (found=%v), want 1", v, ok)
	}
	if _, ok := metricValue(families, "sqlserver_fake_ok", nil); !ok {
		t.Error("sqlserver_fake_ok metric missing, want scraper output present")
	}
}

func TestExporterUpZeroSkipsScrapers(t *testing.T) {
	e := newExporter(fakePinger{err: errors.New("down")}, []Scraper{newFakeScraper("ok", 0, nil)})
	families := gather(t, e)

	if v, ok := metricValue(families, "sqlserver_up", nil); !ok || v != 0 {
		t.Errorf("sqlserver_up = %v (found=%v), want 0", v, ok)
	}
	if _, ok := metricValue(families, "sqlserver_scrape_collector_success", map[string]string{"collector": "ok"}); ok {
		t.Error("scrape_collector_success present, want scrapers skipped when server is down")
	}
	if _, ok := metricValue(families, "sqlserver_fake_ok", nil); ok {
		t.Error("scraper metric present, want scrapers skipped when server is down")
	}
}

func TestExporterPartialFailure(t *testing.T) {
	e := newExporter(fakePinger{}, []Scraper{
		newFakeScraper("good", 0, nil),
		newFakeScraper("bad", 0, errors.New("query failed")),
	})
	families := gather(t, e)

	if v, ok := metricValue(families, "sqlserver_scrape_collector_success", map[string]string{"collector": "good"}); !ok || v != 1 {
		t.Errorf("success{good} = %v (found=%v), want 1", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_scrape_collector_success", map[string]string{"collector": "bad"}); !ok || v != 0 {
		t.Errorf("success{bad} = %v (found=%v), want 0", v, ok)
	}
	if _, ok := metricValue(families, "sqlserver_fake_good", nil); !ok {
		t.Error("metrics from the successful scraper missing, want them exposed despite the failing one")
	}
	if _, ok := metricValue(families, "sqlserver_scrape_collector_duration_seconds", map[string]string{"collector": "bad"}); !ok {
		t.Error("duration for failing scraper missing, want duration reported for every scraper")
	}
}

func TestExporterScrapersRunConcurrently(t *testing.T) {
	const delay = 150 * time.Millisecond
	e := newExporter(fakePinger{}, []Scraper{
		newFakeScraper("a", delay, nil),
		newFakeScraper("b", delay, nil),
		newFakeScraper("c", delay, nil),
		newFakeScraper("d", delay, nil),
	})

	start := time.Now()
	gather(t, e)
	elapsed := time.Since(start)

	// Serial execution would take at least 4*delay (600ms); concurrent
	// execution is bounded by the slowest scraper plus overhead.
	if elapsed >= 3*delay {
		t.Errorf("Collect took %v, want < %v (scrapers must run concurrently)", elapsed, 3*delay)
	}
}

// Scrapes that arrive while a collection is in flight wait for it and expose
// the same result instead of running every scraper again (HML-034).
func TestExporterConcurrentCollectsShareOneRun(t *testing.T) {
	const scrapes = 5
	s := newFakeScraper("shared", 200*time.Millisecond, nil)
	e := newExporter(fakePinger{}, []Scraper{s})

	var wg sync.WaitGroup
	results := make([][]*dto.MetricFamily, scrapes)
	for i := range scrapes {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = gather(t, e)
		}()
	}
	wg.Wait()

	if got := s.calls.Load(); got != 1 {
		t.Errorf("scraper ran %d times for %d simultaneous scrapes, want 1", got, scrapes)
	}
	for i, families := range results {
		if v, ok := metricValue(families, "sqlserver_scrape_collector_success", map[string]string{"collector": "shared"}); !ok || v != 1 {
			t.Errorf("scrape %d: success{shared} = %v (found=%v), want 1", i, v, ok)
		}
		if _, ok := metricValue(families, "sqlserver_fake_shared", nil); !ok {
			t.Errorf("scrape %d: sqlserver_fake_shared missing, want the shared result exposed to every scrape", i)
		}
	}
}

// Sharing only covers scrapes that overlap: the next scrape collects again.
func TestExporterSequentialCollectsRunAgain(t *testing.T) {
	s := newFakeScraper("again", 0, nil)
	e := newExporter(fakePinger{}, []Scraper{s})

	gather(t, e)
	gather(t, e)

	if got := s.calls.Load(); got != 2 {
		t.Errorf("scraper ran %d times for 2 sequential scrapes, want 2", got)
	}
}

// A ping stuck past the deadline (the driver waiting for a cancellation
// that never comes during a failover) must still yield sqlserver_up 0 in
// time, instead of a scrape that outlives Prometheus' scrape_timeout.
func TestExporterPingStuckPastDeadline(t *testing.T) {
	e := newExporter(fakePinger{delay: 2 * time.Second}, []Scraper{newFakeScraper("ok", 0, nil)})
	e.timeout = 50 * time.Millisecond

	start := time.Now()
	families := gather(t, e)
	elapsed := time.Since(start)

	if limit := e.timeout + deadlineGrace + 300*time.Millisecond; elapsed > limit {
		t.Errorf("Collect took %v, want <= %v", elapsed, limit)
	}
	if v, ok := metricValue(families, "sqlserver_up", nil); !ok || v != 0 {
		t.Errorf("sqlserver_up = %v (found=%v), want 0", v, ok)
	}
	if _, ok := metricValue(families, "sqlserver_scrape_collector_success", map[string]string{"collector": "ok"}); ok {
		t.Error("scrape_collector_success present, want scrapers skipped when the ping does not return")
	}
}

// A scraper stuck past the deadline is reported as failed and the others
// are still exposed; what it sends after Collect returns must be discarded
// (the registry no longer reads the channel by then).
func TestExporterAbandonsScraperStuckPastDeadline(t *testing.T) {
	const stuck = time.Second
	e := newExporter(fakePinger{}, []Scraper{
		newFakeScraper("fast", 0, nil),
		newFakeScraper("stuck", stuck, nil),
	})
	e.timeout = 50 * time.Millisecond

	start := time.Now()
	families := gather(t, e)
	elapsed := time.Since(start)

	if limit := e.timeout + deadlineGrace + 300*time.Millisecond; elapsed > limit {
		t.Errorf("Collect took %v, want <= %v", elapsed, limit)
	}
	if v, ok := metricValue(families, "sqlserver_up", nil); !ok || v != 1 {
		t.Errorf("sqlserver_up = %v (found=%v), want 1", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_scrape_collector_success", map[string]string{"collector": "fast"}); !ok || v != 1 {
		t.Errorf("success{fast} = %v (found=%v), want 1", v, ok)
	}
	if v, ok := metricValue(families, "sqlserver_scrape_collector_success", map[string]string{"collector": "stuck"}); !ok || v != 0 {
		t.Errorf("success{stuck} = %v (found=%v), want 0", v, ok)
	}
	if _, ok := metricValue(families, "sqlserver_fake_fast", nil); !ok {
		t.Error("sqlserver_fake_fast missing, want metrics of scrapers that finished in time")
	}
	if _, ok := metricValue(families, "sqlserver_fake_stuck", nil); ok {
		t.Error("sqlserver_fake_stuck present, want the abandoned scraper's late metric discarded")
	}

	// Let the stuck scraper send its metric after Collect returned: a send
	// on the registry's channel would panic or block here.
	time.Sleep(stuck - elapsed + 100*time.Millisecond)
}
