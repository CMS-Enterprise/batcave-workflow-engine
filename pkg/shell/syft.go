package shell

import "os/exec"

// SyftVersion prints the version of the syft CLI
// Requirements: N/A
//
// Ouputs: CLI Version information to STDOUT
func SyftVersion(options ...OptionFunc) ExitCode {
	o := newOptions(options...)
	cmd := exec.Command("syft", "version")
	return run(cmd, o)
}

// SyftScanImage generates an SBOM
//
// Requirements: WithScanImage() option
//
// Ouputs: A JSON vulnerability Report to STDOUT
func SyftScanImage(options ...OptionFunc) ExitCode {
	o := newOptions(options...)
	cmd := exec.Command("syft", "--scope=squashed", "-o", "cyclonedx-json", o.imageTag)
	return run(cmd, o)
}
