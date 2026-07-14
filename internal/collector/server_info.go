package collector

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

type SQLServerProperties struct {
	db                        *sql.DB
	hostname                  *prometheus.Desc
	servername                *prometheus.Desc
	edition                   *prometheus.Desc
	productVersion            *prometheus.Desc
	engineEdition             *prometheus.Desc
	isHadrEnabled             *prometheus.Desc
	visibleCPUs               *prometheus.Desc
	totalPhysicalMemoryKB     *prometheus.Desc
	availablePhysicalMemoryKB *prometheus.Desc
}

func NewServerInfoCollector(db *sql.DB) *SQLServerProperties {
	return &SQLServerProperties{
		db: db,
		hostname: prometheus.NewDesc(
			"sqlserver_hostname",
			"SQL Server hostname",
			[]string{"hostname"},
			nil,
		),

		servername: prometheus.NewDesc(
			"sqlserver_servername",
			"SQL Server instance name",
			[]string{"servername"},
			nil,
		),

		edition: prometheus.NewDesc(
			"sqlserver_edition",
			"SQL Server edition",
			[]string{"edition"},
			nil,
		),

		productVersion: prometheus.NewDesc(
			"sqlserver_product_version",
			"SQL Server product version",
			[]string{"product_version"},
			nil,
		),

		engineEdition: prometheus.NewDesc(
			"sqlserver_engine_edition",
			"SQL Server engine edition",
			nil,
			nil,
		),

		isHadrEnabled: prometheus.NewDesc(
			"sqlserver_is_hadr_enabled",
			"Indicates whether Always On availability groups is enabled",
			nil,
			nil,
		),

		visibleCPUs: prometheus.NewDesc(
			"sqlserver_visible_cpus",
			"Number of visible CPUs to SQL Server",
			nil,
			nil,
		),

		totalPhysicalMemoryKB: prometheus.NewDesc(
			"sqlserver_total_physical_memory_kb",
			"Total physical memory in KB available to SQL Server",
			nil,
			nil,
		),

		availablePhysicalMemoryKB: prometheus.NewDesc(
			"sqlserver_available_physical_memory_kb",
			"Available physical memory in KB to SQL Server",
			nil,
			nil,
		),
	}
}

func (c *SQLServerProperties) Name() string {
	return "server_info"
}

func (c *SQLServerProperties) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.hostname
	ch <- c.servername
	ch <- c.edition
	ch <- c.productVersion
	ch <- c.engineEdition
	ch <- c.isHadrEnabled
	ch <- c.visibleCPUs
	ch <- c.totalPhysicalMemoryKB
	ch <- c.availablePhysicalMemoryKB
}

func (c *SQLServerProperties) Scrape(ctx context.Context, ch chan<- prometheus.Metric) error {
	var errs []error

	row := c.db.QueryRowContext(ctx, `
	SELECT
		SERVERPROPERTY('MachineName') AS hostname,
		SERVERPROPERTY('ServerName') AS servername,
		SERVERPROPERTY('Edition') AS edition,
		SERVERPROPERTY('ProductVersion') AS product_version,
		SERVERPROPERTY('EngineEdition') AS engine_edition,
		SERVERPROPERTY('IsHadrEnabled') AS is_hadr_enabled
	`)

	var hostname, servername, edition, productVersion string
	var engineEdition int
	var isHadrEnabled int

	if err := row.Scan(
		&hostname,
		&servername,
		&edition,
		&productVersion,
		&engineEdition,
		&isHadrEnabled,
	); err != nil {
		errs = append(errs, fmt.Errorf("failed to query server properties: %w", err))
	} else {
		ch <- prometheus.MustNewConstMetric(c.hostname, prometheus.GaugeValue, 1, hostname)
		ch <- prometheus.MustNewConstMetric(c.servername, prometheus.GaugeValue, 1, servername)
		ch <- prometheus.MustNewConstMetric(c.edition, prometheus.GaugeValue, 1, edition)
		ch <- prometheus.MustNewConstMetric(c.productVersion, prometheus.GaugeValue, 1, productVersion)
		ch <- prometheus.MustNewConstMetric(c.engineEdition, prometheus.GaugeValue, float64(engineEdition))
		ch <- prometheus.MustNewConstMetric(c.isHadrEnabled, prometheus.GaugeValue, float64(isHadrEnabled))
	}

	row = c.db.QueryRowContext(ctx, `SELECT cpu_count FROM sys.dm_os_sys_info`)

	var visibleCPUs int
	if err := row.Scan(&visibleCPUs); err != nil {
		errs = append(errs, fmt.Errorf("failed to query dm_os_sys_info: %w", err))
	} else {
		ch <- prometheus.MustNewConstMetric(c.visibleCPUs, prometheus.GaugeValue, float64(visibleCPUs))
	}

	row = c.db.QueryRowContext(ctx, `
	SELECT total_physical_memory_kb as total_physical_memory_mb,
	       available_physical_memory_kb as available_physical_memory_mb
	FROM sys.dm_os_sys_memory
	`)

	var totalPhysicalMemoryKB, availablePhysicalMemoryKB int64
	if err := row.Scan(&totalPhysicalMemoryKB, &availablePhysicalMemoryKB); err != nil {
		errs = append(errs, fmt.Errorf("failed to query dm_os_sys_memory: %w", err))
	} else {
		ch <- prometheus.MustNewConstMetric(c.totalPhysicalMemoryKB, prometheus.GaugeValue, float64(totalPhysicalMemoryKB))
		ch <- prometheus.MustNewConstMetric(c.availablePhysicalMemoryKB, prometheus.GaugeValue, float64(availablePhysicalMemoryKB))
	}

	return errors.Join(errs...)
}
