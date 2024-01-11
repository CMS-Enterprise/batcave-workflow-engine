package pipelines

import (
	"io"
	"os"
)

type PodmanImageBuild struct {
	Stdout io.Writer
	Stderr io.Writer
}

func NewImageBuild(m Mode) *PodmanImageBuild {
	return &PodmanImageBuild{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

func (i *PodmanImageBuild) Run() error {
	// TODO: Add podman shell commands
	return nil
}
