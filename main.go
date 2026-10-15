package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	_ "github.com/microsoft/go-mssqldb"
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

// shutdownTimeout bounds how long in-flight scrapes may take to finish once a
// stop is requested (Ctrl+C, SIGTERM or a Windows service stop).
const shutdownTimeout = 10 * time.Second

func main() {
	args := os.Args[1:]

	if len(args) > 0 && args[0] == "service" {
		os.Exit(runServiceCommand(args[1:]))
	}

	if isWindowsService() {
		os.Exit(runAsWindowsService(args))
	}

	cfg, err := config.Load(args)
	if err != nil {
		if errors.Is(err, config.ErrExitRequested) {
			return
		}
		slog.Error("failed to load configuration", slog.Any("error", err))
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, cfg); err != nil {
		slog.Error("exporter stopped with an error", slog.Any("error", err))
		os.Exit(1)
	}
}

// run starts the exporter and blocks until ctx is cancelled, then shuts the
// HTTP server down gracefully. It returns an error if the exporter could not
// start or the HTTP server failed.
func run(ctx context.Context, cfg *config.Config) error {
	slog.Info("starting mssql-exporter", slog.String("version", version.String()))

	opts := db.DefaultOptions()

	if err := db.Connect(cfg.DSN, opts); err != nil {
		return fmt.Errorf("invalid sql server DSN: %w", err)
	}

	conn := db.DB()

	pingCtx, cancelPing := context.WithTimeout(ctx, db.CurrentOptions().ConnTimeout)
	err := conn.PingContext(pingCtx)
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

	reg.MustRegister(collector.NewExporterWithOptions(conn, cfg.EnabledCollectors, cfg.CollectorOptions))

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

	// ContinueOnError: a single bad metric (e.g. a duplicate series) is logged
	// and dropped instead of failing the whole response with HTTP 500, which
	// would turn every series into a gap.
	metricsHandler := promhttp.InstrumentMetricHandler(
		prometheus.DefaultRegisterer,
		promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{
			ErrorLog:      slog.NewLogLogger(slog.Default().Handler(), slog.LevelError),
			ErrorHandling: promhttp.ContinueOnError,
		}),
	)

	mux := http.NewServeMux()
	mux.Handle(cfg.WebTelemetryPath, metricsHandler)
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
	serveErr := make(chan error, 1)
	go func() { serveErr <- web.ListenAndServe(server, flags, slog.Default()) }()

	select {
	case err := <-serveErr:
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
	}

	slog.Info("shutting down mssql-exporter")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	if err := <-serveErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http server: %w", err)
	}
	return nil
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
