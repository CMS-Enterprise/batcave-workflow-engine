package pipelines

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"path"
	"workflow-engine/pkg/shell"
)

// TODO: pipeline-triggers currently does a curl to a gatecheck repo in GitLab to get the gatecheck.yaml file.
//
//	For now, this gatecheck.yaml file is hardcoded as follows, but it should be changed to point
//	to a gatecheck.yaml specific to the version that is used in omnibus.
var gatecheckYaml = `cyclonedx:
    allowList:
        - id: example allow id
          reason: example reason
    denyList:
        - id: example deny id
          reason: example reason
    required: false
    critical: -1
    high: -1
    medium: -1
    low: -1
    info: -1
    none: -1
    unknown: -1
gitleaks:
    secretsAllowed: false
grype:
    allowList:
        - id: example allow id
          reason: example reason
    denyList:
        - id: example deny id
          reason: example reason
    epssAllowThreshold: 0.01
    epssDenyThreshold: 0.6
    critical: 100
    high: 500
    medium: 1000
    low: 1000
    negligible: -1
    unknown: -1
semgrep:
    error: 1
    warning: 5
    info: -1
`

type imagePublish struct {
	Stdout         io.Writer
	Stderr         io.Writer
	DryRunEnabled  bool
	CLICmd         cliCmd
	logger         *slog.Logger
	imageConfig    ImageConfig
	artifactConfig ArtifactConfig
}

func (p *imagePublish) WithArtifactConfig(config ArtifactConfig) *imagePublish {
	if config.Directory != "" {
		p.artifactConfig.Directory = config.Directory
	}
	if config.AntivirusFilename != "" {
		p.artifactConfig.AntivirusFilename = config.AntivirusFilename
	}
	if config.GatecheckBundleFilename != "" {
		p.artifactConfig.GatecheckBundleFilename = config.GatecheckBundleFilename
	}
	if config.GatecheckConfigFilename != "" {
		p.artifactConfig.GatecheckConfigFilename = config.GatecheckConfigFilename
	}
	if config.GitleaksFilename != "" {
		p.artifactConfig.GitleaksFilename = config.GitleaksFilename
	}
	if config.GrypeFilename != "" {
		p.artifactConfig.GrypeFilename = config.GrypeFilename
	}
	if config.GrypeConfigFilename != "" {
		p.artifactConfig.GrypeConfigFilename = config.GrypeConfigFilename
	}
	if config.GrypeActiveFindingsFilename != "" {
		p.artifactConfig.GrypeActiveFindingsFilename = config.GrypeActiveFindingsFilename
	}
	if config.GrypeAllFindingsFilename != "" {
		p.artifactConfig.GrypeAllFindingsFilename = config.GrypeAllFindingsFilename
	}
	if config.SBOMFilename != "" {
		p.artifactConfig.SBOMFilename = config.SBOMFilename
	}
	if config.SemgrepFilename != "" {
		p.artifactConfig.SemgrepFilename = config.SemgrepFilename
	}
	return p
}

func NewimagePublish(stdout io.Writer, stderr io.Writer) *imagePublish {
	pipeline := &imagePublish{
		Stdout:        stdout,
		Stderr:        stderr,
		DryRunEnabled: false,
		logger:        slog.Default().With("pipeline", "image_package"),
		imageConfig: ImageConfig{
			BuildDockerfile: "Dockerfile",
		},
		artifactConfig: ArtifactConfig{
			Directory:                   ".artifacts",
			AntivirusFilename:           "clamav-report.txt",
			GatecheckBundleFilename:     "gatecheck-bundle.tar.gz",
			GatecheckConfigFilename:     "gatecheck.yaml",
			GrypeFilename:               "grype-image-scan.json",
			GrypeConfigFilename:         ".grype.yaml",
			GrypeActiveFindingsFilename: "active-findings-grype-scan.json",
			GrypeAllFindingsFilename:    "all-findings-grype-scan.json",
			GitleaksFilename:            "gitleaks-secrets-scan-report.json",
			SBOMFilename:                "syft-image-sbom.json",
			SemgrepFilename:             "semgrep-sast-report.json",
		},
	}

	pipeline.CLICmd = shell.DockerCommand(nil, pipeline.Stdout, pipeline.Stderr)

	return pipeline
}

func (i *imagePublish) Run() error {
	var gatecheckBundleError, gatecheckSummaryError error
	antivirusFilename := path.Join(i.artifactConfig.Directory, i.artifactConfig.AntivirusFilename)
	gatecheckBundleFilename := path.Join(i.artifactConfig.Directory, i.artifactConfig.GatecheckBundleFilename)
	gatecheckConfigFilename := path.Join(i.artifactConfig.Directory, i.artifactConfig.GatecheckConfigFilename)
	gitleaksFilename := path.Join(i.artifactConfig.Directory, i.artifactConfig.GitleaksFilename)
	grypeFilename := path.Join(i.artifactConfig.Directory, i.artifactConfig.GrypeFilename)
	grypeConfigFilename := path.Join(i.artifactConfig.Directory, i.artifactConfig.GrypeConfigFilename)
	grypeActiveFindingsFilename := path.Join(i.artifactConfig.Directory, i.artifactConfig.GrypeActiveFindingsFilename)
	grypeAllFindingsFilename := path.Join(i.artifactConfig.Directory, i.artifactConfig.GrypeAllFindingsFilename)
	sbomFilename := path.Join(i.artifactConfig.Directory, i.artifactConfig.SBOMFilename)
	semgrepFilename := path.Join(i.artifactConfig.Directory, i.artifactConfig.SemgrepFilename)
	dockerFilename := i.imageConfig.BuildDockerfile

	l := slog.Default()

	l.Info("start", "dry_run_enabled", i.DryRunEnabled)
	defer l.Info("complete")

	// Create gatecheck.yaml file
	gatecheckConfigFile := []byte(gatecheckYaml)
	gatecheckConfigError := os.WriteFile(gatecheckConfigFilename, gatecheckConfigFile, 0o644)

	// Run gatecheck bundle
	gatecheck := shell.GatecheckCommand(nil, i.Stdout, i.Stderr)
	gatecheckBundleError = gatecheck.Bundle(gatecheckBundleFilename, gitleaksFilename, grypeFilename, grypeConfigFilename, grypeActiveFindingsFilename, grypeAllFindingsFilename, sbomFilename, semgrepFilename, gatecheckConfigFilename, dockerFilename, antivirusFilename).WithDryRun(i.DryRunEnabled).Run()

	// Run gatecheck summary
	gatecheckSummaryError = gatecheck.Summary(gatecheckBundleFilename).WithDryRun(i.DryRunEnabled).Run()

	return errors.Join(gatecheckConfigError, gatecheckBundleError, gatecheckSummaryError)
}
