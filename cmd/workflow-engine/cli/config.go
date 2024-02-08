package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"workflow-engine/pkg/pipelines"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage the configuration file",
	}

	cmd.AddCommand(newConfigInitCommand(), newConfigBuiltinCommand())
	return cmd
}

func newConfigInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Output the default configuration file to stdout",
		RunE:  configInitRun,
	}
	cmd.Flags().StringP("output", "o", "yaml", "config output format (<format>=<file>) empty will write to STDOUT, formats=[json yaml yml toml]")
	return cmd
}

func configInitRun(cmd *cobra.Command, _ []string) error {
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

func newConfigBuiltinCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "builtins",
		Short: "List supported built-in template variables",
		RunE:  configBuiltinRun,
	}
	cmd.Aliases = []string{"vars", "env"}
	cmd.Flags().StringP("output", "o", "table", "config output format (<format>=<file>) empty will write to STDOUT, formats=[table env json yaml yml toml]")
	return cmd
}

func configBuiltinRun(cmd *cobra.Command, _ []string) error {
	output, _ := cmd.Flags().GetString("output")

	return writeBuiltins(cmd.OutOrStdout(), output)
}

func writeRenderedConfigTo(w io.Writer, configTemplateFilename string) error {
	f, err := os.Open(configTemplateFilename)
	if err != nil {
		return err
	}
	return pipelines.RenderTemplate(w, f)
}

func writeBuiltins(w io.Writer, asFormat string) error {
	builtins, err := pipelines.BuiltIns()
	if err != nil {
		return err
	}

	printableBuiltins := printableMap(builtins)

	switch asFormat {
	case "table":
		printableBuiltins.WriteTableTo(w)
	case "json":
		printableBuiltins.encodeJSON(w)
	case "yaml", "yml":
		printableBuiltins.encodeYAML(w)
	case "toml":
		printableBuiltins.encodeTOML(w)
	default:
		return fmt.Errorf("unsupported format: '%s'", asFormat)

	}
	return nil
}

type printableMap map[string]string

func (p printableMap) encodeJSON(w io.Writer) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(p)
}

func (p printableMap) encodeYAML(w io.Writer) {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(4)
	_ = enc.Encode(p)
}

func (p printableMap) encodeTOML(w io.Writer) {
	toml.NewEncoder(w).Encode(p)
}

func (p printableMap) WriteTableTo(w io.Writer) {
	for key, value := range p {
		s := fmt.Sprintf("%-25s %s", key, value)
		fmt.Fprintln(w, s)
	}
}

func writeExampleTo(w io.Writer, asFormat string) error {
	type encoder interface {
		Encode(any) error
	}
	var enc encoder
	switch asFormat {
	case "json":
		jsonEnc := json.NewEncoder(w)
		jsonEnc.SetIndent("", "  ")
		enc = jsonEnc
	case "yaml", "yml":
		yamlEnc := yaml.NewEncoder(w)
		yamlEnc.SetIndent(4)
		enc = yamlEnc
	case "toml":
		enc = toml.NewEncoder(w)
	default:
		return fmt.Errorf("unsupported format: '%s'", asFormat)
	}

	return enc.Encode(pipelines.NewDefaultConfig())
}
