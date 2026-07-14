package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/viniciussouzao/mssql-exporter/internal/collector"
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
	user := os.Getenv("SQL_SERVER_USER")
	password := os.Getenv("SQL_SERVER_PASSWORD")
	host := os.Getenv("SQL_SERVER_HOST")
	port := os.Getenv("SQL_SERVER_PORT")
	database := os.Getenv("SQL_SERVER_DATABASE")
	appName := os.Getenv("EXPORTER_APP_NAME")

	connString := "server=" + host + ";" +
		"user id=" + user + ";" +
		"password=" + password + ";" +
		"port=" + port + ";" +
		"database=" + database + ";" +
		"app name=" + appName

	opts := db.DefaultOptions()

	if err := db.Connect(connString, opts); err != nil {
		slog.Error("failed to connect to sql server:", slog.Any("error", err))
		return
	}

	db := db.DB()

	if err := db.Ping(); err != nil {
		slog.Error("failed to ping sql server:", slog.Any("error", err))
		return
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

	if os.Getenv("ENABLE_OS_COLLECTORS") == "true" {
		oscollector.Register(reg)
		slog.Info("os collectors enabled")
	}

	slog.Info("sql server exporter is running in :9187 /metrics")
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(":9187", nil); err != nil {
		slog.Error("failed to start http server:", slog.Any("error", err))
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
