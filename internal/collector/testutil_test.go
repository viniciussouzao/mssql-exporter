package collector

import (
	"context"
	"database/sql"
	"regexp"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// newSQLMock builds a sqlmock-backed *sql.DB and registers a cleanup that
// verifies every expectation was met and closes the connection, so no test
// can pass because an expected query silently never ran.
func newSQLMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet sqlmock expectations: %v", err)
		}
		// sqlmock treats Close() itself as an expectation in ordered mode
		// unless mock.ExpectClose() was set up; none of these tests care
		// about close ordering, so the error is intentionally discarded.
		_ = db.Close()
	})
	return db, mock
}

// sqlContainsInOrder builds a regex requiring each fragment to appear, in
// that order, anywhere in the query (including across newlines). Used to
// pin down semantically critical SQL (filters, joins, ordering) without
// coupling tests to incidental whitespace.
func sqlContainsInOrder(fragments ...string) string {
	quoted := make([]string, len(fragments))
	for i, f := range fragments {
		quoted[i] = regexp.QuoteMeta(f)
	}
	return "(?s)" + strings.Join(quoted, ".*")
}

// scraperCollector adapts a Scraper into a prometheus.Collector so it can be
// registered on a Pedantic registry and gathered like production code does,
// while capturing the error Scrape returned (Collect has no return value).
type scraperCollector struct {
	s   Scraper
	err error
}

func (w *scraperCollector) Describe(ch chan<- *prometheus.Desc) { w.s.Describe(ch) }

func (w *scraperCollector) Collect(ch chan<- prometheus.Metric) {
	w.err = w.s.Scrape(context.Background(), ch)
}

// gatherScraperErr runs a single scraper exactly once through a Pedantic
// registry and returns both the gathered metric families and the error
// Scrape produced, mirroring how the real Exporter exposes metrics.
func gatherScraperErr(t *testing.T, s Scraper) ([]*dto.MetricFamily, error) {
	t.Helper()
	w := &scraperCollector{s: s}
	reg := prometheus.NewPedanticRegistry()
	if err := reg.Register(w); err != nil {
		t.Fatalf("failed to register scraper: %v", err)
	}
	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("failed to gather: %v", err)
	}
	return families, w.err
}

// counterValue looks up a CounterValue metric's value, mirroring
// metricValue in exporter_test.go which only handles gauges.
func counterValue(families []*dto.MetricFamily, name string, labels map[string]string) (float64, bool) {
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
			return m.GetCounter().GetValue(), true
		}
	}
	return 0, false
}

// gatherScraper is for tests that expect a clean scrape: it fails the test
// immediately if Scrape returns an error, so a happy path can never pass
// just because some metrics happened to be emitted before a late failure.
// Tests that expect an error must use gatherScraperErr or scrapeErr instead.
func gatherScraper(t *testing.T, s Scraper) []*dto.MetricFamily {
	t.Helper()
	families, err := gatherScraperErr(t, s)
	if err != nil {
		t.Fatalf("Scrape() error = %v, want nil", err)
	}
	return families
}

// scrapeErr runs Scrape directly (no registry), draining the metric
// channel, and returns the error it produced. Use only when the test has no
// need to inspect gathered metric families.
func scrapeErr(t *testing.T, s Scraper) error {
	t.Helper()
	ch := make(chan prometheus.Metric, 256)
	err := s.Scrape(context.Background(), ch)
	close(ch)
	for range ch {
	}
	return err
}
