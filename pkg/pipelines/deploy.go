package pipelines

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path"
	"workflow-engine/pkg/shell"
)

type Deploy struct {
	Stdout        io.Writer
	Stderr        io.Writer
	DryRunEnabled bool
	config        *Config
	runtime       struct {
		bundleFilename string
	}
}

func NewDeploy(stdout io.Writer, stderr io.Writer) *Deploy {
	return &Deploy{
		Stdout:        stdout,
		Stderr:        stderr,
		DryRunEnabled: false,
	}
}

func (p *Deploy) WithConfig(config *Config) *Deploy {
	p.config = config
	return p
}

func (p *Deploy) preRun() error {
	p.runtime.bundleFilename = path.Join(p.config.ArtifactDir, p.config.GatecheckBundleFilename)
	return nil
}

func (p *Deploy) Run() error {
	if !p.config.Deploy.Enabled {
		slog.Warn("deployment pipeline disabled, skip.")
		return nil
	}
	if err := p.preRun(); err != nil {
		return errors.New("Deploy Pipeline failed, pre-run error. See logs for details.")
	}

	slog.Warn("deployment pipeline is a beta feature. Only gatecheck validation will be conducted.")

	err := shell.GatecheckValidate(
		shell.WithDryRun(p.DryRunEnabled),
		shell.WithStderr(p.Stderr),
		shell.WithStdout(p.Stdout),
	)
	if err != nil {
		return fmt.Errorf("Deployment Validation failed: %w", err)
	}

	return nil
}
