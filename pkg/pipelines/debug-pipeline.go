package pipelines

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"dagger.io/dagger"
)

type Debug struct {
	Stdout io.Writer
	client *dagger.Client
	cfg    Config
}

func NewDebugPipeline(c *dagger.Client, cfg Config) *Debug {
	return &Debug{client: c, Stdout: io.Discard, cfg: cfg}
}

func (d *Debug) Run() error {
	// Pull the debug container using the dagger client
	container := d.client.Container().From(d.cfg.DebugImage)

	// Set a random env var so the engine doesn't cache
	container.WithEnvVariable("CACHEBUSTER", time.Now().String())

	// Setup a command to execute in the container
	container = container.WithExec([]string{"echo", "sample output from debug container"})

	// Get the output from stdout in the container
	out, err := container.Stdout(context.Background())

	// Print the output to the linked writer
	_, pErr := fmt.Fprint(d.Stdout, out)

	// Combine and return any possible errors, will be nil if no errors happened
	return errors.Join(err, pErr)
}
