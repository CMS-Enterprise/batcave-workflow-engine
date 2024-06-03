package settings

import (
	"errors"
	"reflect"
	"strings"

	"github.com/sagikazarmark/slog-shim"
	"github.com/spf13/cobra"
)

// Config contains all parameters for the various pipelines
type Config struct {
	Version                 string             `mapstructure:"version"`
	ImageTag                string             `mapstructure:"imageTag"                metafield:"ImageTag"`
	ArtifactDir             string             `mapstructure:"artifactDir"             metafield:"ArtifactDir"`
	GatecheckBundleFilename string             `mapstructure:"gatecheckBundleFilename" metafield:"GatecheckBundleFilename"`
	ImageBuild              configImageBuild   `mapstructure:"imageBuild"`
	ImageScan               configImageScan    `mapstructure:"imageScan"`
	CodeScan                configCodeScan     `mapstructure:"codeScan"`
	ImagePublish            configImagePublish `mapstructure:"imagePublish"`
	Validation              configValidation   `mapstructure:"deploy"`
}

func NewConfig() *Config {
	return &Config{
		ImageBuild: configImageBuild{
			Args: []string{},
		},
	}
}

type configImageBuild struct {
	Enabled      bool     `mapstructure:"enabled"      metafield:"ImageBuildEnabled"`
	BuildDir     string   `mapstructure:"buildDir"     metafield:"ImageBuildBuildDir"`
	Dockerfile   string   `mapstructure:"dockerfile"   metafield:"ImageBuildDockerfile"`
	Platform     string   `mapstructure:"platform"     metafield:"ImageBuildPlatform"`
	Target       string   `mapstructure:"target"       metafield:"ImageBuildTarget"`
	CacheTo      string   `mapstructure:"cacheTo"      metafield:"ImageBuildCacheTo"`
	CacheFrom    string   `mapstructure:"cacheFrom"    metafield:"ImageBuildCacheFrom"`
	SquashLayers bool     `mapstructure:"squashLayers" metafield:"ImageBuildSquashLayers"`
	Args         []string `mapstructure:"args"         metafield:"ImageBuildArgs"`
}

type configImageScan struct {
	Enabled             bool   `mapstructure:"enabled"             metafield:"ImageScanEnabled"`
	SyftFilename        string `mapstructure:"syftFilename"        metafield:"ImageScanSyftFilename"`
	GrypeConfigFilename string `mapstructure:"grypeConfigFilename" metafield:"ImageScanGrypeConfigFilename"`
	GrypeFilename       string `mapstructure:"grypeFilename"       metafield:"ImageScanGrypeFilename"`
	ClamavFilename      string `mapstructure:"clamavFilename"      metafield:"ImageScanClamavFilename"`
}

type configCodeScan struct {
	Enabled          bool   `mapstructure:"enabled"          metafield:"CodeScanEnabled"`
	GitleaksFilename string `mapstructure:"gitleaksFilename" metafield:"CodeScanGitleaksFilename"`
	GitleaksSrcDir   string `mapstructure:"gitleaksSrcDir"   metafield:"CodeScanGitleaksSrcDir"`
	SemgrepFilename  string `mapstructure:"semgrepFilename"  metafield:"CodeScanSemgrepFilename"`
	SemgrepRules     string `mapstructure:"semgrepRules"`
}

type configImagePublish struct {
	Enabled              bool   `mapstructure:"enabled"              metafield:"ImagePublishEnabled"`
	BundlePublishEnabled bool   `mapstructure:"bundlePublishEnabled" metafield:"ImagePublishBundleEnabled"`
	BundleTag            string `mapstructure:"bundleTag"            metafield:"ImagePublishBundleTag"`
}

type configValidation struct {
	Enabled                 bool   `mapstructure:"enabled"                 metafield:"ValidationEnabled"`
	GatecheckConfigFilename string `mapstructure:"gatecheckConfigFilename" metafield:"ValidationGatecheckConfigFilename"`
}

type MetaConfig struct {
	ImageTag                          MetaField
	ArtifactDir                       MetaField
	GatecheckBundleFilename           MetaField
	ImageBuildEnabled                 MetaField
	ImageBuildBuildDir                MetaField
	ImageBuildDockerfile              MetaField
	ImageBuildPlatform                MetaField
	ImageBuildTarget                  MetaField
	ImageBuildCacheTo                 MetaField
	ImageBuildCacheFrom               MetaField
	ImageBuildSquashLayers            MetaField
	ImageBuildArgs                    MetaField
	ImageScanEnabled                  MetaField
	ImageScanSyftFilename             MetaField
	ImageScanGrypeConfigFilename      MetaField
	ImageScanGrypeFilename            MetaField
	ImageScanClamavFilename           MetaField
	CodeScanEnabled                   MetaField
	CodeScanGitleaksFilename          MetaField
	CodeScanSemgrepFilename           MetaField
	CodeScanGitleaksSrcDir            MetaField
	ImagePublishEnabled               MetaField
	ImagePublishBundleEnabled         MetaField
	ImagePublishBundleTag             MetaField
	ValidationEnabled                 MetaField
	ValidationGatecheckConfigFilename MetaField
}

func Unmarshal(toConfig *Config, fromMetaConfig *MetaConfig) error {
	slog.Debug("start unmarshal")
	if toConfig == nil || fromMetaConfig == nil {
		return errors.New("dst/src is nil")
	}

	toConfigFields := make(map[string]reflect.Value)

	values := []reflect.Value{reflect.ValueOf(toConfig).Elem()}
	for len(values) != 0 {
		// pop operation
		value := values[len(values)-1]
		values = values[:len(values)-1]

		for i := 0; i < value.NumField(); i++ {

			if value.Field(i).Kind() == reflect.Struct {
				values = append(values, value.Field(i))
				continue
			}

			metaFieldStr := value.Type().Field(i).Tag.Get("metafield")
			if metaFieldStr == "" {
				continue
			}
			toConfigFields[metaFieldStr] = value.Field(i)
		}
	}

	metaConfigValue := reflect.ValueOf(fromMetaConfig).Elem()

	for key, toValue := range toConfigFields {
		slog.Debug("evaluate field", "key", key)
		_, exists := metaConfigValue.Type().FieldByName(key)
		if !exists {
			slog.Error("field not found", "key", key)
			return nil
		}
		field, ok := metaConfigValue.FieldByName(key).Interface().(MetaField)
		if !ok {
			panic("invalid metafield type")
		}

		v := reflect.ValueOf(field.MustEvaluate())
		toValue.Set(v)
	}

	return nil

}

func MustUnmarshal(toConfig *Config, fromMetaConfig *MetaConfig) {
	err := Unmarshal(toConfig, fromMetaConfig)
	if err != nil {
		panic(err)
	}
}

func NewMetaConfig() *MetaConfig {
	m := &MetaConfig{
		ImageTag: MetaField{
			FlagValueP:      new(string),
			FlagName:        "tag",
			FlagDesc:        "The full image tag for the target container image",
			EnvKey:          "WFE_IMAGE_TAG",
			ActionInputName: "tag",
			ActionType:      "String",
			DefaultValue:    "my-app:latest",
			stringDecoder:   stringToStringDecoder,
			cobraFunc: func(f *MetaField, c *cobra.Command) {
				c.Flags().StringVar(f.FlagValueP.(*string), f.FlagName, f.DefaultValue, f.FlagDesc)
			},
		},
		ArtifactDir: MetaField{
			FlagValueP:      new(string),
			FlagName:        "artifact-dir",
			FlagDesc:        "The target directory for all generated artifacts",
			EnvKey:          "WFE_ARTIFACT_DIR",
			ActionInputName: "artifact_dir",
			ActionType:      "String",
			DefaultValue:    "artifacts",
			stringDecoder:   stringToStringDecoder,
			cobraFunc: func(f *MetaField, c *cobra.Command) {
				c.Flags().StringVar(f.FlagValueP.(*string), f.FlagName, f.DefaultValue, f.FlagDesc)
			},
		},
		GatecheckBundleFilename: MetaField{
			FlagValueP:      new(string),
			FlagName:        "bundle-filename",
			FlagDesc:        "The filename for the gatecheck bundle, a validatable archive of security artifacts",
			EnvKey:          "WFE_GATECHECK_BUNDLE_FILENAME",
			ActionInputName: "gatecheck_bundle_filename",
			ActionType:      "String",
			DefaultValue:    "gatecheck-bundle.tar.gz",
			stringDecoder:   stringToStringDecoder,
			cobraFunc: func(f *MetaField, c *cobra.Command) {
				c.Flags().StringVar(f.FlagValueP.(*string), f.FlagName, f.DefaultValue, f.FlagDesc)
			},
		},
		ImageBuildEnabled: MetaField{
			FlagValueP:      new(bool),
			FlagName:        "enabled",
			FlagDesc:        "Enable/Disable the image build pipeline",
			EnvKey:          "WFE_IMAGE_BUILD_ENABLED",
			ActionInputName: "image_build_enabled",
			ActionType:      "Bool",
			DefaultValue:    "true",
			stringDecoder:   stringToBoolDecoder,
			cobraFunc: func(f *MetaField, c *cobra.Command) {
				defaultValue, _ := stringToBoolDecoder(f.DefaultValue)
				c.Flags().BoolVar(f.FlagValueP.(*bool), f.FlagName, defaultValue.(bool), f.FlagDesc)
			},
		},
		ImageBuildBuildDir: MetaField{
			FlagValueP:      new(string),
			FlagName:        "build-dir",
			FlagDesc:        "The build directory to use during an image build",
			EnvKey:          "WFE_IMAGE_BUILD_DIR",
			ActionInputName: "build_dir",
			ActionType:      "String",
			DefaultValue:    ".",
			stringDecoder:   stringToStringDecoder,
			cobraFunc: func(f *MetaField, c *cobra.Command) {
				c.Flags().StringVar(f.FlagValueP.(*string), f.FlagName, f.DefaultValue, f.FlagDesc)
			},
		},
		ImageBuildDockerfile: MetaField{
			FlagValueP:      new(string),
			FlagName:        "dockerfile",
			FlagDesc:        "The Dockerfile/Containerfile to use during an image build",
			EnvKey:          "WFE_IMAGE_BUILD_DOCKERFILE",
			ActionInputName: "dockerfile",
			ActionType:      "String",
			DefaultValue:    "Dockerfile",
			stringDecoder:   stringToStringDecoder,
			cobraFunc: func(f *MetaField, c *cobra.Command) {
				c.Flags().StringVar(f.FlagValueP.(*string), f.FlagName, f.DefaultValue, f.FlagDesc)
			},
		},
		ImageBuildPlatform: MetaField{
			FlagValueP:      new(string),
			FlagName:        "platform",
			FlagDesc:        "The target platform for build",
			EnvKey:          "WFE_IMAGE_BUILD_PLATFORM",
			ActionInputName: "platform",
			ActionType:      "String",
			DefaultValue:    "",
			stringDecoder:   stringToStringDecoder,
			cobraFunc: func(f *MetaField, c *cobra.Command) {
				c.Flags().StringVar(f.FlagValueP.(*string), f.FlagName, f.DefaultValue, f.FlagDesc)
			},
		},
		ImageBuildTarget: MetaField{
			FlagValueP:      new(string),
			FlagName:        "target",
			FlagDesc:        "The target build stage to build (e.g., [linux/amd64])",
			EnvKey:          "WFE_IMAGE_BUILD_TARGET",
			ActionInputName: "target",
			ActionType:      "String",
			DefaultValue:    "",
			stringDecoder:   stringToStringDecoder,
			cobraFunc: func(f *MetaField, c *cobra.Command) {
				c.Flags().StringVar(f.FlagValueP.(*string), f.FlagName, f.DefaultValue, f.FlagDesc)
			},
		},
		ImageBuildCacheTo: MetaField{
			FlagValueP:      new(string),
			FlagName:        "cache-to",
			FlagDesc:        "Cache export destinations (e.g., \"user/app:cache\", \"type=local,src=path/to/dir\")",
			EnvKey:          "WFE_IMAGE_BUILD_CACHE_TO",
			ActionInputName: "cache_to",
			ActionType:      "String",
			DefaultValue:    "",
			stringDecoder:   stringToStringDecoder,
			cobraFunc: func(f *MetaField, c *cobra.Command) {
				c.Flags().StringVar(f.FlagValueP.(*string), f.FlagName, f.DefaultValue, f.FlagDesc)
			},
		},
		ImageBuildCacheFrom: MetaField{
			FlagValueP:      new(string),
			FlagName:        "cache-from",
			FlagDesc:        "External cache sources (e.g., \"user/app:cache\", \"type=local,src=path/to/dir\")",
			EnvKey:          "WFE_IMAGE_BUILD_CACHE_FROM",
			ActionInputName: "cache_from",
			ActionType:      "String",
			DefaultValue:    "",
			stringDecoder:   stringToStringDecoder,
			cobraFunc: func(f *MetaField, c *cobra.Command) {
				c.Flags().StringVar(f.FlagValueP.(*string), f.FlagName, f.DefaultValue, f.FlagDesc)
			},
		},
		ImageBuildSquashLayers: MetaField{
			FlagValueP:      new(bool),
			FlagName:        "squash-layers",
			FlagDesc:        "squash image layers - Only Supported with Podman CLI",
			EnvKey:          "WFE_IMAGE_BUILD_SQUASH_LAYERS",
			ActionInputName: "squash_layers",
			ActionType:      "Bool",
			DefaultValue:    "false",
			stringDecoder:   stringToBoolDecoder,
			cobraFunc: func(f *MetaField, c *cobra.Command) {
				defaultValue, _ := stringToBoolDecoder(f.DefaultValue)
				c.Flags().BoolVar(f.FlagValueP.(*bool), f.FlagName, defaultValue.(bool), f.FlagDesc)
			},
		},
		ImageBuildArgs: MetaField{
			FlagValueP:      new([]string),
			FlagName:        "build-args",
			FlagDesc:        "Comma separated list of build time variables",
			EnvKey:          "WFE_IMAGE_BUILD_ARGS",
			ActionInputName: "build_args",
			ActionType:      "List",
			DefaultValue:    "",
			stringDecoder: func(s string) (any, error) {
				return strings.Split(s, ","), nil
			},
			cobraFunc: func(f *MetaField, c *cobra.Command) {
				// TODO: Image Build Args
				// c.Flags().StringSliceVar(f.FlagValueP.([]string{}), f.FlagName, f.DefaultValue.([]string), f.FlagDesc)
			},
		},
		ImageScanEnabled: MetaField{
			FlagValueP:      new(bool),
			FlagName:        "enabled",
			FlagDesc:        "Enable/Disable the image scan pipeline",
			EnvKey:          "WFE_IMAGE_SCAN_ENABLED",
			ActionInputName: "image_scan_enabled",
			ActionType:      "Bool",
			DefaultValue:    "true",
			stringDecoder:   stringToBoolDecoder,
			cobraFunc: func(f *MetaField, c *cobra.Command) {
				defaultValue, _ := stringToBoolDecoder(f.DefaultValue)
				c.Flags().BoolVar(f.FlagValueP.(*bool), f.FlagName, defaultValue.(bool), f.FlagDesc)
			},
		},
		ImageScanSyftFilename: MetaField{
			FlagValueP:      new(string),
			FlagName:        "syft-filename",
			FlagDesc:        "The filename for the syft SBOM report - must contain 'syft'",
			EnvKey:          "WFE_IMAGE_SCAN_SYFT_FILENAME",
			ActionInputName: "syft_filename",
			ActionType:      "String",
			DefaultValue:    "sbom-report.syft.json",
			stringDecoder:   stringToStringDecoder,
			cobraFunc: func(f *MetaField, c *cobra.Command) {
				c.Flags().StringVar(f.FlagValueP.(*string), f.FlagName, f.DefaultValue, f.FlagDesc)
			},
		},
		ImageScanGrypeConfigFilename: MetaField{
			FlagValueP:      new(string),
			FlagName:        "grype-config-filename",
			FlagDesc:        "The config filename for the grype vulnerability report",
			EnvKey:          "WFE_IMAGE_SCAN_GRYPE_CONFIG_FILENAME",
			ActionInputName: "grype_config_filename",
			ActionType:      "String",
			DefaultValue:    "",
			stringDecoder:   stringToStringDecoder,
			cobraFunc: func(f *MetaField, c *cobra.Command) {
				c.Flags().StringVar(f.FlagValueP.(*string), f.FlagName, f.DefaultValue, f.FlagDesc)
			},
		},
		ImageScanGrypeFilename: MetaField{
			FlagValueP:      new(string),
			FlagName:        "grype-filename",
			FlagDesc:        "The filename for the grype vulnerability report - must contain 'grype'",
			EnvKey:          "WFE_IMAGE_SCAN_GRYPE_FILENAME",
			ActionInputName: "grype_filename",
			ActionType:      "String",
			DefaultValue:    "image-vulnerability-report.grype.json",
			stringDecoder:   stringToStringDecoder,
			cobraFunc: func(f *MetaField, c *cobra.Command) {
				c.Flags().StringVar(f.FlagValueP.(*string), f.FlagName, f.DefaultValue, f.FlagDesc)
			},
		},
		ImageScanClamavFilename: MetaField{
			FlagValueP:      new(string),
			FlagName:        "clamav-filename",
			FlagDesc:        "The filename for the clamscan virus report - must contain 'clamav'",
			EnvKey:          "WFE_IMAGE_SCAN_CLAMAV_FILENAME",
			ActionInputName: "clamav_filename",
			ActionType:      "String",
			DefaultValue:    "virus-report.clamav.txt",
			stringDecoder:   stringToStringDecoder,
			cobraFunc: func(f *MetaField, c *cobra.Command) {
				c.Flags().StringVar(f.FlagValueP.(*string), f.FlagName, f.DefaultValue, f.FlagDesc)
			},
		},
		CodeScanEnabled: MetaField{
			FlagValueP:      new(bool),
			FlagName:        "enabled",
			FlagDesc:        "Enable/Disable the code scan pipeline",
			EnvKey:          "WFE_CODE_SCAN_ENABLED",
			ActionInputName: "code_scan_enabled",
			ActionType:      "Bool",
			DefaultValue:    "true",
			stringDecoder:   stringToBoolDecoder,
			cobraFunc: func(f *MetaField, c *cobra.Command) {
				defaultValue, _ := stringToBoolDecoder(f.DefaultValue)
				c.Flags().BoolVar(f.FlagValueP.(*bool), f.FlagName, defaultValue.(bool), f.FlagDesc)
			},
		},
		CodeScanSemgrepFilename: MetaField{
			FlagValueP:      new(string),
			FlagName:        "semgrep-filename",
			FlagDesc:        "The filename for the semgrep code scan report - must contain 'gitleaks'",
			EnvKey:          "WFE_CODE_SCAN_SEMGREP_FILENAME",
			ActionInputName: "semgrep_filename",
			ActionType:      "String",
			DefaultValue:    "code-scan-report.semgrep.json",
			stringDecoder:   stringToStringDecoder,
			cobraFunc: func(f *MetaField, c *cobra.Command) {
				c.Flags().StringVar(f.FlagValueP.(*string), f.FlagName, f.DefaultValue, f.FlagDesc)
			},
		},
		CodeScanGitleaksFilename: MetaField{
			FlagValueP:      new(string),
			FlagName:        "gitleaks-filename",
			FlagDesc:        "The filename for the gitleaks secret report - must contain 'gitleaks'",
			EnvKey:          "WFE_CODE_SCAN_GITLEAKS_FILENAME",
			ActionInputName: "gitleaks_filename",
			ActionType:      "String",
			DefaultValue:    "secrets-report.gitleaks.json",
			stringDecoder:   stringToStringDecoder,
			cobraFunc: func(f *MetaField, c *cobra.Command) {
				c.Flags().StringVar(f.FlagValueP.(*string), f.FlagName, f.DefaultValue, f.FlagDesc)
			},
		},
		CodeScanGitleaksSrcDir: MetaField{
			FlagValueP:      new(string),
			FlagName:        "gitleaks-src-dir",
			FlagDesc:        "The target directory for the gitleaks scan",
			EnvKey:          "WFE_CODE_SCAN_GITLEAKS_SRC_DIR",
			ActionInputName: "gitleaks_src_dir",
			ActionType:      "String",
			DefaultValue:    ".",
			stringDecoder:   stringToStringDecoder,
			cobraFunc: func(f *MetaField, c *cobra.Command) {
				c.Flags().StringVar(f.FlagValueP.(*string), f.FlagName, f.DefaultValue, f.FlagDesc)
			},
		},
		ImagePublishEnabled: MetaField{
			FlagValueP:      new(bool),
			FlagName:        "enabled",
			FlagDesc:        "Enable/Disable the image publish pipeline",
			EnvKey:          "WFE_IMAGE_PUBLISH_ENABLED",
			ActionInputName: "image_publish_enabled",
			ActionType:      "Bool",
			DefaultValue:    "true",
			stringDecoder:   stringToBoolDecoder,
			cobraFunc: func(f *MetaField, c *cobra.Command) {
				defaultValue, _ := stringToBoolDecoder(f.DefaultValue)
				c.Flags().BoolVar(f.FlagValueP.(*bool), f.FlagName, defaultValue.(bool), f.FlagDesc)
			},
		},
		ImagePublishBundleEnabled: MetaField{
			FlagValueP:      new(bool),
			FlagName:        "enabled",
			FlagDesc:        "Enable/Disable gatecheck artifact bundle publish task",
			EnvKey:          "WFE_IMAGE_BUNDLE_PUBLISH_ENABLED",
			ActionInputName: "bundle_publish_enabled",
			ActionType:      "Bool",
			DefaultValue:    "true",
			stringDecoder:   stringToBoolDecoder,
			cobraFunc: func(f *MetaField, c *cobra.Command) {
				defaultValue, _ := stringToBoolDecoder(f.DefaultValue)
				c.Flags().BoolVar(f.FlagValueP.(*bool), f.FlagName, defaultValue.(bool), f.FlagDesc)
			},
		},
		ImagePublishBundleTag: MetaField{
			FlagValueP:      new(string),
			FlagName:        "bundle-tag",
			FlagDesc:        "The full image tag for the target gatecheck bundle image blob",
			EnvKey:          "WFE_IMAGE_PUBLISH_BUNDLE_TAG",
			ActionInputName: "bundle_publish_tag",
			ActionType:      "String",
			DefaultValue:    "my-app/artifact-bundle:latest",
			stringDecoder:   stringToStringDecoder,
			cobraFunc: func(f *MetaField, c *cobra.Command) {
				c.Flags().StringVar(f.FlagValueP.(*string), f.FlagName, f.DefaultValue, f.FlagDesc)
			},
		},
		ValidationEnabled: MetaField{
			FlagValueP:      new(bool),
			FlagName:        "enabled",
			FlagDesc:        "Enable/Disable the validation pipeline",
			EnvKey:          "WFE_DEPLOY_ENABLED",
			ActionInputName: "validation_enabled",
			ActionType:      "Bool",
			DefaultValue:    "true",
			stringDecoder:   stringToBoolDecoder,
			cobraFunc: func(f *MetaField, c *cobra.Command) {
				defaultValue, _ := stringToBoolDecoder(f.DefaultValue)
				c.Flags().BoolVar(f.FlagValueP.(*bool), f.FlagName, defaultValue.(bool), f.FlagDesc)
			},
		},
		ValidationGatecheckConfigFilename: MetaField{
			FlagValueP:      new(string),
			FlagName:        "gatecheck-config-filename",
			FlagDesc:        "The filename for the gatecheck config",
			EnvKey:          "WFE_DEPLOY_GATECHECK_CONFIG_FILENAME",
			ActionInputName: "gatecheck_config_filename",
			ActionType:      "String",
			DefaultValue:    "",
			stringDecoder:   stringToStringDecoder,
			cobraFunc: func(f *MetaField, c *cobra.Command) {
				c.Flags().StringVar(f.FlagValueP.(*string), f.FlagName, f.DefaultValue, f.FlagDesc)
			},
		},
	}

	return m
}
