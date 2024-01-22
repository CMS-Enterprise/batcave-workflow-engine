package pipelines

type Config struct {
	CacheDir string `json:"cacheDir"`
	Image    Image  `json:"image"`
}

func SetDefaults(c *Config) {
	c.CacheDir = valueOrDefault(c.CacheDir, "./wfe-cache")
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

type Image struct {
	BuildDir        string      `json:"buildDir"`
	BuildDockerfile string      `json:"buildDockerfile"`
	BuildPlatform   string      `json:"buildPlatform"`
	BuildTarget     string      `json:"buildTarget"`
	BuildCacheTo    string      `json:"buildCacheTo"`
	BuildCacheFrom  string      `json:"buildCacheFrom"`
	BuildArgs       [][2]string `json:"buildArgs"`
}

func NewDefaultConfig() Config {
	config := new(Config)
	SetDefaults(config)
	return *config
}
