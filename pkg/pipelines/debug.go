package pipelines

import (
	"errors"
	"io"
	"workflow-engine/pkg/shell"
)

type Debug struct {
	Stdout io.Writer
	Stderr io.Writer
}

func NewDebug(stdoutW io.Writer, stderrW io.Writer) *Debug {
	return &Debug{Stdout: stdoutW, Stderr: stderrW}
}

func (d *Debug) Run() error {
	// Runs all the commands and collects errors. Will not stop if one fails
	errs := errors.Join(
		shell.GrypeCommand(d.Stdout, d.Stderr).Version(),
		shell.SyftCommand(d.Stdout, d.Stderr).Version(),
	)

	return errs
}
