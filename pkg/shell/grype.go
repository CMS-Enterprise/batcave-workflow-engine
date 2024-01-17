package shell

import (
	"io"
)

type grypeCmd struct {
	Executable
	InitCmd func() *Executable
}

// Version outputs the version of the Grype CLI
//
// shell: `grype version`
func (g *grypeCmd) Version() *Command {
	cmd := g.InitCmd().WithArgs("version")
	return &Command{
		RunFunc: func() error {
			return cmd.Run()
		},
		DebugInfo: cmd.String(),
	}
}

// GrypeCommand with custom stdout and stderr
func GrypeCommand(stdout io.Writer, stderr io.Writer) *grypeCmd {
	return &grypeCmd{InitCmd: func() *Executable {
		return NewExecutable("grype").WithOutput(stdout).WithStderr(stderr)
	},
	}
}
