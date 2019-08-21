// +build go1.12

package shell

import "os/exec"

func extractExitCode(err error) int {
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ProcessState.ExitCode()
	}

	return 0
}
