package pipelines

import (
	"fmt"
	"html/template"
	"io"
	"log/slog"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/viper"
)

// Config contains all parameters for the various pipelines
type Config struct {
	Version      string           `json:"version"     toml:"version"     yaml:"version"`
	ImageBuild   configImageBuild `json:"imageBuild"  toml:"imageBuild"  yaml:"imageBuild"`
	ImageScan    configImageScan  `json:"imageScan"   toml:"imageScan"   yaml:"imageScan"`
	CodeScan     configCodeScan   `json:"codeScan"    toml:"codeScan"    yaml:"codeScan"`
	ArtifactsDir string           `json:"artifactDir" toml:"artifactDir" yaml:"artifactDir"`
}

type configImageBuild struct {
	Enabled      bool              `json:"enabled"      toml:"enabled"      yaml:"enabled"`
	BuildDir     string            `json:"buildDir"     toml:"buildDir"     yaml:"buildDir"`
	Dockerfile   string            `json:"dockerfile"   toml:"dockerfile"   yaml:"dockerfile"`
	Tag          string            `json:"tag"          toml:"tag"          yaml:"tag"`
	Platform     string            `json:"platform"     toml:"platform"     yaml:"platform"`
	Target       string            `json:"target"       toml:"target"       yaml:"target"`
	CacheTo      string            `json:"cacheTo"      toml:"cacheTo"      yaml:"cacheTo"`
	CacheFrom    string            `json:"cacheFrom"    toml:"cacheFrom"    yaml:"cacheFrom"`
	SquashLayers bool              `json:"squashLayers" toml:"squashLayers" yaml:"squashLayers"`
	Args         map[string]string `json:"args"         toml:"args"         yaml:"args"`
	ScanTarget   string            `json:"scanTarget"   toml:"scanTarget"   yaml:"scanTarget"`
}

type configImageScan struct {
	Enabled             bool   `json:"enabled"             toml:"enabled"             yaml:"enabled"`
	AntivirusFilename   string `json:"antivirusFilename"   toml:"antivirusFilename"   yaml:"antivirusFilename"`
	SyftFilename        string `json:"syftFilename"        toml:"syftFilename"        yaml:"syftFilename"`
	GrypeConfigFilename string `json:"grypeConfigFilename" toml:"grypeConfigFilename" yaml:"grypeConfigFilename"`
	GrypeActiveFilename string `json:"grypeActiveFilename" toml:"grypeActiveFilename" yaml:"grypeActiveFilename"`
	GrypeFullFilename   string `json:"grypeFullFilename"   toml:"grypeFullFilename"   yaml:"grypeFullFilename"`
	TargetImage         string `json:"targetImage"         toml:"targetImage"         yaml:"targetImage"`
}

type configCodeScan struct {
	Enabled          bool   `json:"enabled"          toml:"enabled"          yaml:"enabled"`
	GitleaksFilename string `json:"gitleaksFilename" toml:"gitleaksFilename" yaml:"gitleaksFilename"`
	GitleaksSrcDir   string `json:"gitleaksSrcDir"   toml:"gitleaksSrcDir"   yaml:"gitleaksSrcDir"`
	SemgrepFilename  string `json:"semgrepFilename"  toml:"semgrepFilename"  yaml:"semgrepFilename"`
	SemgrepRules     string `json:"semgrepRules"     toml:"semgrepRules"     yaml:"semgrepRules"`
}

func SetDefaults(v *viper.Viper) {
	v.MustBindEnv("artifactsdir", "WFE_ARTIFACTS_DIR")
	v.SetDefault("artifactsdir", "artifacts")

	v.MustBindEnv("imagebuild.enabled", "WFE_IMAGE_BUILD_ENABLED")
	v.MustBindEnv("imagebuild.builddir", "WFE_IMAGE_BUILD_DIR")
	v.MustBindEnv("imagebuild.dockerfile", "WFE_IMAGE_BUILD_DOCKERFILE")
	v.MustBindEnv("imagebuild.tag", "WFE_IMAGE_BUILD_TAG")
	v.MustBindEnv("imagebuild.platform", "WFE_BUILD_IMAGE_PLATFORM")
	v.MustBindEnv("imagebuild.target", "WFE_IMAGE_BUILD_TARGET")
	v.MustBindEnv("imagebuild.cacheto", "WFE_IMAGE_BUILD_CACHE_TO")
	v.MustBindEnv("imagebuild.cachefrom", "WFE_IMAGE_BUILD_CACHE_FROM")
	v.MustBindEnv("imagebuild.squashlayers", "WFE_IMAGE_BUILD_SQUASH_LAYERS")
	v.MustBindEnv("imagebuild.scantarget", "WFE_IMAGE_BUILD_SCAN_TARGET")

	v.SetDefault("imagebuild.enabled", "1")
	v.SetDefault("imagebuild.builddir", ".")
	v.SetDefault("imagebuild.dockerfile", "Dockerfile")
	v.SetDefault("imagebuild.tag", "latest")

	v.MustBindEnv("imagescan.enabled", "WFE_IMAGE_SCAN_ENABLED")
	v.MustBindEnv("imagescan.clamavFilename", "WFE_IMAGE_SCAN_CLAMAV_FILENAME")
	v.MustBindEnv("imagescan.syftFilename", "WFE_IMAGE_SCAN_SYFT_FILENAME")
	v.MustBindEnv("imagescan.grypeConfigFilename", "WFE_IMAGE_SCAN_GRYPE_CONFIG_FILENAME")
	v.MustBindEnv("imagescan.grypeActiveFindingsFilename", "WFE_IMAGE_SCAN_GRYPE_ACTIVE_FINDINGS_FILENAME")
	v.MustBindEnv("imagescan.grypeAllFindingsFilename", "WFE_IMAGE_SCAN_GRYPE_ALL_FINDINGS_FILENAME")

	v.SetDefault("imagescan.enabled", "1")
	v.SetDefault("imagescan.clamavFilename", "clamav-virus-report.txt")
	v.SetDefault("imagescan.syftFilename", "syft-sbom-report.json")
	v.SetDefault("imagescan.grypeActiveFilename", "grype-vulnerability-report-active.json")
	v.SetDefault("imagescan.grypeFullFilename", "grype-vulnerability-report-full.json")

	v.MustBindEnv("codescan.enabled", "WFE_CODE_SCAN_ENABLED")
	v.MustBindEnv("codescan.gitleaksFilename", "WFE_CODE_SCAN_GITLEAKS_FILENAME")
	v.MustBindEnv("codescan.gitleaksSrcDir", "WFE_CODE_SCAN_GITLEAKS_SRC_DIR")
	v.MustBindEnv("codescan.semgrepFilename", "WFE_CODE_SCAN_SEMGREP_FILENAME")
	v.MustBindEnv("codescan.semgrepRules", "WFE_CODE_SCAN_SEMGREP_RULES")

	v.SetDefault("codescan.enabled", "1")
	v.SetDefault("codescan.gitleaksFilename", "gitleaks-secrets-report.json")
	v.SetDefault("codescan.gitleaksSrcDir", ".")
	v.SetDefault("codescan.semgrepFilename", "semgrep-sast-report.json")
	v.SetDefault("codescan.semgrepRules", "p/default")
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
