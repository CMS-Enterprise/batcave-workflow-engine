package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"workflow-engine/pkg/pipelines"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// App is the full CLI application
//
// Each command can be written as a method on this struct and attached to the
// root command during the Init function
type App struct {
	cmd                *cobra.Command
	cfg                *pipelines.Config
	flagDryRun         *bool
	flagCLICmd         *string
	flagConfigFilename *string
}

// NewApp bootstrap the CLI Application
func NewApp() *App {
	app := new(App)
	app.Init()
	return app
}

// Init builds the internal command and link all of the functions
//
// Note: This function is automatically called if NewApp is used
func (a *App) Init() {
	// Pipeline Commands
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run a pipeline",
	}

	runDebugCmd := &cobra.Command{
		Use:   "debug",
		Short: "Run the debug pipeline",
		RunE: func(cmd *cobra.Command, args []string) error {
			return debugPipeline(cmdIO(cmd), a.flagDryRun)
		},
	}

	imagebuildCmd := &cobra.Command{
		Use:   "image-build",
		Short: "Build a container image",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.loadConfig(); err != nil {
				return err
			}
			return imageBuildPipeline(cmdIO(cmd), *a.flagDryRun, imageBuildCmd(*a.flagCLICmd), a.cfg.Image)
		},
	}

	imageScanCmd := &cobra.Command{
		Use:   "image-scan",
		Short: "Scan a container image",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.loadConfig(); err != nil {
				return err
			}
			return imageScanPipeline(cmdIO(cmd), *a.flagDryRun, a.cfg.Artifacts, a.cfg.Image.ScanTarget)
		},
	}

	runCmd.AddCommand(runDebugCmd, imagebuildCmd, imageScanCmd)

	// Config Sub Command setup
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage the configuration file",
	}

	configInitCmd := &cobra.Command{
		Use:   "init",
		Short: "Output the default configuration file to stdout",
		RunE: func(cmd *cobra.Command, args []string) error {
			return writeConfigExample(cmd.OutOrStdout())
		},
	}

	configRenderCmd := &cobra.Command{
		Use:   "render",
		Short: "Render a configuration template and output to STDOUT",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return writeRenderedConfig(cmd.OutOrStdout(), args[0])
		},
	}

	configBuiltinsCmd := &cobra.Command{
		Use:   "builtins",
		Short: "List supported built-in template variables",
		RunE: func(cmd *cobra.Command, args []string) error {
			return configBuiltins(cmd.OutOrStdout())
		},
	}

	configCmd.AddCommand(configInitCmd, configRenderCmd, configBuiltinsCmd)

	// Root Command Configuration
	a.cmd = &cobra.Command{
		Use:   "workflow-engine",
		Short: "A portable, opinionate security pipeline",
	}

	// Init flag pointers
	a.flagDryRun = new(bool)
	a.flagConfigFilename = new(string)
	a.flagCLICmd = new(string)

	// Sub Command Flags - image build
	imagebuildCmd.Flags().StringVarP(a.flagCLICmd, "cli-interface", "i", "docker", "[docker|podman] CLI interface to use for image building")
	imagebuildCmd.Flags().String("build-dir", ".", "image build context directory")
	imagebuildCmd.Flags().String("dockerfile", "Dockerfile", "image build custom Dockerfile")
	imagebuildCmd.Flags().String("tag", "", "image build custom tag")
	imagebuildCmd.Flags().String("platform", "", "image build custom platform option")
	imagebuildCmd.Flags().String("target", "", "image build custom target option")
	imagebuildCmd.Flags().String("cache-to", "", "image build custom cache-to option")
	imagebuildCmd.Flags().String("cache-from", "", "image build custom cache-from option")
	imagebuildCmd.Flags().Bool("squash-layers", true, "image build squash all layers into one option")

	// Sub Command Flags - image scan
	imageScanCmd.Flags().String("artifact-directory", "", "the output directory for all artifacts generated in the pipeline")
	imageScanCmd.Flags().String("sbom-filename", "", "the output filename for the syft SBOM")
	imageScanCmd.Flags().String("grype-filename", "", "the output filename for the grype vulnerability report")
	imageScanCmd.Flags().String("gitleaks-filename", "", "the output filename for the gitleaks secrets report")
	imageScanCmd.Flags().String("scan-image-target", "", "scan a specific image")

	// Persistent Flags, available on all commands
	a.cmd.PersistentFlags().BoolVarP(a.flagDryRun, "dry-run", "n", false, "log commands to debug but don't execute")
	a.cmd.PersistentFlags().StringVar(a.flagConfigFilename, "config", "", "workflow engine config file in json, yaml, or toml")

	// Flag marks
	a.cmd.MarkFlagFilename("config", "json")
	a.cmd.MarkFlagDirname("artifact-directory")

	// Other settings
	a.cmd.SilenceUsage = true

	// Viper set up. Viper loads configuration values in this order of precedence
	// 1. explicit call to Set
	// 2. flag
	// 3. env
	// 4. config
	// 5. key/value store
	// 6. default

	// Viper settings
	viper.SetConfigName("workflow-engine")
	viper.AddConfigPath(".")

	// Image Build Bindings
	//   Viper bind config keys to flag values and environment variables
	viper.BindPFlag("image.builddir", imagebuildCmd.Flags().Lookup("build-dir"))
	viper.MustBindEnv("image.builddir", "WFE_BUILD_DIR")

	viper.BindPFlag("image.builddockerfile", imagebuildCmd.Flags().Lookup("dockerfile"))
	viper.MustBindEnv("image.builddockerfile", "WFE_BUILD_DOCKERFILE")

	viper.BindPFlag("image.buildtag", imagebuildCmd.Flags().Lookup("tag"))
	viper.MustBindEnv("image.buildtag", "WFE_BUILD_TAG")

	viper.BindPFlag("image.buildplatform", imagebuildCmd.Flags().Lookup("platform"))
	viper.MustBindEnv("image.buildplatform", "WFE_BUILD_PLATFORM")

	viper.BindPFlag("image.buildtarget", imagebuildCmd.Flags().Lookup("target"))
	viper.MustBindEnv("image.buildtarget", "WFE_BUILD_TARGET")

	viper.BindPFlag("image.buildcacheto", imagebuildCmd.Flags().Lookup("cache-to"))
	viper.MustBindEnv("image.buildcacheto", "WFE_BUILD_CACHE_TO")

	viper.BindPFlag("image.buildcachefrom", imagebuildCmd.Flags().Lookup("cache-from"))
	viper.MustBindEnv("image.buildcachefrom", "WFE_BUILD_CACHE_FROM")

	viper.BindPFlag("image.buildsquashlayers", imagebuildCmd.Flags().Lookup("squash-layers"))
	viper.MustBindEnv("image.buildsquashlayers", "WFE_BUILD_SQUASH_LAYERS")

	// Image Scan Bindings
	viper.BindPFlag("artifacts.directory", imageScanCmd.Flags().Lookup("artifact-directory"))
	viper.MustBindEnv("artifacts.directory", "WFE_ARTIFACT_DIRECTORY")

	viper.BindPFlag("artifacts.sbomfilename", imageScanCmd.Flags().Lookup("sbom-filename"))
	viper.MustBindEnv("artifacts.sbomfilename", "WFE_SBOM_FILENAME")

	viper.BindPFlag("artifacts.grypefilename", imageScanCmd.Flags().Lookup("grype-filename"))
	viper.MustBindEnv("artifacts.grypefilename", "WFE_GRYPE_FILENAME")

	viper.BindPFlag("artifacts.gitleaksfilename", imageScanCmd.Flags().Lookup("gitleaks-filename"))
	viper.MustBindEnv("artifacts.gitleaksfilename", "WFE_GITLEAKS_FILENAME")

	// TODO: need to consider the logic for overthe build tag here
	viper.BindPFlag("image.scantarget", imageScanCmd.Flags().Lookup("scan-image-target"))
	viper.MustBindEnv("image.scantarget", "WFE_SCAN_IMAGE_TARGET")

	a.cmd.AddCommand(runCmd, configCmd)
}

// Execute starts the CLI handler, should be called in the main function
func (a *App) Execute() error {
	return a.cmd.Execute()
}

func (a *App) loadConfig() error {
	l := slog.Default().With("step", "load_config")

	l.Debug("check config file flag value")
	configFile, _ := a.cmd.PersistentFlags().GetString("config")
	if configFile != "" {
		viper.SetConfigFile(configFile)
		l.Debug("viper config file set", "config_file", configFile)
	}

	// viper reads in config values from all sources based on precedence
	l.Debug("viper read-in config", "config_file_flag_value", configFile)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			l.Debug("viper did not find a config file; check other sources")
		} else {
			return err
		}
	}

	var sb strings.Builder
	for _, key := range viper.AllKeys() {
		slog.Debug("config", "key", key, "value", fmt.Sprintf("%v", viper.Get(key)))
	}
	slog.Debug(sb.String())

	// viper will unmarshal the values into the cfg object
	l.Debug("decode configuration file")
	a.cfg = &pipelines.Config{
		Image: pipelines.ImageConfig{
			BuildDir:        viper.GetString("image.builddir"),
			BuildDockerfile: viper.GetString("image.builddockerfile"),
			BuildTag:        viper.GetString("image.buildtag"),
			BuildPlatform:   viper.GetString("image.buildplatform"),
			BuildTarget:     viper.GetString("image.buildtarget"),
			BuildCacheTo:    viper.GetString("image.buildcacheto"),
			BuildCacheFrom:  viper.GetString("image.buildcachefrom"),
			BuildArgs:       make([][2]string, 0),
			ScanTarget:      viper.GetString("image.scantarget"),
		},
		Artifacts: pipelines.ArtifactConfig{
			Directory:        viper.GetString("artifacts.directory"),
			SBOMFilename:     viper.GetString("artifacts.sbomfilename"),
			GrypeFilename:    viper.GetString("artifacts.grypefilename"),
			GitleaksFilename: viper.GetString("artifacts.gitleaksfilename"),
		},
	}

	l.Debug("config file loaded", "content", fmt.Sprintf("%+v", a.cfg))
	return nil
}

type customIO struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

// cmdIO links the cobra command IO defaults to a customIO object
func cmdIO(cmd *cobra.Command) customIO {
	return customIO{
		stdin:  cmd.InOrStdin(),
		stdout: cmd.OutOrStdout(),
		stderr: cmd.ErrOrStderr(),
	}
}

func writeConfigExample(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(pipelines.NewDefaultConfig())
}

func writeRenderedConfig(w io.Writer, configTemplateFilename string) error {
	f, err := os.Open(configTemplateFilename)
	if err != nil {
		return err
	}
	return pipelines.RenderTemplate(w, f)
}

func configBuiltins(w io.Writer) error {
	builtins, err := pipelines.BuiltIns()
	if err != nil {
		return err
	}
	for key, value := range builtins {
		s := fmt.Sprintf("%-25s %s", key, value)
		fmt.Fprintln(w, s)
	}
	return nil
}

func debugPipeline(cio customIO, dryRun *bool) error {
	pipeline := pipelines.NewDebug(cio.stdin, cio.stderr, cio.stdout)
	pipeline.DryRunEnabled = *dryRun
	return pipeline.Run()
}

type imageBuildCmd string

const (
	cliDocker imageBuildCmd = "docker"
	cliPodman               = "podman"
)

func imageBuildPipeline(cio customIO, dryRunEnabled bool, cliCmd imageBuildCmd, config pipelines.ImageConfig) error {
	pipeline := pipelines.NewImageBuild(cio.stdout, cio.stderr)
	pipeline.DryRunEnabled = dryRunEnabled
	if cliCmd == cliPodman {
		pipeline = pipeline.WithPodman()
	}
	return pipeline.WithBuildConfig(config).Run()
}

func imageScanPipeline(cio customIO, dryRunEnabled bool, config pipelines.ArtifactConfig, imageName string) error {
	pipeline := pipelines.NewImageScan(cio.stdin, cio.stdout, cio.stderr)
	pipeline.DryRunEnabled = dryRunEnabled

	return pipeline.WithArtifactConfig(config).WithImageName(imageName).Run()
}
