//go:build linux

package oscollector

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v4/disk"
)

type diskIOCollector struct {
	reads      *prometheus.Desc
	writes     *prometheus.Desc
	readBytes  *prometheus.Desc
	writeBytes *prometheus.Desc
	ioTime     *prometheus.Desc
}

func newDiskIOCollector() *diskIOCollector {
	labels := []string{"device"}
	desc := func(name, help string) *prometheus.Desc {
		return prometheus.NewDesc(name, help, labels, nil)
	}
	return &diskIOCollector{
		reads:      desc("os_disk_reads_total", "Total number of disk read operations"),
		writes:     desc("os_disk_writes_total", "Total number of disk write operations"),
		readBytes:  desc("os_disk_read_bytes_total", "Total bytes read from disk"),
		writeBytes: desc("os_disk_write_bytes_total", "Total bytes written to disk"),
		ioTime:     desc("os_disk_io_time_ms_total", "Total time spent on disk I/O in milliseconds"),
	}
}

func (c *diskIOCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.reads
	ch <- c.writes
	ch <- c.readBytes
	ch <- c.writeBytes
	ch <- c.ioTime
}

func (c *diskIOCollector) Collect(ch chan<- prometheus.Metric) {
	counters, err := disk.IOCounters()
	if err != nil {
		slog.Error("failed to collect disk I/O metrics", slog.Any("error", err))
		return
	}

	for device, stat := range counters {
		counter := func(desc *prometheus.Desc, v uint64) {
			ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, float64(v), device)
		}
		counter(c.reads, stat.ReadCount)
		counter(c.writes, stat.WriteCount)
		counter(c.readBytes, stat.ReadBytes)
		counter(c.writeBytes, stat.WriteBytes)
		counter(c.ioTime, stat.IoTime)
	}
}
