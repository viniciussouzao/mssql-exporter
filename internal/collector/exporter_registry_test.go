package collector

import "testing"

// TestScraperRegistryConsistency guards the wiring between scraperNames and
// scraperFactories: a factory pointing at the wrong constructor, a Name()
// that doesn't match its registration key, or a name missing an entry on
// either side would otherwise compile fine and only misbehave at runtime
// when a specific --collector.<name> flag is toggled.
func TestScraperRegistryConsistency(t *testing.T) {
	db, _ := newSQLMock(t)

	names := ScraperNames()
	if len(names) == 0 {
		t.Fatal("ScraperNames() returned no names")
	}

	seen := make(map[string]bool, len(names))
	for _, name := range names {
		if seen[name] {
			t.Errorf("scraperNames contains duplicate entry %q", name)
		}
		seen[name] = true

		factory, ok := scraperFactories[name]
		if !ok {
			t.Errorf("scraperNames contains %q but scraperFactories has no entry for it", name)
			continue
		}

		scraper := factory(db, DefaultOptions())
		if scraper.Name() != name {
			t.Errorf("scraperFactories[%q] produced a scraper with Name() = %q, want %q (factory wired to the wrong constructor?)", name, scraper.Name(), name)
		}
	}

	for name := range scraperFactories {
		if !seen[name] {
			t.Errorf("scraperFactories contains %q but it is missing from scraperNames (never registered/enabled)", name)
		}
	}
}

// TestNewExporterEnabledMap guards the enable/disable semantics documented
// on NewExporter: missing entries fall back to ScraperEnabledByDefault, and
// an entry only ever toggles the specific collector it names.
func TestNewExporterEnabledMap(t *testing.T) {
	db, _ := newSQLMock(t)

	scraperNamesOf := func(e *Exporter) map[string]bool {
		names := make(map[string]bool, len(e.scrapers))
		for _, s := range e.scrapers {
			names[s.Name()] = true
		}
		return names
	}

	defaultNames := make(map[string]bool, len(scraperNames))
	for _, name := range scraperNames {
		if ScraperEnabledByDefault(name) {
			defaultNames[name] = true
		}
	}

	t.Run("nil map enables exactly the default collectors", func(t *testing.T) {
		e := NewExporter(db, nil)
		got := scraperNamesOf(e)
		if len(got) != len(defaultNames) {
			t.Errorf("got %d scrapers, want %d (collectors enabled by default)", len(got), len(defaultNames))
		}
		for _, name := range scraperNames {
			if got[name] != defaultNames[name] {
				t.Errorf("collector %q enabled = %v, want %v by default", name, got[name], defaultNames[name])
			}
		}
		if got["log_shipping"] {
			t.Error("log_shipping enabled with a nil map, want it disabled by default")
		}
	})

	t.Run("name absent from map defaults to enabled", func(t *testing.T) {
		e := NewExporter(db, map[string]bool{"sessions": true})
		got := scraperNamesOf(e)
		if !got["wait_stats"] {
			t.Error("wait_stats missing, want enabled since it has no entry in the map")
		}
	})

	t.Run("explicit false disables only that collector", func(t *testing.T) {
		e := NewExporter(db, map[string]bool{"sessions": false})
		got := scraperNamesOf(e)
		if got["sessions"] {
			t.Error("sessions present, want disabled")
		}
		if len(got) != len(defaultNames)-1 {
			t.Errorf("got %d scrapers, want %d (every default collector except sessions)", len(got), len(defaultNames)-1)
		}
	})

	t.Run("explicit true enables a collector disabled by default", func(t *testing.T) {
		e := NewExporter(db, map[string]bool{"log_shipping": true})
		if !scraperNamesOf(e)["log_shipping"] {
			t.Error("log_shipping missing, want enabled by an explicit true")
		}
	})

	t.Run("explicit true keeps the collector enabled", func(t *testing.T) {
		e := NewExporter(db, map[string]bool{"sessions": true})
		got := scraperNamesOf(e)
		if !got["sessions"] {
			t.Error("sessions missing, want enabled")
		}
	})
}

// TestNewExporterWithOptionsPassesTopQueriesOptions guards the options
// plumbing: the configured limit must reach the collector.
func TestNewExporterWithOptionsPassesTopQueriesOptions(t *testing.T) {
	db, _ := newSQLMock(t)
	opts := DefaultOptions()
	opts.TopQueries.Limit = 7

	e := NewExporterWithOptions(db, map[string]bool{"top_queries": true}, opts)
	for _, s := range e.scrapers {
		if tq, ok := s.(*SQLServerTopQueries); ok {
			if tq.opts.Limit != 7 {
				t.Errorf("top_queries limit = %d, want 7 from the options", tq.opts.Limit)
			}
			return
		}
	}
	t.Fatal("top_queries scraper not registered")
}
