package shell

import "io"

type clamScan struct {
	InitCmd func() *Executable
}

// Version of the clamscan CLI
func (c *clamScan) Version() *Command {
	exe := c.InitCmd().WithArgs("--version")
	return NewCommand(exe)
}

func (c *clamScan) Scan(targetDirectory string) *Command {
	exe := c.InitCmd().WithArgs(
		"--infected",
		"--recursive",
		"--verbose",
		"--scan-archive=yes",
		"--max-filesize=2000M",
		"--max-scansize=2000M",
		"--stdout",
		targetDirectory,
	)

	return NewCommand(exe)
}

type freshClam struct {
	InitCmd func() *Executable
}

func (c *freshClam) Version() *Command {
	exe := c.InitCmd().WithArgs("--version")
	return NewCommand(exe)
}

func (c *freshClam) FreshClam() *Command {
	exe := c.InitCmd().WithArgs("--verbose")
	return NewCommand(exe)
}

// ClamScanCommand scans files for virus by cross-referencing them to a known virus database
func ClamScanCommand(stdin io.Reader, stdout io.Writer, stderr io.Writer) *clamScan {
	return &clamScan{
		InitCmd: func() *Executable {
			return NewExecutable("clamscan").WithIO(stdin, stdout, stderr)
		},
	}
}

// FreshClam updates the virus definition database
func FreshClamCommand(stdin io.Reader, stdout io.Writer, stderr io.Writer) *freshClam {
	return &freshClam{
		InitCmd: func() *Executable {
			return NewExecutable("freshclam").WithIO(stdin, stdout, stderr)
		},
	}
}
