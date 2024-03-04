package pipelines

import (
	"errors"
	"io"
	"log/slog"
	"workflow-engine/pkg/shell"
)

type imagePublish struct {
	Stdout        io.Writer
	Stderr        io.Writer
	DryRunEnabled bool
	NoPush        bool
	config        *Config
	dockerOrAlias dockerOrAliasCommand
}

func (p *imagePublish) WithConfig(config *Config) *imagePublish {
	p.config = config
	return p
}

func (p *imagePublish) WithPodman() *imagePublish {
	slog.Debug("use podman cli")
	p.dockerOrAlias = shell.PodmanCommand(nil, p.Stdout, p.Stderr)
	return p
}

func NewimagePublish(stdout io.Writer, stderr io.Writer) *imagePublish {
	pipeline := &imagePublish{
		Stdout:        stdout,
		Stderr:        stderr,
		DryRunEnabled: false,
	}

	pipeline.dockerOrAlias = shell.DockerCommand(nil, pipeline.Stdout, pipeline.Stderr)

	return pipeline
}

func (p *imagePublish) Run() error {
	if p.NoPush {
		slog.Warn("pushing is disabled, skip.")
		return nil
	}
	err := p.dockerOrAlias.Push(p.config.ImageBuild.Tag).WithDryRun(p.DryRunEnabled).Run()
	if err != nil {
		slog.Error("failed to push image tag to registry", "image_tag", p.config.ImageBuild.Tag)
		return errors.New("Image Publish Pipeline failed. See log for details.")
	}

	return nil
}
