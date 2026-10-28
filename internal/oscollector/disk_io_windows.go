//go:build windows

package oscollector

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
)

// physicalDiskWMIClass is passed explicitly to wmi.CreateQuery: without it the
// library uses the Go type name as the WMI class, which does not exist.
const physicalDiskWMIClass = "Win32_PerfRawData_PerfDisk_PhysicalDisk"

// win32RawDiskPhysicalDisk mirrors the fields we need from
// Win32_PerfRawData_PerfDisk_PhysicalDisk. Unlike the "Formatted" class,
// these counters are raw, monotonically increasing values, matching the
// cumulative counter semantics used by the Linux collector. The "_Total"
// instance is skipped so only per-disk series are exported.
type win32RawDiskPhysicalDisk struct {
	Name                 string
	DiskReadsPerSec      uint64
	DiskWritesPerSec     uint64
	DiskReadBytesPerSec  uint64
	DiskWriteBytesPerSec uint64
	// PercentDiskTime is a PERF_PRECISION_100NS_TIMER: busy time in 100ns ticks.
	PercentDiskTime        uint64
	CurrentDiskQueueLength uint64
	// AvgDiskReadQueueLength and AvgDiskWriteQueueLength are
	// PERF_COUNTER_100NS_QUEUELEN_TYPE: the queue length summed over time, in
	// 100ns ticks. Their rate is the average queue length, so they are
	// exported as cumulative seconds, like the read/write time on Linux.
	AvgDiskReadQueueLength  uint64
	AvgDiskWriteQueueLength uint64
}

const (
	// ticks100nsPerMs converts 100ns ticks to milliseconds.
	ticks100nsPerMs = 10_000
	// ticks100nsPerSecond converts 100ns ticks to seconds.
	ticks100nsPerSecond = 10_000_000
)

type diskIOCollector struct {
	reads       *prometheus.Desc
	writes      *prometheus.Desc
	readBytes   *prometheus.Desc
	writeBytes  *prometheus.Desc
	ioTime      *prometheus.Desc
	readTime    *prometheus.Desc
	writeTime   *prometheus.Desc
	queueLength *prometheus.Desc
}

func newDiskIOCollector() *diskIOCollector {
	labels := []string{"device"}
	desc := func(name, help string) *prometheus.Desc {
		return prometheus.NewDesc(name, help, labels, nil)
	}
	return &diskIOCollector{
		reads:       desc("os_disk_reads_total", "Total number of disk read operations"),
		writes:      desc("os_disk_writes_total", "Total number of disk write operations"),
		readBytes:   desc("os_disk_read_bytes_total", "Total bytes read from disk"),
		writeBytes:  desc("os_disk_write_bytes_total", "Total bytes written to disk"),
		ioTime:      desc("os_disk_io_time_ms_total", "Total time spent on disk I/O in milliseconds"),
		readTime:    desc("os_disk_read_time_seconds_total", readTimeHelp),
		writeTime:   desc("os_disk_write_time_seconds_total", writeTimeHelp),
		queueLength: desc("os_disk_queue_length", queueLengthHelp),
	}
}

func (c *diskIOCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.reads
	ch <- c.writes
	ch <- c.readBytes
	ch <- c.writeBytes
	ch <- c.ioTime
	ch <- c.readTime
	ch <- c.writeTime
	ch <- c.queueLength
}

func (c *diskIOCollector) Collect(ch chan<- prometheus.Metric) {
	var disks []win32RawDiskPhysicalDisk
	query := wmi.CreateQuery(&disks, "", physicalDiskWMIClass)
	if err := wmi.Query(query, &disks); err != nil {
		slog.Error("failed to query disk performance counters", slog.Any("error", err))
		return
	}

	for _, d := range disks {
		if d.Name == "_Total" {
			continue
		}

		counter := func(desc *prometheus.Desc, v uint64) {
			ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, float64(v), d.Name)
		}
		counter(c.reads, d.DiskReadsPerSec)
		counter(c.writes, d.DiskWritesPerSec)
		counter(c.readBytes, d.DiskReadBytesPerSec)
		counter(c.writeBytes, d.DiskWriteBytesPerSec)

		ioTimeMs := float64(d.PercentDiskTime) / ticks100nsPerMs
		ch <- prometheus.MustNewConstMetric(c.ioTime, prometheus.CounterValue, ioTimeMs, d.Name)
		ch <- prometheus.MustNewConstMetric(c.readTime, prometheus.CounterValue, float64(d.AvgDiskReadQueueLength)/ticks100nsPerSecond, d.Name)
		ch <- prometheus.MustNewConstMetric(c.writeTime, prometheus.CounterValue, float64(d.AvgDiskWriteQueueLength)/ticks100nsPerSecond, d.Name)
		ch <- prometheus.MustNewConstMetric(c.queueLength, prometheus.GaugeValue, float64(d.CurrentDiskQueueLength), d.Name)
	}
}
