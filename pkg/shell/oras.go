package shell

import "os/exec"

// OrasVersion prints version of ORAS CLI
//
// Requirements: N/A
//
// Output: version to STDOUT
func OrasVersion(options ...OptionFunc) ExitCode {
	o := newOptions(options...)
	exe := exec.Command("oras", "version")
	return run(exe, o)
}

// OrasPushBundle push a gatecheck bundle
//
// Requirements: WithArtifactBundle
//
// Output: debug information to STDERR
func OrasPushBundle(options ...OptionFunc) ExitCode {
	o := newOptions(options...)
	exe := exec.Command(
		"oras",
		"push",
		"--disable-path-validation",
		"--artifact-type",
		"application/vnd.gatecheckdev.gatecheck.bundle.tar+gzip",
		o.bundleTag,
		o.gatecheck.bundleFilename,
	)
	return run(exe, o)
}
