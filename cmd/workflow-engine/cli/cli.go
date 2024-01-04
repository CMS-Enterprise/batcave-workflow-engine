package cli

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"
	"workflow-engine/pkg/pipelines"

	"dagger.io/dagger"
	"github.com/spf13/cobra"
)

const (
	pipelineDebug      = "debug"
	pipelineImageBuild = "image-build"
)

// App is the full CLI application
//
// Each command can be written as a method on this struct and attached to the
// root command during the Init function
type App struct {
	cmd    *cobra.Command
	cfg    pipelines.Config
	client *dagger.Client
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
	daggerVersionCmd := &cobra.Command{
		Use:   "dagger-version",
		Short: "Output the version of the dagger CLI",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.daggerVersion(cmd, args)
		},
	}
	debugPipeline := &cobra.Command{
		Use:   "debug-pipeline",
		Short: "Run the debug pipeline",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.pipeline(pipelineDebug, cmd, args)
		},
	}
	imagebuildPipeline := &cobra.Command{
		Use:   "image-build-pipeline",
		Short: "Build a container image",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.pipeline(pipelineImageBuild, cmd, args)
		},
	}

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

	// Flags
	a.cmd.PersistentFlags().StringP("config", "c", "", "Configuration file")
	a.cmd.MarkFlagFilename("config", "json")

	// Other settings
	a.cmd.SilenceUsage = true

	a.cmd.AddCommand(daggerVersionCmd, debugPipeline, imagebuildPipeline, configCmd)
}

// Execute starts the CLI handler, should be called in the main function
func (a *App) Execute() error {
	return a.cmd.Execute()
}

func (a *App) loadConfig(cmd *cobra.Command, args []string) error {
	l := slog.Default().With("flag", "--config")

	l.Debug("check config file flag")
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

func (a *App) daggerVersion(cmd *cobra.Command, args []string) error {
	daggerExecPath, err := exec.LookPath(a.cfg.DaggerExec)
	if err != nil {
		return err
	}

	daggerCmd := exec.CommandContext(context.Background(), daggerExecPath, "version")
	daggerCmd.Stdout = cmd.OutOrStdout()
	daggerCmd.Stderr = cmd.ErrOrStderr()
	return daggerCmd.Run()
}

// pipeline is the wrapper around any pipeline command
//
// This function connects to the dagger client before running the target pipeline.
// This lazy loads the client and prevents the CLI from connecting before every command.
func (a *App) pipeline(target string, cmd *cobra.Command, args []string) error {
	var err error
	if err := a.loadConfig(cmd, args); err != nil {
		return err
	}
	a.client, err = dagger.Connect(context.Background())
	if err != nil {
		slog.Error("failed to connect to dagger client", "error", err)
		return err
	}

	switch target {
	case pipelineDebug:
		return a.debugPipeline(cmd, args)
	case pipelineImageBuild:
		return a.imageBuildPipeline(cmd, args)
	}
	return nil
}

func (a *App) debugPipeline(cmd *cobra.Command, args []string) error {
	pipeline := pipelines.NewDebugPipeline(a.client, pipelines.NewDefaultConfig())
	pipeline.Stdout = cmd.OutOrStdout()
	return pipeline.Run()
}

func (a *App) imageBuildPipeline(cmd *cobra.Command, args []string) error {
	pipeline := pipelines.NewImageBuildPipeline(a.client, a.cfg)
	return pipeline.Run()
}

func (a *App) configInit(cmd *cobra.Command, args []string) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(pipelines.NewDefaultConfig())
}
