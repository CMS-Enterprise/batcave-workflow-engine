package shell

import (
	"fmt"
	"io"
)

type semgrepCmd struct {
	InitCmd func() *Executable
}

// Version outputs the version of the Semgrep CLI
// // shell: `semgrep version`
func (s *semgrepCmd) Version() *Command {
	return NewCommand(s.InitCmd().WithArgs("--version"))
}

// ScanFile runs a Semgrep scan against a target artifact dir
//
// shell: `semgrep ci --json > ${ARTIFACT_FOLDER}/sast/semgrep-sast-report.json || true`
func (s *semgrepCmd) ScanFile(filename string) *Command {
	exe := s.InitCmd().WithArgs(
		fmt.Sprintf("ci --json > %s", filename),
		"|| true",
	)

	return NewCommand(exe)
}

// Semgrep Command with custom stdout and stderr
func SemgrepCommand(stdout io.Writer, stderr io.Writer) *semgrepCmd {
	return &semgrepCmd{
		InitCmd: func() *Executable {
			return NewExecutable("semgrep").WithOutput(stdout).WithStderr(stderr)
		},
	}
}
