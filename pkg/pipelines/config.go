package pipelines

// Config is the main configuration file for all of workflow engine
//
// The file is intended to be represented in json, yaml, or toml which is done via struct field tags
// Note: This is only intended to be the data based representation of values.
// For example, the Image field has values with tags that would represent the file structure of the
// config file. When it's passed to the image build pipeline, additional logic is used to build
// the image build commands.
type Config struct {
	Image     ImageBuildConfig `json:"image" yaml:"image" toml:"image"`
	Artifacts ArtifactConfig   `json:"Artifacts" yaml:"Artifacts" toml:"Artifacts"`
}

type ArtifactConfig struct {
	Directory     string `json:"Directory" yaml:"Directory" toml:"Directory"`
	SBOMFilename  string `json:"sbomFilename" yaml:"sbomFilename" toml:"sbomFilename"`
	GrypeFilename string `json:"grypeFilename" yaml:"grypeFilename" toml:"grypeFilename"`
}

// ImageBuildConfig is a struct representation of the Image field in the Config file
type ImageBuildConfig struct {
	BuildDir          string      `json:"buildDir" yaml:"buildDir" toml:"buildDir"`
	BuildDockerfile   string      `json:"buildDockerfile" yaml:"buildDockerfile" toml:"buildDockerfile"`
	BuildTag          string      `json:"buildTag" yaml:"buildTag" toml:"buildTag"`
	BuildPlatform     string      `json:"buildPlatform" yaml:"buildPlatform" toml:"buildPlatform"`
	BuildTarget       string      `json:"buildTarget" yaml:"buildTarget" toml:"buildTarget"`
	BuildCacheTo      string      `json:"buildCacheTo" yaml:"buildCacheTo" toml:"buildCacheTo"`
	BuildCacheFrom    string      `json:"buildCacheFrom" yaml:"buildCacheFrom" toml:"buildCacheFrom"`
	BuildSquashLayers bool        `json:"buildSquashLayers" yaml:"buildSquashLayers" toml:"buildSquashLayers"`
	BuildArgs         [][2]string `json:"buildArgs" yaml:"buildArgs" toml:"buildArgs"`
}

// NewDefaultConfig creates a new "safe" config object.
// This can be used to prevent nil reference panics
func NewDefaultConfig() *Config {
	// Only fields that are slices need to be inited, the default string value is ""
	return &Config{Image: ImageBuildConfig{BuildArgs: make([][2]string, 0)}}
}
