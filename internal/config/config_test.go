package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
