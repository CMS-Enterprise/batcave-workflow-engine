package pipelines

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path"
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
	dockerFilename := p.config.ImageBuild.Dockerfile
	gitleaksFilename := path.Join(p.config.ArtifactsDir, p.config.CodeScan.GitleaksFilename)
	semgrepFilename := path.Join(p.config.ArtifactsDir, p.config.CodeScan.SemgrepFilename)
	antivirusFilename := path.Join(p.config.ArtifactsDir, p.config.ImageScan.AntivirusFilename)
	grypeFilename := path.Join(p.config.ArtifactsDir, p.config.ImageScan.GrypeFullFilename)
	grypeConfigFilename := path.Join(p.config.ArtifactsDir, p.config.ImageScan.GrypeConfigFilename)
	sbomFilename := path.Join(p.config.ArtifactsDir, p.config.ImageScan.SyftFilename)

	fmt.Fprintln(p.Stderr, dockerFilename, gitleaksFilename, semgrepFilename, antivirusFilename, grypeFilename, grypeConfigFilename, sbomFilename)

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
