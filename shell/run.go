// Package shell runs a command in a subshell. It is a simple wrapper around
// os/exec.
package shell

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
)

// Task defines a command to run.
type Task struct {
	// The command to run.
	Cmd string

	// If set, Dir will be the working directory of the command.
	Dir string

	// Stdin will be attached to stdin in the subshell.
	Stdin io.Reader

	// Stdout will receive anything the process writes to standard out.
	// Defaults to os.Stdout.
	Stdout io.Writer

	// Stderr will receive anything the process writes to standard error.
	// Defaults to os.Stderr.
	Stderr io.Writer
}

func (t Task) setDefaults() Task {
	if t.Stdout == nil {
		t.Stdout = os.Stdout
	}

	if t.Stderr == nil {
		t.Stderr = os.Stderr
	}

	return t
}

// Run executes the task and waits for it to complete. The command will be run
// in a bash subshell. If bash isn't available, sh will used. If neither is
// available an error will be returned.
//
// If the context is canceled the process will be killed. Error will always
// have a type of *RunErr or nil.
func Run(ctx context.Context, t Task) error {
	t = t.setDefaults()

	shell, err := findShell()
	if err != nil {
		return &RunErr{Message: err.Error()}
	}

	cmd := exec.CommandContext(ctx, shell, "-c", t.Cmd)
	cmd.Dir = t.Dir
	cmd.Stdin = t.Stdin
	cmd.Stdout = t.Stdout
	cmd.Stderr = t.Stderr

	err = cmd.Run()
	if err != nil {
		re := &RunErr{
			Message:  err.Error(),
			ExitCode: extractExitCode(err),
		}

		return re
	}

	return nil
}

func findShell() (string, error) {
	bash, err := exec.LookPath("bash")
	if err == nil {
		return bash, nil
	}

	sh, err := exec.LookPath("sh")
	if err == nil {
		return sh, nil
	}

	return "", errors.New("no suitable shell found in PATH")
}

type RunErr struct {
	// Process exit code. Only available in go 1.12 or later. It will
	// always be 0 in earlier versions.
	ExitCode int
	Message  string
}

func (r *RunErr) Error() string {
	return r.Message
}
