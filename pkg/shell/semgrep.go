package shell

import (
	"io"
)

type semgrepCmd struct {
	InitCmd      func() *Executable
	experimental bool
}

// Version outputs the version of the Semgrep CLI
// // shell: `semgrep version`
func (s *semgrepCmd) Version() *Command {
	if s.experimental {
		return NewCommand(s.InitCmd().WithArgs("--help"))
	}
	return NewCommand(s.InitCmd().WithArgs("--version"))
}

// ScanFile runs a Semgrep scan against target artifact dir from env vars
//
// shell: `semgrep ci --json > ${ARTIFACT_FOLDER}/sast/semgrep-sast-report.json || true`
func (s *semgrepCmd) ScanFile() *Command {
	exe := s.InitCmd().WithArgs("ci", "--json")
	if s.experimental {
		exe = s.InitCmd().WithArgs("ci", "--json", "--experimental")
	}
	return NewCommand(exe)
}

// Semgrep Command with custom stdout and stderr
func SemgrepCommand(stdin io.Reader, stdout io.Writer, stderr io.Writer) *semgrepCmd {
	return &semgrepCmd{
		InitCmd: func() *Executable {
			return NewExecutable("semgrep").WithStdin(stdin).WithOutput(stdout).WithStderr(stderr)
		},
	}
}

// OSemgrep Command with custom stdout and stderr
func OSemgrepCommand(stdin io.Reader, stdout io.Writer, stderr io.Writer) *semgrepCmd {
	return &semgrepCmd{
		InitCmd: func() *Executable {
			return NewExecutable("osemgrep").WithStdin(stdin).WithOutput(stdout).WithStderr(stderr)
		},
	}
}
