package pipelines

import (
	"errors"
	"io"
	"log/slog"
	"path"
	"workflow-engine/pkg/shell/legacy"
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
	p.runtime.bundleFilename = path.Join(p.config.ArtifactsDir, p.config.GatecheckBundleFilename)
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

	cmd := shell.GatecheckCommand(nil, p.Stdout, p.Stderr).Validate(p.runtime.bundleFilename)

	if err := cmd.WithDryRun(p.DryRunEnabled).Run(); err != nil {
		return errors.New("Deploy Pipeline failed. See logs for details.")
	}

	return nil
}
