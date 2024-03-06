package shell

import "os/exec"

// GrypeVersion prints the version of the grype CLI
//
// Requirements: N/A
//
// Ouputs: CLI Version information to STDOUT
func GrypeVersion(options ...OptionFunc) ExitCode {
	o := newOptions(options...)
	cmd := exec.Command("grype", "version")
	return run(cmd, o)
}

// GrypeScanSBOM generates a vulnerability report from an SBOM
//
// Requirements: Syft SBOM from STDIN
//
// Ouputs: A JSON vulnerability Report to STDOUT
func GrypeScanSBOM(options ...OptionFunc) ExitCode {
	o := newOptions(options...)
	cmd := exec.Command("grype", "--add-cpes-if-none", "--by-cve", "-o", "json")
	return run(cmd, o)
}
