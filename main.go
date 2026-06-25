package main

import (
	"log/slog"
	"net/http"
	"os"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/viniciussouzao/mssql-exporter/internal/collector"
	"github.com/viniciussouzao/mssql-exporter/internal/db"
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

	prometheus.MustRegister(
		collector.NewWaitStatsCollector(db),
		collector.NewIOFileStatsCollector(db),
		collector.NewSessionsCollector(db),
		collector.NewServerInfoCollector(db),
		collector.NewMemoryClerksCollector(db),
		collector.NewBackupsCollector(db),
		collector.NewTempDBCollector(db),
	)

	slog.Info("sql server exporter is running in :9187 /metrics")
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":9187", nil)
}
