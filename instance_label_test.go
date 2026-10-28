package main

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func labeledGauge() func(prometheus.Registerer) {
	return func(reg prometheus.Registerer) {
		reg.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Name: "sqlserver_test_gauge", Help: "test",
		}, func() float64 { return 1 }))
	}
}

func instanceLabel(t *testing.T, r *instanceRegistry) string {
	t.Helper()
	families, err := r.Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v", err)
	}
	for _, f := range families {
		for _, m := range f.GetMetric() {
			for _, l := range m.GetLabel() {
				if l.GetName() == "sqlserver_instance" {
					return l.GetValue()
				}
			}
		}
	}
	t.Fatal("no sqlserver_instance label gathered")
	return ""
}

func TestInstanceRegistrySetName(t *testing.T) {
	r := newInstanceRegistry("", labeledGauge())
	if got := instanceLabel(t, r); got != "" {
		t.Errorf("sqlserver_instance = %q, want empty before the name is known", got)
	}

	r.setName("SQL01")
	if got := instanceLabel(t, r); got != "SQL01" {
		t.Errorf("sqlserver_instance = %q, want SQL01 after setName", got)
	}
}

// SQL Server down at startup: the name is retried until it can be read, then
// the label is set without a restart (HML-029).
func TestResolveInstanceNameRetriesUntilAvailable(t *testing.T) {
	r := newInstanceRegistry("", labeledGauge())
	var calls atomic.Int32
	query := func(context.Context) (string, error) {
		if calls.Add(1) < 3 {
			return "", errors.New("connection refused")
		}
		return "SQL01", nil
	}

	done := make(chan struct{})
	go func() {
		resolveInstanceName(context.Background(), r, 5*time.Millisecond, query)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("resolveInstanceName did not return after the name became available")
	}

	if got := calls.Load(); got != 3 {
		t.Errorf("query called %d times, want 3 (two failures, then the name)", got)
	}
	if got := instanceLabel(t, r); got != "SQL01" {
		t.Errorf("sqlserver_instance = %q, want SQL01", got)
	}
}

func TestResolveInstanceNameStopsOnNullName(t *testing.T) {
	r := newInstanceRegistry("", labeledGauge())
	var calls atomic.Int32
	query := func(context.Context) (string, error) {
		calls.Add(1)
		return "", errNullServerName
	}

	resolveInstanceName(context.Background(), r, time.Millisecond, query)

	if got := calls.Load(); got != 1 {
		t.Errorf("query called %d times, want 1 (a NULL name is not retried)", got)
	}
	if got := instanceLabel(t, r); got != "" {
		t.Errorf("sqlserver_instance = %q, want empty", got)
	}
}

func TestResolveInstanceNameStopsOnShutdown(t *testing.T) {
	r := newInstanceRegistry("", labeledGauge())
	ctx, cancel := context.WithCancel(context.Background())
	query := func(context.Context) (string, error) { return "", errors.New("down") }

	done := make(chan struct{})
	go func() {
		resolveInstanceName(ctx, r, time.Millisecond, query)
		close(done)
	}()
	time.Sleep(20 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("resolveInstanceName kept retrying after the context was cancelled")
	}
}
