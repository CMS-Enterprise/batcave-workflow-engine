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

func (p *CodeScan) Run() error {
	var gitleaksError, semgrepError error

	gitleaksFilename := path.Join(p.config.ArtifactsDir, p.config.CodeScan.GitleaksFilename)
	semgrepFilename := path.Join(p.config.ArtifactsDir, p.config.CodeScan.SemgrepFilename)
	slog.Info("run image scan pipeline", "dry_run_enabled", p.DryRunEnabled, "artifact_directory", p.config.ArtifactsDir)

	slog.Debug("ensure artifact directory exists")
	if err := os.MkdirAll(p.config.ArtifactsDir, 0o755); err != nil {
		slog.Error("failed to create artifact directory", "directory", p.config.ArtifactsDir)
		return errors.New("Code Scan Pipeline failed to run. See log for details.")
	}

	slog.Debug("open gitleaks artifact for output", "filename", gitleaksFilename)
	gitleaksFile, err := os.OpenFile(gitleaksFilename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		slog.Error("cannot open gitleaks report file", "filename", gitleaksFilename, "error", err)
		return errors.New("Code Scan Pipeline failed, Gitleaks did not run. See log for details.")
	}
	defer gitleaksFile.Close()

	// All of the gatcheck summaries should print at the end
	summaryBuf := new(bytes.Buffer)
	buf := new(bytes.Buffer)
	// MultiWriter will write to the gitleaks file and to the buf so gatecheck can parse it
	mw := io.MultiWriter(gitleaksFile, buf)
	gitleaksError = runGitleaks(mw, p.Stderr, p.config, p.DryRunEnabled)

	slog.Debug("sumarize gitleaks report")
	err = shell.GatecheckCommand(buf, summaryBuf, p.Stderr).List("gitleaks").WithDryRun(p.DryRunEnabled).Run()
	if err != nil {
		slog.Error("cannot run gatecheck list on gitleaks report")
	}
	// Add a new line to separate the reports
	fmt.Fprintln(summaryBuf, "")

	slog.Debug("open semgrep file for output", "filename", semgrepFilename)
	semgrepFile, err := os.OpenFile(semgrepFilename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		slog.Error("failed to create semgrep scan report file")
		return errors.New("Code Scan Pipeline failed, Semgrep did not run. See log for details.")
	}
	defer semgrepFile.Close()

	buf = new(bytes.Buffer)
	mw = io.MultiWriter(semgrepFile, buf)

	semgrepError = runSemgrep(mw, p.Stderr, p.config, p.DryRunEnabled, p.SemgrepExperimental)
	slog.Debug("sumarize gitleaks report")
	err = shell.GatecheckCommand(buf, summaryBuf, p.Stderr).List("semgrep").WithDryRun(p.DryRunEnabled).Run()
	if err != nil {
		slog.Error("cannot run gatecheck list on semgrep report")
	}

	// print the summaries
	_, _ = summaryBuf.WriteTo(p.Stdout)

	return errors.Join(gitleaksError, semgrepError)
}

// runGitleaks the report will be written to stdout
func runGitleaks(reportDst io.Writer, stdErr io.Writer, config *Config, dryRunEnabled bool) error {
	slog.Debug("create temp gitleaks report", "dir", os.TempDir())
	reportFile, err := os.CreateTemp(os.TempDir(), "*-gitleaks-report.json")
	if err != nil {
		return err
	}

	tempReportFilename := reportFile.Name()

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

func runSemgrep(reportDst io.Writer, stdErr io.Writer, config *Config, dryRunEnabled bool, experimental bool) error {
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
