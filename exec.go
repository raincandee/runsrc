package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// runCommand runs argv with the terminal attached and returns its exit code.
// The error is non-nil only when the command could not be started at all.
func runCommand(dir string, argv []string) (int, error) {
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err == nil {
		return 0, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode(), nil
	}
	return 1, err
}

// requireTool fails early if a tool is missing, so the user gets a clear
// message instead of a half-finished build.
func requireTool(name string) error {
	if _, err := exec.LookPath(name); err != nil {
		return fmt.Errorf("'%s' is not installed or not in PATH", name)
	}
	return nil
}
