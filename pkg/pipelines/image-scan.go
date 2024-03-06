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
	runtime       struct {
		sbomFile         *os.File
		grypeFile        *os.File
		clamavFile       *os.File
		gatecheckListBuf *bytes.Buffer
		sbomFilename     string
		grypeFilename    string
		clamavFilename   string
		bundleFilename   string
	}
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

func (p *ImageScan) preRun() error {
	var err error

	if err := MakeDirectoryP(p.config.ArtifactsDir); err != nil {
		slog.Error("failed to create artifact directory", "name", p.config.ArtifactsDir)
		return errors.New("Code Scan Pipeline failed to run. See log for details.")
	}

	p.runtime.sbomFilename = path.Join(p.config.ArtifactsDir, p.config.ImageScan.SyftFilename)
	p.runtime.sbomFile, err = OpenOrCreateFile(p.runtime.sbomFilename)
	if err != nil {
		slog.Error("cannot open syft sbom file", "filename", p.runtime.sbomFilename, "error", err)
		return err
	}

	p.runtime.grypeFilename = path.Join(p.config.ArtifactsDir, p.config.ImageScan.GrypeFullFilename)
	p.runtime.grypeFile, err = OpenOrCreateFile(p.runtime.grypeFilename)
	if err != nil {
		slog.Error("cannot open grype sbom file", "filename", p.runtime.grypeFilename, "error", err)
		return err
	}

	p.runtime.clamavFilename = path.Join(p.config.ArtifactsDir, p.config.ImageScan.ClamavFilename)
	p.runtime.clamavFile, err = OpenOrCreateFile(p.runtime.clamavFilename)
	if err != nil {
		slog.Error("cannot open clamav file", "filename", p.runtime.clamavFilename, "error", err)
		return err
	}

	if err := InitGatecheckBundle(p.config, p.Stderr, p.DryRunEnabled); err != nil {
		slog.Error("cannot initialize gatecheck bundle", "error", err)
		return err
	}

	p.runtime.bundleFilename = path.Join(p.config.ArtifactsDir, p.config.GatecheckBundleFilename)
	p.runtime.gatecheckListBuf = new(bytes.Buffer)

	return nil
}

func (p *ImageScan) Run() error {
	if !p.config.ImageScan.Enabled {
		slog.Warn("image scan pipeline is disabled, skip.")
		return nil
	}

	if err := p.preRun(); err != nil {
		return errors.New("Code Scan Pipeline Pre-Run Failed. See log for details.")
	}

	defer func() {
		_ = p.runtime.sbomFile.Close()
		_ = p.runtime.grypeFile.Close()
		_ = p.runtime.clamavFile.Close()
	}()

	slog.Info("run image scan pipeline", "dry_run_enabled", p.DryRunEnabled, "artifact_directory", p.config.ArtifactsDir)

	syftReportBuf := new(bytes.Buffer)
	syftMW := io.MultiWriter(p.runtime.sbomFile, syftReportBuf)

	if err := RunSyftScan(syftMW, p.Stderr, p.config, p.DryRunEnabled); err != nil {
		slog.Error("syft sbom generation failed")
		return errors.New("image Scan Pipeline failed. See log for details")
	}

	grypeReportBuf := new(bytes.Buffer)
	grypeMW := io.MultiWriter(p.runtime.grypeFile, grypeReportBuf)

	grypeScanError := RunGrypeScanSBOM(grypeMW, syftReportBuf, p.Stderr, p.config, p.DryRunEnabled)
	if grypeScanError != nil {
		return errors.New("image Scan Pipeline failed. See log for details")
	}

	slog.Debug("summarize grype report")
	err := RunGatecheckListAll(p.runtime.gatecheckListBuf, grypeReportBuf, p.Stderr, "grype", p.DryRunEnabled)
	if err != nil {
		slog.Error("cannot run gatecheck list all on grype report")
		return errors.New("image Scan Pipeline failed. See log for details")
	}

	clamavBuf := new(bytes.Buffer)
	clamavMW := io.MultiWriter(p.runtime.clamavFile, clamavBuf)

	// Do a ClamAV (freshclam) update on the CVD database
	slog.Debug("update clamav database")
	freshClamErr := RunFreshClam(clamavMW, clamavBuf, p.Stderr, p.config, p.DryRunEnabled)
	if freshClamErr != nil {
		slog.Error("failed to update clamav database:", freshClamErr)
		return errors.New("image Scan Pipeline failed. See log for details")
	}

	// Do a ClamAV scan on the target directory, fail if the command fails
	slog.Debug("scan target directory with clamav")
	clamScanErr := RunClamavScan(clamavMW, clamavBuf, p.Stderr, p.config, p.DryRunEnabled)
	if clamScanErr != nil {
		slog.Error("clamav failed to scan target directory:", clamScanErr)
		return errors.New("image Scan Pipeline failed. See log for details")
	}
	
	if err := p.postRun(); err != nil {
		return errors.New("Code Scan Pipeline Post-Run Failed. See log for details.")
	}
	return nil
}

func (p *ImageScan) postRun() error {
	files := []string{p.runtime.sbomFilename, p.runtime.grypeFilename, p.runtime.clamavFilename}
	err := RunGatecheckBundleAdd(p.runtime.bundleFilename, p.Stderr, p.DryRunEnabled, files...)
	if err != nil {
		slog.Error("cannot run gatecheck bundle add", "error", err)
	}

	// print the Gatecheck List Content
	_, _ = p.runtime.gatecheckListBuf.WriteTo(p.Stdout)

	return err
}

func RunSyftScan(reportDst io.Writer, stdErr io.Writer, config *Config, dryRunEnabled bool) error {
	return shell.SyftCommand(nil, reportDst, stdErr).ScanImage(config.ImageScan.TargetImage).WithDryRun(dryRunEnabled).Run()
}

func RunGrypeScanSBOM(reportDst io.Writer, syftSrc io.Reader, stdErr io.Writer, config *Config, dryRunEnabled bool) error {
	return shell.GrypeCommand(syftSrc, reportDst, stdErr).ScanSBOM().WithDryRun(dryRunEnabled).Run()
}

func RunFreshClam(reportDst io.Writer, clamavSrc io.Reader, stdErr io.Writer, config *Config, dryRunEnabled bool) error {
	return shell.FreshClamCommand(clamavSrc, reportDst, stdErr).Run().WithDryRun(dryRunEnabled).Run()
}

func RunClamavScan(reportDst io.Writer, clamavSrc io.Reader, stdErr io.Writer, config *Config, dryRunEnabled bool) error {
	return shell.ClamScanCommand(clamavSrc, reportDst, stdErr).Scan(".").WithDryRun(dryRunEnabled).Run()
}
