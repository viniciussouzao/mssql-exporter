package collector

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

type SQLServerSchedulers struct {
	db            *sql.DB
	runnableTasks *prometheus.Desc
	currentTasks  *prometheus.Desc
	activeWorkers *prometheus.Desc
	workQueue     *prometheus.Desc
	online        *prometheus.Desc
}

func NewSchedulersCollector(db *sql.DB) *SQLServerSchedulers {
	return &SQLServerSchedulers{
		db: db,
		runnableTasks: prometheus.NewDesc(
			"sqlserver_schedulers_runnable_tasks",
			"Number of tasks waiting for CPU across the online visible schedulers (sustained values indicate CPU pressure)",
			nil, nil,
		),
		currentTasks: prometheus.NewDesc(
			"sqlserver_schedulers_current_tasks",
			"Number of current tasks across the online visible schedulers",
			nil, nil,
		),
		activeWorkers: prometheus.NewDesc(
			"sqlserver_schedulers_active_workers",
			"Number of active workers across the online visible schedulers",
			nil, nil,
		),
		workQueue: prometheus.NewDesc(
			"sqlserver_schedulers_work_queue",
			"Number of tasks waiting for a worker across the online visible schedulers",
			nil, nil,
		),
		online: prometheus.NewDesc(
			"sqlserver_schedulers_online",
			"Number of online visible schedulers (the ones that run user work)",
			nil, nil,
		),
	}
}

func (c *SQLServerSchedulers) Name() string {
	return "schedulers"
}

func (c *SQLServerSchedulers) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.runnableTasks
	ch <- c.currentTasks
	ch <- c.activeWorkers
	ch <- c.workQueue
	ch <- c.online
}

func (c *SQLServerSchedulers) Scrape(ctx context.Context, ch chan<- prometheus.Metric) error {
	// VISIBLE ONLINE: hidden schedulers (DAC, internal work) do not run
	// user queries, and VISIBLE OFFLINE ones are kept out by affinity or
	// licensing limits; counting them would dilute runnable tasks per CPU.
	row := c.db.QueryRowContext(ctx, `
		SELECT
			SUM(runnable_tasks_count),
			SUM(current_tasks_count),
			SUM(active_workers_count),
			SUM(work_queue_count),
			COUNT(*)
		FROM sys.dm_os_schedulers
		WHERE status = 'VISIBLE ONLINE'
	`)

	var runnable, current, workers, workQueue, online float64
	if err := row.Scan(&runnable, &current, &workers, &workQueue, &online); err != nil {
		return fmt.Errorf("failed to query dm_os_schedulers: %w", err)
	}

	ch <- prometheus.MustNewConstMetric(c.runnableTasks, prometheus.GaugeValue, runnable)
	ch <- prometheus.MustNewConstMetric(c.currentTasks, prometheus.GaugeValue, current)
	ch <- prometheus.MustNewConstMetric(c.activeWorkers, prometheus.GaugeValue, workers)
	ch <- prometheus.MustNewConstMetric(c.workQueue, prometheus.GaugeValue, workQueue)
	ch <- prometheus.MustNewConstMetric(c.online, prometheus.GaugeValue, online)

	return nil
}
