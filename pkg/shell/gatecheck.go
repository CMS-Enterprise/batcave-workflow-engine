package shell

import "io"

type gatecheckCmd struct {
	InitCmd func() *Executable
}

// Version outputs the version of the gatecheck CLI
//
// shell: `gatecheck version`
func (s *gatecheckCmd) Version() *Command {
	return NewCommand(s.InitCmd().WithArgs("version"))
}

// List will print a table of vulnerabilities in a report
//
// shell: `cat grype-report.json | gatecheck list -i grype`
func (s *gatecheckCmd) List(inputFileType string) *Command {
	exe := s.InitCmd().WithArgs("list", "--input-type", inputFileType)
	return NewCommand(exe)
}

// ListAll will print a table of vulnerabilities in a report
//
// shell: `cat grype-report.json | gatecheck list --all -i grype`
func (s *gatecheckCmd) ListAll(inputFileType string) *Command {
	exe := s.InitCmd().WithArgs("list", "--all", "--input-type", inputFileType)
	return NewCommand(exe)
}

func (s *gatecheckCmd) BundleCreate(bundleFilename string, targetFilename string) *Command {
	exe := s.InitCmd().WithArgs("bundle", "create", bundleFilename, targetFilename)
	return NewCommand(exe)
}

func (s *gatecheckCmd) BundleAdd(bundleFilename string, targetFilename string) *Command {
	exe := s.InitCmd().WithArgs("bundle", "add", bundleFilename, targetFilename)
	return NewCommand(exe)
}

// GatecheckCommand with custom stdin, stdout, and stderr
//
// stdin must be provided even though it isn't used because without it grype exits immediately
func GatecheckCommand(stdin io.Reader, stdout io.Writer, stderr io.Writer) *gatecheckCmd {
	return &gatecheckCmd{
		InitCmd: func() *Executable {
			return NewExecutable("gatecheck").WithIO(stdin, stdout, stderr)
		},
	}
}
