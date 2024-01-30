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
	Image     ImageConfig    `json:"image" yaml:"image" toml:"image"`
	Artifacts ArtifactConfig `json:"artifacts" yaml:"artifacts" toml:"artifacts"`
}

type ArtifactConfig struct {
	Directory     string `json:"directory" yaml:"directory" toml:"directory"`
	SBOMFilename  string `json:"sbomFilename" yaml:"sbomFilename" toml:"sbomFilename"`
	GrypeFilename string `json:"grypeFilename" yaml:"grypeFilename" toml:"grypeFilename"`
}

// ImageConfig is a struct representation of the Image field in the Config file
type ImageConfig struct {
	BuildDir          string      `json:"buildDir" yaml:"buildDir" toml:"buildDir"`
	BuildDockerfile   string      `json:"buildDockerfile" yaml:"buildDockerfile" toml:"buildDockerfile"`
	BuildTag          string      `json:"buildTag" yaml:"buildTag" toml:"buildTag"`
	BuildPlatform     string      `json:"buildPlatform" yaml:"buildPlatform" toml:"buildPlatform"`
	BuildTarget       string      `json:"buildTarget" yaml:"buildTarget" toml:"buildTarget"`
	BuildCacheTo      string      `json:"buildCacheTo" yaml:"buildCacheTo" toml:"buildCacheTo"`
	BuildCacheFrom    string      `json:"buildCacheFrom" yaml:"buildCacheFrom" toml:"buildCacheFrom"`
	BuildSquashLayers bool        `json:"buildSquashLayers" yaml:"buildSquashLayers" toml:"buildSquashLayers"`
	BuildArgs         [][2]string `json:"buildArgs" yaml:"buildArgs" toml:"buildArgs"`
	ScanTarget        string      `json:"scanTarget" yaml:"scanTarget" toml:"scanTarget"`
}

// NewDefaultConfig creates a new "safe" config object.
// This can be used to prevent nil reference panics
func NewDefaultConfig() *Config {
	// Only fields that are slices need to be inited, the default string value is ""
	return &Config{Image: ImageConfig{BuildArgs: make([][2]string, 0)}}
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
