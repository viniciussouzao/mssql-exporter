package main

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

func TestParseServiceCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    *serviceCommand
		wantErr string
	}{
		{
			name: "install with config",
			args: []string{"install", "--config.file", `C:\ProgramData\mssql-exporter\config.yml`},
			want: &serviceCommand{Action: "install", Name: defaultServiceName, ConfigFile: `C:\ProgramData\mssql-exporter\config.yml`},
		},
		{
			name: "flags before action and custom name",
			args: []string{"--name", "mssql-exporter-inst02", "start"},
			want: &serviceCommand{Action: "start", Name: "mssql-exporter-inst02"},
		},
		{name: "uninstall", args: []string{"uninstall"}, want: &serviceCommand{Action: "uninstall", Name: defaultServiceName}},
		{name: "stop", args: []string{"stop"}, want: &serviceCommand{Action: "stop", Name: defaultServiceName}},
		{name: "status", args: []string{"status"}, want: &serviceCommand{Action: "status", Name: defaultServiceName}},
		{name: "missing action", args: nil, wantErr: "missing service action"},
		{name: "unknown action", args: []string{"restart"}, wantErr: `unknown service action "restart"`},
		{name: "extra arguments", args: []string{"start", "now"}, wantErr: "unexpected arguments: now"},
		{name: "install without config", args: []string{"install"}, wantErr: "--config.file is required"},
		{name: "config on non-install", args: []string{"start", "--config.file", "c.yml"}, wantErr: "only used by install"},
		{name: "invalid name", args: []string{"--name", `bad name\x`, "status"}, wantErr: "invalid service name"},
		{name: "unknown flag", args: []string{"status", "--bogus"}, wantErr: "unknown flag"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseServiceCommand(tt.args, &bytes.Buffer{})
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("parseServiceCommand(%q) error = %v, want it to contain %q", tt.args, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseServiceCommand(%q) unexpected error: %v", tt.args, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseServiceCommand(%q) = %+v, want %+v", tt.args, got, tt.want)
			}
		})
	}
}

func TestParseServiceCommandHelp(t *testing.T) {
	var out bytes.Buffer
	_, err := parseServiceCommand([]string{"--help"}, &out)
	if !errors.Is(err, pflag.ErrHelp) {
		t.Fatalf("error = %v, want pflag.ErrHelp", err)
	}
	for _, want := range []string{"install", "uninstall", "--config.file", "--name"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("usage output missing %q:\n%s", want, out.String())
		}
	}
}

func TestServiceRunArgsRoundTrip(t *testing.T) {
	cfg := `C:\Program Files\mssql exporter\config.yml`
	args := serviceRunArgs("mssql-exporter-inst02", cfg)

	name, rest := splitServiceName(args)
	if name != "mssql-exporter-inst02" {
		t.Errorf("service name = %q, want %q", name, "mssql-exporter-inst02")
	}
	if want := []string{"--config.file", cfg}; !reflect.DeepEqual(rest, want) {
		t.Errorf("remaining args = %q, want %q", rest, want)
	}
}

func TestSplitServiceNameDefault(t *testing.T) {
	name, rest := splitServiceName([]string{"--config.file", "c.yml"})
	if name != defaultServiceName {
		t.Errorf("service name = %q, want %q", name, defaultServiceName)
	}
	if want := []string{"--config.file", "c.yml"}; !reflect.DeepEqual(rest, want) {
		t.Errorf("remaining args = %q, want %q", rest, want)
	}
}
