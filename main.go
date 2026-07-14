package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/viniciussouzao/mssql-exporter/internal/collector"
	"github.com/viniciussouzao/mssql-exporter/internal/config"
	"github.com/viniciussouzao/mssql-exporter/internal/db"
	"github.com/viniciussouzao/mssql-exporter/internal/oscollector"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		slog.Error("Error loading .env file", slog.Any("error", err))
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

	opts := db.DefaultOptions()

	if err := db.Connect(cfg.DSN, opts); err != nil {
		slog.Error("failed to connect to sql server", slog.Any("error", err))
		os.Exit(1)
	}

	db := db.DB()

	if err := db.Ping(); err != nil {
		slog.Error("failed to ping sql server", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("successfully connected and pinged sql server")

	instanceName := queryServerName(db)
	slog.Info("sql server instance", slog.String("instance", instanceName))

	reg := prometheus.WrapRegistererWith(
		prometheus.Labels{"sqlserver_instance": instanceName},
		prometheus.DefaultRegisterer,
	)

	reg.MustRegister(
		collector.NewWaitStatsCollector(db),
		collector.NewIOFileStatsCollector(db),
		collector.NewSessionsCollector(db),
		collector.NewServerInfoCollector(db),
		collector.NewMemoryClerksCollector(db),
		collector.NewBackupsCollector(db),
		collector.NewTempDBCollector(db),
	)

	if cfg.EnableOSCollectors {
		oscollector.Register(reg)
		slog.Info("os collectors enabled")
	}

	slog.Info("sql server exporter is running", slog.String("address", cfg.WebListenAddress), slog.String("telemetry_path", cfg.WebTelemetryPath))
	http.Handle(cfg.WebTelemetryPath, promhttp.Handler())
	if err := http.ListenAndServe(cfg.WebListenAddress, nil); err != nil {
		slog.Error("failed to start http server", slog.Any("error", err))
		os.Exit(1)
	}
}

func queryServerName(db *sql.DB) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var name string
	err := db.QueryRowContext(ctx, "SELECT @@SERVERNAME").Scan(&name)
	if err != nil {
		slog.Error("failed to query @@SERVERNAME, using empty label", slog.Any("error", err))
		return ""
	}
	return name
}
