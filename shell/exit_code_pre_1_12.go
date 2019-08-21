// +build !go1.12

package shell

func extractExitCode(err error) int {
	return 0
}
