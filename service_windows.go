//go:build windows

package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/pflag"
	"github.com/viniciussouzao/mssql-exporter/internal/config"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

const (
	serviceDisplayName = "mssql-exporter (Prometheus SQL Server exporter)"
	serviceDescription = "Exposes SQL Server metrics for Prometheus. https://github.com/viniciussouzao/mssql-exporter"

	// stateWaitTimeout bounds how long start/stop/uninstall wait for the
	// service to reach the requested state.
	stateWaitTimeout = 30 * time.Second
)

// serviceRecoveryActions restarts the service after a crash or a non-zero
// exit; the failure count resets after a day without failures.
var serviceRecoveryActions = []mgr.RecoveryAction{
	{Type: mgr.ServiceRestart, Delay: 5 * time.Second},
	{Type: mgr.ServiceRestart, Delay: 30 * time.Second},
	{Type: mgr.ServiceRestart, Delay: 60 * time.Second},
}

const serviceRecoveryResetSeconds = 24 * 60 * 60

func isWindowsService() bool {
	ok, err := svc.IsWindowsService()
	return err == nil && ok
}

// runAsWindowsService runs the exporter under the Service Control Manager,
// logging to the Windows Event Log since a service has no console.
func runAsWindowsService(args []string) int {
	name, rest := splitServiceName(args)

	elog, err := eventlog.Open(name)
	if err == nil {
		defer func() { _ = elog.Close() }()
		slog.SetDefault(slog.New(newEventLogHandler(elog, slog.LevelInfo)))
	}

	if err := svc.Run(name, &exporterService{args: rest}); err != nil {
		slog.Error("windows service failed", slog.Any("error", err))
		return 1
	}
	return 0
}

type exporterService struct {
	args []string
}

// Execute implements svc.Handler.
func (s *exporterService) Execute(_ []string, requests <-chan svc.ChangeRequest, status chan<- svc.Status) (bool, uint32) {
	status <- svc.Status{State: svc.StartPending}

	cfg, err := config.Load(s.args)
	if err != nil {
		slog.Error("failed to load configuration", slog.Any("error", err))
		return true, 1
	}

	// Services start in System32: make paths relative to the config file
	// (e.g. web.config_file) resolve next to it instead.
	if cfg.ConfigFile != "" {
		if err := os.Chdir(filepath.Dir(cfg.ConfigFile)); err != nil {
			slog.Warn("failed to change to the config file directory", slog.Any("error", err))
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runErr := make(chan error, 1)
	go func() { runErr <- run(ctx, cfg) }()

	status <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}

	for {
		select {
		case err := <-runErr:
			// The exporter stopped on its own: report a service-specific
			// error so the SCM recovery actions restart it.
			slog.Error("exporter stopped unexpectedly", slog.Any("error", err))
			return true, 1
		case req := <-requests:
			switch req.Cmd {
			case svc.Interrogate:
				status <- req.CurrentStatus
			case svc.Stop, svc.Shutdown:
				status <- svc.Status{State: svc.StopPending, WaitHint: uint32((shutdownTimeout + 5*time.Second).Milliseconds())}
				cancel()
				select {
				case err := <-runErr:
					if err != nil {
						slog.Error("exporter shutdown failed", slog.Any("error", err))
						return true, 1
					}
				case <-time.After(shutdownTimeout + 5*time.Second):
					slog.Warn("exporter did not stop in time")
				}
				return false, 0
			}
		}
	}
}

func runServiceCommand(args []string) int {
	cmd, err := parseServiceCommand(args, os.Stderr)
	if err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			return 0
		}
		fmt.Fprintln(os.Stderr, "error:", err)
		return 2
	}

	switch cmd.Action {
	case "install":
		err = installService(cmd.Name, cmd.ConfigFile)
	case "uninstall":
		err = uninstallService(cmd.Name)
	case "start":
		err = startService(cmd.Name)
	case "stop":
		err = stopService(cmd.Name)
	case "status":
		err = printServiceStatus(cmd.Name)
	}
	if err != nil {
		if errors.Is(err, windows.ERROR_ACCESS_DENIED) {
			err = fmt.Errorf("%w (run from an elevated prompt: \"Run as administrator\")", err)
		}
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	return 0
}

func connectManager() (*mgr.Mgr, error) {
	m, err := mgr.Connect()
	if err != nil {
		return nil, fmt.Errorf("connect to the service control manager: %w", err)
	}
	return m, nil
}

func openService(m *mgr.Mgr, name string) (*mgr.Service, error) {
	s, err := m.OpenService(name)
	if err != nil {
		if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
			return nil, fmt.Errorf("service %q is not installed", name)
		}
		return nil, fmt.Errorf("open service %q: %w", name, err)
	}
	return s, nil
}

func installService(name, configFile string) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate executable: %w", err)
	}
	if exe, err = filepath.Abs(exe); err != nil {
		return fmt.Errorf("locate executable: %w", err)
	}
	if configFile, err = filepath.Abs(configFile); err != nil {
		return fmt.Errorf("resolve config file: %w", err)
	}

	// Fail now rather than in the Event Log on first start.
	if _, err := config.Load([]string{"--config.file", configFile}); err != nil {
		return fmt.Errorf("invalid config file %s: %w", configFile, err)
	}

	m, err := connectManager()
	if err != nil {
		return err
	}
	defer func() { _ = m.Disconnect() }()

	if s, err := m.OpenService(name); err == nil {
		_ = s.Close()
		return fmt.Errorf("service %q already exists; run `service uninstall --name %s` first", name, name)
	}

	s, err := m.CreateService(name, exe, mgr.Config{
		DisplayName:      serviceDisplayName,
		Description:      serviceDescription,
		StartType:        mgr.StartAutomatic,
		DelayedAutoStart: true,
	}, serviceRunArgs(name, configFile)...)
	if err != nil {
		return fmt.Errorf("create service: %w", err)
	}
	defer func() { _ = s.Close() }()

	if err := s.SetRecoveryActions(serviceRecoveryActions, serviceRecoveryResetSeconds); err != nil {
		_ = s.Delete()
		return fmt.Errorf("configure recovery actions: %w", err)
	}
	// Also restart when the exporter exits with an error, not only on crashes.
	if err := s.SetRecoveryActionsOnNonCrashFailures(true); err != nil {
		_ = s.Delete()
		return fmt.Errorf("configure recovery actions: %w", err)
	}

	if err := eventlog.InstallAsEventCreate(name, eventlog.Error|eventlog.Warning|eventlog.Info); err != nil {
		// Already registered (e.g. a previous install): logging still works.
		fmt.Fprintf(os.Stderr, "warning: event log source %q not registered: %v\n", name, err)
	}

	fmt.Printf("Service %q installed (automatic delayed start, restarts on failure).\n", name)
	fmt.Printf("  executable:  %s\n  config file: %s\n", exe, configFile)
	fmt.Printf("Start it with: mssql-exporter service start --name %s\n", name)
	fmt.Printf("Logs go to the Windows Event Log (Application, source %q).\n", name)
	return nil
}

func uninstallService(name string) error {
	m, err := connectManager()
	if err != nil {
		return err
	}
	defer func() { _ = m.Disconnect() }()

	s, err := openService(m, name)
	if err != nil {
		return err
	}
	defer func() { _ = s.Close() }()

	if st, err := s.Query(); err == nil && st.State != svc.Stopped {
		if err := controlAndWait(s, svc.Stop, svc.Stopped); err != nil {
			return err
		}
	}

	if err := s.Delete(); err != nil {
		return fmt.Errorf("delete service: %w", err)
	}
	if err := eventlog.Remove(name); err != nil {
		fmt.Fprintf(os.Stderr, "warning: event log source %q not removed: %v\n", name, err)
	}
	fmt.Printf("Service %q removed.\n", name)
	return nil
}

func startService(name string) error {
	m, err := connectManager()
	if err != nil {
		return err
	}
	defer func() { _ = m.Disconnect() }()

	s, err := openService(m, name)
	if err != nil {
		return err
	}
	defer func() { _ = s.Close() }()

	if err := s.Start(); err != nil {
		return fmt.Errorf("start service: %w", err)
	}
	if err := waitForState(s, svc.Running); err != nil {
		return err
	}
	fmt.Printf("Service %q is running.\n", name)
	return nil
}

func stopService(name string) error {
	m, err := connectManager()
	if err != nil {
		return err
	}
	defer func() { _ = m.Disconnect() }()

	s, err := openService(m, name)
	if err != nil {
		return err
	}
	defer func() { _ = s.Close() }()

	if err := controlAndWait(s, svc.Stop, svc.Stopped); err != nil {
		return err
	}
	fmt.Printf("Service %q stopped.\n", name)
	return nil
}

func printServiceStatus(name string) error {
	m, err := connectManager()
	if err != nil {
		return err
	}
	defer func() { _ = m.Disconnect() }()

	s, err := openService(m, name)
	if err != nil {
		return err
	}
	defer func() { _ = s.Close() }()

	st, err := s.Query()
	if err != nil {
		return fmt.Errorf("query service: %w", err)
	}
	c, err := s.Config()
	if err != nil {
		return fmt.Errorf("read service config: %w", err)
	}
	fmt.Printf("Service:  %s\nState:    %s\nPID:      %d\nCommand:  %s\n", name, stateName(st.State), st.ProcessId, c.BinaryPathName)
	return nil
}

func controlAndWait(s *mgr.Service, c svc.Cmd, want svc.State) error {
	if _, err := s.Control(c); err != nil {
		if errors.Is(err, windows.ERROR_SERVICE_NOT_ACTIVE) {
			return nil
		}
		return fmt.Errorf("control service: %w", err)
	}
	return waitForState(s, want)
}

func waitForState(s *mgr.Service, want svc.State) error {
	deadline := time.Now().Add(stateWaitTimeout)
	for {
		st, err := s.Query()
		if err != nil {
			return fmt.Errorf("query service: %w", err)
		}
		if st.State == want {
			return nil
		}
		if want == svc.Running && st.State == svc.Stopped {
			return errors.New("service stopped right after starting; check the Windows Event Log (Application) for the cause")
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for state %s (current: %s)", stateName(want), stateName(st.State))
		}
		time.Sleep(300 * time.Millisecond)
	}
}

func stateName(s svc.State) string {
	switch s {
	case svc.Stopped:
		return "stopped"
	case svc.StartPending:
		return "start pending"
	case svc.StopPending:
		return "stop pending"
	case svc.Running:
		return "running"
	case svc.ContinuePending:
		return "continue pending"
	case svc.PausePending:
		return "pause pending"
	case svc.Paused:
		return "paused"
	default:
		return fmt.Sprintf("unknown (%d)", s)
	}
}
