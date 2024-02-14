package pipelines

import (
	"io"
	"log/slog"
	"os"
)

type CodeScan struct {
	Stdin          io.Reader
	Stdout         io.Writer
	Stderr         io.Writer
	logger         *slog.Logger
	DryRunEnabled  bool
	artifactConfig ArtifactConfig
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
	p.logger = p.logger.With("dry_run_enabled", p.DryRunEnabled)
	p.logger = p.logger.With(
		"artifact_config.directory", p.artifactConfig.Directory,
		"artifact_config.gitleaks_filename", p.artifactConfig.GitleaksFilename,
		"artifact_config.semgrep_filename", p.artifactConfig.SemgrepFilename,
	)
	// TODO: Add workflow
	return nil
}
