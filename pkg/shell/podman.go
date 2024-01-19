package shell

import (
	"io"
)

type dockerCLICmd struct {
	Executable
	InitCmd func() *Executable
}

// Version outputs the version of the CLI
//
// shell: `[docker|podman] version`
func (p *dockerCLICmd) Version() *Command {
	cmd := p.InitCmd().WithArgs("version")
	return &Command{
		RunFunc: func() error {
			return cmd.Run()
		},
		DebugInfo: cmd.String(),
	}
}

// Info tests the connection to the container runtime daemon
//
// shell: `[docker|podman] info`
func (p *dockerCLICmd) Info() *Command {
	cmd := p.InitCmd().WithArgs("info")
	return &Command{
		RunFunc: func() error {
			return cmd.Run()
		},
		DebugInfo: cmd.String(),
	}
}

// PodmanCommand with custom stdout and stderr
func PodmanCommand(stdout io.Writer, stderr io.Writer) *dockerCLICmd {
	return &dockerCLICmd{
		Executable: Executable{},
		InitCmd: func() *Executable {
			return NewExecutable("podman").WithOutput(stdout).WithStderr(stderr)
		},
	}
}

// DockerCommand with custom stdout and stderr
func DockerCommand(stdout io.Writer, stderr io.Writer) *dockerCLICmd {
	return &dockerCLICmd{
		Executable: Executable{},
		InitCmd: func() *Executable {
			return NewExecutable("docker").WithOutput(stdout).WithStderr(stderr)
		},
	}
}
