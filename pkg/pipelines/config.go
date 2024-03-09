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
	Version                 string             `json:"version"                 toml:"version"                 yaml:"version"`
	ImageBuild              configImageBuild   `json:"imageBuild"              toml:"imageBuild"              yaml:"imageBuild"`
	ImageScan               configImageScan    `json:"imageScan"               toml:"imageScan"               yaml:"imageScan"`
	CodeScan                configCodeScan     `json:"codeScan"                toml:"codeScan"                yaml:"codeScan"`
	ImagePublish            configImagePublish `json:"imagePublish"            toml:"imagePublish"            yaml:"imagePublish"`
	Deploy                  configDeploy       `json:"deploy"                  toml:"deploy"                  yaml:"deploy"`
	ArtifactDir             string             `json:"artifactDir"             toml:"artifactDir"             yaml:"artifactDir"`
	GatecheckBundleFilename string             `json:"gatecheckBundleFilename" toml:"gatecheckBundleFilename" yaml:"gatecheckBundleFilename"`
}

type configImageBuild struct {
	Enabled      bool     `json:"enabled"      toml:"enabled"      yaml:"enabled"`
	BuildDir     string   `json:"buildDir"     toml:"buildDir"     yaml:"buildDir"`
	Dockerfile   string   `json:"dockerfile"   toml:"dockerfile"   yaml:"dockerfile"`
	Tag          string   `json:"tag"          toml:"tag"          yaml:"tag"`
	Platform     string   `json:"platform"     toml:"platform"     yaml:"platform"`
	Target       string   `json:"target"       toml:"target"       yaml:"target"`
	CacheTo      string   `json:"cacheTo"      toml:"cacheTo"      yaml:"cacheTo"`
	CacheFrom    string   `json:"cacheFrom"    toml:"cacheFrom"    yaml:"cacheFrom"`
	SquashLayers bool     `json:"squashLayers" toml:"squashLayers" yaml:"squashLayers"`
	Args         []string `json:"args"         toml:"args"         yaml:"args"`
	ScanTarget   string   `json:"scanTarget"   toml:"scanTarget"   yaml:"scanTarget"`
}

type configImageScan struct {
	Enabled             bool   `json:"enabled"             toml:"enabled"             yaml:"enabled"`
	SyftFilename        string `json:"syftFilename"        toml:"syftFilename"        yaml:"syftFilename"`
	GrypeConfigFilename string `json:"grypeConfigFilename" toml:"grypeConfigFilename" yaml:"grypeConfigFilename"`
	GrypeActiveFilename string `json:"grypeActiveFilename" toml:"grypeActiveFilename" yaml:"grypeActiveFilename"`
	GrypeFullFilename   string `json:"grypeFullFilename"   toml:"grypeFullFilename"   yaml:"grypeFullFilename"`
	ClamavFilename      string `json:"clamavFilename"      toml:"clamavFilename"      yaml:"clamavFilename"`
	TargetImage         string `json:"targetImage"         toml:"targetImage"         yaml:"targetImage"`
}

type configCodeScan struct {
	Enabled          bool   `json:"enabled"          toml:"enabled"          yaml:"enabled"`
	GitleaksFilename string `json:"gitleaksFilename" toml:"gitleaksFilename" yaml:"gitleaksFilename"`
	GitleaksSrcDir   string `json:"gitleaksSrcDir"   toml:"gitleaksSrcDir"   yaml:"gitleaksSrcDir"`
	SemgrepFilename  string `json:"semgrepFilename"  toml:"semgrepFilename"  yaml:"semgrepFilename"`
	SemgrepRules     string `json:"semgrepRules"     toml:"semgrepRules"     yaml:"semgrepRules"`
}

type configImagePublish struct {
	Enabled              bool   `json:"enabled"              toml:"enabled"              yaml:"enabled"`
	BundlePublishEnabled bool   `json:"bundlePublishEnabled" toml:"bundlePublishEnabled" yaml:"bundlePublishEnabled"`
	BundleTag            string `json:"bundleTag"            toml:"bundleTag"            yaml:"bundleTag"`
}

type configDeploy struct {
	Enabled bool `json:"enabled" toml:"enabled" yaml:"enabled"`
}

func BindEnvs(v *viper.Viper) {
	v.MustBindEnv("artifactdir", "WFE_ARTIFACT_DIR")
	v.MustBindEnv("gatecheckBundleFilename", "WFE_GATECHECK_BUNDLE_FILENAME")

	v.MustBindEnv("imagebuild.enabled", "WFE_IMAGE_BUILD_ENABLED")
	v.MustBindEnv("imagebuild.builddir", "WFE_IMAGE_BUILD_DIR")
	v.MustBindEnv("imagebuild.dockerfile", "WFE_IMAGE_BUILD_DOCKERFILE")
	v.MustBindEnv("imagebuild.tag", "WFE_IMAGE_BUILD_TAG")
	v.MustBindEnv("imagebuild.platform", "WFE_BUILD_IMAGE_PLATFORM")
	v.MustBindEnv("imagebuild.target", "WFE_IMAGE_BUILD_TARGET")
	v.MustBindEnv("imagebuild.cacheto", "WFE_IMAGE_BUILD_CACHE_TO")
	v.MustBindEnv("imagebuild.cachefrom", "WFE_IMAGE_BUILD_CACHE_FROM")
	v.MustBindEnv("imagebuild.squashlayers", "WFE_IMAGE_BUILD_SQUASH_LAYERS")

	v.MustBindEnv("imagescan.enabled", "WFE_IMAGE_SCAN_ENABLED")
	v.MustBindEnv("imagescan.clamavFilename", "WFE_IMAGE_SCAN_CLAMAV_FILENAME")
	v.MustBindEnv("imagescan.syftFilename", "WFE_IMAGE_SCAN_SYFT_FILENAME")
	v.MustBindEnv("imagescan.grypeConfigFilename", "WFE_IMAGE_SCAN_GRYPE_CONFIG_FILENAME")
	v.MustBindEnv("imagescan.grypeActiveFindingsFilename", "WFE_IMAGE_SCAN_GRYPE_ACTIVE_FINDINGS_FILENAME")
	v.MustBindEnv("imagescan.grypeAllFindingsFilename", "WFE_IMAGE_SCAN_GRYPE_ALL_FINDINGS_FILENAME")
	v.MustBindEnv("imagescan.targetimage", "WFE_IMAGE_SCAN_TARGET_IMAGE")

	v.MustBindEnv("codescan.enabled", "WFE_CODE_SCAN_ENABLED")
	v.MustBindEnv("codescan.gitleaksFilename", "WFE_CODE_SCAN_GITLEAKS_FILENAME")
	v.MustBindEnv("codescan.gitleaksSrcDir", "WFE_CODE_SCAN_GITLEAKS_SRC_DIR")
	v.MustBindEnv("codescan.semgrepFilename", "WFE_CODE_SCAN_SEMGREP_FILENAME")
	v.MustBindEnv("codescan.semgrepRules", "WFE_CODE_SCAN_SEMGREP_RULES")

	v.MustBindEnv("imagepublish.enabled", "WFE_IMAGE_PUBLISH_ENABLED")
	v.MustBindEnv("imagepublish.bundlepublishenabled", "WFE_IMAGE_PUBLISH_BUNDLE_PUBLISH_ENABLED")
	v.MustBindEnv("imagepublish.bundletag", "WFE_BUNDLE_TAG")

	v.MustBindEnv("deploy.enabled", "WFE_DEPLOY_ENABLED")
}

func SetDefaults(v *viper.Viper) {
	v.SetDefault("version", "1")
	v.SetDefault("artifactdir", "artifacts")

	v.SetDefault("gatecheckBundleFilename", "gatecheck-bundle.tar.gz")

	v.SetDefault("imagebuild.enabled", "1")
	v.SetDefault("imagebuild.builddir", ".")
	v.SetDefault("imagebuild.dockerfile", "Dockerfile")
	v.SetDefault("imagebuild.tag", "my-app:latest")

	v.SetDefault("imagescan.enabled", "1")
	v.SetDefault("imagescan.clamavFilename", "clamav-virus-report.txt")
	v.SetDefault("imagescan.syftFilename", "syft-sbom-report.json")
	v.SetDefault("imagescan.grypeActiveFilename", "grype-vulnerability-report-active.json")
	v.SetDefault("imagescan.grypeFullFilename", "grype-vulnerability-report-full.json")

	v.SetDefault("codescan.enabled", "1")
	v.SetDefault("codescan.gitleaksFilename", "gitleaks-secrets-report.json")
	v.SetDefault("codescan.gitleaksSrcDir", ".")
	v.SetDefault("codescan.semgrepFilename", "semgrep-sast-report.json")
	v.SetDefault("codescan.semgrepRules", "p/default")

	v.SetDefault("imagepublish.enabled", "1")
	v.SetDefault("imagepublish.bundlepublishenabled", "1")
	v.SetDefault("imagepublish.bundletag", "my-app/artifact-bundle:latest")

	v.SetDefault("deploy.enabled", "1")
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
