package pipelines

import (
	"errors"
	"io"
	"log/slog"
	"strings"
	"workflow-engine/pkg/shell"
)

type ImageBuild struct {
	Stdout        io.Writer
	Stderr        io.Writer
	DockerAlias   string
	DryRunEnabled bool
	config        *Config
}

func NewImageBuild(stdout io.Writer, stderr io.Writer) *ImageBuild {
	pipeline := &ImageBuild{
		Stdout:        stdout,
		Stderr:        stderr,
		DryRunEnabled: false,
		DockerAlias:   "docker",
	}

	return pipeline
}

func (p *ImageBuild) WithBuildConfig(config *Config) *ImageBuild {
	slog.Debug("apply build config")
	p.config = config
	return p
}

func (p *ImageBuild) Run() error {
	slog.Info("run image build pipeline", "dry_run_enabled", p.DryRunEnabled, "artifact_directory", p.config.ArtifactsDir, "alias", p.DockerAlias)

	alias := shell.DockerAliasDocker
	// print the connection information, exit pipeline if failed
	switch strings.ToLower(p.DockerAlias) {
	case "podman":
		alias = shell.DockerAliasPodman
	case "docker":
		alias = shell.DockerAliasDocker
	}

	// "" values will be stripped out
	buildOpts := shell.ImageBuildOptions{
		ImageName:    p.config.ImageBuild.Tag,
		BuildDir:     p.config.ImageBuild.BuildDir,
		Dockerfile:   p.config.ImageBuild.Dockerfile,
		Target:       p.config.ImageBuild.Target,
		Platform:     p.config.ImageBuild.Platform,
		SquashLayers: p.config.ImageBuild.SquashLayers,
		CacheTo:      p.config.ImageBuild.CacheTo,
		CacheFrom:    p.config.ImageBuild.CacheFrom,
	}

	opts := []shell.OptionFunc{
		shell.WithDockerAlias(alias),
		shell.WithStdout(p.Stdout),
		shell.WithStderr(p.Stderr),
		shell.WithDryRun(p.DryRunEnabled),
	}

	exitCode := shell.DockerInfo()

	if exitCode != shell.ExitOK {
		return errors.New("Image Build Pipeline ran but failed. See log for details.")
	}

	opts = append(opts, shell.WithBuildImageOptions(buildOpts))
	exitCode = shell.DockerBuild(opts...)

	if exitCode != shell.ExitOK {
		return errors.New("Image Build Pipeline ran but failed. See log for details.")
	}

	return nil
}
