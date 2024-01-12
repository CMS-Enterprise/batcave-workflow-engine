package shell

import (
	"io"
	"log/slog"
)

type podmanCmd struct {
	Command
	InitCmd func() *Command
}

// Version outputs the version of the podman CLI
//
// shell: `podman version`
func (p *podmanCmd) Version() error {
	cmd := p.InitCmd().WithArgs("version")
	slog.Debug("run", "command", cmd.String())
	return cmd.Run()
}

// PodmanComand with custom stdout and stder
func PodmanComand(stdout io.Writer, stderr io.Writer) *podmanCmd {
	return &podmanCmd{
		Command: Command{},
		InitCmd: func() *Command {
			return NewCommand("podman").WithOutput(stdout).WithStderr(stderr)
		},
	}
}
