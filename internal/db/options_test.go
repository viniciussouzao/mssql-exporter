package db

import (
	"testing"
	"time"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.MaxOpenConns != 5 {
		t.Errorf("MaxOpenConns = %d, want 5", opts.MaxOpenConns)
	}
	if opts.MaxIdleConns != 5 {
		t.Errorf("MaxIdleConns = %d, want 5", opts.MaxIdleConns)
	}
	if opts.ConnMaxLifetime != 5*time.Minute {
		t.Errorf("ConnMaxLifetime = %v, want 5m", opts.ConnMaxLifetime)
	}
	if opts.ConnMaxIdleTime != 2*time.Minute {
		t.Errorf("ConnMaxIdleTime = %v, want 2m", opts.ConnMaxIdleTime)
	}
	if opts.ConnTimeout != 5*time.Second {
		t.Errorf("ConnTimeout = %v, want 5s", opts.ConnTimeout)
	}
	// ConnTimeout is not a pool setter: database/sql has no equivalent knob,
	// it is only read back via CurrentOptions() by callers building their
	// own context deadline (see collector.Exporter.Collect).
}

func TestCurrentOptionsFallsBackToDefaults(t *testing.T) {
	t.Cleanup(withActualOpts())
	actualOpts = nil

	got := CurrentOptions()
	want := DefaultOptions()
	if *got != *want {
		t.Errorf("CurrentOptions() = %+v, want defaults %+v", got, want)
	}
}

func TestCurrentOptionsReturnsConfiguredOptions(t *testing.T) {
	t.Cleanup(withActualOpts())
	custom := &Options{MaxOpenConns: 42}
	actualOpts = custom

	if CurrentOptions() != custom {
		t.Errorf("CurrentOptions() = %v, want the options set via Connect", CurrentOptions())
	}
}

// fakePoolConfigurer records every pool-tuning call so applyPoolOptions can
// be verified without a real *sql.DB.
type fakePoolConfigurer struct {
	maxOpenConns    int
	maxIdleConns    int
	connMaxLifetime time.Duration
	connMaxIdleTime time.Duration

	connMaxLifetimeCalled bool
	connMaxIdleTimeCalled bool
}

func (f *fakePoolConfigurer) SetMaxOpenConns(n int) { f.maxOpenConns = n }
func (f *fakePoolConfigurer) SetMaxIdleConns(n int) { f.maxIdleConns = n }
func (f *fakePoolConfigurer) SetConnMaxLifetime(d time.Duration) {
	f.connMaxLifetime = d
	f.connMaxLifetimeCalled = true
}
func (f *fakePoolConfigurer) SetConnMaxIdleTime(d time.Duration) {
	f.connMaxIdleTime = d
	f.connMaxIdleTimeCalled = true
}

func TestApplyPoolOptionsSetsEverySetter(t *testing.T) {
	f := &fakePoolConfigurer{}
	opts := &Options{
		MaxOpenConns:    3,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Minute,
		ConnMaxIdleTime: 30 * time.Second,
	}

	applyPoolOptions(f, opts)

	if f.maxOpenConns != opts.MaxOpenConns {
		t.Errorf("maxOpenConns = %d, want %d", f.maxOpenConns, opts.MaxOpenConns)
	}
	if f.maxIdleConns != opts.MaxIdleConns {
		t.Errorf("maxIdleConns = %d, want %d", f.maxIdleConns, opts.MaxIdleConns)
	}
	if !f.connMaxLifetimeCalled || f.connMaxLifetime != opts.ConnMaxLifetime {
		t.Errorf("connMaxLifetime = %v (called=%v), want %v", f.connMaxLifetime, f.connMaxLifetimeCalled, opts.ConnMaxLifetime)
	}
	if !f.connMaxIdleTimeCalled || f.connMaxIdleTime != opts.ConnMaxIdleTime {
		t.Errorf("connMaxIdleTime = %v (called=%v), want %v", f.connMaxIdleTime, f.connMaxIdleTimeCalled, opts.ConnMaxIdleTime)
	}
}

func TestApplyPoolOptionsSkipsZeroLifetimeAndIdleTime(t *testing.T) {
	f := &fakePoolConfigurer{}
	opts := &Options{MaxOpenConns: 1, MaxIdleConns: 1}

	applyPoolOptions(f, opts)

	if f.connMaxLifetimeCalled {
		t.Error("SetConnMaxLifetime called, want it skipped for a zero ConnMaxLifetime (0 means \"no limit\" in database/sql, not \"apply now\")")
	}
	if f.connMaxIdleTimeCalled {
		t.Error("SetConnMaxIdleTime called, want it skipped for a zero ConnMaxIdleTime")
	}
}
