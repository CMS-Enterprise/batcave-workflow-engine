package shell

import (
	"io"
)

type syftCmd struct {
	Executable
	InitCmd func() *Executable
}

// Version outputs the version of the syft CLI
//
// shell: `syft version`
func (s *syftCmd) Version() *Command {
	cmd := s.InitCmd().WithArgs("version")
	return &Command{
		RunFunc: func() error {
			return cmd.Run()
		},
		DebugInfo: cmd.String(),
	}
}

// SyftCommand with custom stdout and stderr
func SyftCommand(stdout io.Writer, stderr io.Writer) *syftCmd {
	return &syftCmd{InitCmd: func() *Executable {
		return NewExecutable("syft").WithOutput(stdout).WithStderr(stderr)
	},
	}
}
