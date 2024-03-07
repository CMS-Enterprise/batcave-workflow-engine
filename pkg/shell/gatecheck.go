package shell

import "os/exec"

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

// GatecheckBundleAdd add a file to an existing bundle
//
// Requirement: WithBundleFile
//
// Output: debug to STDERR
func GatecheckBundleAdd(options ...OptionFunc) ExitCode {
	o := newOptions(options...)
	cmd := exec.Command("gatecheck", "bundle", "add",
		o.gatecheck.bundleFilename, o.gatecheck.targetFile)
	return run(cmd, o)
}

// GatecheckBundleCreate new bundle and add a file
//
// Requirement: WithBundleFile
//
// Output: debug to STDERR
func GatecheckBundleCreate(options ...OptionFunc) ExitCode {
	o := newOptions(options...)
	cmd := exec.Command("gatecheck", "bundle", "create",
		o.gatecheck.bundleFilename, o.gatecheck.targetFile)
	return run(cmd, o)
}
