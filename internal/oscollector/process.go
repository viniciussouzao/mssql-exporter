//go:build linux

package oscollector

import (
	"log/slog"
	"slices"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v4/process"
)

// sqlServerProcessNames are the process image names (case-insensitive,
// without extension) that identify a SQL Server engine process.
var sqlServerProcessNames = []string{"sqlservr"}

type processCollector struct {
	cpuSeconds *prometheus.Desc
	memoryRSS  *prometheus.Desc
	openFDs    *prometheus.Desc
}

func newProcessCollector() *processCollector {
	labels := []string{"process", "pid"}
	desc := func(name, help string) *prometheus.Desc {
		return prometheus.NewDesc(name, help, labels, nil)
	}
	return &processCollector{
		cpuSeconds: desc("os_process_cpu_seconds_total", "Total user and system CPU time spent by the process in seconds"),
		memoryRSS:  desc("os_process_memory_rss_bytes", "Resident set size (physical memory in use) of the process in bytes"),
		openFDs:    desc("os_process_open_fds", "Number of open file descriptors (or handles on Windows) held by the process"),
	}
}

func (c *processCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.cpuSeconds
	ch <- c.memoryRSS
	ch <- c.openFDs
}

func (c *processCollector) Collect(ch chan<- prometheus.Metric) {
	procs, err := process.Processes()
	if err != nil {
		slog.Error("failed to list processes", slog.Any("error", err))
		return
	}

	for _, p := range procs {
		name, err := p.Name()
		if err != nil || !isSQLServerProcess(name) {
			continue
		}

		pid := strconv.Itoa(int(p.Pid))

		if times, err := p.Times(); err != nil {
			slog.Error("failed to collect process cpu times", slog.Int("pid", int(p.Pid)), slog.Any("error", err))
		} else {
			ch <- prometheus.MustNewConstMetric(c.cpuSeconds, prometheus.CounterValue, times.User+times.System, name, pid)
		}

		if mem, err := p.MemoryInfo(); err != nil {
			slog.Error("failed to collect process memory info", slog.Int("pid", int(p.Pid)), slog.Any("error", err))
		} else {
			ch <- prometheus.MustNewConstMetric(c.memoryRSS, prometheus.GaugeValue, float64(mem.RSS), name, pid)
		}

		if fds, err := p.NumFDs(); err != nil {
			slog.Error("failed to collect process open fd count", slog.Int("pid", int(p.Pid)), slog.Any("error", err))
		} else {
			ch <- prometheus.MustNewConstMetric(c.openFDs, prometheus.GaugeValue, float64(fds), name, pid)
		}
	}
}

func isSQLServerProcess(name string) bool {
	name = strings.ToLower(strings.TrimSuffix(name, ".exe"))
	return slices.Contains(sqlServerProcessNames, name)
}
