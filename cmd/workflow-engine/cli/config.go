package cli

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"workflow-engine/pkg/pipelines"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newConfigCommand() *cobra.Command {
	// config init
	initCmd := newBasicCommand("init", "write the default configuration file", runConfigInit)
	initCmd.Flags().StringP("output", "o", "yaml", "config output format (<format>=<file>) empty will write to STDOUT, formats=[json yaml yml toml]")

	// config info
	infoCmd := newBasicCommand("info <CONFIG FILE>", "print the loaded configuration values", runConfigInfo)
	infoCmd.Args = cobra.ExactArgs(1)

	// config vars
	varsCmd := newBasicCommand("vars", "list supported builtin variables that can be used in templates", runConfigVars)
	varsCmd.Flags().StringP("output", "o", "toml", "config output format (<format>=<file>) empty will write to STDOUT, formats=[json yaml yml toml]")

	// config render
	renderCmd := newBasicCommand("render", "render a configuration template config using builtin variables", runConfigRender)
	renderCmd.Flags().StringP("render", "o", "yaml", "config output format (<format>=<file>) empty will write to STDOUT, formats=[json yaml yml toml]")

	// config convert
	convertCmd := newBasicCommand("convert <TEMPLATE CONFIG FILE>", "convert a configuration file", runConfigConvert)
	convertCmd.Flags().StringP("output", "o", "json", "config output format (<format>=<file>) empty will write to STDOUT, formats=[json yaml yml toml]")
	_ = convertCmd.MarkFlagFilename("file")
	convertCmd.Args = cobra.ExactArgs(1)

	// config
	cmd := &cobra.Command{Use: "config", Short: "manage the workflow engine config file"}

	// add sub commands
	cmd.AddCommand(infoCmd, initCmd, varsCmd, renderCmd, convertCmd)

	return cmd
}

// Run Functions - Parsing flags and arguments at command runtime
func runConfigInfo(cmd *cobra.Command, args []string) error {
	config := new(pipelines.Config)
	if err := LoadOrDefault(args[0], config, viper.GetViper()); err != nil {
		return err
	}
	return ListConfig(cmd.OutOrStdout(), viper.GetViper())
}

func runConfigInit(cmd *cobra.Command, _ []string) error {
	var targetWriter io.Writer

	output, _ := cmd.Flags().GetString("output")

	format, filename := ParsedOutput(output)

	switch {
	case filename == "":
		targetWriter = cmd.OutOrStdout()
	default:
		slog.Debug("open", "filename", filename)
		f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		defer f.Close()
		targetWriter = f

	}

	config := new(pipelines.Config)

	if err := viper.Unmarshal(config); err != nil {
		slog.Error("viper unmarshal defaults into a new config object")
		return errors.New("Config Init Failed.")
	}

	if err := NewAbstractEncoder(targetWriter, config).Encode(format); err != nil {
		slog.Error("cannot encode default config object to stdout", "format", format)
		return errors.New("Config Init Failed.")
	}

	return nil
}

func runConfigVars(cmd *cobra.Command, _ []string) error {
	output, _ := cmd.Flags().GetString("output")

	return writeBuiltins(cmd.OutOrStdout(), output)
}

func runConfigRender(cmd *cobra.Command, _ []string) error {
	tmplFilename, _ := cmd.Flags().GetString("file")
	switch {
	case tmplFilename == "":
		return writeRenderConfigToFrom(cmd.OutOrStdout(), cmd.InOrStdin())
	default:
		return writeRenderedConfigTo(cmd.OutOrStdout(), tmplFilename)
	}
}

func runConfigConvert(cmd *cobra.Command, args []string) error {
	configFilename := args[0]

	output, _ := cmd.Flags().GetString("output")

	format, filename := ParsedOutput(output)

	slog.Debug("config convert", "config_filename", configFilename, "output_format",
		format, "output_filename", filename, "output_flag_value", output)

	// Let viper handle unmarshalling from the various file types without env or flag values
	tempViper := viper.New()

	tempViper.SetConfigFile(configFilename)
	if err := tempViper.ReadInConfig(); err != nil {
		return err
	}

	// check for a target destination filename or default to STDOUT
	var outputWriter io.Writer
	switch {
	case filename != "":
		f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("failed to open output file: %w", err)
		}
		outputWriter = f
	default:
		outputWriter = cmd.OutOrStdout()
	}

	config := new(pipelines.Config)
	// Force load without default configuration
	if err := LoadOrDefault("", config, tempViper); err != nil {
		return err
	}

	return NewAbstractEncoder(outputWriter, config).Encode(format)
}

// Execution functions - Logic for command execution

func writeRenderedConfigTo(w io.Writer, configTemplateFilename string) error {
	slog.Debug("open render src", "src_filename", configTemplateFilename)
	f, err := os.Open(configTemplateFilename)
	if err != nil {
		return err
	}
	return pipelines.RenderTemplate(w, f)
}

func writeRenderConfigToFrom(out io.Writer, in io.Reader) error {
	return pipelines.RenderTemplate(out, in)
}

func writeBuiltins(w io.Writer, asFormat string) error {
	builtins, err := pipelines.BuiltIns()
	if err != nil {
		return err
	}

	return NewAbstractEncoder(w, builtins).Encode(asFormat)
}
