//go:build !windows

package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

func isWindowsService() bool { return false }

func runAsWindowsService([]string) int { return 1 }

func runServiceCommand(args []string) int {
	if _, err := parseServiceCommand(args, os.Stderr); err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			return 0
		}
		fmt.Fprintln(os.Stderr, "error:", err)
		return 2
	}
	fmt.Fprintln(os.Stderr, "error: the service subcommand is only supported on Windows; use your init system (e.g. a systemd unit) elsewhere")
	return 1
}
