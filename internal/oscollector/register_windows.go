//go:build windows

package oscollector

import "github.com/prometheus/client_golang/prometheus"

// Register adds all Windows OS collectors to the given registerer.
func Register(r prometheus.Registerer) {
	r.MustRegister(
		newCPUCollector(),
		newMemoryCollector(),
		newDiskIOCollector(),
		newDiskUsageCollector(),
		newNetworkCollector(),
		newProcessCollector(),
		newBootTimeCollector(),
	)
}
