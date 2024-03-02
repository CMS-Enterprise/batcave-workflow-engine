package pipelines

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"os"
	"path"
	"workflow-engine/pkg/shell"
)

type ImageScan struct {
	Stdout        io.Writer
	Stderr        io.Writer
	DryRunEnabled bool
	config        *Config
}

func (p *ImageScan) WithConfig(config *Config) *ImageScan {
	p.config = config
	return p
}

func NewImageScan(stdout io.Writer, stderr io.Writer) *ImageScan {
	return &ImageScan{
		Stdout:        stdout,
		Stderr:        stderr,
		DryRunEnabled: false,
		config:        new(Config),
	}
}

func (p *ImageScan) Run() error {
	slog.Info("run image scan pipeline", "dry_run_enabled", p.DryRunEnabled, "artifact_directory", p.config.ArtifactsDir)

	slog.Debug("ensure artifact directory exists")
	if err := os.MkdirAll(p.config.ArtifactsDir, 0o755); err != nil {
		slog.Error("failed to create artifact directory", "directory", p.config.ArtifactsDir)
		return errors.New("Code Scan Pipeline failed to run. See log for details.")
	}

	sbomFilename := path.Join(p.config.ArtifactsDir, p.config.ImageScan.TargetImage)
	slog.Info("open sbom dest file for write", "dest", sbomFilename)

	sbomFile, err := os.OpenFile(sbomFilename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		slog.Error("failed to open syft sbom file", "filename", sbomFilename, "error", err)
		return err
	}

	grypeFilename := path.Join(p.config.ArtifactsDir, p.config.ImageScan.GrypeFullFilename)
	slog.Info("open grype dest file for write", "dest", grypeFilename)

	grypeFile, err := os.OpenFile(grypeFilename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		slog.Error("failed to open grype file", "filename", grypeFilename, "error", err)
		return errors.New("image Scan Pipeline failed. See log for details")
	}

	syftReportBuf := new(bytes.Buffer)
	syftMW := io.MultiWriter(sbomFile, syftReportBuf)

	if err = runSyft(syftMW, p.Stderr, p.config, p.DryRunEnabled); err != nil {
		slog.Error("syft sbom generation failed")
		return errors.New("image Scan Pipeline failed. See log for details")
	}

	grypeReportBuf := new(bytes.Buffer)
	grypeMW := io.MultiWriter(grypeFile, grypeReportBuf)

	if err := runGrype(grypeMW, syftReportBuf, p.Stderr, p.config, p.DryRunEnabled); err != nil {
		return err
	}

	// TODO: use the buffers for gatecheck sumamry and bundle

	return nil
}

func runSyft(reportDst io.Writer, stdErr io.Writer, config *Config, dryRunEnabled bool) error {
	return shell.SyftCommand(nil, reportDst, stdErr).ScanImage(config.ImageScan.TargetImage).WithDryRun(dryRunEnabled).Run()
}

func runGrype(reportDst io.Writer, syftSrc io.Reader, stdErr io.Writer, config *Config, dryRunEnabled bool) error {
	return shell.GrypeCommand(syftSrc, reportDst, stdErr).ScanSBOM().WithDryRun(dryRunEnabled).Run()
}
