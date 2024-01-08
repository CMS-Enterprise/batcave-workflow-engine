package pipelines

type Config struct {
	CacheDir         string `json:"cacheDir"`
	DaggerDebugImage string `json:"debugImage"`
	DaggerExec       string `json:"daggerExec"`
	Image            Image  `json:"image"`
}

func SetDefaults(c *Config) {
	c.CacheDir = valueOrDefault(c.CacheDir, "./wfe-cache")
	c.DaggerDebugImage = valueOrDefault(c.DaggerDebugImage, "alpine:latest")
	c.DaggerExec = valueOrDefault(c.DaggerExec, "dagger")
	c.Image.BuildDir = valueOrDefault(c.Image.BuildDir, ".")
	c.Image.BuildDockerfile = valueOrDefault(c.Image.BuildDockerfile, "Dockerfile")
	c.Image.BuildPlatform = valueOrDefault(c.Image.BuildPlatform, "")
	c.Image.BuildTarget = valueOrDefault(c.Image.BuildTarget, "")
}

func valueOrDefault(value string, d string) string {
	if value == "" {
		return d
	}
	return value
}

type Image struct {
	BuildDir        string `json:"buildDir"`
	BuildDockerfile string `json:"buildDockerfile"`
	BuildPlatform   string `json:"buildPlatform"`
	BuildTarget     string `json:"buildTarget"`
}

func NewDefaultConfig() Config {
	config := new(Config)
	SetDefaults(config)
	return *config
}
