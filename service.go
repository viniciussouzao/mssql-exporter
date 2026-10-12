package main

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"

	"github.com/spf13/pflag"
)

const defaultServiceName = "mssql-exporter"

// serviceNameFlag is added to the service's command line at install time so
// the running service knows its own name (used as the Event Log source). It
// is consumed before the remaining arguments reach config.Load.
const serviceNameFlag = "--service.name="

var (
	serviceActions   = []string{"install", "uninstall", "start", "stop", "status"}
	validServiceName = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,80}$`)
)

// serviceCommand is a parsed `mssql-exporter service <action>` invocation.
type serviceCommand struct {
	Action     string
	Name       string
	ConfigFile string
}

func serviceUsage(w io.Writer, fs *pflag.FlagSet) {
	_, _ = fmt.Fprintf(w, `Usage: mssql-exporter service <action> [flags]

Manage mssql-exporter as a Windows service (run from an elevated prompt).

Actions:
  install     Register the service (automatic start, restart on failure)
  uninstall   Stop and remove the service
  start       Start the service
  stop        Stop the service
  status      Show the service state

Flags:
%s`, fs.FlagUsages())
}

// parseServiceCommand parses the arguments following `service`. It returns
// pflag.ErrHelp when help was requested (usage already written to out).
func parseServiceCommand(args []string, out io.Writer) (*serviceCommand, error) {
	fs := pflag.NewFlagSet("mssql-exporter service", pflag.ContinueOnError)
	fs.SetOutput(out)
	name := fs.String("name", defaultServiceName, "Windows service name (allows several exporters on one host).")
	configFile := fs.String("config.file", "", "YAML config file the service runs with (required for install).")
	fs.Usage = func() { serviceUsage(out, fs) }

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if fs.NArg() == 0 {
		fs.Usage()
		return nil, errors.New("missing service action")
	}
	if fs.NArg() > 1 {
		return nil, fmt.Errorf("unexpected arguments: %s", strings.Join(fs.Args()[1:], " "))
	}

	cmd := &serviceCommand{Action: fs.Arg(0), Name: *name, ConfigFile: *configFile}

	if !slices.Contains(serviceActions, cmd.Action) {
		return nil, fmt.Errorf("unknown service action %q (want one of: %s)", cmd.Action, strings.Join(serviceActions, ", "))
	}
	if !validServiceName.MatchString(cmd.Name) {
		return nil, fmt.Errorf("invalid service name %q: use letters, digits, '.', '_' or '-'", cmd.Name)
	}
	if cmd.Action == "install" && cmd.ConfigFile == "" {
		return nil, errors.New("--config.file is required for install (the DSN has no CLI flag, so the service needs a config file)")
	}
	if cmd.Action != "install" && cmd.ConfigFile != "" {
		return nil, errors.New("--config.file is only used by install")
	}
	return cmd, nil
}

// serviceRunArgs builds the arguments the installed service starts with.
// configFile must already be an absolute path: a service's working
// directory is System32.
func serviceRunArgs(name, configFile string) []string {
	return []string{serviceNameFlag + name, "--config.file", configFile}
}

// splitServiceName removes the --service.name flag added at install time and
// returns the service name (defaultServiceName when absent) and the rest.
func splitServiceName(args []string) (string, []string) {
	name := defaultServiceName
	rest := make([]string, 0, len(args))
	for _, a := range args {
		if v, ok := strings.CutPrefix(a, serviceNameFlag); ok {
			name = v
			continue
		}
		rest = append(rest, a)
	}
	return name, rest
}
