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
	Stdin                         io.Reader
	Stdout                        io.Writer
	Stderr                        io.Writer
	DryRunEnabled                 bool
	SemgrepExperimental           bool
	SemgrepErrorOnFindingsEnabled bool
	SemgrepRules                  string
	artifactConfig                ConfigArtifacts
}

func (s *CodeScan) WithConfig(config ConfigArtifacts) *CodeScan {
	if config.Directory != "" {
		s.artifactConfig.Directory = config.Directory
	}
	if config.GitleaksFilename != "" {
		s.artifactConfig.GitleaksFilename = config.GitleaksFilename
	}
	if config.SemgrepFilename != "" {
		s.artifactConfig.SemgrepFilename = config.SemgrepFilename
	}
	if config.GatecheckBundleFilename != "" {
		s.artifactConfig.GatecheckBundleFilename = config.GatecheckBundleFilename
	}
	return s
}

func NewCodeScan(stdout io.Writer, stderr io.Writer) *CodeScan {
	return &CodeScan{
		Stdin:          os.Stdin, // Default to OS stdin TODO: possible vulnerability, config injection
		Stdout:         stdout,
		Stderr:         stderr,
		artifactConfig: NewDefaultConfig().Artifacts,
		DryRunEnabled:  false,
	}
}

func (p *CodeScan) Run() error {
	var gitleaksError, semgrepError error

	semgrepFilename := path.Join(p.artifactConfig.Directory, p.artifactConfig.SemgrepFilename)
	gitleaksFilename := path.Join(p.artifactConfig.Directory, p.artifactConfig.GitleaksFilename)
	slog.Info("run image scan pipeline",
		"dry_run_enabled", p.DryRunEnabled,
		"artifact_config.directory", p.artifactConfig.Directory,
		"artifact_config.gitleaks_filename", p.artifactConfig.GitleaksFilename,
		"artifact_config.semgrep_filename", p.artifactConfig.SemgrepFilename,
	)

	if err := os.MkdirAll(p.artifactConfig.Directory, 0o755); err != nil {
		slog.Error("failed to create artifact directory", "directory", p.artifactConfig.Directory)
		return errors.New("Code Scan Pipeline failed to run. See log for details.")
	}

	gitleaksFile, err := os.OpenFile(gitleaksFilename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		slog.Error("cannot open gitleaks report file", "filename", gitleaksFilename, "error", err)
		return errors.New("Code Scan Pipeline failed, Gitleaks did not run. See log for details.")
	}
	defer gitleaksFile.Close()

	gitleaksError = runGitleaks(gitleaksFile, p.Stderr, p.artifactConfig, p.DryRunEnabled)
	if gitleaksError != nil {
		slog.Error("gitleaks scan failed. continue pipeline")
	}

	slog.Debug("open semgrep file for output", "filename", semgrepFilename)
	semgrepFile, err := os.OpenFile(semgrepFilename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		slog.Error("failed to create semgrep scan report file")
		return errors.New("Code Scan Pipeline failed, Semgrep did not run. See log for details.")
	}

	semgrepError = runSemgrep(semgrepFile, p.Stderr, p.DryRunEnabled, p.SemgrepExperimental, p.SemgrepRules, p.SemgrepErrorOnFindingsEnabled)
	if semgrepError != nil {
		return errors.New("Code Scan Pipeline failed")
	}

	return errors.Join(gitleaksError, semgrepError)
}

// runGitleaks the report will be written to stdout
func runGitleaks(reportDst io.Writer, stdErr io.Writer, config ConfigArtifacts, dryRunEnabled bool) error {
	slog.Debug("create temp gitleaks report", "dir", os.TempDir())
	reportFile, err := os.CreateTemp(os.TempDir(), "*-gitleaks-report.json")
	if err != nil {
		return err
	}

	tempReportFilename := reportFile.Name()

	cmd := shell.GitleaksCommand(nil, nil, stdErr).DetectSecrets(config.GitleaksSrcDir, tempReportFilename)
	err = cmd.WithDryRun(dryRunEnabled).Run()
	if err != nil {
		return err
	}

	// Seek errors are really unlikely, just join with the copy error in the rare case that it occurs
	_, seekErr := reportFile.Seek(0, io.SeekStart)

	_, copyErr := io.Copy(reportDst, reportFile)
	return errors.Join(seekErr, copyErr)
}

func runSemgrep(reportDst io.Writer, stdErr io.Writer, dryRunEnabled bool, experimental bool, rules string, errOnFindings bool) error {
	var semgrep interface {
		Scan(string) *shell.Command
	}

	if experimental {
		semgrep = shell.OSemgrepCommand(nil, reportDst, stdErr)
	} else {
		semgrep = shell.SemgrepCommand(nil, reportDst, stdErr)
	}

	// manually suppress errors for findings, convert to warnings
	// https://semgrep.dev/docs/semgrep-ci/configuring-blocking-and-errors-in-ci/
	// error code documentation: https://semgrep.dev/docs/cli-reference/
	semgrepError := semgrep.Scan(rules).WithDryRun(dryRunEnabled).Run()

	// Note: Golang switch statements will only excute the first matching case
	switch {
	case semgrepError == nil:
		break
	// error with suppression disabled
	case semgrepError.Error() == "exit status 1" && errOnFindings:
		return errors.New("Code Scan Pipeline failed. See log for details.")
	// error with suppression enabled (default)
	case semgrepError.Error() == "exit status 1":
		slog.Warn("Semgrep findings detected. See log for details.")
	}

	return nil
}
