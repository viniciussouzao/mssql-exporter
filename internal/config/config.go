package config

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

const (
	defaultWebListenAddress = ":9187"
	defaultWebTelemetryPath = "/metrics"
)

// Config holds the resolved exporter configuration, merged from CLI flags,
// environment variables and an optional YAML config file.
//
// Precedence per field, highest first: CLI flag > environment variable >
// config file > default. DSN has no CLI flag on purpose: process arguments
// are visible via `ps`/task manager, so it's only accepted via env var or
// config file.
type Config struct {
	DSN                string
	WebListenAddress   string
	WebTelemetryPath   string
	EnableOSCollectors bool
	ConfigFile         string
}

type fileConfig struct {
	DSN string `yaml:"dsn"`
	Web struct {
		ListenAddress string `yaml:"listen_address"`
		TelemetryPath string `yaml:"telemetry_path"`
	} `yaml:"web"`
	Collectors struct {
		OSCollectors bool `yaml:"os_collectors"`
	} `yaml:"collectors"`
}

// Load parses CLI flags from args (typically os.Args[1:]) and merges them
// with environment variables and an optional YAML config file (--config.file)
// to produce the final Config.
func Load(args []string) (*Config, error) {
	fs := pflag.NewFlagSet("mssql-exporter", pflag.ContinueOnError)

	webListenAddress := fs.String("web.listen-address", defaultWebListenAddress, "Address to listen on for the web interface and metrics.")
	webTelemetryPath := fs.String("web.telemetry-path", defaultWebTelemetryPath, "Path under which to expose metrics.")
	enableOSCollectors := fs.Bool("collector.os-collectors", false, "Enable OS-level collectors (CPU, memory, disk).")
	configFile := fs.String("config.file", "", "Path to an optional YAML config file.")
	showVersion := fs.Bool("version", false, "Print version information and exit.")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			os.Exit(0)
		}
		return nil, err
	}

	if *showVersion {
		fmt.Println(versionString())
		os.Exit(0)
	}

	var fc fileConfig
	if *configFile != "" {
		data, err := os.ReadFile(*configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		if err := yaml.Unmarshal(data, &fc); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	cfg := &Config{ConfigFile: *configFile}

	cfg.WebListenAddress = defaultWebListenAddress
	if fc.Web.ListenAddress != "" {
		cfg.WebListenAddress = fc.Web.ListenAddress
	}
	if fs.Changed("web.listen-address") {
		cfg.WebListenAddress = *webListenAddress
	}

	cfg.WebTelemetryPath = defaultWebTelemetryPath
	if fc.Web.TelemetryPath != "" {
		cfg.WebTelemetryPath = fc.Web.TelemetryPath
	}
	if fs.Changed("web.telemetry-path") {
		cfg.WebTelemetryPath = *webTelemetryPath
	}

	cfg.EnableOSCollectors = fc.Collectors.OSCollectors
	if v := os.Getenv("ENABLE_OS_COLLECTORS"); v != "" {
		cfg.EnableOSCollectors = v == "true"
	}
	if fs.Changed("collector.os-collectors") {
		cfg.EnableOSCollectors = *enableOSCollectors
	}

	cfg.DSN = fc.DSN
	if v := os.Getenv("SQL_SERVER_DSN"); v != "" {
		cfg.DSN = v
	}
	if cfg.DSN == "" {
		return nil, fmt.Errorf("SQL_SERVER_DSN not set via environment variable nor config file")
	}

	return cfg, nil
}

func versionString() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "mssql-exporter (unknown version)"
	}
	return fmt.Sprintf("mssql-exporter %s", info.Main.Version)
}
