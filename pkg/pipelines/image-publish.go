package pipelines

import (
	"fmt"
	"io"
	"path"
	"workflow-engine/pkg/shell"
)

type imagePublish struct {
	Stdout        io.Writer
	Stderr        io.Writer
	DryRunEnabled bool
	CLICmd        cliCmd
	config        *Config
}

func (p *imagePublish) WithConfig(config *Config) *imagePublish {
	p.config = config
	return p
}

func NewimagePublish(stdout io.Writer, stderr io.Writer) *imagePublish {
	pipeline := &imagePublish{
		Stdout:        stdout,
		Stderr:        stderr,
		DryRunEnabled: false,
	}

	pipeline.CLICmd = shell.DockerCommand(nil, pipeline.Stdout, pipeline.Stderr)

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

	// TODO: Add gatecheck bundle
	fmt.Fprintln(p.Stderr, dockerFilename, gitleaksFilename, semgrepFilename, antivirusFilename, grypeFilename, grypeConfigFilename, sbomFilename)

	return nil
}
