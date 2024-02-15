package shell

import (
	"fmt"
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

// ScanSBOM runs a grype scan against a target sbom file and produces a JSON report
//
// shell: `grype sbom:<filename> --add-cpes-if-none --by-cve -o json
func (g *grypeCmd) ScanSBOM(filename string) *Command {
	// TODO: At some point it would be more efficient to read the file once rather than saving to disk

	exe := g.InitCmd().WithArgs(
		fmt.Sprintf("sbom:%s", filename),
		"--add-cpes-if-none",
		"--by-cve",
		"-o",
		"json",
	)

	return NewCommand(exe)
}

// GrypeCommand with custom stdin, stdout, and stderr
// stdin must be provided even though it isn't used because without it grype exits immediately
func GrypeCommand(stdin io.Reader, stdout io.Writer, stderr io.Writer) *grypeCmd {
	return &grypeCmd{
		InitCmd: func() *Executable {
			return NewExecutable("grype").WithIO(stdin, stdout, stderr)
		},
	}
}
