// Package shell is an abstraction layer around the exec package
//
// It turns commands into Go objects so a specific sub-set of that command
// can be called without having to deal with string manipulation as is typically done
// with bash.
package shell

import (
	"context"
	"errors"
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

// WithIO attaches all of the necessary IO to the command
func (e *Executable) WithIO(stdin io.Reader, stdout io.Writer, stderr io.Writer) *Executable {
	e.Stdin = stdin
	e.Stdout = stdout
	e.Stderr = stderr
	return e
}

// WithArgs attaches given arguments to a command
//
// All other arguments will be overridden except for the base command/binary
func (e *Executable) WithArgs(args ...string) *Executable {
	e.Args = append(e.Args[:1], args...)
	return e
}

// NewExecutable creates an Executable initialized with the given executable name
// it does not check if the given command is present on the $PATH
func NewExecutable(executableName string) *Executable {
	exe := exec.Command(executableName)
	exe.Stderr = os.Stderr
	exe.Stdin = os.Stdin
	exe.Stdout = os.Stdout
	return &Executable{Cmd: *exe}
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
	executable    *Executable
}

func NewCommand(exe *Executable) *Command {
	c := &Command{
		runFunc: func() error {
			return exe.Run()
		},
		executable:    exe,
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

// RunWithIO can be used to adjust the IO at run time which is useful for redirects
func (c *Command) WithIO(stdin io.Reader, stdout io.Writer, stderr io.Writer) *Command {
	c.executable.Stdin = stdin
	c.executable.Stdout = stdout
	c.executable.Stderr = stderr
	return c
}

// RunWithContext will run and kill the process if ctx.Done happens before the command completes
func (c *Command) RunWithContext(ctx context.Context) error {
	c.logger.Info("run with context", "command", c.String())
	if err := c.executable.Start(); err != nil {
		return err
	}
	doneChan := make(chan struct{}, 1)
	var runError error
	go func() {
		runError = c.executable.Wait()
		doneChan <- struct{}{}
	}()

	select {
	case <-doneChan:
		return runError
	case <-ctx.Done():
		err := c.executable.Process.Kill()
		return errors.Join(errors.New("command canceled"), err)
	}
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
	err := echo.WithIO(nil, os.Stdout, os.Stderr).WithArgs("howdy world").Run()
	if err != nil {
		panic(err)
	}
	// Output: howdy world
}
