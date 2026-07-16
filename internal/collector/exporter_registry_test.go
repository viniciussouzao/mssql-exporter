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

		scraper := factory(db)
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
// on NewExporter: entries default to enabled, and the enabled map can only
// ever turn off the specific collector it names.
func TestNewExporterEnabledMap(t *testing.T) {
	db, _ := newSQLMock(t)

	scraperNamesOf := func(e *Exporter) map[string]bool {
		names := make(map[string]bool, len(e.scrapers))
		for _, s := range e.scrapers {
			names[s.Name()] = true
		}
		return names
	}

	t.Run("nil map enables all registered collectors", func(t *testing.T) {
		e := NewExporter(db, nil)
		got := scraperNamesOf(e)
		if len(got) != len(scraperNames) {
			t.Errorf("got %d scrapers, want %d (all registered collectors)", len(got), len(scraperNames))
		}
		for _, name := range scraperNames {
			if !got[name] {
				t.Errorf("collector %q missing, want enabled by default", name)
			}
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
		if len(got) != len(scraperNames)-1 {
			t.Errorf("got %d scrapers, want %d (every collector except sessions)", len(got), len(scraperNames)-1)
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
