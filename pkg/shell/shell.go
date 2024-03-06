package shell

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os/exec"
)

type ExitCode int

const (
	ExitOK               ExitCode = 0
	ExitUnknown                   = 232
	ExitContextCancel             = 231
	ExitKillFailure               = 230
	ExitBadConfiguration          = 299
)

// Command is any function that accepts optionFuncs and returns an exit code
//
// Most commands can take advantage of the run function which automatically
// parses the options to configure the exec.Cmd
//
// It also handles early termination of the command with a context and logging
type Command func(...OptionFunc) ExitCode

// Options are flexible parameters for any command
type Options struct {
	dryRunEnabled bool
	stdin         io.Reader
	stdout        io.Writer
	stderr        io.Writer
	ctx           context.Context
	scanImage     string
	tarFilename   string
	dockerAlias   DockerAlias
	imageName     string
}

// apply should be called before the exec.Cmd is run
func (o *Options) apply(options ...OptionFunc) {
	for _, optionFunc := range options {
		optionFunc(o)
	}
}

// newOptions is used to generate an Options struct and automatically apply optionFuncs
func newOptions(options ...OptionFunc) *Options {
	o := new(Options)
	o.apply(options...)
	return o
}

// OptionFunc are used to set option values in a flexible way
type OptionFunc func(o *Options)

// WithDryRun sets the dryRunEnabled option which will print the command that would run and exitOK
func WithDryRun(enabled bool) OptionFunc {
	return func(o *Options) {
		o.dryRunEnabled = enabled
	}
}

// WithIO sets input and output for a command
func WithIO(stdin io.Reader, stdout io.Writer, stderr io.Writer) OptionFunc {
	return func(o *Options) {
		o.stdin = stdin
		o.stdout = stdout
		o.stderr = stderr
	}
}

// WithStdin only sets STDIN for the command
func WithStdin(r io.Reader) OptionFunc {
	return func(o *Options) {
		o.stdin = r
	}
}

// WithStdout only sets STDOUT for the command
func WithStdout(w io.Writer) OptionFunc {
	return func(o *Options) {
		o.stdout = w
	}
}

// WithStderr only sets STDERR for the command
func WithStderr(w io.Writer) OptionFunc {
	return func(o *Options) {
		o.stderr = w
	}
}

func WithScanImage(image string) OptionFunc {
	return func(o *Options) {
		o.scanImage = image
	}
}

func WithCtx(ctx context.Context) OptionFunc {
	return func(o *Options) {
		o.ctx = ctx
	}
}

// WithDockerAlias can be used to configure an alternative docker compatible CLI
//
// For example, `docker build` and `podman build` can be used interchangably
func WithDockerAlias(a DockerAlias) OptionFunc {
	return func(o *Options) {
		o.dockerAlias = a
	}
}

// WithImage can be used for multiple commands to define a target image as a parameter
//
// This will include the full image and tag for example `alpine:latest`
func WithImage(image string) OptionFunc {
	return func(o *Options) {
		o.imageName = image
	}
}

// run handles the execution of the command
//
// context will be set to background if not provided in the o.ctx
// this enables the command to be terminated before completion
// if ctx fires done.
//
// Setting the dry run option will always return ExitOK
func run(cmd *exec.Cmd, o *Options) ExitCode {
	slog.Info("shell exec", "dry_run", o.dryRunEnabled, "command", cmd.String())
	if o.dryRunEnabled {
		return ExitOK
	}

	cmd.Stdin = o.stdin
	cmd.Stdout = o.stdout
	cmd.Stderr = o.stderr

	if err := cmd.Start(); err != nil {
		return ExitUnknown
	}
	if o.ctx == nil {
		o.ctx = context.Background()
	}
	var runError error
	doneChan := make(chan struct{}, 1)
	go func() {
		runError = cmd.Wait()
		doneChan <- struct{}{}
	}()

	select {
	case <-o.ctx.Done():
		if err := cmd.Process.Kill(); err != nil {
			return ExitKillFailure
		}
		return ExitContextCancel
	case <-doneChan:
		var exitCodeError *exec.ExitError
		if errors.As(runError, &exitCodeError) {
			return ExitCode(exitCodeError.ExitCode())
		}
		if runError != nil {
			return ExitUnknown
		}
	}

	return ExitOK
}
