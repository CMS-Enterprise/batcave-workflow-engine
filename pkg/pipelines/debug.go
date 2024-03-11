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
		return errors.New("Debug Pipeline failed to Run.")
	}
	slog.Info(fmt.Sprintf("Current directory: %s", wd))

	// Collect errors for mandatory commands
	commonOptions := []shell.OptionFunc{shell.WithDryRun(d.DryRunEnabled), shell.WithStdout(d.Stdout)}
	errs := errors.Join(
		shell.GrypeVersion(commonOptions...),
		shell.SyftVersion(commonOptions...),
		shell.FreshClamVersion(commonOptions...),
		shell.ClamScanVersion(commonOptions...),
		shell.GitLeaksVersion(commonOptions...),
		shell.SemgrepVersion(commonOptions...),
		shell.OrasVersion(commonOptions...),
	)

	return errs
}
