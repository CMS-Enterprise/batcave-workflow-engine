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
// shell: `semgrep ci --json`
func (s *semgrepCmd) Scan(rules string) *Command {
	args := []string{"ci", "--json"}
	if rules != "" {
		args = append(args, "--config", rules)
	}
	exe := s.InitCmd().WithArgs(args...)
	if s.experimental {
		exe = s.InitCmd().WithArgs(append(args, "--experimental")...)
	}
	return NewCommand(exe)
}

// Semgrep Command with custom stdout and stderr
func SemgrepCommand(stdin io.Reader, stdout io.Writer, stderr io.Writer) *semgrepCmd {
	return &semgrepCmd{
		InitCmd: func() *Executable {
			return NewExecutable("semgrep").WithIO(stdin, stdout, stderr)
		},
	}
}

// OSemgrep Command with custom stdout and stderr
func OSemgrepCommand(stdin io.Reader, stdout io.Writer, stderr io.Writer) *semgrepCmd {
	return &semgrepCmd{
		InitCmd: func() *Executable {
			return NewExecutable("osemgrep").WithIO(stdin, stdout, stderr)
		},
	}
}
