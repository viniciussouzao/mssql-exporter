//go:build linux

package oscollector

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v4/disk"
)

type diskUsageCollector struct {
	total       *prometheus.Desc
	used        *prometheus.Desc
	free        *prometheus.Desc
	inodesTotal *prometheus.Desc
	inodesUsed  *prometheus.Desc
}

func newDiskUsageCollector() *diskUsageCollector {
	labels := []string{"device", "mountpoint", "fstype"}
	desc := func(name, help string) *prometheus.Desc {
		return prometheus.NewDesc(name, help, labels, nil)
	}
	return &diskUsageCollector{
		total:       desc("os_filesystem_size_bytes", "Total filesystem size in bytes"),
		used:        desc("os_filesystem_used_bytes", "Used filesystem space in bytes"),
		free:        desc("os_filesystem_free_bytes", "Free filesystem space in bytes"),
		inodesTotal: desc("os_filesystem_inodes_total", "Total number of inodes"),
		inodesUsed:  desc("os_filesystem_inodes_used", "Used number of inodes"),
	}
}

func (c *diskUsageCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.total
	ch <- c.used
	ch <- c.free
	ch <- c.inodesTotal
	ch <- c.inodesUsed
}

func (c *diskUsageCollector) Collect(ch chan<- prometheus.Metric) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		slog.Error("failed to list disk partitions", slog.Any("error", err))
		return
	}

	for _, p := range partitions {
		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			slog.Error("failed to get disk usage", slog.String("mountpoint", p.Mountpoint), slog.Any("error", err))
			continue
		}

		labels := []string{p.Device, p.Mountpoint, p.Fstype}
		gauge := func(desc *prometheus.Desc, v uint64) {
			ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, float64(v), labels...)
		}
		gauge(c.total, usage.Total)
		gauge(c.used, usage.Used)
		gauge(c.free, usage.Free)
		gauge(c.inodesTotal, usage.InodesTotal)
		gauge(c.inodesUsed, usage.InodesUsed)
	}
}
