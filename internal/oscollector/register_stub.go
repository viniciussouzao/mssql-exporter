//go:build !linux

package oscollector

import "github.com/prometheus/client_golang/prometheus"

// Register is a no-op on non-Linux platforms.
func Register(_ prometheus.Registerer) {}
