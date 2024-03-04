package pipelines

import (
	"errors"
	"io"
	"log/slog"
	"workflow-engine/pkg/shell"
)

type dockerOrAliasCommand interface {
	Version() *shell.Command
	Info() *shell.Command
	Push(string) *shell.Command
	Build(*shell.ImageBuildOptions) *shell.Command
}

type ImageBuild struct {
	Stdout        io.Writer
	Stderr        io.Writer
	DryRunEnabled bool
	DockerOrAlias dockerOrAliasCommand
	config        *Config
}

func NewImageBuild(stdout io.Writer, stderr io.Writer) *ImageBuild {
	pipeline := &ImageBuild{
		Stdout:        stdout,
		Stderr:        stderr,
		DryRunEnabled: false,
		DockerOrAlias: shell.DockerCommand(nil, stdout, stderr),
	}

	pipeline.DockerOrAlias = shell.DockerCommand(nil, pipeline.Stdout, pipeline.Stderr)

	return pipeline
}

func (p *ImageBuild) WithPodman() *ImageBuild {
	slog.Debug("use podman cli")
	p.DockerOrAlias = shell.PodmanCommand(nil, p.Stdout, p.Stderr)
	return p
}

func (p *ImageBuild) WithBuildConfig(config *Config) *ImageBuild {
	slog.Debug("apply build config")
	p.config = config
	return p
}

func (p *ImageBuild) Run() error {
	slog.Info("run image build pipeline", "dry_run_enabled", p.DryRunEnabled, "artifact_directory", p.config.ArtifactsDir)

	// print the connection information, exit pipeline if failed
	err := p.DockerOrAlias.Info().WithDryRun(p.DryRunEnabled).Run()
	if err != nil {
		slog.Error("cannot print docker/podman system information which is likely due to bad engine connection")
		return errors.New("Image Build Pipeline failed to run. See log for details.")
	}

	buildOpts := shell.NewImageBuildOptions().
		WithBuildDir(p.config.ImageBuild.BuildDir).
		WithBuildFile(p.config.ImageBuild.Dockerfile).
		WithBuildArgs(p.config.ImageBuild.Args).
		WithTag(p.config.ImageBuild.Tag).
		WithBuildPlatform(p.config.ImageBuild.Platform).
		WithBuildTarget(p.config.ImageBuild.Target).
		WithCache(p.config.ImageBuild.CacheTo, p.config.ImageBuild.CacheFrom)

	err = p.DockerOrAlias.Build(buildOpts).WithDryRun(p.DryRunEnabled).Run()
	if err != nil {
		return errors.New("Image Build Pipeline ran but failed. See log for details.")
	}
	return nil
}
