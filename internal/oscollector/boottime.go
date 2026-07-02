//go:build linux

package oscollector

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v4/host"
)

type bootTimeCollector struct {
	bootTime *prometheus.Desc
}

func newBootTimeCollector() *bootTimeCollector {
	return &bootTimeCollector{
		bootTime: prometheus.NewDesc(
			"os_boot_time_seconds",
			"Unix timestamp (seconds) of the last host boot",
			nil,
			nil,
		),
	}
}

func (c *bootTimeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.bootTime
}

func (c *bootTimeCollector) Collect(ch chan<- prometheus.Metric) {
	boot, err := host.BootTime()
	if err != nil {
		slog.Error("failed to collect host boot time", slog.Any("error", err))
		return
	}
	ch <- prometheus.MustNewConstMetric(c.bootTime, prometheus.GaugeValue, float64(boot))
}
