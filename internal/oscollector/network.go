//go:build linux

package oscollector

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	gopsnet "github.com/shirou/gopsutil/v4/net"
)

type networkCollector struct {
	rxBytes   *prometheus.Desc
	txBytes   *prometheus.Desc
	rxPackets *prometheus.Desc
	txPackets *prometheus.Desc
	rxErrors  *prometheus.Desc
	txErrors  *prometheus.Desc
	rxDrops   *prometheus.Desc
	txDrops   *prometheus.Desc
}

func newNetworkCollector() *networkCollector {
	labels := []string{"interface"}
	desc := func(name, help string) *prometheus.Desc {
		return prometheus.NewDesc(name, help, labels, nil)
	}
	return &networkCollector{
		rxBytes:   desc("os_network_receive_bytes_total", "Total bytes received on network interface"),
		txBytes:   desc("os_network_transmit_bytes_total", "Total bytes transmitted on network interface"),
		rxPackets: desc("os_network_receive_packets_total", "Total packets received on network interface"),
		txPackets: desc("os_network_transmit_packets_total", "Total packets transmitted on network interface"),
		rxErrors:  desc("os_network_receive_errors_total", "Total receive errors on network interface"),
		txErrors:  desc("os_network_transmit_errors_total", "Total transmit errors on network interface"),
		rxDrops:   desc("os_network_receive_drops_total", "Total receive drops on network interface"),
		txDrops:   desc("os_network_transmit_drops_total", "Total transmit drops on network interface"),
	}
}

func (c *networkCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.rxBytes
	ch <- c.txBytes
	ch <- c.rxPackets
	ch <- c.txPackets
	ch <- c.rxErrors
	ch <- c.txErrors
	ch <- c.rxDrops
	ch <- c.txDrops
}

func (c *networkCollector) Collect(ch chan<- prometheus.Metric) {
	counters, err := gopsnet.IOCounters(true)
	if err != nil {
		slog.Error("failed to collect network I/O metrics", slog.Any("error", err))
		return
	}

	for _, iface := range counters {
		counter := func(desc *prometheus.Desc, v uint64) {
			ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, float64(v), iface.Name)
		}
		counter(c.rxBytes, iface.BytesRecv)
		counter(c.txBytes, iface.BytesSent)
		counter(c.rxPackets, iface.PacketsRecv)
		counter(c.txPackets, iface.PacketsSent)
		counter(c.rxErrors, iface.Errin)
		counter(c.txErrors, iface.Errout)
		counter(c.rxDrops, iface.Dropin)
		counter(c.txDrops, iface.Dropout)
	}
}
