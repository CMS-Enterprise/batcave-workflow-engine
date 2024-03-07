package pipelines

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"sync"
	"workflow-engine/pkg/shell"
	legacyShell "workflow-engine/pkg/shell/legacy"
)

type CodeScan struct {
	Stdout                        io.Writer
	Stderr                        io.Writer
	DryRunEnabled                 bool
	SemgrepExperimental           bool
	SemgrepErrorOnFindingsEnabled bool
	SemgrepRules                  string
	config                        *Config
	runtime                       struct {
		gitleaksFile      *os.File
		semgrepFile       *os.File
		bundleFilename    string
		gitleaksFilename  string
		semgrepFilename   string
		postSummaryBuffer *bytes.Buffer
		summaryMutex      sync.Mutex
	}
}

func (p *CodeScan) WithConfig(config *Config) *CodeScan {
	p.config = config
	return p
}

func NewCodeScan(stdout io.Writer, stderr io.Writer) *CodeScan {
	return &CodeScan{
		Stdout:        stdout,
		Stderr:        stderr,
		DryRunEnabled: false,
		config:        new(Config),
	}
}

// preRun is responsible for opening file needed during the run
func (p *CodeScan) preRun() error {
	var err error

	if err := MakeDirectoryP(p.config.ArtifactsDir); err != nil {
		slog.Error("failed to create artifact directory", "name", p.config.ArtifactsDir)
		return errors.New("Code Scan Pipeline failed to run. See log for details.")
	}

	p.runtime.gitleaksFilename = path.Join(p.config.ArtifactsDir, p.config.CodeScan.GitleaksFilename)
	p.runtime.gitleaksFile, err = OpenOrCreateFile(p.runtime.gitleaksFilename)
	if err != nil {
		slog.Error("cannot open gitleaks report file", "filename", p.runtime.gitleaksFile, "error", err)
		return err
	}

	p.runtime.semgrepFilename = path.Join(p.config.ArtifactsDir, p.config.CodeScan.SemgrepFilename)
	p.runtime.semgrepFile, err = OpenOrCreateFile(p.runtime.semgrepFilename)
	if err != nil {
		slog.Error("cannot open semgrep report file", "filename", p.runtime.semgrepFilename, "error", err)
		return err
	}

	if err := InitGatecheckBundle(p.config, p.Stderr, p.DryRunEnabled); err != nil {
		slog.Error("cannot initialize gatecheck bundle", "error", err)
		return err
	}

	p.runtime.postSummaryBuffer = new(bytes.Buffer)
	p.runtime.bundleFilename = path.Join(p.config.ArtifactsDir, p.config.GatecheckBundleFilename)

	return nil
}

func (p *CodeScan) Run() error {
	if !p.config.CodeScan.Enabled {
		slog.Warn("code scan pipeline is disabled, skip.")
		return nil
	}

	if err := p.preRun(); err != nil {
		return errors.New("Code Scan Pipeline Pre-Run Failed. See log for details.")
	}

	defer func() {
		_ = p.runtime.gitleaksFile.Close()
		_ = p.runtime.semgrepFile.Close()
	}()

	slog.Info("run image scan pipeline", "dry_run_enabled", p.DryRunEnabled, "artifact_directory", p.config.ArtifactsDir)

	slog.Debug("open gatecheck bundle file for output", "filename", p.runtime.bundleFilename)

	// Add a new line to separate the reports
	fmt.Fprintln(p.runtime.postSummaryBuffer, "")

	semgrepTask := NewAsyncTask("semgrep")
	go func() {
		defer semgrepTask.stdErrPipeWriter.Close()
		buf := new(bytes.Buffer)
		mw := io.MultiWriter(p.runtime.semgrepFile, buf)
		exitCode := shell.SemgrepScan(
			shell.WithDryRun(p.DryRunEnabled),
			shell.WithIO(nil, mw, semgrepTask.stdErrPipeWriter),
			shell.WithSemgrep(p.config.CodeScan.SemgrepRules, p.SemgrepExperimental),
		)
		switch exitCode {
		case shell.ExitOK:
			semgrepTask.logger.Debug("no semgrep findings")
		case 1:
			semgrepTask.logger.Debug("semgrep findings, suppress error")
		default:
			// Don't gatecheck list
			semgrepTask.exitError = exitCode.GetError("semgrep")
			return
		}
		// locking prevents writing at the same time
		p.runtime.summaryMutex.Lock()
		defer p.runtime.summaryMutex.Unlock()

		fmt.Fprintf(p.runtime.postSummaryBuffer, "%50s\n", "Semgrep Findings")
		// list report
		exitCode = shell.GatecheckList(
			shell.WithDryRun(p.DryRunEnabled),
			shell.WithIO(buf, p.runtime.postSummaryBuffer, nil),
			shell.WithReportType("semgrep"),
			shell.WithErrorOnly(semgrepTask.stdErrPipeWriter),
		)
		// Join errors, will be nil or both are nil
		semgrepTask.exitError = errors.Join(semgrepTask.exitError, exitCode.GetError("gatcheck list"))
	}()

	gitleaksTask := NewAsyncTask("gitleaks")
	go func() {
		defer gitleaksTask.stdErrPipeWriter.Close()
		exitCode := shell.GitLeaksDetect(
			shell.WithDryRun(p.DryRunEnabled),
			shell.WithStderr(gitleaksTask.stdErrPipeWriter),
			shell.WithGitleaks(p.config.CodeScan.GitleaksSrcDir, p.runtime.gitleaksFilename),
		)

		gitleaksTask.exitError = exitCode.GetError("gitleaks")

		// Gitleaks annoyingly doesn't output the json to stdout, so no piping into gatecheck list
		_ = p.runtime.gitleaksFile.Close()

		if gitleaksTask.exitError != nil {
			return
		}

		// locking prevents writing at the same time
		p.runtime.summaryMutex.Lock()
		defer p.runtime.summaryMutex.Unlock()
		fmt.Fprintf(p.runtime.postSummaryBuffer, "%30s\n", "Gitleaks Findings")
		// list report
		exitCode = shell.GatecheckList(
			shell.WithDryRun(p.DryRunEnabled),
			shell.WithIO(nil, p.runtime.postSummaryBuffer, nil),
			shell.WithListTarget(p.runtime.gitleaksFilename),
			shell.WithErrorOnly(semgrepTask.stdErrPipeWriter),
		)
		gitleaksTask.exitError = exitCode.GetError("gatecheck list gitleaks report")
	}()

	var gitleaksError, semgrepError error

	// Wait order determines the stderr print order
	if err := semgrepTask.Wait(p.Stderr); err != nil {
		semgrepError = fmt.Errorf("semgrep run failure: %v", err)
	}

	if err := gitleaksTask.Wait(p.Stderr); err != nil {
		semgrepError = fmt.Errorf("gitleaks run failure: %v", err)
	}

	var postRunError error

	if err := p.postRun(); err != nil {
		postRunError = errors.New("Code Scan Pipeline Post-Run Failed. See log for details.")
	}

	return errors.Join(gitleaksError, semgrepError, postRunError)
}

func (p *CodeScan) postRun() error {
	files := []string{p.runtime.gitleaksFilename, p.runtime.semgrepFilename}
	err := RunGatecheckBundleAdd(p.runtime.bundleFilename, p.Stderr, p.DryRunEnabled, files...)
	if err != nil {
		slog.Error("cannot run gatecheck bundle add", "error", err)
	}

	// print the Gatecheck List Content
	_, _ = p.runtime.postSummaryBuffer.WriteTo(p.Stdout)
	return err
}

// RunGitleaksDetect the report will be written to stdout
//
// Gitleaks is a special case because the command does not support writing to a file that doesn't exist
// It also doesn't write the contents of the report to stdout which means piping isn't possible.
// This function creates a temporary file for the report and then copies the content to the dst writer
func RunGitleaksDetect(reportDst io.Writer, stdErr io.Writer, config *Config, dryRunEnabled bool) error {
	slog.Debug("create temp gitleaks report", "dir", os.TempDir())
	reportFile, err := os.CreateTemp(os.TempDir(), "*-gitleaks-report.json")
	if err != nil {
		return err
	}
	tempReportFilename := reportFile.Name()

	defer func() {
		_ = reportFile.Close()
		_ = os.Remove(tempReportFilename)
	}()

	cmd := legacyShell.GitleaksCommand(nil, nil, stdErr).DetectSecrets(config.CodeScan.GitleaksSrcDir, tempReportFilename)
	err = cmd.WithDryRun(dryRunEnabled).Run()
	if err != nil {
		return errors.New("Code Scan Pipeline failed: Gitleaks execution failure. See log for details.")
	}

	// Seek errors are really unlikely, just join with the copy error in the rare case that it occurs
	_, seekErr := reportFile.Seek(0, io.SeekStart)

	_, copyErr := io.Copy(reportDst, reportFile)
	return errors.Join(seekErr, copyErr)
}

func RunSemgrep(reportDst io.Writer, stdErr io.Writer, config *Config, dryRunEnabled bool, experimental bool) error {
	var semgrep interface {
		Scan(rules string) *legacyShell.Command
	}

	if experimental {
		semgrep = legacyShell.OSemgrepCommand(nil, reportDst, stdErr)
	} else {
		semgrep = legacyShell.SemgrepCommand(nil, reportDst, stdErr)
	}

	// manually suppress errors for findings, convert to warnings
	// https://semgrep.dev/docs/semgrep-ci/configuring-blocking-and-errors-in-ci/
	// error code documentation: https://semgrep.dev/docs/cli-reference/
	if err := semgrep.Scan(config.CodeScan.SemgrepRules).WithDryRun(dryRunEnabled).Run(); err != nil {
		return errors.New("Code Scan Pipeline failed: Semgrep findings detected. See log for details.")
	}

	return nil
}
