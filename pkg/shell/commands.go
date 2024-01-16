// Package shell is an abstraction layer around the exec package
//
// It turns commands into Go objects so a specific sub-set of that command
// can be called without having to deal with string manipulation as is typically done
// with bash.
package shell

import (
	"io"
	"os"
	"os/exec"
)

// Executable is a wrapper around exec.Cmd from the standard library
//
// # It add additional capabilities and simplifies the API by using a builder pattern
//
// Example:
//
// err := NewCommand("echo").WithStdout(os.Stdout).WithArgs("howdy world").Run()
type Executable struct {
	exec.Cmd
}

// WithStdout attaches the command standard output to the given writer
func (c *Executable) WithStdout(w io.Writer) *Executable {
	c.Stdout = w
	return c
}

// WithStderr attaches the command standard error to the given writer
func (c *Executable) WithStderr(w io.Writer) *Executable {
	c.Stderr = w
	return c
}

// WithStdin attaches the given reader to the commands standard input
func (c *Executable) WithStdin(r io.Reader) *Executable {
	c.Stdin = r
	return c
}

// WithOutput attaches comand standard output and standard error to the given writer
func (c *Executable) WithOutput(w io.Writer) *Executable {
	return c.WithStdout(w).WithStderr(w)
}

// WithArgs attaches given arguments to a command
func (c *Executable) WithArgs(args ...string) *Executable {
	c.Args = append(c.Args[:1], args...)
	return c
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
	RunE      func() error
	DebugInfo string
}

// Run runs the command function
func (c *Command) Run() error {
	return c.RunE()
}

// String provides debug information about the command
func (c *Command) String() string {
	return c.DebugInfo
}

func ExampleEcho() {
	echo := NewExecutable("echo")
	err := echo.WithStdout(os.Stdout).WithArgs("howdy world").Run()
	if err != nil {
		panic(err)
	}
	// Output: howdy world
}
