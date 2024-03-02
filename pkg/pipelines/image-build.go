package pipelines

import (
	"io"
	"log/slog"
	"workflow-engine/pkg/shell"
)

type cliCmd interface {
	Version() *shell.Command
	Info() *shell.Command
	Push(string) *shell.Command
	Build(*shell.ImageBuildOptions) *shell.Command
}

type ImageBuild struct {
	Stdout        io.Writer
	Stderr        io.Writer
	DryRunEnabled bool
	CLICmd        cliCmd
	logger        *slog.Logger
	cfg           ConfigImage
}

func NewImageBuild(stdout io.Writer, stderr io.Writer) *ImageBuild {
	pipeline := &ImageBuild{
		Stdout:        stdout,
		Stderr:        stderr,
		DryRunEnabled: false,
		logger:        slog.Default().With("pipeline", "image_build"),
	}

	pipeline.CLICmd = shell.DockerCommand(nil, pipeline.Stdout, pipeline.Stderr)

	return pipeline
}

func (i *ImageBuild) WithPodman() *ImageBuild {
	i.logger.Debug("use podman cli")
	i.CLICmd = shell.PodmanCommand(nil, i.Stdout, i.Stderr)
	return i
}

func (i *ImageBuild) WithBuildConfig(buildConfig ConfigImage) *ImageBuild {
	i.logger.Debug("apply build config")
	i.cfg = buildConfig
	return i
}

func (i *ImageBuild) Run() error {
	l := slog.Default()

	l.Info("start", "dry_run_enabled", i.DryRunEnabled)
	// defer will run right before the return of this function, even for early returns due to errors
	defer l.Info("complete")

	// print the connection information, exit pipeline if failed
	err := i.CLICmd.Info().WithDryRun(i.DryRunEnabled).Run()
	if err != nil {
		return err
	}

	buildOpts := shell.NewImageBuildOptions().
		WithBuildDir(i.cfg.BuildDir).
		WithBuildFile(i.cfg.BuildDockerfile).
		WithBuildArgs(i.cfg.BuildArgs).
		WithTag(i.cfg.BuildTag).
		WithBuildPlatform(i.cfg.BuildPlatform).
		WithBuildTarget(i.cfg.BuildTarget).
		WithCache(i.cfg.BuildCacheTo, i.cfg.BuildCacheFrom)

	return i.CLICmd.Build(buildOpts).WithDryRun(i.DryRunEnabled).Run()
}
