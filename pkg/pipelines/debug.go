package pipelines

import (
	"errors"
	"io"
	"log/slog"
	"workflow-engine/pkg/shell"
)

type Debug struct {
	Stdout        io.Writer
	Stderr        io.Writer
	DryRunEnabled bool
}

// NewDebug creates a new debug pipeline with custom stdout and stderr
func NewDebug(stdoutW io.Writer, stderrW io.Writer) *Debug {
	return &Debug{Stdout: stdoutW, Stderr: stderrW, DryRunEnabled: false}
}

// Run prints the version for all expected commands
//
// All commands will run in sequence, stopping if one of the commands fail
func (d *Debug) Run() error {
	l := slog.Default().With("pipeline", "debug", "dry_run", d.DryRunEnabled)
	l.Info("start pipeline")
	errs := errors.Join(
		shell.GrypeCommand(d.Stdout, d.Stderr).Version().RunOptional(d.DryRunEnabled),
		shell.SyftCommand(d.Stdout, d.Stderr).Version().RunOptional(d.DryRunEnabled),
		shell.PodmanComand(d.Stdout, d.Stderr).Version().RunOptional(d.DryRunEnabled),
	)

	l.Info("complete pipeline")
	return errs
}
