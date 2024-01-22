package pipelines

import (
	"io"
	"log/slog"
	"workflow-engine/pkg/shell"
)

type cliCmd interface {
	Version() *shell.Command
	Info() *shell.Command
}

type ImageBuild struct {
	Stdout        io.Writer
	Stderr        io.Writer
	DryRunEnabled bool
	CLICmd        cliCmd
}

func NewImageBuild(stdout io.Writer, stderr io.Writer) *ImageBuild {
	pipeline := &ImageBuild{
		Stdout:        stdout,
		Stderr:        stderr,
		DryRunEnabled: false,
	}

	pipeline.CLICmd = shell.DockerCommand(pipeline.Stdout, pipeline.Stderr)

	return pipeline
}

func (i *ImageBuild) WithPodman() *ImageBuild {
	i.CLICmd = shell.PodmanCommand(i.Stdout, i.Stderr)
	return i
}

func (i *ImageBuild) Run() error {
	l := slog.Default().With("pipeline", "image_build", "dry_run", i.DryRunEnabled)

	l.Info("start")

	// print the connection information
	err := i.CLICmd.Info().WithDryRun(i.DryRunEnabled).Run()
	l.Info("complete")

	return err
}
