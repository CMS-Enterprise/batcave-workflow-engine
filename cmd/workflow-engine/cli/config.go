package cli

import (
	"io"
	"log/slog"
	"os"
	"strings"
	"workflow-engine/pkg/pipelines"

	"github.com/spf13/cobra"
)

func newConfigCommand() *cobra.Command {

	// config init
	initCmd := newBasicCommand("init", "write the default configuration file", runConfigInit)
	initCmd.Flags().StringP("output", "o", "yaml", "config output format (<format>=<file>) empty will write to STDOUT, formats=[json yaml yml toml]")

	// config vars
	varsCmd := newBasicCommand("vars", "list supported builtin variables that can be used in templates", runConfigVars)
	varsCmd.Flags().StringP("output", "o", "table", "config output format (<format>=<file>) empty will write to STDOUT, formats=[table json yaml yml toml]")

	// config render
	renderCmd := newBasicCommand("render", "render a configuration template from `--file` flag or STDIN", runConfigRender)
	renderCmd.Flags().StringP("file", "f", "", "a text file that contains placeholders")
	renderCmd.MarkFlagFilename("file")

	// config
	cmd := &cobra.Command{Use: "config", Short: "manage the workflow engine config file"}

	// add sub commands
	cmd.AddCommand(initCmd, varsCmd, renderCmd)

	return cmd
}

// Run Functions - responsible for parsing flags at runtime

func runConfigInit(cmd *cobra.Command, _ []string) error {
	var targetWriter io.Writer
	var format, filename string

	output, _ := cmd.Flags().GetString("output")
	if strings.Contains(output, "=") {
		parts := strings.Split(output, "=")
		format, filename = parts[0], parts[1]
	}

	switch filename {
	case "":
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

// Execution functions - contains runtime logic

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

func writeExampleTo(w io.Writer, asFormat string) error {
	return NewAbstractEncoder(w, pipelines.NewDefaultConfig()).Encode(asFormat)
}
