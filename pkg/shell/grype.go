package shell

import (
	"io"
	"log/slog"
)

type grypeCmd struct {
	Command
	InitCmd func() *Command
}

// Version outputs the version of the Grype CLI
//
// shell: `grype version`
func (g *grypeCmd) Version() error {
	cmd := g.InitCmd().WithArgs("version")
	slog.Debug("run", "command", cmd.String())
	return cmd.Run()
}

// GrypeCommand with custom stdout and stderr
func GrypeCommand(stdout io.Writer, stderr io.Writer) *grypeCmd {
	return &grypeCmd{InitCmd: func() *Command {
		return NewCommand("grype").WithOutput(stdout).WithStderr(stderr)
	},
	}
}
