package shell

import "os/exec"

// GitLeaksVersion prints version of GitLeaks CLI
//
// Requirements: N/A
//
// Output: version to STDOUT
func GitLeaksVersion(options ...OptionFunc) ExitCode {
	o := newOptions(options...)
	exe := exec.Command("gitleaks", "version")
	return run(exe, o)
}

// GitLeaksDetect prints version of GitLeaks CLI
//
// Requirements: WithGitleaks
//
// Output: debug to STDERR
func GitLeaksDetect(options ...OptionFunc) ExitCode {
	o := newOptions(options...)
	exe := exec.Command("gitleaks",
		"detect",
		"--exit-code",
		"0",
		"--verbose",
		"--source",
		o.gitleaks.targetDirectory,
		"--report-path",
		o.gitleaks.reportPath,
	)
	return run(exe, o)
}
