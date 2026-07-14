package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

type fakePinger struct {
	err error
}

func (p fakePinger) PingContext(_ context.Context) error {
	return p.err
}

type fakeScraper struct {
	name  string
	delay time.Duration
	err   error
	desc  *prometheus.Desc
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
