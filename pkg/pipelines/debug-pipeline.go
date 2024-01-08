package pipelines

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"dagger.io/dagger"
)

type DebugExecFunc func(stdout io.Writer) error

func NewLocalDebugExec(cfg Config) DebugExecFunc {
	return func(w io.Writer) error {
		_, err := fmt.Fprintln(w, "sample output from debug local execution")
		return err
	}
}

func NewDaggerDebugExec(client *dagger.Client, cfg Config) DebugExecFunc {
	return func(w io.Writer) error {
		// Pull the debug container using the dagger client
		container := client.Container().From(cfg.DaggerDebugImage)

		// Set a random env var so the engine doesn't cache
		container.WithEnvVariable("CACHEBUSTER", time.Now().String())

		// Setup a command to execute in the container
		container = container.WithExec([]string{"echo", "sample output from debug container"})

		// Get the output from stdout in the container
		out, err := container.Stdout(context.Background())

		// Print the output to the linked writer
		_, pErr := fmt.Fprint(w, out)

		// Combine and return any possible errors, will be nil if no errors happened
		return errors.Join(err, pErr)
	}
}
