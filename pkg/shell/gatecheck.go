package shell

import "os/exec"

// GatecheckList will print a summarized view of a a report
//
// Requirement: supported report from STDIN WithReportType
//
// Output: table to STDOUT
func GatecheckList(options ...OptionFunc) ExitCode {
	o := newOptions(options...)
	cmd := exec.Command("gatecheck", "list", "--input-type", o.reportType)
	if o.listTargetFilename != "" {
		cmd = exec.Command("gatecheck", "list", o.listTargetFilename)
	}
	return run(cmd, o)
}

// GatecheckListAll will print a summarized view of a a report with EPSS scores
//
// Requirement: supported report from STDIN
//
// Output: table to STDOUT
func GatecheckListAll(options ...OptionFunc) ExitCode {
	o := newOptions(options...)
	cmd := exec.Command("gatecheck", "list", "--all", "--input-type", o.reportType)
	return run(cmd, o)
}

// GatecheckVersion print version information
//
// Requirement: N/A
//
// Output: to STDOUT
func GatecheckVersion(options ...OptionFunc) ExitCode {
	o := newOptions(options...)
	cmd := exec.Command("gatecheck", "version")
	return run(cmd, o)
}
