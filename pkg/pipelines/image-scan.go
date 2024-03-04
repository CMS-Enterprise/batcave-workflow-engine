package pipelines

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
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

	if err := MakeDirectoryP(p.config.ArtifactsDir); err != nil {
		slog.Error("failed to create artifact directory", "directory", p.config.ArtifactsDir)
		return errors.New("Code Scan Pipeline failed to run. See log for details.")
	}

	sbomFilename := path.Join(p.config.ArtifactsDir, p.config.ImageScan.TargetImage)
	slog.Info("open sbom dest file for write", "dest", sbomFilename)

	sbomFile, err := OpenOrCreateFile(sbomFilename)
	if err != nil {
		slog.Error("failed to open syft sbom file", "filename", sbomFilename, "error", err)
		return err
	}

	grypeFilename := path.Join(p.config.ArtifactsDir, p.config.ImageScan.GrypeFullFilename)
	slog.Info("open grype dest file for write", "dest", grypeFilename)

	grypeFile, err := OpenOrCreateFile(grypeFilename)
	if err != nil {
		slog.Error("failed to open grype file", "filename", grypeFilename, "error", err)
		return errors.New("image Scan Pipeline failed. See log for details")
	}

	syftReportBuf := new(bytes.Buffer)
	syftMW := io.MultiWriter(sbomFile, syftReportBuf)

	if err = RunSyftScan(syftMW, p.Stderr, p.config, p.DryRunEnabled); err != nil {
		slog.Error("syft sbom generation failed")
		return errors.New("image Scan Pipeline failed. See log for details")
	}

	grypeReportBuf := new(bytes.Buffer)
	grypeMW := io.MultiWriter(grypeFile, grypeReportBuf)

	if err := RunGrypeScanSBOM(grypeMW, syftReportBuf, p.Stderr, p.config, p.DryRunEnabled); err != nil {

		return errors.New("image Scan Pipeline failed. See log for details")
	}

	slog.Debug("summarize grype report")
	err = RunGatecheckListAll(p.Stdout, grypeReportBuf, p.Stderr, "grype", p.DryRunEnabled)
	if err != nil {
		slog.Error("cannot run gatecheck list all on grype report")
	}

	// Holds the ClamAV scan output TODO: multi writer to the artifact directory and gatecheck
	clamavBuf := new(bytes.Buffer)

	// Do a ClamAV (freshclam) update on the CVD database
	err = shell.FreshClamCommand(nil, p.Stdout, p.Stderr).Run().WithDryRun(p.DryRunEnabled).Run()
	if err != nil {
		return err
	}

	// Do a ClamAV scan on the target directory, fail if the command fails
	err = shell.ClamScanCommand(nil, clamavBuf, p.Stderr).Scan(".").WithDryRun(p.DryRunEnabled).Run()
	if err != nil {
		return err
	}

	// Save the ClamAV file to the artifact directory
	clamavFilename := path.Join(p.artifactConfig.Directory, p.artifactConfig.ClamavFilename)
	p.logger.Debug("open clamav artifact", "dest", clamavFilename)
	clamavFile, err := os.OpenFile(clamavFilename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer clamavFile.Close()

	p.logger.Debug("save clamav artifact", "dest", clamavFilename)
	if _, err := io.Copy(clamavFile, buf); err != nil {
		return err
	}
	
	return nil
}

func RunSyftScan(reportDst io.Writer, stdErr io.Writer, config *Config, dryRunEnabled bool) error {
	return shell.SyftCommand(nil, reportDst, stdErr).ScanImage(config.ImageScan.TargetImage).WithDryRun(dryRunEnabled).Run()
}

func RunGrypeScanSBOM(reportDst io.Writer, syftSrc io.Reader, stdErr io.Writer, config *Config, dryRunEnabled bool) error {
	return shell.GrypeCommand(syftSrc, reportDst, stdErr).ScanSBOM().WithDryRun(dryRunEnabled).Run()
}
