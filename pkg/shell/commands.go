// Package shell is an abstraction layer around the exec package
//
// It turns commands into Go objects so a specific sub-set of that command
// can be called without having to deal with string manipulation as is typically done
// with bash.
package shell

import (
	"io"
	"log/slog"
	"os"
	"os/exec"
)

// Executable is a wrapper around exec.Cmd from the standard library
//
// # It add additional capabilities and simplifies the API by using a builder pattern
//
// Example:
//
// err := NewExecutable("echo").WithStdout(os.Stdout).WithArgs("howdy world").Run()
type Executable struct {
	exec.Cmd
}

// WithStdout attaches the command standard output to the given writer
func (e *Executable) WithStdout(w io.Writer) *Executable {
	e.Stdout = w
	return e
}

// WithStderr attaches the command standard error to the given writer
func (e *Executable) WithStderr(w io.Writer) *Executable {
	e.Stderr = w
	return e
}

// WithStdin attaches the given reader to the commands standard input
func (e *Executable) WithStdin(r io.Reader) *Executable {
	e.Stdin = r
	return e
}

// WithOutput attaches comand standard output and standard error to the given writer
func (e *Executable) WithOutput(w io.Writer) *Executable {
	return e.WithStdout(w).WithStderr(w)
}

// WithArgs attaches given arguments to a command
func (e *Executable) WithArgs(args ...string) *Executable {
	e.Args = append(e.Args[:1], args...)
	return e
}

// NewExecutable creates an Executable initialized with the given executable name
// it does not check if the given command is present on the $PATH
func NewExecutable(executableName string) *Executable {
	return &Executable{Cmd: *exec.Command(executableName)}
}

// Runner is the interface that wraps the Run method.
//
// Run could be a single command execution or a sequence of commands
// with it's own execution logic.
// String should return debug information that can optionally be printed
// by the caller.
type Runner interface {
	Run() error
	String() error
}

// Command implements the Runner interface for basic use cases.
type Command struct {
	runFunc       func() error
	info          string
	dryRunEnabled bool
	logger        *slog.Logger
}

func NewCommand(exe *Executable) *Command {
	c := &Command{
		runFunc: func() error {
			return exe.Run()
		},
		info:          exe.String(),
		dryRunEnabled: false,
		logger:        slog.Default(),
	}
	return c
}

func (c *Command) WithRunFunc(f func() error) *Command {
	c.runFunc = f
	return c
}

// WithDryRun will only setup the command to only execute if dryRun is not enabled
//
// The reason for this function is to optionally run after debugging
// the command that would run if dry run is set to true
func (c *Command) WithDryRun(enabled bool) *Command {
	c.dryRunEnabled = enabled
	return c
}

// Run the command function set at init
func (c *Command) Run() error {
	c.logger.Info("run", "command", c.String())
	if c.dryRunEnabled {
		return nil
	}
	return c.runFunc()
}

// RunLogError runs the command function and logs any potential errors
// it will also debug before the run
func (c *Command) RunLogError() {
	err := c.Run()
	if err != nil {
		c.logger.Error("command failed", "command", c.String(), "error", err)
	}
}

func (c *Command) RunLogErrorAsWarning() {
	err := c.Run()
	if err != nil {
		c.logger.Warn("command failed", "command", c.String(), "error", err)
	}
}

// String provides debug information about the command
func (c *Command) String() string {
	return c.info
}

func ExampleEcho() {
	echo := NewExecutable("echo")
	err := echo.WithStdout(os.Stdout).WithArgs("howdy world").Run()
	if err != nil {
		panic(err)
	}
	// Output: howdy world
}
