package config

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/viniciussouzao/mssql-exporter/internal/collector"
)

const testDSN = "sqlserver://sa:pass@localhost:1433?database=master"

func writeConfigFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadDefaultsWithEnvDSN(t *testing.T) {
	t.Setenv("SQL_SERVER_DSN", testDSN)

	cfg, err := Load(nil)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DSN != testDSN {
		t.Errorf("DSN = %q, want %q", cfg.DSN, testDSN)
	}
	if cfg.WebListenAddress != defaultWebListenAddress {
		t.Errorf("WebListenAddress = %q, want %q", cfg.WebListenAddress, defaultWebListenAddress)
	}
	if cfg.WebTelemetryPath != defaultWebTelemetryPath {
		t.Errorf("WebTelemetryPath = %q, want %q", cfg.WebTelemetryPath, defaultWebTelemetryPath)
	}
	if cfg.EnableOSCollectors {
		t.Error("EnableOSCollectors = true, want false")
	}
}

func TestLoadMissingDSN(t *testing.T) {
	t.Setenv("SQL_SERVER_DSN", "")

	_, err := Load(nil)
	if err == nil {
		t.Fatal("Load() error = nil, want missing DSN error")
	}
	if !strings.Contains(err.Error(), "SQL_SERVER_DSN") {
		t.Errorf("Load() error = %q, want it to mention SQL_SERVER_DSN", err)
	}
}

func TestLoadFromConfigFile(t *testing.T) {
	t.Setenv("SQL_SERVER_DSN", "")
	t.Setenv("ENABLE_OS_COLLECTORS", "")

	path := writeConfigFile(t, `
dsn: "`+testDSN+`"
web:
  listen_address: ":9300"
  telemetry_path: "/custom"
collectors:
  os_collectors: true
`)

	cfg, err := Load([]string{"--config.file", path})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DSN != testDSN {
		t.Errorf("DSN = %q, want %q", cfg.DSN, testDSN)
	}
	if cfg.WebListenAddress != ":9300" {
		t.Errorf("WebListenAddress = %q, want :9300", cfg.WebListenAddress)
	}
	if cfg.WebTelemetryPath != "/custom" {
		t.Errorf("WebTelemetryPath = %q, want /custom", cfg.WebTelemetryPath)
	}
	if !cfg.EnableOSCollectors {
		t.Error("EnableOSCollectors = false, want true")
	}
}

func TestLoadFlagOverridesConfigFile(t *testing.T) {
	t.Setenv("SQL_SERVER_DSN", "")

	path := writeConfigFile(t, `
dsn: "`+testDSN+`"
web:
  listen_address: ":9300"
`)

	cfg, err := Load([]string{"--config.file", path, "--web.listen-address", ":9200"})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.WebListenAddress != ":9200" {
		t.Errorf("WebListenAddress = %q, want flag value :9200", cfg.WebListenAddress)
	}
}

func TestLoadEnvDSNOverridesConfigFile(t *testing.T) {
	envDSN := "sqlserver://sa:pass@envhost:1433?database=master"
	t.Setenv("SQL_SERVER_DSN", envDSN)

	path := writeConfigFile(t, `dsn: "`+testDSN+`"`)

	cfg, err := Load([]string{"--config.file", path})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DSN != envDSN {
		t.Errorf("DSN = %q, want env value %q", cfg.DSN, envDSN)
	}
}

func TestLoadOSCollectorsEnvVar(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		fileBool bool
		want     bool
	}{
		{"true enables", "true", false, true},
		{"1 enables", "1", false, true},
		{"TRUE enables", "TRUE", false, true},
		{"false overrides file", "false", true, false},
		{"invalid value is ignored, file wins", "yes", true, true},
		{"empty is unset, file wins", "", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SQL_SERVER_DSN", "")
			t.Setenv("ENABLE_OS_COLLECTORS", tt.envValue)

			content := `dsn: "` + testDSN + `"`
			if tt.fileBool {
				content += "\ncollectors:\n  os_collectors: true"
			}
			path := writeConfigFile(t, content)

			cfg, err := Load([]string{"--config.file", path})
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if cfg.EnableOSCollectors != tt.want {
				t.Errorf("EnableOSCollectors = %v, want %v", cfg.EnableOSCollectors, tt.want)
			}
		})
	}
}

func TestLoadWebConfigFilePrecedence(t *testing.T) {
	t.Setenv("SQL_SERVER_DSN", "")

	path := writeConfigFile(t, `
dsn: "`+testDSN+`"
web:
  config_file: "/etc/from-yaml.yml"
`)

	cfg, err := Load([]string{"--config.file", path})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.WebConfigFile != "/etc/from-yaml.yml" {
		t.Errorf("WebConfigFile = %q, want yaml value", cfg.WebConfigFile)
	}

	cfg, err = Load([]string{"--config.file", path, "--web.config.file", "/etc/from-flag.yml"})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.WebConfigFile != "/etc/from-flag.yml" {
		t.Errorf("WebConfigFile = %q, want flag to override yaml", cfg.WebConfigFile)
	}
}

func TestLoadCollectorsDefaultEnabled(t *testing.T) {
	t.Setenv("SQL_SERVER_DSN", testDSN)

	cfg, err := Load(nil)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.EnabledCollectors) == 0 {
		t.Fatal("EnabledCollectors is empty, want every collector listed")
	}
	for name, enabled := range cfg.EnabledCollectors {
		if want := collector.ScraperEnabledByDefault(name); enabled != want {
			t.Errorf("collector %q enabled = %v by default, want %v", name, enabled, want)
		}
	}
	if cfg.EnabledCollectors["log_shipping"] {
		t.Error("log_shipping enabled by default, want opt-in (needs extra msdb grants)")
	}
	if !cfg.EnabledCollectors["sessions"] {
		t.Error("sessions disabled by default, want enabled")
	}
}

func TestLoadDefaultDisabledCollectorEnabledViaFlagAndConfigFile(t *testing.T) {
	t.Setenv("SQL_SERVER_DSN", testDSN)

	cfg, err := Load([]string{"--collector.log_shipping"})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !cfg.EnabledCollectors["log_shipping"] {
		t.Error("log_shipping = disabled, want enabled via flag")
	}

	t.Setenv("SQL_SERVER_DSN", "")
	path := writeConfigFile(t, `
dsn: "`+testDSN+`"
collectors:
  log_shipping: true
`)
	cfg, err = Load([]string{"--config.file", path})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !cfg.EnabledCollectors["log_shipping"] {
		t.Error("log_shipping = disabled, want enabled via config file")
	}
}

func TestLoadCollectorDisabledViaFlag(t *testing.T) {
	t.Setenv("SQL_SERVER_DSN", testDSN)

	cfg, err := Load([]string{"--collector.backups=false"})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.EnabledCollectors["backups"] {
		t.Error("backups = enabled, want disabled via flag")
	}
	if !cfg.EnabledCollectors["sessions"] {
		t.Error("sessions = disabled, want untouched collectors to stay enabled")
	}
}

func TestLoadCollectorDisabledViaConfigFile(t *testing.T) {
	t.Setenv("SQL_SERVER_DSN", "")

	path := writeConfigFile(t, `
dsn: "`+testDSN+`"
collectors:
  wait_stats: false
`)

	cfg, err := Load([]string{"--config.file", path})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.EnabledCollectors["wait_stats"] {
		t.Error("wait_stats = enabled, want disabled via config file")
	}
}

func TestLoadCollectorFlagOverridesConfigFile(t *testing.T) {
	t.Setenv("SQL_SERVER_DSN", "")

	path := writeConfigFile(t, `
dsn: "`+testDSN+`"
collectors:
  wait_stats: false
`)

	cfg, err := Load([]string{"--config.file", path, "--collector.wait_stats=true"})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !cfg.EnabledCollectors["wait_stats"] {
		t.Error("wait_stats = disabled, want flag to re-enable over config file")
	}
}

func TestLoadOSCollectorsFlagOverridesEnv(t *testing.T) {
	t.Setenv("SQL_SERVER_DSN", testDSN)
	t.Setenv("ENABLE_OS_COLLECTORS", "false")

	cfg, err := Load([]string{"--collector.os-collectors"})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !cfg.EnableOSCollectors {
		t.Error("EnableOSCollectors = false, want flag value true")
	}
}

func TestLoadEmptyTelemetryPath(t *testing.T) {
	t.Setenv("SQL_SERVER_DSN", testDSN)

	_, err := Load([]string{"--web.telemetry-path", ""})
	if err == nil {
		t.Fatal("Load() error = nil, want empty telemetry path error")
	}
}

func TestLoadTelemetryPathWithoutLeadingSlash(t *testing.T) {
	t.Setenv("SQL_SERVER_DSN", testDSN)

	_, err := Load([]string{"--web.telemetry-path", "metrics"})
	if err == nil {
		t.Fatal("Load() error = nil, want leading slash error")
	}
	if !strings.Contains(err.Error(), "must start with '/'") {
		t.Errorf("Load() error = %q, want it to mention the leading slash", err)
	}
}

func TestLoadEmptyListenAddress(t *testing.T) {
	t.Setenv("SQL_SERVER_DSN", testDSN)

	_, err := Load([]string{"--web.listen-address", ""})
	if err == nil {
		t.Fatal("Load() error = nil, want empty listen address error")
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	t.Setenv("SQL_SERVER_DSN", testDSN)

	_, err := Load([]string{"--config.file", filepath.Join(t.TempDir(), "missing.yml")})
	if err == nil {
		t.Fatal("Load() error = nil, want read error")
	}
}

func TestLoadInvalidConfigFile(t *testing.T) {
	t.Setenv("SQL_SERVER_DSN", testDSN)

	path := writeConfigFile(t, "dsn: [not: valid")

	_, err := Load([]string{"--config.file", path})
	if err == nil {
		t.Fatal("Load() error = nil, want parse error")
	}
}

func TestLoadVersionReturnsExitRequested(t *testing.T) {
	_, err := Load([]string{"--version"})
	if !errors.Is(err, ErrExitRequested) {
		t.Errorf("Load(--version) error = %v, want ErrExitRequested", err)
	}
}

func TestLoadHelpReturnsExitRequested(t *testing.T) {
	_, err := Load([]string{"--help"})
	if !errors.Is(err, ErrExitRequested) {
		t.Errorf("Load(--help) error = %v, want ErrExitRequested", err)
	}
}

func TestLoadUnknownFlag(t *testing.T) {
	_, err := Load([]string{"--does-not-exist"})
	if err == nil {
		t.Fatal("Load() error = nil, want unknown flag error")
	}
	if errors.Is(err, ErrExitRequested) {
		t.Error("unknown flag should not be ErrExitRequested")
	}
}

func TestLoadTopQueriesOptionsDefaults(t *testing.T) {
	t.Setenv("SQL_SERVER_DSN", testDSN)

	cfg, err := Load(nil)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got, want := cfg.CollectorOptions, collector.DefaultOptions(); !reflect.DeepEqual(got, want) {
		t.Errorf("CollectorOptions = %+v, want defaults %+v", got, want)
	}
	if cfg.EnabledCollectors["top_queries"] {
		t.Error("top_queries enabled by default, want it opt-in")
	}
}

func TestLoadTopQueriesOptionsFromConfigFileAndFlags(t *testing.T) {
	t.Setenv("SQL_SERVER_DSN", testDSN)
	path := writeConfigFile(t, `
collectors:
  top_queries: true
top_queries:
  limit: 20
  include_text: true
  text_length: 200
`)

	cfg, err := Load([]string{"--config.file", path})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	want := collector.TopQueriesOptions{Limit: 20, IncludeText: true, TextLength: 200}
	if cfg.CollectorOptions.TopQueries != want {
		t.Errorf("TopQueries = %+v, want %+v from the config file", cfg.CollectorOptions.TopQueries, want)
	}
	if !cfg.EnabledCollectors["top_queries"] {
		t.Error("top_queries disabled, want it enabled by the config file")
	}

	// Flags win over the file; unset fields keep the file's values.
	cfg, err = Load([]string{"--config.file", path, "--collector.top_queries.limit=5", "--collector.top_queries.include_text=false"})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	want = collector.TopQueriesOptions{Limit: 5, IncludeText: false, TextLength: 200}
	if cfg.CollectorOptions.TopQueries != want {
		t.Errorf("TopQueries = %+v, want %+v (flags over file)", cfg.CollectorOptions.TopQueries, want)
	}
}

func TestLoadTopQueriesOptionsValidation(t *testing.T) {
	t.Setenv("SQL_SERVER_DSN", testDSN)

	for _, args := range [][]string{
		{"--collector.top_queries.limit=0"},
		{"--collector.top_queries.limit=1001"},
		{"--collector.top_queries.text_length=15"},
		{"--collector.top_queries.text_length=4001"},
	} {
		if _, err := Load(args); err == nil || !strings.Contains(err.Error(), "collector.top_queries.") {
			t.Errorf("Load(%v) error = %v, want a collector.top_queries range error", args, err)
		}
	}
	for _, args := range [][]string{
		{"--collector.top_queries.limit=1"},
		{"--collector.top_queries.limit=1000"},
		{"--collector.top_queries.text_length=16"},
		{"--collector.top_queries.text_length=4000"},
	} {
		if _, err := Load(args); err != nil {
			t.Errorf("Load(%v) error = %v, want the boundary accepted", args, err)
		}
	}
}

func TestLoadIndexFragmentationOptions(t *testing.T) {
	t.Setenv("SQL_SERVER_DSN", testDSN)

	cfg, err := Load(nil)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.EnabledCollectors["index_fragmentation"] {
		t.Error("index_fragmentation enabled by default, want it opt-in")
	}

	path := writeConfigFile(t, `
collectors:
  index_fragmentation: true
index_fragmentation:
  interval: 30m
  min_page_count: 200
  databases: [AppDB, Sales]
`)
	cfg, err = Load([]string{"--config.file", path})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	want := collector.IndexFragmentationOptions{Interval: 30 * time.Minute, MinPageCount: 200, Databases: []string{"AppDB", "Sales"}}
	if !reflect.DeepEqual(cfg.CollectorOptions.IndexFragmentation, want) {
		t.Errorf("IndexFragmentation = %+v, want %+v from the config file", cfg.CollectorOptions.IndexFragmentation, want)
	}
	if !cfg.EnabledCollectors["index_fragmentation"] {
		t.Error("index_fragmentation disabled, want it enabled by the config file")
	}

	// Flags win over the file; unset fields keep the file's values.
	cfg, err = Load([]string{"--config.file", path, "--collector.index_fragmentation.interval=2h", "--collector.index_fragmentation.databases=Only"})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	want = collector.IndexFragmentationOptions{Interval: 2 * time.Hour, MinPageCount: 200, Databases: []string{"Only"}}
	if !reflect.DeepEqual(cfg.CollectorOptions.IndexFragmentation, want) {
		t.Errorf("IndexFragmentation = %+v, want %+v (flags over file)", cfg.CollectorOptions.IndexFragmentation, want)
	}
}

func TestLoadIndexFragmentationOptionsValidation(t *testing.T) {
	t.Setenv("SQL_SERVER_DSN", testDSN)

	for _, args := range [][]string{
		{"--collector.index_fragmentation.interval=59s"},
		{"--collector.index_fragmentation.min_page_count=-1"},
	} {
		if _, err := Load(args); err == nil || !strings.Contains(err.Error(), "collector.index_fragmentation.") {
			t.Errorf("Load(%v) error = %v, want a collector.index_fragmentation range error", args, err)
		}
	}
	for _, args := range [][]string{
		{"--collector.index_fragmentation.interval=1m"},
		{"--collector.index_fragmentation.min_page_count=0"},
	} {
		if _, err := Load(args); err != nil {
			t.Errorf("Load(%v) error = %v, want the boundary accepted", args, err)
		}
	}
}

func TestLoadQueryStoreOptions(t *testing.T) {
	t.Setenv("SQL_SERVER_DSN", testDSN)

	cfg, err := Load(nil)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.EnabledCollectors["query_store"] {
		t.Error("query_store enabled by default, want it opt-in")
	}
	if want := collector.DefaultOptions().QueryStore; !reflect.DeepEqual(cfg.CollectorOptions.QueryStore, want) {
		t.Errorf("QueryStore = %+v, want the defaults %+v", cfg.CollectorOptions.QueryStore, want)
	}

	path := writeConfigFile(t, `
collectors:
  query_store: true
query_store:
  interval: 15m
  databases: [AppDB, Sales]
`)
	cfg, err = Load([]string{"--config.file", path})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	want := collector.QueryStoreOptions{Interval: 15 * time.Minute, Databases: []string{"AppDB", "Sales"}}
	if !reflect.DeepEqual(cfg.CollectorOptions.QueryStore, want) {
		t.Errorf("QueryStore = %+v, want %+v from the config file", cfg.CollectorOptions.QueryStore, want)
	}
	if !cfg.EnabledCollectors["query_store"] {
		t.Error("query_store disabled, want it enabled by the config file")
	}

	// Flags win over the file.
	cfg, err = Load([]string{"--config.file", path, "--collector.query_store.interval=1h", "--collector.query_store.databases=Only"})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	want = collector.QueryStoreOptions{Interval: time.Hour, Databases: []string{"Only"}}
	if !reflect.DeepEqual(cfg.CollectorOptions.QueryStore, want) {
		t.Errorf("QueryStore = %+v, want %+v (flags over file)", cfg.CollectorOptions.QueryStore, want)
	}

	if _, err := Load([]string{"--collector.query_store.interval=59s"}); err == nil || !strings.Contains(err.Error(), "collector.query_store.interval") {
		t.Errorf("Load(interval=59s) error = %v, want a collector.query_store.interval range error", err)
	}
	if _, err := Load([]string{"--collector.query_store.interval=1m"}); err != nil {
		t.Errorf("Load(interval=1m) error = %v, want the boundary accepted", err)
	}
}

func TestLoadInstanceLabelPrecedence(t *testing.T) {
	t.Setenv("SQL_SERVER_DSN", "")
	t.Setenv("SQL_SERVER_INSTANCE_LABEL", "")

	path := writeConfigFile(t, `
dsn: "`+testDSN+`"
instance_label: "from-yaml"
`)

	cfg, err := Load([]string{"--config.file", path})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.InstanceLabel != "from-yaml" {
		t.Errorf("InstanceLabel = %q, want yaml value", cfg.InstanceLabel)
	}

	t.Setenv("SQL_SERVER_INSTANCE_LABEL", "from-env")
	cfg, err = Load([]string{"--config.file", path})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.InstanceLabel != "from-env" {
		t.Errorf("InstanceLabel = %q, want env to override yaml", cfg.InstanceLabel)
	}

	cfg, err = Load([]string{"--config.file", path, "--sqlserver.instance-label", "from-flag"})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.InstanceLabel != "from-flag" {
		t.Errorf("InstanceLabel = %q, want flag to override env and yaml", cfg.InstanceLabel)
	}

	// An explicitly empty flag goes back to the automatic @@SERVERNAME label.
	cfg, err = Load([]string{"--config.file", path, "--sqlserver.instance-label="})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.InstanceLabel != "" {
		t.Errorf("InstanceLabel = %q, want empty (automatic) from an empty flag", cfg.InstanceLabel)
	}
}

func TestLoadInstanceLabelDefaultsToAutomatic(t *testing.T) {
	t.Setenv("SQL_SERVER_DSN", testDSN)
	t.Setenv("SQL_SERVER_INSTANCE_LABEL", "")

	cfg, err := Load(nil)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.InstanceLabel != "" {
		t.Errorf("InstanceLabel = %q, want empty (automatic) by default", cfg.InstanceLabel)
	}
}

func TestLoadInstanceLabelRejectsBlank(t *testing.T) {
	t.Setenv("SQL_SERVER_DSN", testDSN)
	t.Setenv("SQL_SERVER_INSTANCE_LABEL", "")

	_, err := Load([]string{"--sqlserver.instance-label", "   "})
	if err == nil || !strings.Contains(err.Error(), "instance-label") {
		t.Fatalf("Load() error = %v, want a blank instance label error", err)
	}
}

func TestLoadMaxConnectionsPrecedence(t *testing.T) {
	t.Setenv("SQL_SERVER_DSN", "")
	t.Setenv("SQL_SERVER_MAX_CONNECTIONS", "")

	path := writeConfigFile(t, `
dsn: "`+testDSN+`"
max_connections: 8
`)

	cfg, err := Load([]string{"--config.file", path})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.MaxConnections != 8 {
		t.Errorf("MaxConnections = %d, want yaml value 8", cfg.MaxConnections)
	}

	t.Setenv("SQL_SERVER_MAX_CONNECTIONS", "12")
	cfg, err = Load([]string{"--config.file", path})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.MaxConnections != 12 {
		t.Errorf("MaxConnections = %d, want env to override yaml", cfg.MaxConnections)
	}

	cfg, err = Load([]string{"--config.file", path, "--sqlserver.max-connections", "20"})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.MaxConnections != 20 {
		t.Errorf("MaxConnections = %d, want flag to override env and yaml", cfg.MaxConnections)
	}
}

func TestLoadMaxConnectionsDefault(t *testing.T) {
	t.Setenv("SQL_SERVER_DSN", testDSN)
	t.Setenv("SQL_SERVER_MAX_CONNECTIONS", "")

	cfg, err := Load(nil)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if want := 5; cfg.MaxConnections != want {
		t.Errorf("MaxConnections = %d, want default %d", cfg.MaxConnections, want)
	}
}

func TestLoadMaxConnectionsValidation(t *testing.T) {
	t.Setenv("SQL_SERVER_DSN", testDSN)
	t.Setenv("SQL_SERVER_MAX_CONNECTIONS", "")

	for _, v := range []string{"0", "-1", "101"} {
		if _, err := Load([]string{"--sqlserver.max-connections", v}); err == nil || !strings.Contains(err.Error(), "max-connections") {
			t.Errorf("Load(max-connections=%s) error = %v, want a range error", v, err)
		}
	}
	for _, v := range []string{"1", "100"} {
		if _, err := Load([]string{"--sqlserver.max-connections", v}); err != nil {
			t.Errorf("Load(max-connections=%s) error = %v, want the boundary accepted", v, err)
		}
	}

	t.Setenv("SQL_SERVER_MAX_CONNECTIONS", "ten")
	if _, err := Load(nil); err == nil || !strings.Contains(err.Error(), "SQL_SERVER_MAX_CONNECTIONS") {
		t.Errorf("Load(SQL_SERVER_MAX_CONNECTIONS=ten) error = %v, want a parse error", err)
	}
}

func TestLoadWaitStatsIncludeIdle(t *testing.T) {
	t.Setenv("SQL_SERVER_DSN", "")

	cfg, err := Load([]string{"--config.file", writeConfigFile(t, `dsn: "`+testDSN+`"`)})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.CollectorOptions.WaitStats.IncludeIdle {
		t.Error("WaitStats.IncludeIdle = true, want idle waits filtered by default")
	}

	path := writeConfigFile(t, `
dsn: "`+testDSN+`"
wait_stats:
  include_idle: true
`)
	cfg, err = Load([]string{"--config.file", path})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !cfg.CollectorOptions.WaitStats.IncludeIdle {
		t.Error("WaitStats.IncludeIdle = false, want yaml value true")
	}

	cfg, err = Load([]string{"--config.file", path, "--collector.wait_stats.include_idle=false"})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.CollectorOptions.WaitStats.IncludeIdle {
		t.Error("WaitStats.IncludeIdle = true, want the flag to override the config file")
	}
}

func TestLoadMissingIndexesLimit(t *testing.T) {
	t.Setenv("SQL_SERVER_DSN", "")

	path := writeConfigFile(t, `
dsn: "`+testDSN+`"
missing_indexes:
  limit: 50
`)
	cfg, err := Load([]string{"--config.file", path})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.CollectorOptions.MissingIndexes.Limit != 50 {
		t.Errorf("MissingIndexes.Limit = %d, want yaml value 50", cfg.CollectorOptions.MissingIndexes.Limit)
	}
	cfg, err = Load([]string{"--config.file", path, "--collector.missing_indexes.limit", "5"})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.CollectorOptions.MissingIndexes.Limit != 5 {
		t.Errorf("MissingIndexes.Limit = %d, want the flag to override the config file", cfg.CollectorOptions.MissingIndexes.Limit)
	}

	for _, v := range []string{"0", "601"} {
		if _, err := Load([]string{"--config.file", path, "--collector.missing_indexes.limit", v}); err == nil || !strings.Contains(err.Error(), "missing_indexes.limit") {
			t.Errorf("Load(limit=%s) error = %v, want a range error", v, err)
		}
	}
}
