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
	Stdin          io.Reader
	Stdout         io.Writer
	Stderr         io.Writer
	logger         *slog.Logger
	DryRunEnabled  bool
	artifactConfig ArtifactConfig
	imageName      string
}

func (p *ImageScan) WithArtifactConfig(config ArtifactConfig) *ImageScan {
	if config.Directory != "" {
		p.artifactConfig.Directory = config.Directory
	}
	if config.SBOMFilename != "" {
		p.artifactConfig.SBOMFilename = config.SBOMFilename
	}
	if config.GrypeFilename != "" {
		p.artifactConfig.GrypeFilename = config.GrypeFilename
	}
	if config.GitleaksFilename != "" {
		p.artifactConfig.GitleaksFilename = config.GitleaksFilename
	}
	if config.SemgrepFilename != "" {
		p.artifactConfig.SemgrepFilename = config.SemgrepFilename
	}
	return p
}

func (p *ImageScan) WithImageName(imageName string) *ImageScan {
	p.imageName = imageName
	return p
}

func NewImageScan(stdout io.Writer, stderr io.Writer) *ImageScan {
	return &ImageScan{
		Stdin:  os.Stdin, // Default to OS stdin
		Stdout: stdout,
		Stderr: stderr,
		artifactConfig: ArtifactConfig{
			Directory:        os.TempDir(),
			SBOMFilename:     "sbom.json",
			GrypeFilename:    "scan.json",
			GitleaksFilename: "gitleaks.json",
			SemgrepFilename:  "semgrep-sast-report.json",
		},
		DryRunEnabled: false,
		logger:        slog.Default().With("pipeline", "image_scan"),
	}
}

func (p *ImageScan) Run() error {
	p.logger = p.logger.With("dry_run_enabled", p.DryRunEnabled)
	p.logger = p.logger.With(
		"artifact_config.directory", p.artifactConfig.Directory,
		"artifact_config.sbom_filename", p.artifactConfig.SBOMFilename,
		"artifact_config.grype_filename", p.artifactConfig.GrypeFilename,
		"artifact_config.gitleaks_filename", p.artifactConfig.GitleaksFilename,
	)

	dir, err := os.Stat(p.artifactConfig.Directory)
	if err != nil && os.IsNotExist(err) {
		err := os.MkdirAll(p.artifactConfig.Directory, 0755 /* rwxr-xr-x */)
		if err != nil {
			return err
		}
	} else if !dir.IsDir() {
		return errors.New("ArtifactConfig.Directory must be a directory, but it is a file")
	}

	// Do a gitleaks scan on the source directory, fail if the command fails
	gitleaksDirectory := p.artifactConfig.Directory
	gitleaksFilename := path.Join(p.artifactConfig.Directory, p.artifactConfig.GitleaksFilename)
	p.logger.Debug("open gitleaks dest file for write", "dest", gitleaksFilename)
	err = shell.GitleaksCommand(p.Stdin, p.Stdout, p.Stderr).DetectSecrets(gitleaksDirectory, gitleaksFilename).WithDryRun(p.DryRunEnabled).Run()
	if err != nil {
		return err
	}

	// Do a semgrep scan on the source directory, fail if the command fails
	semgrepFilename := path.Join(p.artifactConfig.Directory, p.artifactConfig.SemgrepFilename)
	p.logger.Debug("open semgrep dest file for write", "dest", semgrepFilename)

	semgrepFile, semgreperr := os.OpenFile(semgrepFilename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)

	if semgreperr != nil {
		return err
	}

	err = shell.SemgrepCommand(p.Stdin, semgrepFile, p.Stderr).ScanFile().WithDryRun(p.DryRunEnabled).Run()
	if err != nil {
		return err
	}

	// TODO: need syft SBOM output filename, it'll have to be saved in the artifact directory
	sbomFilename := path.Join(p.artifactConfig.Directory, p.artifactConfig.SBOMFilename)
	p.logger.Info("open sbom dest file for write", "dest", sbomFilename)

	sbomFile, err := os.OpenFile(sbomFilename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)

	if err != nil {
		return err
	}

	err = shell.SyftCommand(p.Stdin, sbomFile, p.Stderr).
		ScanImage(p.imageName).
		WithDryRun(p.DryRunEnabled).Run()

	if err != nil {
		sbomFile.Close()
		return err
	}

	sbomFile.Close()

	// Holds the grype scan output TODO: multi writer to the artifact directory and gatecheck
	buf := new(bytes.Buffer)

	// Do a grype scan on the SBOM, fail if the command fails
	err = shell.GrypeCommand(p.Stdin, buf, p.Stderr).ScanSBOM(sbomFilename).WithDryRun(p.DryRunEnabled).Run()
	if err != nil {
		return err
	}

	// Save the grype file to the artifact directory
	grypeFilename := path.Join(p.artifactConfig.Directory, p.artifactConfig.GrypeFilename)
	p.logger.Debug("open grype artifact", "dest", grypeFilename)
	grypeFile, err := os.OpenFile(grypeFilename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer grypeFile.Close()

	p.logger.Debug("save grype artifact", "dest", grypeFilename)
	if _, err := io.Copy(grypeFile, buf); err != nil {
		return err
	}

	return nil
}
