package pipelines

import (
	"errors"
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

	semgrepFilename := path.Join(p.config.ArtifactsDir, p.config.CodeScan.SemgrepFilename)
	gitleaksFilename := path.Join(p.config.ArtifactsDir, p.config.CodeScan.GitleaksFilename)
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

	gitleaksError = runGitleaks(gitleaksFile, p.Stderr, p.config, p.DryRunEnabled)

	slog.Debug("open semgrep file for output", "filename", semgrepFilename)
	semgrepFile, err := os.OpenFile(semgrepFilename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		slog.Error("failed to create semgrep scan report file")
		return errors.New("Code Scan Pipeline failed, Semgrep did not run. See log for details.")
	}

	semgrepError = runSemgrep(semgrepFile, p.Stderr, p.config, p.DryRunEnabled, p.SemgrepExperimental)

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
