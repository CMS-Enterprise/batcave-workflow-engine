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

func (p *imagePublish) WithArtifactConfig(config ArtifactConfig) *imagePublish {
	if config.Directory != "" {
		p.artifactConfig.Directory = config.Directory
	}
	if config.AntivirusFilename != "" {
		p.artifactConfig.AntivirusFilename = config.AntivirusFilename
	}
	if config.GatecheckBundleFilename != "" {
		p.artifactConfig.GatecheckBundleFilename = config.GatecheckBundleFilename
	}
	if config.GatecheckConfigFilename != "" {
		p.artifactConfig.GatecheckConfigFilename = config.GatecheckConfigFilename
	}
	if config.GitleaksFilename != "" {
		p.artifactConfig.GitleaksFilename = config.GitleaksFilename
	}
	if config.GrypeFilename != "" {
		p.artifactConfig.GrypeFilename = config.GrypeFilename
	}
	if config.GrypeConfigFilename != "" {
		p.artifactConfig.GrypeConfigFilename = config.GrypeConfigFilename
	}
	if config.GrypeActiveFindingsFilename != "" {
		p.artifactConfig.GrypeActiveFindingsFilename = config.GrypeActiveFindingsFilename
	}
	if config.GrypeAllFindingsFilename != "" {
		p.artifactConfig.GrypeAllFindingsFilename = config.GrypeAllFindingsFilename
	}
	if config.SBOMFilename != "" {
		p.artifactConfig.SBOMFilename = config.SBOMFilename
	}
	if config.SemgrepFilename != "" {
		p.artifactConfig.SemgrepFilename = config.SemgrepFilename
	}
	if config.ClamavFilename != "" {
		p.artifactConfig.ClamavFilename = config.ClamavFilename
	}

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
	clamavFilename := path.Join(p.config.ArtifactsDir, p.config.ImageScan.ClamavFilename)

	fmt.Fprintln(p.Stderr, dockerFilename, gitleaksFilename, semgrepFilename, antivirusFilename, grypeFilename, grypeConfigFilename, sbomFilename, clamavFilename)

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
