package pipelines

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"workflow-engine/pkg/shell"
)

type CodeScan struct {
	Stdin                         io.Reader
	Stdout                        io.Writer
	Stderr                        io.Writer
	logger                        *slog.Logger
	DryRunEnabled                 bool
	SemgrepExperimental           bool
	SemgrepErrorOnFindingsEnabled bool
	SemgrepRules                  string
	artifactConfig                ArtifactConfig
}

type semgrepCLI interface {
	Scan(string) *shell.Command
}

func (s *CodeScan) WithArtifactConfig(config ArtifactConfig) *CodeScan {
	if config.Directory != "" {
		s.artifactConfig.Directory = config.Directory
	}
	if config.GitleaksFilename != "" {
		s.artifactConfig.GitleaksFilename = config.GitleaksFilename
	}
	if config.SemgrepFilename != "" {
		s.artifactConfig.SemgrepFilename = config.SemgrepFilename
	}
	return s
}

func NewCodeScan(stdout io.Writer, stderr io.Writer) *CodeScan {
	return &CodeScan{
		Stdin:  os.Stdin, // Default to OS stdin
		Stdout: stdout,
		Stderr: stderr,
		artifactConfig: ArtifactConfig{
			Directory:        os.TempDir(),
			GitleaksFilename: "gitleaks-secrets-scan-report.json",
			SemgrepFilename:  "semgrep-sast-report.json",
		},
		DryRunEnabled: false,
		logger:        slog.Default().With("pipeline", "code_scan"),
	}
}

func (p *CodeScan) Run() error {
	var semgrepFileError, gitleaksError, semgrepError error
	var semgrepReportFile *os.File

	p.logger = p.logger.With("dry_run_enabled", p.DryRunEnabled)
	p.logger = p.logger.With(
		"artifact_config.directory", p.artifactConfig.Directory,
		"artifact_config.gitleaks_filename", p.artifactConfig.GitleaksFilename,
		"artifact_config.semgrep_filename", p.artifactConfig.SemgrepFilename,
	)
	gitleaks := shell.GitleaksCommand(nil, p.Stdout, p.Stderr)
	gitleaksError = gitleaks.DetectSecrets(".", p.artifactConfig.GitleaksFilename).WithDryRun(p.DryRunEnabled).Run()

	if gitleaksError != nil {
		slog.Error("gitleaks detect", "error", gitleaksError)
	}

	var semgrep semgrepCLI

	slog.Debug("open semgrep file for output", "filename", p.artifactConfig.SemgrepFilename)
	semgrepReportFile, semgrepFileError = os.OpenFile(p.artifactConfig.SemgrepFilename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if semgrepFileError != nil {
		return errors.Join(gitleaksError, semgrepError)
	}

	defer semgrepReportFile.Close()

	if p.SemgrepExperimental {
		semgrep = shell.OSemgrepCommand(nil, semgrepReportFile, p.Stderr)
	} else {
		semgrep = shell.SemgrepCommand(nil, semgrepReportFile, p.Stderr)
	}

	// manually suppress errors for findings, convert to warnings
	// https://semgrep.dev/docs/semgrep-ci/configuring-blocking-and-errors-in-ci/
	semgrepError = semgrep.Scan(p.SemgrepRules).WithDryRun(p.DryRunEnabled).Run()
	if semgrepError != nil {
		// Note: Golang switch statements will only excute the first matching case
		switch {
		// error with suppression disabled
		case semgrepError.Error() == "exit status 1" && p.SemgrepErrorOnFindingsEnabled:
			return errors.Join(fmt.Errorf("Semgrep Findings: %w", semgrepError), gitleaksError)
		// error with suppression enabled (default)
		case semgrepError.Error() == "exit status 1":
			slog.Warn("Semgrep findings detected. See log for details.")
			semgrepError = nil
		default:
			// error code documentation: https://semgrep.dev/docs/cli-reference/
			slog.Error("semgrep unexpected command failure. See log for details.", "error", semgrepError)
		}

	}

	return errors.Join(gitleaksError, semgrepError, semgrepFileError)
}