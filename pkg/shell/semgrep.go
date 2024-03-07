package shell

import "os/exec"

// SemgrepVersion prints version of Semgrep CLI
//
// Requirements: N/A
//
// Output: version to STDOUT
func SemgrepVersion(options ...OptionFunc) ExitCode {
	o := newOptions(options...)
	exe := exec.Command("semgrep", "--version")
	if o.semgrep.experimental {
		exe = exec.Command("osemgrep", "--help")
	}
	return run(exe, o)
}

// SemgrepScan runs a Semgrep scan against target artifact dir from env vars
//
// Requirements: WithSemgrep
//
// Output: JSON report to STDOUT
func SemgrepScan(options ...OptionFunc) ExitCode {
	o := newOptions(options...)
	exe := exec.Command("semgrep", "ci", "--json", "--config", o.semgrep.rules)
	if o.semgrep.experimental {
		exe = exec.Command("osemgrep", "ci", "--json", "--experimental", "--config", o.semgrep.rules)
	}
	return run(exe, o)
}
