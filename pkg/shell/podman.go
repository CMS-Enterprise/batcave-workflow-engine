package shell

import (
	"io"
)

type podmanCmd struct {
	Executable
	InitCmd func() *Executable
}

// Version outputs the version of the podman CLI
//
// shell: `podman version`
func (p *podmanCmd) Version() *Command {
	cmd := p.InitCmd().WithArgs("version")
	return &Command{
		RunFunc: func() error {
			return cmd.Run()
		},
		DebugInfo: cmd.String(),
	}
}

// PodmanComand with custom stdout and stderr
func PodmanComand(stdout io.Writer, stderr io.Writer) *podmanCmd {
	return &podmanCmd{
		Executable: Executable{},
		InitCmd: func() *Executable {
			return NewExecutable("podman").WithOutput(stdout).WithStderr(stderr)
		},
	}
}
