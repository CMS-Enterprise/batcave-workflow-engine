package shell

import (
	"io"
)

type gitleaksCmd struct {
	InitCmd func() *Executable
}

// Version outputs the version of the Gitleaks CLI
//
// shell: `gitleaks version`
func (g *gitleaksCmd) Version() *Command {
	return NewCommand(g.InitCmd().WithArgs("version"))
}

// DetectSecrets runs a gitleaks scan against the current repo and produces a JSON report
//
// shell: gitleaks detect --exit-code 0 --verbose --source ${TARGET_DIRECTORY} --report-path ${GITLEAKS_REPORT}
func (g *gitleaksCmd) DetectSecrets(sourceDirectory string, reportPath string) *Command {
	exe := g.InitCmd().WithArgs(
		"detect",
		"exit-code",
		"0",
		"--verbose",
		"--source",
		sourceDirectory,
		"--report-path",
		reportPath,
	)

	return NewCommand(exe)
}

// GitleaksCommand with custom stdin, stdout, and stderr
func GitleaksCommand(stdin io.Reader, stdout io.Writer, stderr io.Writer) *gitleaksCmd {
	return &gitleaksCmd{
		InitCmd: func() *Executable {
			return NewExecutable("gitleaks").WithIO(stdin, stdout, stderr)
		},
	}
}
