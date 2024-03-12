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

	err := shell.DockerPush(
		shell.WithDryRun(p.DryRunEnabled),
		shell.WithImageTag(p.config.ImageTag),
		shell.WithStderr(p.Stderr),
		shell.WithStdout(p.Stdout),
		shell.WithDockerAlias(alias),
	)

	if err != nil {
		slog.Error("failed to push image tag to registry", "image_tag", p.config.ImageTag)
		return errors.New("Image Publish Pipeline failed.")
	}

	if !p.config.ImagePublish.BundlePublishEnabled {
		slog.Warn("bundle image publish is disabled, skip")
		return nil

	}

	if p.config.ImagePublish.BundleTag == "" {
		return errors.New("Image Publish Pipeline failed: no artifact image defined for image publish")
	}

	bundleFilename := path.Join(p.config.ArtifactDir, p.config.GatecheckBundleFilename)

	err = shell.OrasPushBundle(
		shell.WithDryRun(p.DryRunEnabled),
		shell.WithIO(nil, p.Stdout, p.Stderr),
		shell.WithBundleImage(p.config.ImagePublish.BundleTag, bundleFilename),
	)

	if err != nil {
		slog.Error("failed to push image artifact bundle to registry",
			"image_tag", p.config.ImagePublish.BundleTag, "bundle_filename", bundleFilename)
		return errors.New("Image Publish Pipeline failed.")
	}

	return nil
}
