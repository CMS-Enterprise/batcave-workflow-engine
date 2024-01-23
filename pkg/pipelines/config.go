package pipelines

type Config struct {
	Image ImageBuildConfig `json:"image" yaml:"image" toml:"image"`
}

func SetDefaults(c *Config) {
	c.Image.BuildDir = valueOrDefault(c.Image.BuildDir, ".")
	c.Image.BuildDockerfile = valueOrDefault(c.Image.BuildDockerfile, "Dockerfile")
	c.Image.BuildPlatform = valueOrDefault(c.Image.BuildPlatform, "")
	c.Image.BuildTarget = valueOrDefault(c.Image.BuildTarget, "")
	if c.Image.BuildArgs == nil {
		c.Image.BuildArgs = make([][2]string, 0)
	}
}

func valueOrDefault(value string, d string) string {
	if value == "" {
		return d
	}
	return value
}

type ImageBuildConfig struct {
	BuildDir        string      `json:"buildDir" yaml:"buildDir" toml:"buildDir"`
	BuildDockerfile string      `json:"buildDockerfile" yaml:"buildDockerfile" toml:"buildDockerfile"`
	BuildTag        string      `json:"buildTag" yaml:"buildTag" toml:"buildTag"`
	BuildPlatform   string      `json:"buildPlatform" yaml:"buildPlatform" toml:"buildPlatform"`
	BuildTarget     string      `json:"buildTarget" yaml:"buildTarget" toml:"buildTarget"`
	BuildCacheTo    string      `json:"buildCacheTo" yaml:"buildCacheTo" toml:"buildCacheTo"`
	BuildCacheFrom  string      `json:"buildCacheFrom" yaml:"buildCacheFrom" toml:"buildCacheFrom"`
	BuildArgs       [][2]string `json:"buildArgs" yaml:"buildArgs" toml:"buildArgs"`
}

func NewDefaultConfig() *Config {
	config := new(Config)
	SetDefaults(config)
	return config
}
