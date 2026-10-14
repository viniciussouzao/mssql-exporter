package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
	"github.com/viniciussouzao/mssql-exporter/internal/collector"
	"github.com/viniciussouzao/mssql-exporter/internal/version"
	"gopkg.in/yaml.v3"
)

const (
	defaultWebListenAddress = ":9187"
	defaultWebTelemetryPath = "/metrics"
)

// ErrExitRequested is returned by Load when the requested action was fully
// handled by the flag parsing itself (--help or --version) and the program
// should exit successfully without starting the exporter.
var ErrExitRequested = errors.New("exit requested")

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
	WebConfigFile      string
	EnableOSCollectors bool
	ConfigFile         string

	// EnabledCollectors maps each SQL Server collector name (see
	// collector.ScraperNames) to whether it should run. Defaults come from
	// collector.ScraperEnabledByDefault.
	EnabledCollectors map[string]bool
}

type fileConfig struct {
	DSN string `yaml:"dsn"`
	Web struct {
		ListenAddress string `yaml:"listen_address"`
		TelemetryPath string `yaml:"telemetry_path"`
		ConfigFile    string `yaml:"config_file"`
	} `yaml:"web"`
	// Collectors toggles collectors by name; `os_collectors` toggles the
	// OS-level collectors, every other key must match a SQL Server collector.
	Collectors map[string]bool `yaml:"collectors"`
}

// Load parses CLI flags from args (typically os.Args[1:]) and merges them
// with environment variables and an optional YAML config file (--config.file)
// to produce the final Config. It returns ErrExitRequested when --help or
// --version was handled and the caller should exit with status 0.
func Load(args []string) (*Config, error) {
	fs := pflag.NewFlagSet("mssql-exporter", pflag.ContinueOnError)

	webListenAddress := fs.String("web.listen-address", defaultWebListenAddress, "Address to listen on for the web interface and metrics.")
	webTelemetryPath := fs.String("web.telemetry-path", defaultWebTelemetryPath, "Path under which to expose metrics.")
	webConfigFile := fs.String("web.config.file", "", "Path to a exporter-toolkit web config file enabling TLS and/or basic auth (see https://github.com/prometheus/exporter-toolkit/blob/master/docs/web-configuration.md).")
	enableOSCollectors := fs.Bool("collector.os-collectors", false, "Enable OS-level collectors (CPU, memory, disk).")
	configFile := fs.String("config.file", "", "Path to an optional YAML config file.")
	showVersion := fs.Bool("version", false, "Print version information and exit.")

	collectorFlags := make(map[string]*bool, len(collector.ScraperNames()))
	for _, name := range collector.ScraperNames() {
		collectorFlags[name] = fs.Bool("collector."+name, collector.ScraperEnabledByDefault(name), "Enable the "+name+" collector.")
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			return nil, ErrExitRequested
		}
		return nil, err
	}

	if *showVersion {
		fmt.Println(version.String())
		return nil, ErrExitRequested
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
	if cfg.WebListenAddress == "" {
		return nil, fmt.Errorf("web.listen-address must not be empty")
	}

	cfg.WebTelemetryPath = defaultWebTelemetryPath
	if fc.Web.TelemetryPath != "" {
		cfg.WebTelemetryPath = fc.Web.TelemetryPath
	}
	if fs.Changed("web.telemetry-path") {
		cfg.WebTelemetryPath = *webTelemetryPath
	}
	if !strings.HasPrefix(cfg.WebTelemetryPath, "/") {
		return nil, fmt.Errorf("web.telemetry-path must start with '/'")
	}

	cfg.WebConfigFile = fc.Web.ConfigFile
	if fs.Changed("web.config.file") {
		cfg.WebConfigFile = *webConfigFile
	}

	cfg.EnableOSCollectors = fc.Collectors["os_collectors"]
	if v := os.Getenv("ENABLE_OS_COLLECTORS"); v != "" {
		enabled, err := strconv.ParseBool(v)
		if err != nil {
			slog.Warn("invalid ENABLE_OS_COLLECTORS value, ignoring", slog.String("value", v))
		} else {
			cfg.EnableOSCollectors = enabled
		}
	}
	if fs.Changed("collector.os-collectors") {
		cfg.EnableOSCollectors = *enableOSCollectors
	}

	scraperNames := collector.ScraperNames()
	known := make(map[string]bool, len(scraperNames)+1)
	known["os_collectors"] = true
	cfg.EnabledCollectors = make(map[string]bool, len(scraperNames))
	for _, name := range scraperNames {
		known[name] = true

		enabled := collector.ScraperEnabledByDefault(name)
		if v, ok := fc.Collectors[name]; ok {
			enabled = v
		}
		if fs.Changed("collector." + name) {
			enabled = *collectorFlags[name]
		}
		cfg.EnabledCollectors[name] = enabled
	}
	for name := range fc.Collectors {
		if !known[name] {
			slog.Warn("unknown collector in config file, ignoring", slog.String("collector", name))
		}
	}

	cfg.DSN = fc.DSN
	if v := os.Getenv("SQL_SERVER_DSN"); v != "" {
		cfg.DSN = v
	}
	if cfg.DSN == "" {
		return nil, fmt.Errorf("SQL_SERVER_DSN not set via environment variable or config file")
	}

	return cfg, nil
}
