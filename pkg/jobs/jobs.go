package jobs

import (
	"context"
	"time"

	"dagger.io/dagger"
)

// RunDebug is the debug job with specific execution logic
func RunDebug(container *dagger.Container) (string, error) {
	return container.
		WithEnvVariable("CACHEBUSTER", time.Now().String()).
		WithExec([]string{"echo", "sample output from debug container"}).
		Stdout(context.Background())
}

// RunBuildImage TODO: add build image command logic
func RunBuildImage(container *dagger.Container) (string, error) {
	return "build image", nil
}

// RunGrype TODO: add syft command logic
func RunSyft(container *dagger.Container) (string, error) {
	return "syft report", nil
}

// RunGrype TODO: add grype command logic
func RunGrype(container *dagger.Container) (string, error) {
	return "grype report", nil
}
