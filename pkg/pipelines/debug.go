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
	l.Info("start")

	// Collect errors for mandatory commands
	errs := errors.Join(
		shell.GrypeCommand(d.Stdout, d.Stderr).Version().WithDryRun(d.DryRunEnabled).Run(),
		shell.SyftCommand(d.Stdout, d.Stderr).Version().WithDryRun(d.DryRunEnabled).Run(),
	)

	// Just log errors for optional commands
	shell.PodmanCommand(d.Stdout, d.Stderr).Version().WithDryRun(d.DryRunEnabled).RunLogErrorAsWarning()
	shell.DockerCommand(d.Stdout, d.Stderr).Version().WithDryRun(d.DryRunEnabled).RunLogErrorAsWarning()

	l.Info("complete")
	return errs
}
