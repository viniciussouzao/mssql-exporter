package main

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"runtime"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/exporter-toolkit/web"
	"github.com/viniciussouzao/mssql-exporter/internal/collector"
	"github.com/viniciussouzao/mssql-exporter/internal/config"
	"github.com/viniciussouzao/mssql-exporter/internal/db"
	"github.com/viniciussouzao/mssql-exporter/internal/oscollector"
	"github.com/viniciussouzao/mssql-exporter/internal/version"
)

func init() {
	// .env is a local-development convenience; its absence is the normal case
	// when running from --config.file or environment variables (e.g. as a service).
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		slog.Warn("failed to load .env file", slog.Any("error", err))
	}
}

func main() {
	cfg, err := config.Load(os.Args[1:])
	if err != nil {
		if errors.Is(err, config.ErrExitRequested) {
			return
		}
		slog.Error("failed to load configuration", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("starting mssql-exporter", slog.String("version", version.String()))

	opts := db.DefaultOptions()

	if err := db.Connect(cfg.DSN, opts); err != nil {
		slog.Error("invalid sql server DSN", slog.Any("error", err))
		os.Exit(1)
	}

	conn := db.DB()

	pingCtx, cancelPing := context.WithTimeout(context.Background(), db.CurrentOptions().ConnTimeout)
	err = conn.PingContext(pingCtx)
	cancelPing()

	var instanceName string
	if err != nil {
		slog.Warn("sql server is unreachable at startup, retrying on every scrape (sqlserver_up will report 0)", slog.Any("error", err))
		slog.Warn("sqlserver_instance label will stay empty until the exporter restarts with sql server reachable")
	} else {
		slog.Info("successfully connected and pinged sql server")
		instanceName = queryServerName(conn)
		slog.Info("sql server instance", slog.String("instance", instanceName))
	}

	reg := prometheus.WrapRegistererWith(
		prometheus.Labels{"sqlserver_instance": instanceName},
		prometheus.DefaultRegisterer,
	)

	reg.MustRegister(collector.NewExporter(conn, cfg.EnabledCollectors))

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "mssql_exporter_build_info",
			Help: "A metric with a constant '1' value labeled by version, commit, build date and Go version from which mssql-exporter was built.",
			ConstLabels: prometheus.Labels{
				"version":   version.Version,
				"commit":    version.Commit,
				"date":      version.Date,
				"goversion": runtime.Version(),
			},
		},
		func() float64 { return 1 },
	))

	if cfg.EnableOSCollectors {
		oscollector.Register(reg)
		slog.Info("os collectors enabled")
	}

	mux := http.NewServeMux()
	mux.Handle(cfg.WebTelemetryPath, promhttp.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	slog.Info("sql server exporter is running", slog.String("address", cfg.WebListenAddress), slog.String("telemetry_path", cfg.WebTelemetryPath))

	server := &http.Server{Handler: mux}
	flags := &web.FlagConfig{
		WebListenAddresses: &[]string{cfg.WebListenAddress},
		WebSystemdSocket:   new(bool),
		WebConfigFile:      &cfg.WebConfigFile,
	}
	if err := web.ListenAndServe(server, flags, slog.Default()); err != nil {
		slog.Error("failed to start http server", slog.Any("error", err))
		os.Exit(1)
	}
}

func queryServerName(conn *sql.DB) string {
	ctx, cancel := context.WithTimeout(context.Background(), db.CurrentOptions().ConnTimeout)
	defer cancel()

	var name string
	err := conn.QueryRowContext(ctx, "SELECT @@SERVERNAME").Scan(&name)
	if err != nil {
		slog.Error("failed to query @@SERVERNAME, using empty label", slog.Any("error", err))
		return ""
	}
	return name
}
