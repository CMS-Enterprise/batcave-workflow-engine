package pipelines

import (
	"context"
	"log/slog"
	"path"

	"dagger.io/dagger"
)

type ImageBuild struct {
	client *dagger.Client
	cfg    Config
}

func NewImageBuildPipeline(c *dagger.Client, cfg Config) *ImageBuild {
	return &ImageBuild{client: c, cfg: cfg}
}

func (p *ImageBuild) Run() error {
	slog.SetDefault(slog.Default().With("pipeline", "image-build"))
	// Load the directory on the Host machine where the files are
	src := p.client.Host().Directory(p.cfg.Image.BuildDir)

	// Generate a new scratch container
	newContainer := p.client.Container()

	// Init the container from the Dockerfile in host directory
	slog.Debug("build image from dockerfile", "dockerfile", p.cfg.Image.BuildDockerfile, "build_dir", p.cfg.Image.BuildDir)
	newContainer = newContainer.Build(src, dagger.ContainerBuildOpts{Dockerfile: p.cfg.Image.BuildDockerfile})

	// Export the container as a tarball to the cache directory
	_, err := newContainer.Export(context.Background(), path.Join(p.cfg.CacheDir, "image.tar"))
	if err != nil {
		return err
	}
	return nil
}
