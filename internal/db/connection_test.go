package db

import (
	"testing"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
)

// withActualOpts snapshots the package-level actualOpts before a test
// mutates it and returns a func restoring it, so tests don't leak state
// into each other (important once t.Parallel is used, and cheap insurance
// even without it).
func withActualOpts() func() {
	prev := actualOpts
	return func() { actualOpts = prev }
}

// withDB snapshots the package-level db before a test mutates it, closes
// whatever connection the test leaves behind, and restores the previous
// value.
func withDB() func() {
	prev := db
	return func() {
		if db != nil && db != prev {
			_ = db.Close()
		}
		db = prev
	}
}

// The go-mssqldb "sqlserver" driver's Connector type does not satisfy
// database/sql/driver.DriverContext (it returns its own *Connector rather
// than the driver.Connector interface), so sql.Open never validates the DSN
// eagerly here — it only fails later, on the first dial. Connect must not
// paper over that by rejecting anything itself.
func TestConnectNeverDialsEvenWithGarbageDSN(t *testing.T) {
	t.Cleanup(withDB())
	t.Cleanup(withActualOpts())

	if err := Connect("not a valid dsn at all", DefaultOptions()); err != nil {
		t.Errorf("Connect() error = %v, want nil (Open never dials, so it can't fail on DSN content)", err)
	}
}

func TestConnectStoresOptionsForCurrentOptions(t *testing.T) {
	t.Cleanup(withDB())
	t.Cleanup(withActualOpts())

	opts := &Options{
		MaxOpenConns:    3,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Minute,
		ConnMaxIdleTime: 30 * time.Second,
		ConnTimeout:     time.Second,
	}

	if err := Connect("sqlserver://user:pass@localhost:1433?database=test", opts); err != nil {
		t.Fatalf("Connect() error = %v, want nil (Open only validates the DSN, it does not dial)", err)
	}

	if CurrentOptions() != opts {
		t.Errorf("CurrentOptions() = %v, want the exact options passed to Connect", CurrentOptions())
	}
	// The setters themselves are covered by TestApplyPoolOptionsSetsEverySetter;
	// *sql.DB only exposes MaxOpenConnections back via Stats(), so that's the
	// one setter worth re-checking end-to-end through the real Connect path.
	if stats := DB().Stats(); stats.MaxOpenConnections != opts.MaxOpenConns {
		t.Errorf("MaxOpenConnections = %d, want %d", stats.MaxOpenConnections, opts.MaxOpenConns)
	}
}

func TestDBReturnsConnectedInstance(t *testing.T) {
	t.Cleanup(withDB())
	t.Cleanup(withActualOpts())

	if err := Connect("sqlserver://user:pass@localhost:1433?database=test", DefaultOptions()); err != nil {
		t.Fatalf("Connect() error = %v, want nil", err)
	}

	if DB() == nil {
		t.Error("DB() = nil, want the connection established by Connect")
	}
}
