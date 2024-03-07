package pipelines

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	shell "workflow-engine/pkg/shell"
	legacyShell "workflow-engine/pkg/shell/legacy"
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
	commonOptions := []shell.OptionFunc{shell.WithDryRun(d.DryRunEnabled), shell.WithStdout(d.Stdout)}
	errs := errors.Join(
		legacyShell.GrypeCommand(nil, d.Stdout, d.Stderr).Version().WithDryRun(d.DryRunEnabled).Run(),
		legacyShell.SyftCommand(nil, d.Stdout, d.Stderr).Version().WithDryRun(d.DryRunEnabled).Run(),
		legacyShell.GitleaksCommand(nil, d.Stdout, d.Stderr).Version().WithDryRun(d.DryRunEnabled).Run(),
		legacyShell.GatecheckCommand(nil, d.Stdout, d.Stderr).Version().WithDryRun(d.DryRunEnabled).Run(),
		legacyShell.OrasCommand(nil, d.Stdout, d.Stderr).Version().WithDryRun(d.DryRunEnabled).Run(),
		legacyShell.ClamScanCommand(nil, d.Stdout, d.Stderr).Version().WithDryRun(d.DryRunEnabled).Run(),
		shell.GrypeVersion(commonOptions...).GetError("grype"),
		shell.SyftVersion(commonOptions...).GetError("syft"),
		shell.ClamScanVersion(commonOptions...).GetError("clamscan"),
		shell.FreshClamVersion(commonOptions...).GetError("freshclam"),
	)

	// Just log errors for optional commands
	legacyShell.PodmanCommand(nil, d.Stdout, d.Stderr).Version().WithDryRun(d.DryRunEnabled).RunLogErrorAsWarning()
	legacyShell.DockerCommand(nil, d.Stdout, d.Stderr).Version().WithDryRun(d.DryRunEnabled).RunLogErrorAsWarning()

	return errs
}
