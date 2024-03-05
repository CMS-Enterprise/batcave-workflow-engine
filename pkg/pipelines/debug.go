package pipelines

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"workflow-engine/pkg/shell"
)

type Debug struct {
	Stdin         io.Reader
	Stdout        io.Writer
	Stderr        io.Writer
	DryRunEnabled bool
}

// NewDebug creates a new debug pipeline with custom stdout and stderr
func NewDebug(stdoutW io.Writer, stderrW io.Writer) *Debug {
	return &Debug{Stdin: os.Stdin, Stdout: stdoutW, Stderr: stderrW, DryRunEnabled: false}
}

// Run prints the version for all expected commands
//
// All commands will run in sequence, stopping if one of the commands fail
func (d *Debug) Run() error {
	slog.Info("start debug pipeline", "dry_run", d.DryRunEnabled)

	// Get current directory
	wd, err := os.Getwd()
	if err != nil {
		slog.Error("cannot get current working directory", "error", err)
		return errors.New("Debug Pipeline failed to Run. See log for details.")
	}
	slog.Info(fmt.Sprintf("Current directory: %s", wd))

	// Collect errors for mandatory commands
	errs := errors.Join(
		shell.GrypeCommand(nil, d.Stdout, d.Stderr).Version().WithDryRun(d.DryRunEnabled).Run(),
		shell.SyftCommand(nil, d.Stdout, d.Stderr).Version().WithDryRun(d.DryRunEnabled).Run(),
		shell.GitleaksCommand(nil, d.Stdout, d.Stderr).Version().WithDryRun(d.DryRunEnabled).Run(),
		shell.GatecheckCommand(nil, d.Stdout, d.Stderr).Version().WithDryRun(d.DryRunEnabled).Run(),
		shell.OrasCommand(nil, d.Stdout, d.Stderr).Version().WithDryRun(d.DryRunEnabled).Run(),
	)

	// Just log errors for optional commands
	shell.PodmanCommand(nil, d.Stdout, d.Stderr).Version().WithDryRun(d.DryRunEnabled).RunLogErrorAsWarning()
	shell.DockerCommand(nil, d.Stdout, d.Stderr).Version().WithDryRun(d.DryRunEnabled).RunLogErrorAsWarning()

	return errs
}
