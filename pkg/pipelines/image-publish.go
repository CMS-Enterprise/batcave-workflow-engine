package pipelines

import (
	"errors"
	"io"
	"log/slog"
	"path"
	"workflow-engine/pkg/shell"
)

type ImagePublish struct {
	Stdout        io.Writer
	Stderr        io.Writer
	DryRunEnabled bool
	config        *Config
	dockerOrAlias dockerOrAliasCommand
	runtime       struct {
		bundleFilename string
	}
}

func (p *ImagePublish) WithConfig(config *Config) *ImagePublish {
	p.config = config
	return p
}

func (p *ImagePublish) WithPodman() *ImagePublish {
	slog.Debug("use podman cli")
	p.dockerOrAlias = shell.PodmanCommand(nil, p.Stdout, p.Stderr)
	return p
}

func NewimagePublish(stdout io.Writer, stderr io.Writer) *ImagePublish {
	pipeline := &ImagePublish{
		Stdout:        stdout,
		Stderr:        stderr,
		DryRunEnabled: false,
	}

	pipeline.dockerOrAlias = shell.DockerCommand(nil, pipeline.Stdout, pipeline.Stderr)

	return pipeline
}

func (p *ImagePublish) preRun() error {
	p.runtime.bundleFilename = path.Join(p.config.ArtifactsDir, p.config.GatecheckBundleFilename)
	return nil
}

func (p *ImagePublish) Run() error {
	if !p.config.ImagePublish.Enabled {
		slog.Warn("image publish pipeline is disabled, skip.")
		return nil
	}

	if err := p.preRun(); err != nil {
		return errors.New("Code Scan Pipeline Pre-Run Failed. See log for details.")
	}

	err := p.dockerOrAlias.Push(p.config.ImageBuild.Tag).WithDryRun(p.DryRunEnabled).Run()
	if err != nil {
		slog.Error("failed to push image tag to registry", "image_tag", p.config.ImageBuild.Tag)
		return errors.New("Image Publish Pipeline failed. See log for details.")
	}

	cmd := shell.OrasCommand(nil, p.Stdout, p.Stderr).PushBundle(p.config.ImagePublish.ArtifactsImage, p.runtime.bundleFilename)
	err = cmd.WithDryRun(p.DryRunEnabled).Run()
	if err != nil {
		slog.Error("failed to push image artifact bundle to registry",
			"image_tag", p.config.ImagePublish.ArtifactsImage, "bundle_filename", p.runtime.bundleFilename)
		return errors.New("Image Publish Pipeline failed. See log for details.")
	}
	return nil
}
