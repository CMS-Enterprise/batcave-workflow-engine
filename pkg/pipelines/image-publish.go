package pipelines

import (
	"errors"
	"io"
	"log/slog"
	"path"
	"strings"
	"workflow-engine/pkg/shell"
)

type ImagePublish struct {
	Stdout        io.Writer
	Stderr        io.Writer
	DryRunEnabled bool
	DockerAlias   string
	config        *Config
	runtime       struct {
		bundleFilename       string
		dockerfileKey        string
		sbomKey              string
		clamavKey            string
		gatecheckConfigKey   string
		gatecheckManifestKey string
		grypeReportKey       string
		semgrepReportKey     string
		gitleaksKey          string
	}
}

func (p *ImagePublish) WithConfig(config *Config) *ImagePublish {
	p.config = config
	return p
}

func NewimagePublish(stdout io.Writer, stderr io.Writer) *ImagePublish {
	pipeline := &ImagePublish{
		Stdout:        stdout,
		Stderr:        stderr,
		DryRunEnabled: false,
	}
	return pipeline
}

func (p *ImagePublish) preRun() error {
	// numbers for date format is From the docs: https://go.dev/src/time/format.go
	p.runtime.bundleFilename = path.Join(p.config.ArtifactDir, p.config.GatecheckBundleFilename)
	return nil
}

func (p *ImagePublish) Run() error {
	if !p.config.ImagePublish.Enabled {
		slog.Warn("image publish pipeline is disabled, skip.")
		return nil
	}

	alias := shell.DockerAliasDocker
	switch strings.ToLower(p.DockerAlias) {
	case "podman":
		alias = shell.DockerAliasPodman
	case "docker":
		alias = shell.DockerAliasDocker
	}

	if err := p.preRun(); err != nil {
		return errors.New("Code Scan Pipeline Pre-Run Failed.")
	}

	exitCode := shell.DockerPush(
		shell.WithDryRun(p.DryRunEnabled),
		shell.WithImage(p.config.ImageBuild.Tag),
		shell.WithStderr(p.Stderr),
		shell.WithDockerAlias(alias),
	)

	if exitCode != shell.ExitOK {
		slog.Error("failed to push image tag to registry", "image_tag", p.config.ImageBuild.Tag)
		return errors.New("Image Publish Pipeline failed.")
	}

	image, bundle := p.config.ImagePublish.ArtifactImage, p.runtime.bundleFilename
	exitCode = shell.OrasPushBundle(
		shell.WithIO(nil, p.Stdout, p.Stderr),
		shell.WithArtifactBundle(image, bundle),
	)

	if exitCode != shell.ExitOK {
		slog.Error("failed to push image artifact bundle to registry", "image_tag", image, "bundle_filename", bundle)
		return errors.New("Image Publish Pipeline failed.")
	}

	return nil
}
