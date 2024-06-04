package pipelines

import (
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
	if !p.config.ImageBuild.Enabled {
		slog.Warn("image build pipeline is disabled, skip.")
		return nil
	}

	slog.Info("run image build pipeline", "dry_run_enabled", p.DryRunEnabled, "artifact_directory", p.config.ArtifactDir, "alias", p.DockerAlias)

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
		Tag:          p.config.ImageTag,
		BuildDir:     p.config.ImageBuild.BuildDir,
		Dockerfile:   p.config.ImageBuild.Dockerfile,
		Target:       p.config.ImageBuild.Target,
		Platform:     p.config.ImageBuild.Platform,
		SquashLayers: p.config.ImageBuild.SquashLayers,
		CacheTo:      p.config.ImageBuild.CacheTo,
		CacheFrom:    p.config.ImageBuild.CacheFrom,
		BuildArgs:    p.config.ImageBuild.Args,
	}

	opts := []shell.OptionFunc{
		shell.WithDockerAlias(alias),
		shell.WithStdout(p.Stdout),
		shell.WithStderr(p.Stderr),
		shell.WithDryRun(p.DryRunEnabled),
		shell.WithBuildImageOptions(buildOpts),
	}

	err := shell.DockerInfo(opts...)
	if err != nil {
		return err
	}

	opts = append(opts, shell.WithBuildImageOptions(buildOpts))
	err = shell.DockerBuild(opts...)

	if err != nil {
		return err
	}

	return nil
}
