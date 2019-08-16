package shell

import (
	"bytes"
	"context"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	cases := []struct {
		cmd      string
		stdin    string
		stdout   *regexp.Regexp
		stderr   *regexp.Regexp
		exitCode int
		timeout  time.Duration
	}{
		{
			cmd:    "tr a-z A-Z",
			stdin:  "the littlest monkey",
			stdout: regexp.MustCompile("^THE LITTLEST MONKEY$"),
			stderr: nil,
		},
		{
			cmd:    "tr a-z A-Z && echo -n man",
			stdin:  "the littlest monkey",
			stdout: regexp.MustCompile("THE LITTLEST MONKEYman"),
			stderr: nil,
		},
		{
			cmd:      "echo invalid &&",
			stdin:    "",
			stdout:   nil,
			stderr:   regexp.MustCompile("syntax error"),
			exitCode: 2,
		},
		{
			cmd:      "echo -n stderr test 1>&2",
			stdin:    "",
			stdout:   nil,
			stderr:   regexp.MustCompile("^stderr test$"),
			exitCode: 0,
		},
		{
			cmd:      "exit 255",
			stdin:    "",
			stdout:   nil,
			stderr:   nil,
			exitCode: 255,
		},
		{
			cmd:      "sleep 2",
			stdin:    "",
			stdout:   nil,
			stderr:   nil,
			exitCode: -1,
			timeout:  100 * time.Millisecond,
		},
	}

	for i, c := range cases {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		task := Task{
			Cmd:    c.cmd,
			Stdout: stdout,
			Stderr: stderr,
		}

		if c.stdin != "" {
			task.Stdin = strings.NewReader(c.stdin)
		}

		ctx := context.Background()
		if c.timeout > 0 {
			var cancel func()
			ctx, cancel = context.WithTimeout(ctx, c.timeout)
			defer cancel()
		}

		err := Run(ctx, task)
		if c.exitCode == 0 {
			if err != nil {
				t.Errorf("%d: unexpected error %v", i, err)
			}
		} else {
			exitCode := 0
			if re, ok := err.(*RunErr); ok {
				exitCode = re.ExitCode
			}

			if exitCode != c.exitCode {
				t.Errorf("%d: got return code %d, want %d", i, exitCode, c.exitCode)
			}
		}

		if c.stdout == nil {
			if stdout.String() != "" {
				t.Errorf("%d: got stdout %q, want %q", i, stdout.String(), "")
			}
		} else {
			if !c.stdout.Match(stdout.Bytes()) {
				t.Errorf("%d: got stdout %q, want match to %q", i, stdout.String(), c.stdout)
			}
		}

		if c.stderr == nil {
			if stderr.String() != "" {
				t.Errorf("%d: got stderr %q, want %q", i, stderr.String(), "")
			}
		} else {
			if !c.stderr.Match(stderr.Bytes()) {
				t.Errorf("%d: got stderr %q, want match to %q", i, stderr.String(), c.stderr)
			}
		}
	}
}
