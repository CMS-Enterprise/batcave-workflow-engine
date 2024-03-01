package pipelines

import (
	"fmt"
	"html/template"
	"io"
	"log/slog"

	"github.com/go-git/go-git/v5"
)

// Config is the main configuration file for all of workflow engine
//
// The file is intended to be represented in json, yaml, or toml which is done via struct field tags
// Note: This is only intended to be the data based representation of values.
// For example, the Image field has values with tags that would represent the file structure of the
// config file. When it's passed to the image build pipeline, additional logic is used to build
// the image build commands.
type Config struct {
	Image     ImageConfig    `json:"image"     toml:"image"     yaml:"image"`
	Artifacts ArtifactConfig `json:"artifacts" toml:"artifacts" yaml:"artifacts"`
}

type ArtifactConfig struct {
	Directory                   string `json:"directory"                   toml:"directory"                   yaml:"directory"`
	BundleDirectory             string `json:"bundleDirectory"             toml:"bundleDirectory"             yaml:"bundleDirectory"`
	AntivirusFilename           string `json:"antivirusFilename"           toml:"antivirusFilename"           yaml:"antivirusFilename"`
	SBOMFilename                string `json:"sbomFilename"                toml:"sbomFilename"                yaml:"sbomFilename"`
	GrypeFilename               string `json:"grypeFilename"               toml:"grypeFilename"               yaml:"grypeFilename"`
	GrypeConfigFilename         string `json:"grypeConfigFilename"         toml:"grypeConfigFilename"         yaml:"grypeConfigFilename"`
	GrypeActiveFindingsFilename string `json:"grypeActiveFindingsFilename" toml:"grypeActiveFindingsFilename" yaml:"grypeActiveFindingsFilename"`
	GrypeAllFindingsFilename    string `json:"grypeAllFindingsFilename"    toml:"grypeAllFindingsFilename"    yaml:"grypeAllFindingsFilename"`
	GitleaksFilename            string `json:"gitleaksFilename"            toml:"gitleaksFilename"            yaml:"gitleaksFilename"`
	SemgrepFilename             string `json:"semgrepFilename"             toml:"semgrepFilename"             yaml:"semgrepFilename"`
	GatecheckBundleFilename     string `json:"gatecheckBundleFilename"     toml:"gatecheckBundleFilename"     yaml:"gatecheckBundleFilename"`
	GatecheckConfigFilename     string `json:"gatecheckConfigFilename"     toml:"gatecheckConfigFilename"     yaml:"gatecheckConfigFilename"`
}

// ImageConfig is a struct representation of the Image field in the Config file
type ImageConfig struct {
	BuildDir          string            `json:"buildDir"          toml:"buildDir"          yaml:"buildDir"`
	BuildDockerfile   string            `json:"buildDockerfile"   toml:"buildDockerfile"   yaml:"buildDockerfile"`
	BuildTag          string            `json:"buildTag"          toml:"buildTag"          yaml:"buildTag"`
	BuildPlatform     string            `json:"buildPlatform"     toml:"buildPlatform"     yaml:"buildPlatform"`
	BuildTarget       string            `json:"buildTarget"       toml:"buildTarget"       yaml:"buildTarget"`
	BuildCacheTo      string            `json:"buildCacheTo"      toml:"buildCacheTo"      yaml:"buildCacheTo"`
	BuildCacheFrom    string            `json:"buildCacheFrom"    toml:"buildCacheFrom"    yaml:"buildCacheFrom"`
	BuildSquashLayers bool              `json:"buildSquashLayers" toml:"buildSquashLayers" yaml:"buildSquashLayers"`
	BuildArgs         map[string]string `json:"buildArgs"         toml:"buildArgs"         yaml:"buildArgs"`
	ScanTarget        string            `json:"scanTarget"        toml:"scanTarget"        yaml:"scanTarget"`
}

// NewDefaultConfig creates a new "safe" config object.
// This can be used to prevent nil reference panics
func NewDefaultConfig() *Config {
	// Only fields that are slices need to be inited, the default string value is ""
	return &Config{
		Image: ImageConfig{
			BuildDir:        ".",
			BuildDockerfile: "Dockerfile",
			BuildArgs:       map[string]string{},
		},
		Artifacts: ArtifactConfig{
			Directory:                   "artifacts",
			BundleDirectory:             "artifacts",
			AntivirusFilename:           "clamav-report.txt",
			SBOMFilename:                "syft-sbom.json",
			GrypeFilename:               "grype-report.json",
			GrypeConfigFilename:         ".grype.yaml",
			GrypeActiveFindingsFilename: "active-findings-grype-scan.json",
			GrypeAllFindingsFilename:    "all-findings-grype-scan.json",
			GitleaksFilename:            "gitleaks-secrets-report.json",
			SemgrepFilename:             "semgrep-sast-report.json",
			GatecheckBundleFilename:     "gatecheck-bundle.tar.gz",
			GatecheckConfigFilename:     "gatecheck.yaml",
		},
	}
}

func RenderTemplate(dst io.Writer, templateSrc io.Reader) error {
	builtins, err := BuiltIns()
	if err != nil {
		return fmt.Errorf("template rendering failed: could not load built-in values: %w", err)
	}
	tmpl := template.New("workflow-engine config")

	content, err := io.ReadAll(templateSrc)
	if err != nil {
		return fmt.Errorf("template rendering failed: could not load template content: %w", err)
	}

	tmpl, err = tmpl.Parse(string(content))
	if err != nil {
		return fmt.Errorf("template rendering failed: could not parse template input: %w", err)
	}

	return tmpl.Execute(dst, builtins)
}

func BuiltIns() (map[string]string, error) {
	builtins := map[string]string{}

	slog.Debug("open current repo", "step", "builtins")
	r, err := git.PlainOpen(".")
	if err != nil {
		return builtins, err
	}

	slog.Debug("get repo HEAD")
	ref, err := r.Head()
	if err != nil {
		return builtins, err
	}

	builtins["GitCommitSHA"] = ref.Hash().String()
	builtins["GitCommitShortSHA"] = ref.Hash().String()[:8]
	builtins["GitCommitBranch"] = ref.Name().Short()

	return builtins, nil
}
