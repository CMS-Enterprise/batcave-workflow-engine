package shell

import (
	"io"
)

type syftCmd struct {
	InitCmd func() *Executable
}

// Version outputs the version of the syft CLI
// // shell: `syft version`
func (s *syftCmd) Version() *Command {
	return NewCommand(s.InitCmd().WithArgs("version"))
}

// Run executes the Syft CLI
//
// shell: `syft <image> --scope=squashed -o cyclonedx-json`
func (s *syftCmd) ScanImage(image string) *Command {
	cmd := s.InitCmd().WithArgs(image, "--scope=squashed", "-o", "cyclonedx-json")

	return NewCommand(cmd)
}

// SyftCommand with custom stdin, stdout, and stderr
// stdin must be provided even though it isn't used because without it syft exits immediately
func SyftCommand(stdin io.Reader, stdout io.Writer, stderr io.Writer) *syftCmd {
	return &syftCmd{
		InitCmd: func() *Executable {
			return NewExecutable("syft").WithStdin(stdin).WithOutput(stdout).WithStderr(stderr)
		},
	}
}
