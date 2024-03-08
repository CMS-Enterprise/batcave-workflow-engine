package shell

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
)

type ExitCode int

const (
	ExitOK               ExitCode = 0
	ExitUnknown          ExitCode = 232
	ExitContextCancel    ExitCode = 231
	ExitKillFailure      ExitCode = 230
	ExitBadConfiguration ExitCode = 299
)

func (e ExitCode) GetError(name string) error {
	if e == ExitOK {
		return nil
	}
	return fmt.Errorf("%s non-zero exit code: %d", name, e)
}

// Command is any function that accepts optionFuncs and returns an exit code
//
// Most commands can take advantage of the run function which automatically
// parses the options to configure the exec.Cmd
//
// It also handles early termination of the command with a context and logging
type Command func(...OptionFunc) ExitCode

// Options are flexible parameters for any command
type Options struct {
	dryRunEnabled   bool
	stdin           io.Reader
	stdout          io.Writer
	stderr          io.Writer
	errorOnlyStderr io.Writer
	errorOnly       bool
	ctx             context.Context
	failTriggerFunc func()
	scanImage       string
	tarFilename     string
	dockerAlias     DockerAlias
	imageName       string
	reportType      string
	artifactImage   string

	imageBuildOptions ImageBuildOptions

	gatecheck struct {
		bundleFilename string
		targetFile     string
	}

	semgrep struct {
		rules        string
		experimental bool
	}

	gitleaks struct {
		targetDirectory string
		reportPath      string
	}

	listTargetFilename string
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
	o.failTriggerFunc = func() {}
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

// WithErrorOnly buffers stderr unless there is a non-zero exit code.
//
// If there is a non-zero exit, the error buffer will dump to stderr
func WithErrorOnly(stderr io.Writer) OptionFunc {
	return func(o *Options) {
		o.errorOnly = true
		o.errorOnlyStderr = stderr
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

func WithGitleaks(targetDirectory string, reportPath string) OptionFunc {
	return func(o *Options) {
		o.gitleaks.targetDirectory = targetDirectory
		o.gitleaks.reportPath = reportPath
	}
}

// WithFailTrigger will call the provided function for non-zero exit
//
// This can be useful if running multiple commands async and you want
// to early termination with a context cancel should either command fail
func WithFailTrigger(f func()) OptionFunc {
	return func(o *Options) {
		o.failTriggerFunc = f
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

// WithImage can be used for multiple commands to define a archive/tar filename
//
// should include the full filename including extension
func WithTarFilename(filename string) OptionFunc {
	return func(o *Options) {
		o.tarFilename = filename
	}
}

func WithReportType(reportType string) OptionFunc {
	return func(o *Options) {
		o.reportType = reportType
	}
}

func WithArtifactBundle(artifactImage string, bundleFilename string) OptionFunc {
	return func(o *Options) {
		o.artifactImage = artifactImage
		o.gatecheck.bundleFilename = bundleFilename
	}
}

func WithBundleFile(bundleFilename string, targetFilename string) OptionFunc {
	return func(o *Options) {
		o.gatecheck.bundleFilename = bundleFilename
		o.gatecheck.targetFile = targetFilename
	}
}

func WithSemgrep(rules string, experimental bool) OptionFunc {
	return func(o *Options) {
		o.semgrep.rules = rules
		o.semgrep.experimental = experimental
	}
}

func WithListTarget(filename string) OptionFunc {
	return func(o *Options) {
		o.listTargetFilename = filename
	}
}

func WithBuildImageOptions(options ImageBuildOptions) OptionFunc {
	return func(o *Options) {
		o.imageBuildOptions = options
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

	slog.Info("shell exec", "dry_run", o.dryRunEnabled, "command", cmd.String(), "errors_only", o.errorOnly)
	if o.dryRunEnabled {
		return ExitOK
	}

	cmd.Stdin = o.stdin
	cmd.Stdout = o.stdout
	cmd.Stderr = o.stderr

	stdErrBuf := new(bytes.Buffer)
	if o.errorOnly {
		cmd.Stderr = stdErrBuf
	}

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

	var exitCode ExitCode

	// Either context will cancel or the command will finish before
	// capture the exit code
	select {
	case <-o.ctx.Done():
		exitCode = ExitContextCancel
		if err := cmd.Process.Kill(); err != nil {
			exitCode = ExitKillFailure
		}
	case <-doneChan:

		exitCode = ExitOK

		var exitCodeError *exec.ExitError
		if errors.As(runError, &exitCodeError) {
			exitCode = ExitCode(exitCodeError.ExitCode())
			break
		}
		if runError != nil {
			exitCode = ExitUnknown
		}
	}

	if exitCode != ExitOK {
		o.failTriggerFunc()
		if o.errorOnly {
			slog.Info("non-zero exit and error only, dumping log")
			if _, err := io.Copy(o.errorOnlyStderr, stdErrBuf); err != nil {
				slog.Warn("cannot dump stderr to destination", "error", err)
			}
		}
	}

	return exitCode
}
