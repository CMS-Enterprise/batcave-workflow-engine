package shell

import (
	"io"
)

type grypeCmd struct {
	InitCmd func() *Executable
}

// Version outputs the version of the Grype CLI
//
// shell: `grype version`
func (g *grypeCmd) Version() *Command {
	return NewCommand(g.InitCmd().WithArgs("version"))
}

// GrypeCommand with custom stdout and stderr
func GrypeCommand(stdout io.Writer, stderr io.Writer) *grypeCmd {
	return &grypeCmd{
		InitCmd: func() *Executable {
			return NewExecutable("grype").WithOutput(stdout).WithStderr(stderr)
		},
	}
}
