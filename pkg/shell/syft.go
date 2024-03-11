package shell

import (
	"io"
	"os"
	"os/exec"
)

// SyftVersion prints the version of the syft CLI
// Requirements: N/A
//
// Ouputs: CLI Version information to STDOUT
func SyftVersion(options ...OptionFunc) error {
	o := newOptions(options...)
	cmd := exec.Command("syft", "version")
	return run(cmd, o)
}

// SyftScanImage generates an SBOM
//
// Requirements: WithTarfilename() option
//
// Ouputs: A JSON vulnerability Report to STDOUT
func SyftScanImage(options ...OptionFunc) error {
	o := newOptions(options...)
	cmd := exec.Command("syft", "--scope=squashed", "-o", "cyclonedx-json", "--from", "docker-archive", o.tarFilename)
	return run(cmd, o)
}
