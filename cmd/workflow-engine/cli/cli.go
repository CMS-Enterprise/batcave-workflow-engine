package cli

import (
	"encoding/json"
	"fmt"
	"log/slog"
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
			return debugPipeline(cmd, a.flagDryRun)
		},
	}
	imagebuildCmd := &cobra.Command{
		Use:   "image-build",
		Short: "Build a container image",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.loadConfig(cmd, args); err != nil {
				return err
			}
			return imageBuildPipeline(cmd, a.flagDryRun, imageBuildCmd(*a.flagCLICmd), a.cfg.Image)
		},
	}

	runCmd.AddCommand(runDebugCmd, imagebuildCmd)

	// Config Sub Command setup
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage the configuration file",
	}

	configInitCmd := &cobra.Command{
		Use:   "init",
		Short: "Output the default configuration file to stdout",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.configInit(cmd)
		},
	}

	configCmd.AddCommand(configInitCmd)

	// Root Command Configuration
	a.cmd = &cobra.Command{
		Use:   "workflow-engine",
		Short: "A portable, opinionate security pipeline",
	}

	// Init flag pointers
	a.flagDryRun = new(bool)
	a.flagConfigFilename = new(string)
	a.flagCLICmd = new(string)

	// Sub Command Flags
	imagebuildCmd.Flags().StringVarP(a.flagCLICmd, "cli-interface", "i", "docker", "[docker|podman] CLI interface to use for image building")
	imagebuildCmd.Flags().String("build-dir", ".", "image build context directory")
	imagebuildCmd.Flags().String("dockerfile", "Dockerfile", "image build custom Dockerfile")
	imagebuildCmd.Flags().String("tag", "", "image build custom tag")
	imagebuildCmd.Flags().String("platform", "", "image build custom platform option")
	imagebuildCmd.Flags().String("target", "", "image build custom target option")
	imagebuildCmd.Flags().String("cache-to", "", "image build custom cache-to option")
	imagebuildCmd.Flags().String("cache-from", "", "image build custom cache-from option")
	imagebuildCmd.Flags().Bool("squash-layers", true, "image build squash all layers into one option")

	// Persistent Flags, available on all commands
	a.cmd.PersistentFlags().BoolVarP(a.flagDryRun, "dry-run", "n", false, "log commands to debug but don't execute")
	a.cmd.PersistentFlags().StringVar(a.flagConfigFilename, "config", "", "workflow engine config file in json, yaml, or toml")

	// Flag marks
	a.cmd.MarkFlagFilename("config", "json")

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

	// Viper bind config keys to flag values and environment variables
	viper.BindPFlag("buildDir", imagebuildCmd.Flags().Lookup("build-dir"))
	viper.MustBindEnv("buildDir", "WFE_BUILD_DIR")

	viper.BindPFlag("buildDockerfile", imagebuildCmd.Flags().Lookup("dockerfile"))
	viper.MustBindEnv("buildDockerfile", "WFE_BUILD_DOCKERFILE")

	viper.BindPFlag("buildTag", imagebuildCmd.Flags().Lookup("tag"))
	viper.MustBindEnv("buildTag", "WFE_BUILD_TAG")

	viper.BindPFlag("buildPlatform", imagebuildCmd.Flags().Lookup("platform"))
	viper.MustBindEnv("buildPlatform", "WFE_BUILD_PLATFORM")

	viper.BindPFlag("buildTarget", imagebuildCmd.Flags().Lookup("target"))
	viper.MustBindEnv("buildTarget", "WFE_BUILD_TARGET")

	viper.BindPFlag("buildCacheTo", imagebuildCmd.Flags().Lookup("cache-to"))
	viper.MustBindEnv("buildCacheTo", "WFE_BUILD_CACHE_TO")

	viper.BindPFlag("buildCacheFrom", imagebuildCmd.Flags().Lookup("cache-from"))
	viper.MustBindEnv("buildCacheFrom", "WFE_BUILD_CACHE_FROM")

	viper.BindPFlag("buildSquashLayers", imagebuildCmd.Flags().Lookup("squash-layers"))
	viper.MustBindEnv("buildSquashLayers", "WFE_BUILD_SQUASH_LAYERS")

	a.cmd.AddCommand(runCmd, configCmd)
}

// Execute starts the CLI handler, should be called in the main function
func (a *App) Execute() error {
	return a.cmd.Execute()
}

func (a *App) loadConfig(cmd *cobra.Command, args []string) error {
	l := slog.Default().With("step", "load_config")

	l.Debug("check config file flag value")
	configFile, _ := a.cmd.PersistentFlags().GetString("config")
	if configFile != "" {
		viper.SetConfigFile(configFile)
		return nil
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

	// viper will unmarshal the values into the cfg object
	l.Debug("decode configuration file")
	a.cfg = &pipelines.Config{
		Image: pipelines.ImageBuildConfig{
			BuildDir:        viper.GetString("buildDir"),
			BuildDockerfile: viper.GetString("buildDockerfile"),
			BuildTag:        viper.GetString("buildTag"),
			BuildPlatform:   viper.GetString("buildPlatform"),
			BuildTarget:     viper.GetString("buildTarget"),
			BuildCacheTo:    viper.GetString("buildCacheTo"),
			BuildCacheFrom:  viper.GetString("buildCacheFrom"),
			BuildArgs:       make([][2]string, 0),
		},
	}

	l.Debug("config file loaded", "content", fmt.Sprintf("%+v", a.cfg))
	return nil
}

func (a *App) configInit(cmd *cobra.Command) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(pipelines.NewDefaultConfig())
}

func debugPipeline(cmd *cobra.Command, dryRun *bool) error {
	pipeline := pipelines.NewDebug(cmd.OutOrStdout(), cmd.ErrOrStderr())
	pipeline.DryRunEnabled = *dryRun
	return pipeline.Run()
}

type imageBuildCmd string

const (
	cliDocker imageBuildCmd = "docker"
	cliPodman               = "podman"
)

func imageBuildPipeline(cmd *cobra.Command, dryRun *bool, cliCmd imageBuildCmd, config pipelines.ImageBuildConfig) error {
	pipeline := pipelines.NewImageBuild(cmd.OutOrStdout(), cmd.ErrOrStderr())
	pipeline.DryRunEnabled = *dryRun
	if cliCmd == cliPodman {
		pipeline = pipeline.WithPodman()
	}
	return pipeline.WithBuildConfig(config).Run()
}

func imageScanPipeline(cmd *cobra.Command, dryRun *bool, config pipelines.Config) error {
	pipeline := pipelines.NewImageScan(cmd.OutOrStdout(), cmd.ErrOrStderr())
	pipeline.DryRunEnabled = *dryRun

	return pipeline.WithArtifactConfig(config.Artifacts).WithImageName(config.Image.BuildTag).Run()
}
