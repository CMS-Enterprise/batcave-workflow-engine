package cli

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
	"workflow-engine/pkg/pipelines"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newConfigCommand() *cobra.Command {

	// config init
	initCmd := newBasicCommand("init", "write the default configuration file", runConfigInit)
	initCmd.Flags().StringP("output", "o", "yaml", "config output format (<format>=<file>) empty will write to STDOUT, formats=[json yaml yml toml]")

	// config vars
	varsCmd := newBasicCommand("vars", "list supported builtin variables that can be used in templates", runConfigVars)
	varsCmd.Flags().StringP("output", "o", "table", "config output format (<format>=<file>) empty will write to STDOUT, formats=[table json yaml yml toml]")

	// config render
	renderCmd := newBasicCommand("render", "render a configuration template (`--file` flag or STDIN) and write to STDOUT", runConfigRender)
	renderCmd.Flags().StringP("file", "f", "", "a text file that contains placeholders")
	renderCmd.MarkFlagFilename("file")

	// config convert
	convertCmd := newBasicCommand("convert", "convert a configuration file (`--file` or STDIN) to another format", runConfigConvert)
	convertCmd.Flags().StringP("file", "f", "", "input file to use as source")
	convertCmd.Flags().StringP("input", "i", "json", "the input file format [json yaml yml toml]")
	convertCmd.Flags().StringP("output", "o", "json", "config output format (<format>=<file>) empty will write to STDOUT, formats=[json yaml yml toml]")
	convertCmd.MarkFlagFilename("file")
	convertCmd.MarkFlagsOneRequired("file", "input")

	// config
	cmd := &cobra.Command{Use: "config", Short: "manage the workflow engine config file"}

	// add sub commands
	cmd.AddCommand(initCmd, varsCmd, renderCmd, convertCmd)

	return cmd
}

// Run Functions - Parsing flags and arguments at command runtime

func runConfigInit(cmd *cobra.Command, _ []string) error {
	var targetWriter io.Writer

	output, _ := cmd.Flags().GetString("output")

	format, filename := ParsedOutput(output)

	switch {
	case filename == "":
		targetWriter = cmd.OutOrStdout()
	default:
		slog.Debug("open", "filename", filename)
		f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		targetWriter = f

	}

	if err := writeExampleTo(targetWriter, format); err != nil {
		slog.Error("failed to encode configuration file", "error", err)
		return err
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

func runConfigConvert(cmd *cobra.Command, _ []string) error {
	configFilename, _ := cmd.Flags().GetString("file")

	output, _ := cmd.Flags().GetString("output")
	input, _ := cmd.Flags().GetString("input")

	format, filename := ParsedOutput(output)

	slog.Debug("config convert", "config_filename", configFilename, "output_format",
		format, "output_filename", filename, "output_flag_value", output)

	// Let viper handle unmarshalling from the various file types without env or flag values
	tempViper := viper.New()

	var readErr error

	// Use config filename from flag or default to reading from STDIN
	switch {
	case configFilename != "":
		tempViper.SetConfigFile(configFilename)
		readErr = tempViper.ReadInConfig()
	default:
		if !slices.Contains([]string{"json", "yaml", "yml", "toml"}, input) {
			return fmt.Errorf("unsupported input format '%s'", input)
		}
		slog.Debug("config convert read config from stdin")
		tempViper.SetConfigType(input)
		readErr = tempViper.ReadConfig(cmd.InOrStdin())
	}

	if readErr != nil {
		return readErr
	}
	config := ConfigFromViper(tempViper)

	// check for a target destination filename or default to STDOUT
	var outputWriter io.Writer
	switch {
	case filename != "":
		f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open output file: %w", err)
		}
		outputWriter = f
	default:
		outputWriter = cmd.OutOrStdout()
	}

	return writeConfigTo(outputWriter, config, format)
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

func writeConfigTo(w io.Writer, config pipelines.Config, asFormat string) error {
	return NewAbstractEncoder(w, config).Encode(asFormat)
}

func writeBuiltins(w io.Writer, asFormat string) error {
	builtins, err := pipelines.BuiltIns()
	if err != nil {
		return err
	}

	return NewAbstractEncoder(w, builtins).Encode(asFormat)
}

func writeExampleTo(w io.Writer, asFormat string) error {
	return NewAbstractEncoder(w, pipelines.NewDefaultConfig()).Encode(asFormat)
}
