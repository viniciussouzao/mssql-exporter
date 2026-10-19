package collector

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// SQLServerIndexFragmentation reports logical fragmentation and size of the
// indexes of every accessible user database. sys.dm_db_index_physical_stats
// reads the index pages, which takes minutes on large databases, so unlike
// the other collectors it runs in the background on its own interval (see
// IndexFragmentationOptions) and Scrape only serves the last result.
// Disabled by default: even in LIMITED mode it generates I/O on big objects.
type SQLServerIndexFragmentation struct {
	db   *sql.DB
	opts IndexFragmentationOptions
	now  func() time.Time

	startOnce sync.Once

	mu   sync.Mutex
	last *fragmentationResult // nil until the first collection ends

	fragmentation   *prometheus.Desc
	pageCount       *prometheus.Desc
	lastTimestamp   *prometheus.Desc
	lastDuration    *prometheus.Desc
	lastSuccess     *prometheus.Desc
	lastDatabaseCnt *prometheus.Desc
}

type indexFragmentation struct {
	database, schema, table, index string
	fragmentationPercent, pages    float64
}

type fragmentationResult struct {
	indexes   []indexFragmentation
	finished  time.Time
	duration  time.Duration
	databases int
	err       error
}

func NewIndexFragmentationCollector(db *sql.DB, opts IndexFragmentationOptions) *SQLServerIndexFragmentation {
	labels := []string{"db_name", "schema_name", "table_name", "index_name"}
	return &SQLServerIndexFragmentation{
		db:   db,
		opts: opts,
		now:  time.Now,
		fragmentation: prometheus.NewDesc(
			"sqlserver_index_fragmentation_percent",
			"Logical fragmentation of the index leaf level (LIMITED mode), weighted by page count across partitions; index_name is empty for heaps",
			labels, nil,
		),
		pageCount: prometheus.NewDesc(
			"sqlserver_index_page_count",
			"Number of 8 KB pages of the index leaf level (in-row data)",
			labels, nil,
		),
		lastTimestamp: prometheus.NewDesc(
			"sqlserver_index_fragmentation_last_collection_timestamp_seconds",
			"Unix timestamp of the end of the last background index fragmentation collection",
			nil, nil,
		),
		lastDuration: prometheus.NewDesc(
			"sqlserver_index_fragmentation_last_collection_duration_seconds",
			"Duration of the last background index fragmentation collection",
			nil, nil,
		),
		lastSuccess: prometheus.NewDesc(
			"sqlserver_index_fragmentation_last_collection_success",
			"Whether the last background index fragmentation collection succeeded for every database (1) or not (0); failures are logged",
			nil, nil,
		),
		lastDatabaseCnt: prometheus.NewDesc(
			"sqlserver_index_fragmentation_last_collection_databases",
			"Number of databases whose indexes were read by the last background collection",
			nil, nil,
		),
	}
}

func (c *SQLServerIndexFragmentation) Name() string {
	return "index_fragmentation"
}

func (c *SQLServerIndexFragmentation) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.fragmentation
	ch <- c.pageCount
	ch <- c.lastTimestamp
	ch <- c.lastDuration
	ch <- c.lastSuccess
	ch <- c.lastDatabaseCnt
}

// Scrape starts the background loop on first use and emits the result of
// the last finished collection (nothing before the first one ends). It never
// fails: a failed background collection is reported by
// sqlserver_index_fragmentation_last_collection_success and logged once,
// instead of failing every scrape until the next collection hours later.
func (c *SQLServerIndexFragmentation) Scrape(_ context.Context, ch chan<- prometheus.Metric) error {
	c.startOnce.Do(func() { go c.loop() })

	c.mu.Lock()
	last := c.last
	c.mu.Unlock()
	if last == nil {
		return nil
	}

	for _, ix := range last.indexes {
		labels := []string{ix.database, ix.schema, ix.table, ix.index}
		ch <- prometheus.MustNewConstMetric(c.fragmentation, prometheus.GaugeValue, ix.fragmentationPercent, labels...)
		ch <- prometheus.MustNewConstMetric(c.pageCount, prometheus.GaugeValue, ix.pages, labels...)
	}
	success := 1.0
	if last.err != nil {
		success = 0
	}
	ch <- prometheus.MustNewConstMetric(c.lastTimestamp, prometheus.GaugeValue, float64(last.finished.UnixMilli())/1000)
	ch <- prometheus.MustNewConstMetric(c.lastDuration, prometheus.GaugeValue, last.duration.Seconds())
	ch <- prometheus.MustNewConstMetric(c.lastSuccess, prometheus.GaugeValue, success)
	ch <- prometheus.MustNewConstMetric(c.lastDatabaseCnt, prometheus.GaugeValue, float64(last.databases))
	return nil
}

// loop collects right away, then every Interval. Each collection is bounded
// by the interval, so collections never overlap.
func (c *SQLServerIndexFragmentation) loop() {
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
func (c *SQLServerIndexFragmentation) refresh(ctx context.Context) {
	start := c.now()
	res := &fragmentationResult{}

	databases, err := listAccessibleDatabases(ctx, c.db, c.opts.Databases)
	if err != nil {
		res.err = err
	}
	var dbErrs []error
	for _, name := range databases {
		indexes, err := c.collectDatabase(ctx, name)
		if err != nil {
			dbErrs = append(dbErrs, fmt.Errorf("database %s: %w", name, err))
			continue
		}
		res.indexes = append(res.indexes, indexes...)
		res.databases++
	}
	res.err = errors.Join(append([]error{res.err}, dbErrs...)...)
	res.finished = c.now()
	res.duration = res.finished.Sub(start)

	if res.err != nil {
		slog.Warn("index fragmentation collection failed", slog.String("collector", c.Name()), slog.Any("error", res.err))
	}

	c.mu.Lock()
	c.last = res
	c.mu.Unlock()
}

// collectDatabase reads one database. The catalog views are referenced with
// three-part names because dm_db_index_physical_stats takes the database as
// a parameter but sys.indexes/objects/schemas are per database; the name
// comes from sys.databases and is bracket-quoted. Partitions are combined
// per index (page-weighted fragmentation) so each index is one series; a
// heap has no name and gets an empty index_name, which no named index can
// collide with (a placeholder like "HEAP" could, and a duplicate series
// fails the whole /metrics response).
func (c *SQLServerIndexFragmentation) collectDatabase(ctx context.Context, name string) ([]indexFragmentation, error) {
	db := quoteIdentifier(name)
	rows, err := c.db.QueryContext(ctx, `
		SELECT
			s.name,
			o.name,
			ISNULL(i.name, ''),
			SUM(ps.avg_fragmentation_in_percent * ps.page_count) / NULLIF(SUM(ps.page_count), 0),
			SUM(ps.page_count)
		FROM sys.dm_db_index_physical_stats(DB_ID(@p1), NULL, NULL, NULL, 'LIMITED') ps
		INNER JOIN `+db+`.sys.indexes i
			ON i.object_id = ps.object_id AND i.index_id = ps.index_id
		INNER JOIN `+db+`.sys.objects o
			ON o.object_id = ps.object_id
		INNER JOIN `+db+`.sys.schemas s
			ON s.schema_id = o.schema_id
		WHERE ps.alloc_unit_type_desc = 'IN_ROW_DATA'
			AND o.is_ms_shipped = 0
		GROUP BY s.name, o.name, i.name
		HAVING SUM(ps.page_count) >= @p2
	`, name, c.opts.MinPageCount)
	if err != nil {
		return nil, fmt.Errorf("failed to query dm_db_index_physical_stats: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var indexes []indexFragmentation
	var scanErrs []error
	for rows.Next() {
		ix := indexFragmentation{database: name}
		var fragmentation sql.NullFloat64
		if err := rows.Scan(&ix.schema, &ix.table, &ix.index, &fragmentation, &ix.pages); err != nil {
			scanErrs = append(scanErrs, fmt.Errorf("failed to scan index fragmentation row: %w", err))
			continue
		}
		ix.fragmentationPercent = fragmentation.Float64
		indexes = append(indexes, ix)
	}
	return indexes, errors.Join(append(scanErrs, rows.Err())...)
}

// quoteIdentifier bracket-quotes a SQL Server identifier, like QUOTENAME.
func quoteIdentifier(name string) string {
	return "[" + strings.ReplaceAll(name, "]", "]]") + "]"
}
