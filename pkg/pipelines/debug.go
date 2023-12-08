package pipelines

import (
	"io"
	"strings"
	"workflow-engine/pkg/environments"
	"workflow-engine/pkg/jobs"

	"dagger.io/dagger"
)

// Debug pipeline is designed for smoke testing features
type Debug struct {
	name   string
	client *dagger.Client
	stdout io.Writer
}

// NewDebugPipeline setups an instance of the debug pipeline
func NewDebugPipeline(c *dagger.Client, stdoutWriter io.Writer) *Debug {
	pipeline := &Debug{name: "Debug Pipeline", client: c, stdout: stdoutWriter}
	return pipeline
}

// Execute runs the full debug pipeline
func (p *Debug) Execute() error {
	alpine := environments.NewAlpine(p.client)
	// Get the debug system information
	debugOutput, err := jobs.RunDebug(alpine.Container())
	strings.NewReader(debugOutput).WriteTo(p.stdout)
	return err
}
