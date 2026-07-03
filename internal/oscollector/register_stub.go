//go:build !linux && !windows

package oscollector

import "github.com/prometheus/client_golang/prometheus"

// Register is a no-op on unsupported platforms.
func Register(_ prometheus.Registerer) {}
