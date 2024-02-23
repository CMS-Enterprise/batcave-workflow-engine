package shell

import (
	"io"
)

type gatecheckCmd struct {
	InitCmd func() *Executable
}

// Version outputs the version of the gatecheck CLI
// // shell: `gatecheck version`
func (s *gatecheckCmd) Version() *Command {
	return NewCommand(s.InitCmd().WithArgs("version"))
}

// Run executes the Gatecheck CLI bundle command
//
// shell: `gatecheck bundle -o ${GATECHECK_BUNDLE} ${GITLEAKS_REPORT} ${SEMGREP_REPORT} --skip-missing
func (s *gatecheckCmd) Bundle(bundle string, files ...string) *Command {
	// cmd := s.InitCmd().WithArgs("bundle", "-o", bundle, "--skip-missing", files...)
	args := []string{"bundle", "-o", bundle, "--skip-missing"}
	args = append(args, files...)
	cmd := s.InitCmd().WithArgs(args...)
	return NewCommand(cmd)
}

// Run executes the Gatecheck CLI print (summary) command
//
// shell: `gatecheck print ${GATECHECK_BUNDLE}
func (s *gatecheckCmd) Summary(bundle string) *Command {
	cmd := s.InitCmd().WithArgs("print", bundle)
	return NewCommand(cmd)
}

// GatecheckCommand with custom stdin, stdout, and stderr
// stdin must be provided even though it isn't used because without it gatecheck exits immediately
func GatecheckCommand(stdin io.Reader, stdout io.Writer, stderr io.Writer) *gatecheckCmd {
	return &gatecheckCmd{
		InitCmd: func() *Executable {
			return NewExecutable("gatecheck").WithIO(stdin, stdout, stderr)
		},
	}
}
