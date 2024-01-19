package cli

import (
	"encoding/json"
	"log/slog"
	"os"
	"workflow-engine/pkg/pipelines"

	"github.com/spf13/cobra"
)

// App is the full CLI application
//
// Each command can be written as a method on this struct and attached to the
// root command during the Init function
type App struct {
	cmd                *cobra.Command
	cfg                pipelines.Config
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
			return debugPipeline(cmd, args, a.flagDryRun)
		},
	}
	imagebuildCmd := &cobra.Command{
		Use:   "image-build",
		Short: "Build a container image",
		RunE: func(cmd *cobra.Command, args []string) error {
			return imageBuildPipeline(cmd, args, a.flagDryRun, imageBuildCmd(*a.flagCLICmd))
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
			return a.configInit(cmd, args)
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

	// Persistent Flags
	a.cmd.PersistentFlags().BoolVarP(a.flagDryRun, "dry-run", "n", false, "log commands to debug but don't execute")
	a.cmd.PersistentFlags().StringVarP(a.flagConfigFilename, "config", "c", "", "Configuration file")

	// Flag marks
	a.cmd.MarkFlagFilename("config", "json")

	// Other settings
	a.cmd.SilenceUsage = true

	a.cmd.AddCommand(runCmd, configCmd)
}

// Execute starts the CLI handler, should be called in the main function
func (a *App) Execute() error {
	return a.cmd.Execute()
}

func (a *App) loadConfig(cmd *cobra.Command, args []string) error {
	l := slog.Default()

	l.Debug("check config file flag value")
	configFile, _ := a.cmd.PersistentFlags().GetString("config")
	if configFile == "" {
		a.cfg = pipelines.NewDefaultConfig()
		return nil
	}
	l = l.With("flag", "--config", "value", configFile)

	l.Debug("open configuration file")
	f, err := os.Open(configFile)
	if err != nil {
		l.Error("cannot open configuration file", "error", err)
		return err
	}

	l.Debug("decode configuration file")
	var cfg pipelines.Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		l.Error("cannot decode configuration file", "error", err)
		return err
	}

	a.cfg = cfg
	l.Debug("config file loaded", "content", a.cfg)
	return nil
}

func (a *App) configInit(cmd *cobra.Command, args []string) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(pipelines.NewDefaultConfig())
}

func debugPipeline(cmd *cobra.Command, args []string, dryRun *bool) error {
	pipeline := pipelines.NewDebug(cmd.OutOrStdout(), cmd.ErrOrStderr())
	pipeline.DryRunEnabled = *dryRun
	return pipeline.Run()
}

type imageBuildCmd string

const (
	cliDocker imageBuildCmd = "docker"
	cliPodman               = "podman"
)

func imageBuildPipeline(cmd *cobra.Command, args []string, dryRun *bool, cliCmd imageBuildCmd) error {
	pipeline := pipelines.NewImageBuild(cmd.OutOrStdout(), cmd.ErrOrStderr())
	pipeline.DryRunEnabled = *dryRun
	if cliCmd == cliPodman {
		pipeline = pipeline.WithPodman()
	}
	return pipeline.Run()
}
