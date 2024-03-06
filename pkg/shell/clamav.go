package shell

import "os/exec"

// ClamScanVersion print version information
//
// Requirements: N/A
//
// Outputs: version information to STDOUT
func ClamScanVersion(optionFuncs ...OptionFunc) ExitCode {
	o := newOptions(optionFuncs...)
	cmd := exec.Command("clamscan", "--version")
	return run(cmd, o)
}

// FreshClamVersion print version information
//
// Requirements: N/A
//
// Outputs: version information to STDOUT
func FreshClamVersion(optionFuncs ...OptionFunc) ExitCode {
	o := newOptions(optionFuncs...)
	cmd := exec.Command("freshclam", "--version")
	return run(cmd, o)
}

// Clamscan runs ClamAV virus scan on an image archive
//
// Requirements: WithTarFilename() option
//
// Outputs: Virus Report to STDOUT
func Clamscan(optionFuncs ...OptionFunc) ExitCode {
	o := newOptions(optionFuncs...)
	cmd := exec.Command(
		"clamscan",
		"--infected",
		"--recursive",
		"--verbose",
		"--scan-archive=yes",
		"--max-filesize=2000M",
		"--max-scansize=2000M",
		"--stdout",
		o.tarFilename,
	)
	return run(cmd, o)
}

// Freshclam runs ClamAV virus definition database update
//
// Requirements: N/A
//
// Outputs: Debugging information to STDOUT
func Freshclam(optionFuncs ...OptionFunc) ExitCode {
	o := newOptions(optionFuncs...)
	cmd := exec.Command("freshclam", "--verbose")
	return run(cmd, o)
}
