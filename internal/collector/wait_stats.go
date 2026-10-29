package collector

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// idleWaitTypes are background and idle waits: they accumulate while SQL
// Server has nothing to do (lazy writer, checkpoint, Always On and mirroring
// queues, Query Store and trace flushes, ...) and do not slow any query, but
// they dominate any "top waits" view (on Azure SQL Managed Instance the top
// waits are all of this kind, HML-039). Based on the commonly used benign
// wait list; waits that do slow queries, such as HADR_SYNC_COMMIT, are kept.
// SLEEP%, XE% and BROKER_% are always excluded by prefix.
var idleWaitTypes = []string{
	"CHECKPOINT_QUEUE", "CHKPT", "CLR_AUTO_EVENT", "CLR_MANUAL_EVENT", "CLR_SEMAPHORE",
	"DBMIRROR_DBM_EVENT", "DBMIRROR_EVENTS_QUEUE", "DBMIRROR_WORKER_QUEUE", "DBMIRRORING_CMD",
	"DIRTY_PAGE_POLL", "DISPATCHER_QUEUE_SEMAPHORE", "EXECSYNC", "FSAGENT",
	"FT_IFTS_SCHEDULER_IDLE_WAIT", "FT_IFTSHC_MUTEX",
	"HADR_CLUSAPI_CALL", "HADR_FILESTREAM_IOMGR_IOCOMPLETION", "HADR_LOGCAPTURE_WAIT",
	"HADR_NOTIFICATION_DEQUEUE", "HADR_TIMER_TASK", "HADR_WORK_QUEUE",
	"KSOURCE_WAKEUP", "LAZYWRITER_SLEEP", "LOGMGR_QUEUE", "MEMORY_ALLOCATION_EXT", "ONDEMAND_TASK_QUEUE",
	"PARALLEL_REDO_DRAIN_WORKER", "PARALLEL_REDO_LOG_CACHE", "PARALLEL_REDO_TRAN_LIST",
	"PARALLEL_REDO_WORKER_SYNC", "PARALLEL_REDO_WORKER_WAIT_WORK",
	"PREEMPTIVE_OS_FLUSHFILEBUFFERS", "PREEMPTIVE_XE_GETTARGETSTATE", "PVS_PREALLOCATE",
	"PWAIT_ALL_COMPONENTS_INITIALIZED", "PWAIT_DIRECTLOGCONSUMER_GETNEXT", "PWAIT_EXTENSIBILITY_CLEANUP_TASK",
	"QDS_ASYNC_QUEUE", "QDS_CLEANUP_STALE_QUERIES_TASK_MAIN_LOOP_SLEEP",
	"QDS_PERSIST_TASK_MAIN_LOOP_SLEEP", "QDS_SHUTDOWN_QUEUE",
	"REDO_THREAD_PENDING_WORK", "REQUEST_FOR_DEADLOCK_SEARCH", "RESOURCE_QUEUE", "SERVER_IDLE_CHECK",
	"SNI_HTTP_ACCEPT", "SOS_WORK_DISPATCHER", "SP_SERVER_DIAGNOSTICS_SLEEP",
	"SQLTRACE_BUFFER_FLUSH", "SQLTRACE_INCREMENTAL_FLUSH_SLEEP", "SQLTRACE_WAIT_ENTRIES",
	"STARTUP_DEPENDENCY_MANAGER", "UCS_SESSION_REGISTRATION", "VDI_CLIENT_OTHER",
	"WAIT_FOR_RESULTS", "WAITFOR", "WAITFOR_TASKSHUTDOWN",
	"WAIT_XTP_CKPT_CLOSE", "WAIT_XTP_HOST_WAIT", "WAIT_XTP_OFFLINE_CKPT_NEW_LOG", "WAIT_XTP_RECOVERY",
}

// waitStatsQuery builds the query from static strings only. Unless idle
// waits are included, it also skips the idle wait types and the wait types
// that never happened (most of the ~1,000 rows of the DMV).
func waitStatsQuery(opts WaitStatsOptions) string {
	query := `
	SELECT wait_type, wait_time_ms, waiting_tasks_count, max_wait_time_ms
	FROM sys.dm_os_wait_stats
	WHERE wait_type NOT LIKE 'SLEEP%'
	AND wait_type NOT LIKE 'XE%'
	AND wait_type NOT LIKE 'BROKER_%'`
	if opts.IncludeIdle {
		return query
	}
	return query + `
	AND waiting_tasks_count > 0
	AND wait_type NOT IN (N'` + strings.Join(idleWaitTypes, "', N'") + `')`
}

type SQLServerWaitStats struct {
	db            *sql.DB
	query         string
	waitTimeMs    *prometheus.Desc
	waitingTasks  *prometheus.Desc
	maxWaitTimeMs *prometheus.Desc
}

func NewWaitStatsCollector(db *sql.DB, opts WaitStatsOptions) *SQLServerWaitStats {
	return &SQLServerWaitStats{
		db:    db,
		query: waitStatsQuery(opts),
		waitTimeMs: prometheus.NewDesc(
			"sqlserver_wait_time_ms_total",
			"Wait time in milliseconds by wait type since SQL Server started (or the wait statistics were cleared)",
			[]string{"wait_type"},
			nil,
		),

		waitingTasks: prometheus.NewDesc(
			"sqlserver_waiting_tasks_total",
			"Waits by wait type since SQL Server started (or the wait statistics were cleared)",
			[]string{"wait_type"},
			nil,
		),

		maxWaitTimeMs: prometheus.NewDesc(
			"sqlserver_max_wait_time_ms",
			"Longest single wait in milliseconds by wait type since SQL Server started (or the wait statistics were cleared)",
			[]string{"wait_type"},
			nil,
		),
	}
}

func (c *SQLServerWaitStats) Name() string {
	return "wait_stats"
}

func (c *SQLServerWaitStats) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.waitTimeMs
	ch <- c.waitingTasks
	ch <- c.maxWaitTimeMs
}

func (c *SQLServerWaitStats) Scrape(ctx context.Context, ch chan<- prometheus.Metric) error {
	rows, err := c.db.QueryContext(ctx, c.query)

	if err != nil {
		return fmt.Errorf("failed to query dm_os_wait_stats: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var scanErrs []error
	for rows.Next() {
		var waitType string
		var waitTimeMs float64
		var waitingTasks float64
		var maxWaitTimeMs float64

		err := rows.Scan(&waitType, &waitTimeMs, &waitingTasks, &maxWaitTimeMs)
		if err != nil {
			scanErrs = append(scanErrs, fmt.Errorf("failed to scan wait stats row: %w", err))
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.waitTimeMs,
			prometheus.CounterValue,
			waitTimeMs,
			waitType,
		)

		ch <- prometheus.MustNewConstMetric(
			c.waitingTasks,
			prometheus.CounterValue,
			waitingTasks,
			waitType,
		)

		ch <- prometheus.MustNewConstMetric(
			c.maxWaitTimeMs,
			prometheus.GaugeValue,
			maxWaitTimeMs,
			waitType,
		)
	}

	return errors.Join(append(scanErrs, rows.Err())...)
}
