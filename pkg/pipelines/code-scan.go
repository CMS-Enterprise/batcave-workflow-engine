package pipelines

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"workflow-engine/pkg/shell"
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
		gitleaksFile     *os.File
		semgrepFile      *os.File
		bundleFilename   string
		gitleaksFilename string
		semgrepFilename  string
		gatecheckListBuf *bytes.Buffer
	}
}

func (s *CodeScan) WithConfig(config *Config) *CodeScan {
	s.config = config
	return s
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

	p.runtime.gatecheckListBuf = new(bytes.Buffer)
	p.runtime.bundleFilename = path.Join(p.config.ArtifactsDir, p.config.GatecheckBundleFilename)

	return nil
}

func (p *CodeScan) Run() error {
	if err := p.preRun(); err != nil {
		return errors.New("Code Scan Pipeline Pre-Run Failed. See log for details.")
	}

	defer func() {
		_ = p.runtime.gitleaksFile.Close()
		_ = p.runtime.semgrepFile.Close()
	}()

	slog.Info("run image scan pipeline", "dry_run_enabled", p.DryRunEnabled, "artifact_directory", p.config.ArtifactsDir)

	slog.Debug("open gatecheck bundle file for output", "filename", p.runtime.bundleFilename)

	// All of the gatcheck summaries should print at the end
	buf := new(bytes.Buffer)
	// MultiWriter will write to the gitleaks file and to the buf so gatecheck can parse it
	mw := io.MultiWriter(p.runtime.gitleaksFile, buf)
	gitleaksError := RunGitleaksDetect(mw, p.Stderr, p.config, p.DryRunEnabled)

	slog.Debug("summarize gitleaks report")
	if err := RunGatecheckList(p.runtime.gatecheckListBuf, buf, p.Stderr, "gitleaks", p.DryRunEnabled); err != nil {
		slog.Error("cannot run gatecheck list on gitleaks report")
	}

	// Add a new line to separate the reports
	fmt.Fprintln(p.runtime.gatecheckListBuf, "")

	buf = new(bytes.Buffer)
	mw = io.MultiWriter(p.runtime.semgrepFile, buf)

	semgrepError := RunSemgrep(mw, p.Stderr, p.config, p.DryRunEnabled, p.SemgrepExperimental)
	slog.Debug("summarize semgrep report")
	if err := RunGatecheckList(p.runtime.gatecheckListBuf, buf, p.Stderr, "semgrep", p.DryRunEnabled); err != nil {
		slog.Error("cannot run gatecheck list on semgrep report")
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
	_, _ = p.runtime.gatecheckListBuf.WriteTo(p.Stdout)

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

	cmd := shell.GitleaksCommand(nil, nil, stdErr).DetectSecrets(config.CodeScan.GitleaksSrcDir, tempReportFilename)
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
		Scan(rules string) *shell.Command
	}

	if experimental {
		semgrep = shell.OSemgrepCommand(nil, reportDst, stdErr)
	} else {
		semgrep = shell.SemgrepCommand(nil, reportDst, stdErr)
	}

	// manually suppress errors for findings, convert to warnings
	// https://semgrep.dev/docs/semgrep-ci/configuring-blocking-and-errors-in-ci/
	// error code documentation: https://semgrep.dev/docs/cli-reference/
	if err := semgrep.Scan(config.CodeScan.SemgrepRules).WithDryRun(dryRunEnabled).Run(); err != nil {
		return errors.New("Code Scan Pipeline failed: Semgrep findings detected. See log for details.")
	}

	return nil
}
