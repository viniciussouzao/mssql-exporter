package collector

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// SQLServerQueryStore reports the health of Query Store in every accessible
// user database: whether it is still capturing (it silently switches to
// READ_ONLY when it fills up), how much of its storage is used and whether
// forced plans keep failing. Query Store is per database and its catalog
// views get slow on large stores, so like index_fragmentation it collects in
// the background on its own interval (see QueryStoreOptions) and Scrape
// only serves the last result. Disabled by default: it needs access to each
// database. No per-query series (see ROADMAP).
type SQLServerQueryStore struct {
	db   *sql.DB
	opts QueryStoreOptions
	now  func() time.Time

	startOnce sync.Once

	mu   sync.Mutex
	last *queryStoreResult // nil until the first collection ends

	actualState     *prometheus.Desc
	desiredState    *prometheus.Desc
	readonlyReason  *prometheus.Desc
	storageUsed     *prometheus.Desc
	storageMax      *prometheus.Desc
	forcedPlans     *prometheus.Desc
	forcedFailing   *prometheus.Desc
	lastTimestamp   *prometheus.Desc
	lastDuration    *prometheus.Desc
	lastSuccess     *prometheus.Desc
	lastDatabaseCnt *prometheus.Desc
}

type queryStoreDatabase struct {
	database                  string
	actualState, desiredState string
	readonlyReason            sql.NullFloat64
	storageUsedMB             sql.NullFloat64
	storageMaxMB              sql.NullFloat64
	forcedPlans               float64
	forcedPlansFailing        float64
}

type queryStoreResult struct {
	databases []queryStoreDatabase
	finished  time.Time
	duration  time.Duration
	err       error
}

func NewQueryStoreCollector(db *sql.DB, opts QueryStoreOptions) *SQLServerQueryStore {
	return &SQLServerQueryStore{
		db:   db,
		opts: opts,
		now:  time.Now,
		actualState: prometheus.NewDesc(
			"sqlserver_query_store_actual_state",
			"Current Query Store state of the database (1 for the current state, labeled by state: OFF, READ_ONLY, READ_WRITE, ERROR)",
			[]string{"db_name", "state"}, nil,
		),
		desiredState: prometheus.NewDesc(
			"sqlserver_query_store_desired_state",
			"Query Store state configured for the database (1 for the configured state, labeled by state)",
			[]string{"db_name", "state"}, nil,
		),
		readonlyReason: prometheus.NewDesc(
			"sqlserver_query_store_readonly_reason",
			"Bitmask of the reasons Query Store is READ_ONLY (0 = not read-only; 1 = database read-only, 2 = single user, 4 = emergency, 8 = secondary replica, 65536 = max storage size reached, 131072 = statement limit, 262144 = in-memory limit, 524288 = database disk size limit)",
			[]string{"db_name"}, nil,
		),
		storageUsed: prometheus.NewDesc(
			"sqlserver_query_store_storage_used_bytes",
			"Space used by Query Store in the database, in bytes",
			[]string{"db_name"}, nil,
		),
		storageMax: prometheus.NewDesc(
			"sqlserver_query_store_storage_max_bytes",
			"Maximum space Query Store may use in the database (MAX_STORAGE_SIZE_MB), in bytes",
			[]string{"db_name"}, nil,
		),
		forcedPlans: prometheus.NewDesc(
			"sqlserver_query_store_forced_plans",
			"Number of plans forced in Query Store",
			[]string{"db_name"}, nil,
		),
		forcedFailing: prometheus.NewDesc(
			"sqlserver_query_store_forced_plans_failing",
			"Number of forced plans whose forcing has failed since they were forced (force_failure_count > 0)",
			[]string{"db_name"}, nil,
		),
		lastTimestamp: prometheus.NewDesc(
			"sqlserver_query_store_last_collection_timestamp_seconds",
			"Unix timestamp of the end of the last background Query Store collection",
			nil, nil,
		),
		lastDuration: prometheus.NewDesc(
			"sqlserver_query_store_last_collection_duration_seconds",
			"Duration of the last background Query Store collection",
			nil, nil,
		),
		lastSuccess: prometheus.NewDesc(
			"sqlserver_query_store_last_collection_success",
			"Whether the last background Query Store collection succeeded for every database (1) or not (0); failures are logged",
			nil, nil,
		),
		lastDatabaseCnt: prometheus.NewDesc(
			"sqlserver_query_store_last_collection_databases",
			"Number of databases read by the last background Query Store collection",
			nil, nil,
		),
	}
}

func (c *SQLServerQueryStore) Name() string {
	return "query_store"
}

func (c *SQLServerQueryStore) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.actualState
	ch <- c.desiredState
	ch <- c.readonlyReason
	ch <- c.storageUsed
	ch <- c.storageMax
	ch <- c.forcedPlans
	ch <- c.forcedFailing
	ch <- c.lastTimestamp
	ch <- c.lastDuration
	ch <- c.lastSuccess
	ch <- c.lastDatabaseCnt
}

// Scrape starts the background loop on first use and emits the result of
// the last finished collection (nothing before the first one ends). It never
// fails: a failed background collection is reported by
// sqlserver_query_store_last_collection_success and logged.
func (c *SQLServerQueryStore) Scrape(_ context.Context, ch chan<- prometheus.Metric) error {
	c.startOnce.Do(func() { go c.loop() })

	c.mu.Lock()
	last := c.last
	c.mu.Unlock()
	if last == nil {
		return nil
	}

	const mb = 1024 * 1024
	for _, d := range last.databases {
		ch <- prometheus.MustNewConstMetric(c.actualState, prometheus.GaugeValue, 1, d.database, d.actualState)
		ch <- prometheus.MustNewConstMetric(c.desiredState, prometheus.GaugeValue, 1, d.database, d.desiredState)
		if d.readonlyReason.Valid {
			ch <- prometheus.MustNewConstMetric(c.readonlyReason, prometheus.GaugeValue, d.readonlyReason.Float64, d.database)
		}
		if d.storageUsedMB.Valid {
			ch <- prometheus.MustNewConstMetric(c.storageUsed, prometheus.GaugeValue, d.storageUsedMB.Float64*mb, d.database)
		}
		if d.storageMaxMB.Valid {
			ch <- prometheus.MustNewConstMetric(c.storageMax, prometheus.GaugeValue, d.storageMaxMB.Float64*mb, d.database)
		}
		ch <- prometheus.MustNewConstMetric(c.forcedPlans, prometheus.GaugeValue, d.forcedPlans, d.database)
		ch <- prometheus.MustNewConstMetric(c.forcedFailing, prometheus.GaugeValue, d.forcedPlansFailing, d.database)
	}
	success := 1.0
	if last.err != nil {
		success = 0
	}
	ch <- prometheus.MustNewConstMetric(c.lastTimestamp, prometheus.GaugeValue, float64(last.finished.UnixMilli())/1000)
	ch <- prometheus.MustNewConstMetric(c.lastDuration, prometheus.GaugeValue, last.duration.Seconds())
	ch <- prometheus.MustNewConstMetric(c.lastSuccess, prometheus.GaugeValue, success)
	ch <- prometheus.MustNewConstMetric(c.lastDatabaseCnt, prometheus.GaugeValue, float64(len(last.databases)))
	return nil
}

// loop collects right away, then every Interval. Each collection is bounded
// by the interval, so collections never overlap.
func (c *SQLServerQueryStore) loop() {
	ticker := time.NewTicker(c.opts.Interval)
	defer ticker.Stop()
	for {
		ctx, cancel := context.WithTimeout(context.Background(), c.opts.Interval)
		c.refresh(ctx)
		cancel()
		<-ticker.C
	}
}

// refresh runs one collection and publishes its result.
func (c *SQLServerQueryStore) refresh(ctx context.Context) {
	start := c.now()
	res := &queryStoreResult{}

	databases, err := listAccessibleDatabases(ctx, c.db, c.opts.Databases)
	if err != nil {
		res.err = err
	}
	var dbErrs []error
	for _, name := range databases {
		d, err := c.collectDatabase(ctx, name)
		if err != nil {
			dbErrs = append(dbErrs, fmt.Errorf("database %s: %w", name, err))
			continue
		}
		res.databases = append(res.databases, d)
	}
	res.err = errors.Join(append([]error{res.err}, dbErrs...)...)
	res.finished = c.now()
	res.duration = res.finished.Sub(start)

	if res.err != nil {
		slog.Warn("query store collection failed", slog.String("collector", c.Name()), slog.Any("error", res.err))
	}

	c.mu.Lock()
	c.last = res
	c.mu.Unlock()
}

// collectDatabase reads one database's Query Store options and forced plans.
// Both views are per database, so they are referenced with three-part names
// (the name comes from sys.databases and is bracket-quoted). The views exist
// since SQL Server 2016; on older versions every database fails.
func (c *SQLServerQueryStore) collectDatabase(ctx context.Context, name string) (queryStoreDatabase, error) {
	db := quoteIdentifier(name)
	d := queryStoreDatabase{database: name}
	err := c.db.QueryRowContext(ctx, `
		SELECT
			o.actual_state_desc,
			o.desired_state_desc,
			o.readonly_reason,
			o.current_storage_size_mb,
			o.max_storage_size_mb,
			p.forced_plans,
			p.forced_plans_failing
		FROM `+db+`.sys.database_query_store_options o
		CROSS JOIN (
			SELECT
				COUNT(*) AS forced_plans,
				COUNT(CASE WHEN force_failure_count > 0 THEN 1 END) AS forced_plans_failing
			FROM `+db+`.sys.query_store_plan
			WHERE is_forced_plan = 1
		) p
	`).Scan(&d.actualState, &d.desiredState, &d.readonlyReason, &d.storageUsedMB, &d.storageMaxMB,
		&d.forcedPlans, &d.forcedPlansFailing)
	if err != nil {
		return d, fmt.Errorf("failed to query database_query_store_options: %w", err)
	}
	return d, nil
}
