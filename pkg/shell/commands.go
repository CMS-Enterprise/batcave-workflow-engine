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

// Command is a wrapper around exec.Cmd from the standard library
//
// # It add additional capabilities and simplifies the API by using a builder pattern
//
// Example:
//
// err := NewCommand("echo").WithStdout(os.Stdout).WithArgs("howdy world").Run()
type Command struct {
	exec.Cmd
}

// WithStdout attaches the command standard output to the given writer
func (c *Command) WithStdout(w io.Writer) *Command {
	c.Stdout = w
	return c
}

// WithStderr attaches the command standard error to the given writer
func (c *Command) WithStderr(w io.Writer) *Command {
	c.Stderr = w
	return c
}

// WithStdin attaches the given reader to the commands standard input
func (c *Command) WithStdin(r io.Reader) *Command {
	c.Stdin = r
	return c
}

// WithOutput attaches comand standard output and standard error to the given writer
func (c *Command) WithOutput(w io.Writer) *Command {
	return c.WithStdout(w).WithStderr(w)
}

// WithArgs attaches given arguments to a command
func (c *Command) WithArgs(args ...string) *Command {
	c.Args = append(c.Args[:1], args...)
	return c
}

// NewCommand creates a Command initialized with the given executable name
// it does not check if the given command is present on the $PATH
func NewCommand(executableName string) *Command {
	return &Command{Cmd: *exec.Command(executableName)}
}

func ExampleEcho() {
	echo := NewCommand("echo")
	err := echo.WithStdout(os.Stdout).WithArgs("howdy world").Run()
	if err != nil {
		panic(err)
	}
	// Output: howdy world
}
