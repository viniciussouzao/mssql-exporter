//go:build windows

package oscollector

import (
	"fmt"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v4/cpu"
)

type cpuCollector struct {
	usagePercent *prometheus.Desc
}

func newCPUCollector() *cpuCollector {
	return &cpuCollector{
		usagePercent: prometheus.NewDesc(
			"os_cpu_usage_percent",
			"CPU usage percentage per core (0-100)",
			[]string{"cpu"},
			nil,
		),
	}
}

func (c *cpuCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.usagePercent
}

func (c *cpuCollector) Collect(ch chan<- prometheus.Metric) {
	perCore, err := cpu.Percent(0, true)
	if err != nil {
		slog.Error("failed to collect per-core cpu usage", slog.Any("error", err))
		return
	}

	var total float64
	for i, p := range perCore {
		ch <- prometheus.MustNewConstMetric(c.usagePercent, prometheus.GaugeValue, p, fmt.Sprintf("cpu%d", i))
		total += p
	}

	if len(perCore) > 0 {
		ch <- prometheus.MustNewConstMetric(c.usagePercent, prometheus.GaugeValue, total/float64(len(perCore)), "total")
	}
}
