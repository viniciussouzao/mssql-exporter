package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/viniciussouzao/mssql-exporter/internal/db"
)

// instanceNameRetryInterval is how often the exporter retries reading
// @@SERVERNAME when SQL Server was unreachable at startup.
const instanceNameRetryInterval = 15 * time.Second

// instanceRegistry gathers the collectors that carry the sqlserver_instance
// label. The label is a constant label of the registry, so changing it means
// registering the collectors again in a new registry and swapping it in:
// when SQL Server is down at startup the label starts empty and is set once
// @@SERVERNAME can be read (HML-029).
type instanceRegistry struct {
	register func(prometheus.Registerer)
	current  atomic.Pointer[prometheus.Registry]
}

// newInstanceRegistry builds the registry with the given label value;
// register adds the labeled collectors to the registerer it receives.
func newInstanceRegistry(name string, register func(prometheus.Registerer)) *instanceRegistry {
	r := &instanceRegistry{register: register}
	r.setName(name)
	return r
}

// setName replaces the registry with one whose collectors carry name.
// A Gather in progress finishes on the previous registry.
func (r *instanceRegistry) setName(name string) {
	reg := prometheus.NewRegistry()
	r.register(prometheus.WrapRegistererWith(prometheus.Labels{"sqlserver_instance": name}, reg))
	r.current.Store(reg)
}

func (r *instanceRegistry) Gather() ([]*dto.MetricFamily, error) {
	return r.current.Load().Gather()
}

// errNullServerName is returned when @@SERVERNAME is NULL; retrying does not
// help, so the label stays empty.
var errNullServerName = errors.New("@@SERVERNAME is NULL")

// queryServerName reads @@SERVERNAME within the query timeout.
func queryServerName(ctx context.Context, conn *sql.DB) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, db.CurrentOptions().ConnTimeout)
	defer cancel()

	var name sql.NullString
	if err := conn.QueryRowContext(ctx, "SELECT @@SERVERNAME").Scan(&name); err != nil {
		return "", err
	}
	if !name.Valid {
		return "", errNullServerName
	}
	return name.String, nil
}

// resolveInstanceName calls query every interval until it returns a name,
// then sets it on r. It gives up when ctx is done or the name is NULL.
func resolveInstanceName(ctx context.Context, r *instanceRegistry, interval time.Duration, query func(context.Context) (string, error)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		name, err := query(ctx)
		switch {
		case errors.Is(err, errNullServerName):
			slog.Warn("@@SERVERNAME is NULL, sqlserver_instance label stays empty (set instance_label to avoid it)")
			return
		case err != nil:
			slog.Debug("sql server instance name not available yet", slog.Any("error", err))
			continue
		}
		r.setName(name)
		slog.Info("sql server instance resolved, sqlserver_instance label set", slog.String("instance", name))
		return
	}
}
