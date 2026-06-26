//go:build linux

package oscollector

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v4/mem"
)

type memoryCollector struct {
	total     *prometheus.Desc
	used      *prometheus.Desc
	free      *prometheus.Desc
	available *prometheus.Desc
	cached    *prometheus.Desc
	buffers   *prometheus.Desc
	swapTotal *prometheus.Desc
	swapUsed  *prometheus.Desc
	swapFree  *prometheus.Desc
}

func newMemoryCollector() *memoryCollector {
	desc := func(name, help string) *prometheus.Desc {
		return prometheus.NewDesc(name, help, nil, nil)
	}
	return &memoryCollector{
		total:     desc("os_memory_total_bytes", "Total physical memory in bytes"),
		used:      desc("os_memory_used_bytes", "Used physical memory in bytes"),
		free:      desc("os_memory_free_bytes", "Free physical memory in bytes"),
		available: desc("os_memory_available_bytes", "Available physical memory in bytes (free + reclaimable)"),
		cached:    desc("os_memory_cached_bytes", "Cached memory in bytes"),
		buffers:   desc("os_memory_buffers_bytes", "Buffered memory in bytes"),
		swapTotal: desc("os_memory_swap_total_bytes", "Total swap space in bytes"),
		swapUsed:  desc("os_memory_swap_used_bytes", "Used swap space in bytes"),
		swapFree:  desc("os_memory_swap_free_bytes", "Free swap space in bytes"),
	}
}

func (c *memoryCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.total
	ch <- c.used
	ch <- c.free
	ch <- c.available
	ch <- c.cached
	ch <- c.buffers
	ch <- c.swapTotal
	ch <- c.swapUsed
	ch <- c.swapFree
}

func (c *memoryCollector) Collect(ch chan<- prometheus.Metric) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		slog.Error("failed to collect virtual memory metrics", slog.Any("error", err))
	} else {
		g := func(desc *prometheus.Desc, v uint64) {
			ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, float64(v))
		}
		g(c.total, vm.Total)
		g(c.used, vm.Used)
		g(c.free, vm.Free)
		g(c.available, vm.Available)
		g(c.cached, vm.Cached)
		g(c.buffers, vm.Buffers)
	}

	swap, err := mem.SwapMemory()
	if err != nil {
		slog.Error("failed to collect swap memory metrics", slog.Any("error", err))
		return
	}
	g := func(desc *prometheus.Desc, v uint64) {
		ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, float64(v))
	}
	g(c.swapTotal, swap.Total)
	g(c.swapUsed, swap.Used)
	g(c.swapFree, swap.Free)
}
