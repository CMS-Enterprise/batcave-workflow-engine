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
func (s *syftCmd) ScanImage(image string, sbomFilename string) *Command {
	cmd := s.InitCmd().WithArgs(image, "--scope=squashed", "-o", "cyclonedx-json")

	return NewCommand(cmd)
}

// SyftCommand with custom stdout and stderr
func SyftCommand(stdout io.Writer, stderr io.Writer) *syftCmd {
	return &syftCmd{
		InitCmd: func() *Executable {
			return NewExecutable("syft").WithOutput(stdout).WithStderr(stderr)
		},
	}
}
