package shell

import (
	"io"
)

type clamavCmd struct {
	InitCmd      func() *Executable
	experimental bool
}

// Version outputs the version of the ClamAV CLI
// shell: `clamscan --version`
func (s *clamavCmd) Version() *Command {
	return NewCommand(s.InitCmd().WithArgs("--version"))
}

// Run runs a FreshClam command to update the CVD database
//
// shell: `freshclam`
func (s *clamavCmd) Run() *Command {
	exe := s.InitCmd()
	return NewCommand(exe)
}

// ScanFile runs ClamAV Scan
//
// shell: `clamscan -irv --scan-archive=yes --max-filesize=4000M --max-scansize=4000M --stdout ${TARGET_DIR}`
func (s *clamavCmd) Scan(directory string) *Command {
	args := []string{"-irv", "--scan-archive=yes", "--max-filesize=2000M", "--max-scansize=2000M", "--stdout", "--debug", directory}
	exe := s.InitCmd().WithArgs(args...)
	return NewCommand(exe)
}

// FreshClam Command with custom stdout and stderr
func FreshClamCommand(stdin io.Reader, stdout io.Writer, stderr io.Writer) *clamavCmd {
	return &clamavCmd{
		InitCmd: func() *Executable {
			return NewExecutable("freshclam").WithIO(stdin, stdout, stderr)
		},
	}
}

// ClamScan Command with custom stdout and stderr
func ClamScanCommand(stdin io.Reader, stdout io.Writer, stderr io.Writer) *clamavCmd {
	return &clamavCmd{
		InitCmd: func() *Executable {
			return NewExecutable("clamscan").WithIO(stdin, stdout, stderr)
		},
	}
}
