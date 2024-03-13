package pipelines

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"slices"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config contains all parameters for the various pipelines
type Config struct {
	Version                 string             `mapstructure:"version"`
	ImageTag                string             `mapstructure:"imageTag"`
	ArtifactDir             string             `mapstructure:"artifactDir"`
	GatecheckBundleFilename string             `mapstructure:"gatecheckBundleFilename"`
	ImageBuild              configImageBuild   `mapstructure:"imageBuild"`
	ImageScan               configImageScan    `mapstructure:"imageScan"`
	CodeScan                configCodeScan     `mapstructure:"codeScan"`
	ImagePublish            configImagePublish `mapstructure:"imagePublish"`
	Deploy                  configDeploy       `mapstructure:"deploy"`
}

type configImageBuild struct {
	Enabled      bool     `mapstructure:"enabled"`
	BuildDir     string   `mapstructure:"buildDir"`
	Dockerfile   string   `mapstructure:"dockerfile"`
	Platform     string   `mapstructure:"platform"`
	Target       string   `mapstructure:"target"`
	CacheTo      string   `mapstructure:"cacheTo"`
	CacheFrom    string   `mapstructure:"cacheFrom"`
	SquashLayers bool     `mapstructure:"squashLayers"`
	Args         []string `mapstructure:"args"`
}

type configImageScan struct {
	Enabled             bool   `mapstructure:"enabled"`
	SyftFilename        string `mapstructure:"syftFilename"`
	GrypeConfigFilename string `mapstructure:"grypeConfigFilename"`
	GrypeFilename       string `mapstructure:"grypeFilename"`
	ClamavFilename      string `mapstructure:"clamavFilename"`
}

type configCodeScan struct {
	Enabled          bool   `mapstructure:"enabled"`
	GitleaksFilename string `mapstructure:"gitleaksFilename"`
	GitleaksSrcDir   string `mapstructure:"gitleaksSrcDir"`
	SemgrepFilename  string `mapstructure:"semgrepFilename"`
	SemgrepRules     string `mapstructure:"semgrepRules"`
}

type configImagePublish struct {
	Enabled              bool   `mapstructure:"enabled"`
	BundlePublishEnabled bool   `mapstructure:"bundlePublishEnabled"`
	BundleTag            string `mapstructure:"bundleTag"`
}

type configDeploy struct {
	Enabled                 bool   `mapstructure:"enabled"`
	GatecheckConfigFilename string `mapstructure:"gatecheckConfigFilename"`
}

// metaConfigField is used to map viper values to env variables and their associated default values
type metaConfigField struct {
	Key         string
	Env         string
	Default     any
	Description string
}

var metaConfig = []metaConfigField{
	{Key: "imagetag", Env: "WFE_IMAGE_TAG", Default: "my-app:latest",
		Description: "The full image tag for the target container image"},

	{Key: "artifactdir", Env: "WFE_ARTIFACT_DIR", Default: "artifacts",
		Description: "The target directory for all generated artifacts",
	},
	{Key: "gatecheckbundlefilename", Env: "WFE_GATECHECK_BUNDLE_FILENAME", Default: "artifacts/gatecheck-bundle.tar.gz",
		Description: "The filename for the gatecheck bundle, a validatable archive of security artifacts",
	},

	{Key: "imagebuild.enabled", Env: "WFE_IMAGE_BUILD_ENABLED", Default: true,
		Description: "Enable/Disable the image build pipeline",
	},
	{Key: "imagebuild.builddir", Env: "WFE_IMAGE_BUILD_DIR", Default: ".",
		Description: "The build directory to using during an image build",
	},
	{Key: "imagebuild.dockerfile", Env: "WFE_IMAGE_BUILD_DOCKERFILE", Default: "Dockerfile",
		Description: "The Dockerfile/Containerfile to use during an image build",
	},
	{Key: "imagebuild.platform", Env: "WFE_IMAGE_BUILD_PLATFORM", Default: nil,
		Description: "The target platform for build",
	},
	{Key: "imagebuild.target", Env: "WFE_IMAGE_BUILD_TARGET", Default: nil,
		Description: "The target build stage to build (e.g., [linux/amd64])",
	},
	{Key: "imagebuild.cacheto", Env: "WFE_IMAGE_BUILD_CACHE_TO", Default: nil,
		Description: "Cache export destinations (e.g., \"user/app:cache\", \"type=local,src=path/to/dir\")",
	},
	{Key: "imagebuild.cachefrom", Env: "WFE_IMAGE_BUILD_CACHE_FROM", Default: nil,
		Description: "External cache sources (e.g., \"user/app:cache\", \"type=local,src=path/to/dir\")",
	},
	{Key: "imagebuild.squashlayers", Env: "WFE_IMAGE_BUILD_SQUASH_LAYERS", Default: false,
		Description: "squash image layers - Only Supported with Podman CLI",
	},
	{Key: "imagebuild.args", Env: "WFE_IMAGE_BUILD_ARGS", Default: nil,
		Description: "Comma seperated list of build time variables",
	},

	{Key: "imagescan.enabled", Env: "WFE_IMAGE_SCAN_ENABLED", Default: true,
		Description: "Enable/Disable the image scan pipeline",
	},
	{Key: "imagescan.syftfilename", Env: "WFE_IMAGE_SCAN_SYFT_FILENAME", Default: "syft-sbom-report.json",
		Description: "The filename for the syft SBOM report - must contain 'syft'",
	},
	{Key: "imagescan.grypeconfigfilename", Env: "WFE_IMAGE_SCAN_GRYPE_CONFIG_FILENAME", Default: nil,
		Description: "The config filename for the grype vulnerability report",
	},
	{Key: "imagescan.grypefilename", Env: "WFE_IMAGE_SCAN_GRYPE_FILENAME", Default: "grype-vulnerability-report-full.json",
		Description: "The filename for the grype vulnerability report - must contain 'grype'",
	},
	{Key: "imagescan.clamavfilename", Env: "WFE_IMAGE_SCAN_CLAMAV_FILENAME", Default: "clamav-virus-report.txt",
		Description: "The filename for the clamscan virus report - must contain 'clamav'",
	},

	{Key: "codescan.enabled", Env: "WFE_CODE_SCAN_ENABLED", Default: true,
		Description: "Enable/Disable the code scan pipeline",
	},
	{Key: "codescan.gitleaksfilename", Env: "WFE_CODE_SCAN_GITLEAKS_FILENAME", Default: "gitleaks-secrets-report.json",
		Description: "The filename for the gitleaks secret report - must contain 'gitleaks'",
	},
	{Key: "codescan.gitleakssrcdir", Env: "WFE_CODE_SCAN_GITLEAKS_SRC_DIR", Default: ".",
		Description: "The target directory for the gitleaks scan",
	},
	{Key: "codescan.semgrepfilename", Env: "WFE_CODE_SCAN_SEMGREP_FILENAME", Default: "semgrep-sast-report.json",
		Description: "The filename for the semgrep SAST report - must contain 'semgrep'",
	},
	{Key: "codescan.semgreprules", Env: "WFE_CODE_SCAN_SEMGREP_RULES", Default: "p/default",
		Description: "Semgrep ruleset manual override",
	},

	{Key: "imagepublish.enabled", Env: "WFE_IMAGE_PUBLISH_ENABLED", Default: true,
		Description: "Enable/Disable the image publish pipeline",
	},
	{Key: "imagepublish.bundlepublishenabled", Env: "WFE_IMAGE_BUNDLE_PUBLISH_ENABLED", Default: true,
		Description: "Enable/Disable gatecheck artifact bundle publish task",
	},
	{Key: "imagepublish.bundletag", Env: "WFE_IMAGE_PUBLISH_BUNDLE_TAG", Default: "my-app/artifact-bundle:latest",
		Description: "The full image tag for the target gatecheck bundle image blob",
	},

	{Key: "deploy.enabled", Env: "WFE_IMAGE_PUBLISH_ENABLED", Default: true,
		Description: "Enable/Disable the deploy pipeline",
	},
	{Key: "deploy.gatecheckconfigfilename", Env: "WFE_DEPLOY_GATECHECK_CONFIG_FILENAME", Default: nil,
		Description: "The filename for the gatecheck config",
	},
}

func metaConfigLookup(key string) metaConfigField {
	for _, value := range metaConfig {
		if value.Key == key {
			return value
		}
	}
	return metaConfigField{}
}

func BindViper(v *viper.Viper) {
	for _, field := range metaConfig {
		viper.MustBindEnv(field.Key, field.Env)
		if field.Default != nil {
			viper.SetDefault(field.Key, field.Env)
		}
	}
}

type githubAction struct {
	Name        string                      `yaml:"name"`
	Description string                      `yaml:"description"`
	Inputs      map[string]actionInputField `yaml:"inputs"`
	Runs        actionRunsConfig            `yaml:"runs"`
}

type actionInputField struct {
	Description string `yaml:"description"`
	Default     string `yaml:"default,omitempty"`
}

type actionRunsConfig struct {
	Using string            `yaml:"using"`
	Image string            `yaml:"image"`
	Args  []string          `yaml:"args,flow"`
	Env   map[string]string `yaml:"env"`
}

type supportedField struct {
	key        string
	inputField string
}

func WriteGithubActionCodeScan(dst io.Writer, image string) error {
	supportedFields := []supportedField{
		{key: "artifactdir", inputField: "artifact_dir"},
		{key: "gatecheckbundlefilename", inputField: "gatecheck_bundle_filename"},
		{key: "codescan.gitleaksfilename", inputField: "gitleaks_filename"},
		{key: "codescan.gitleakssrcdir", inputField: "gitleaks_src_dir"},
		{key: "codescan.semgrepfilename", inputField: "semgrep_filename"},
		{key: "codescan.semgreprules", inputField: "semgrep_rules"},
	}

	action := githubAction{
		Name:        "Code Scan",
		Description: "Scan a code repository with Workflow Engine",
		Inputs: map[string]actionInputField{
			"config_file": {
				Description: "The workflow engine config file name",
				Default:     "",
			},
		},
		Runs: actionRunsConfig{
			Using: "docker",
			Image: image,
			Args:  []string{"run", "code-scan", "--config", "${{ inputs.config_file }}", "--verbose"},
			Env:   map[string]string{},
		},
	}

	return writeAction(action, supportedFields, dst)
}

func WriteGithubActionImageBuild(dst io.Writer, image string) error {
	supportedFields := []supportedField{
		{key: "imagetag", inputField: "image_tag"},
		{key: "artifactdir", inputField: "artifact_dir"},
		{key: "gatecheckbundlefilename", inputField: "gatecheck_bundle_filename"},
		{key: "imagebuild.builddir", inputField: "build_dir"},
		{key: "imagebuild.dockerfile", inputField: "dockerfile"},
		{key: "imagebuild.args", inputField: "args"},
		{key: "imagebuild.platform", inputField: "platform"},
		{key: "imagebuild.target", inputField: "target"},
		{key: "imagebuild.cacheto", inputField: "cache_to"},
		{key: "imagebuild.cachefrom", inputField: "cache_from"},
		{key: "imagebuild.squashlayers", inputField: "squash_layers"},
	}

	action := githubAction{
		Name:        "Build Image",
		Description: "Build a container image with Workflow Engine",
		Inputs: map[string]actionInputField{
			"config_file": {
				Description: "The workflow engine config file name",
				Default:     "",
			},
		},
		Runs: actionRunsConfig{
			Using: "docker",
			Image: image,
			Args:  []string{"run", "image-build", "--config", "${{ inputs.config_file }}", "--verbose"},
			Env:   map[string]string{},
		},
	}

	return writeAction(action, supportedFields, dst)
}

func WriteGithubActionImageScan(dst io.Writer, image string) error {
	supportedFields := []supportedField{
		{key: "artifactdir", inputField: "artifact_dir"},
		{key: "gatecheckbundlefilename", inputField: "gatecheck_bundle_filename"},
		{key: "imagescan.syftfilename", inputField: "syft_filename"},
		{key: "imagescan.grypeconfigfilename", inputField: "grype_config_filename"},
		{key: "imagescan.grypefilename", inputField: "grype_filename"},
		{key: "imagescan.clamavfilename", inputField: "clamav_filename"},
	}

	action := githubAction{
		Name:        "Image Scan",
		Description: "Scan a container image with Workflow Engine",
		Inputs: map[string]actionInputField{
			"config_file": {
				Description: "The workflow engine config file name",
				Default:     "",
			},
		},
		Runs: actionRunsConfig{
			Using: "docker",
			Image: image,
			Args:  []string{"run", "image-scan", "--config", "${{ inputs.config_file }}", "--verbose"},
			Env:   map[string]string{},
		},
	}

	return writeAction(action, supportedFields, dst)
}

func WriteGithubActionImagePublish(dst io.Writer, image string) error {
	supportedFields := []supportedField{
		{key: "imagepublish.bundletag", inputField: "bundle_tag"},
		{key: "imagepublish.bundlepublishenabled", inputField: "bundle_publish_enabled"},
	}

	action := githubAction{
		Name:        "Image Publish",
		Description: "Publish a container image with Workflow Engine",
		Inputs: map[string]actionInputField{
			"config_file": {
				Description: "The workflow engine config file name",
				Default:     "",
			},
		},
		Runs: actionRunsConfig{
			Using: "docker",
			Image: image,
			Args:  []string{"run", "image-publish", "--config", "${{ inputs.config_file }}", "--verbose"},
			Env:   map[string]string{},
		},
	}

	return writeAction(action, supportedFields, dst)
}

func WriteGithubActionDeploy(dst io.Writer, image string) error {
	supportedFields := []supportedField{
		{key: "deploy.gatecheckconfigfilename", inputField: "gatecheck_config_filename"},
	}

	action := githubAction{
		Name:        "Deploy Validation",
		Description: "Validate Artifacts with Workflow Engine for Deployment",
		Inputs:      map[string]actionInputField{},
		Runs: actionRunsConfig{
			Using: "docker",
			Image: image,
			Args:  []string{"run", "deploy", "--verbose"},
			Env:   map[string]string{},
		},
	}

	return writeAction(action, supportedFields, dst)
}

func writeAction(action githubAction, supportedFields []supportedField, dst io.Writer) error {

	for _, field := range supportedFields {
		configField := metaConfigLookup(field.key)
		defaultValue, ok := configField.Default.(string)
		if !ok {
			defaultValue = ""
		}
		action.Inputs[field.inputField] = actionInputField{
			Description: configField.Description,
			Default:     defaultValue,
		}
		action.Runs.Env[configField.Env] = fmt.Sprintf("${{ inputs.%s }}", field.inputField)
	}

	enc := yaml.NewEncoder(dst)
	enc.SetIndent(2)
	return enc.Encode(action)
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

func paddedMetaConfigData() [][4]string {
	data := [][4]string{{"Config Key", "Environment Variable", "Default Value", "Description"}}

	maxLens := []int{
		len(data[0][0]),
		len(data[0][1]),
		len(data[0][2]),
		len(data[0][3]),
	}
	printableMetaConfig := slices.Clone(metaConfig)

	// find the max length for each field in the slice
	for i := range printableMetaConfig {
		if printableMetaConfig[i].Default == nil {
			printableMetaConfig[i].Default = "-"
		}
		enabled, ok := printableMetaConfig[i].Default.(bool)
		if ok {
			printableMetaConfig[i].Default = "0"
			if enabled {
				printableMetaConfig[i].Default = "1"
			}
		}

		maxLens[0] = max(len(printableMetaConfig[i].Key), maxLens[0])
		maxLens[1] = max(len(printableMetaConfig[i].Env), maxLens[1])
		maxLens[2] = max(len(printableMetaConfig[i].Default.(string)), maxLens[2])
		maxLens[3] = max(len(printableMetaConfig[i].Description), maxLens[3])
	}

	slices.SortFunc(printableMetaConfig, func(a, b metaConfigField) int {
		return strings.Compare(a.Key, b.Key)
	})

	for i := 0; i < 4; i++ {
		// adjust header row
		format := fmt.Sprintf("%%-%ds", maxLens[i])
		data[0][i] = fmt.Sprintf(format, data[0][i])
	}

	for i := 1; i < len(printableMetaConfig); i++ {

		data = append(data, [4]string{})
		format := fmt.Sprintf("%%-%ds", maxLens[0])
		data[i][0] = fmt.Sprintf(format, printableMetaConfig[i].Key)

		format = fmt.Sprintf("%%-%ds", maxLens[1])
		data[i][1] = fmt.Sprintf(format, printableMetaConfig[i].Env)

		format = fmt.Sprintf("%%-%ds", maxLens[2])
		data[i][2] = fmt.Sprintf(format, printableMetaConfig[i].Default)

		format = fmt.Sprintf("%%-%ds", maxLens[3])
		data[i][3] = fmt.Sprintf(format, printableMetaConfig[i].Description)
	}

	return data
}

func WriteMarkdownEnv(dst io.Writer) error {
	data := paddedMetaConfigData()

	// header
	var headerErr error
	_, headerErr = fmt.Fprintf(dst, "| %s | %s | %s | %s |\n",
		data[0][0],
		data[0][1],
		data[0][2],
		data[0][3],
	)

	// header seperator
	_, err := fmt.Fprintf(dst, "| %s | %s | %s | %s |\n",
		strings.Repeat("-", len(data[0][0])),
		strings.Repeat("-", len(data[0][1])),
		strings.Repeat("-", len(data[0][2])),
		strings.Repeat("-", len(data[0][3])),
	)

	headerErr = errors.Join(headerErr, err)
	if headerErr != nil {
		return headerErr
	}

	// Rows of content, skip header
	for i := 1; i < len(data)-1; i++ {
		_, err = fmt.Fprintf(dst, "| %s | %s | %s | %s |\n", data[i][0], data[i][1], data[i][2], data[i][3])
		if err != nil {
			return err
		}
	}
	return nil
}
