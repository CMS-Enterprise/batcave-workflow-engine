package shell

import (
	"io"
	"log/slog"
)

type syftCmd struct {
	Command
	InitCmd func() *Command
}

// Version outputs the version of the syft CLI
//
// shell: `syft version`
func (s *syftCmd) Version() error {
	cmd := s.InitCmd().WithArgs("version")
	slog.Debug("run", "command", cmd.String())
	return cmd.Run()
}

// SyftCommand with custom stdout and stderr
func SyftCommand(stdout io.Writer, stderr io.Writer) *syftCmd {
	return &syftCmd{InitCmd: func() *Command {
		return NewCommand("syft").WithOutput(stdout).WithStderr(stderr)
	},
	}
}
