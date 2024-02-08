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

	cmd.AddCommand(newConfigInitCommand())
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

	if err := writeConfigExample(format, targetWriter); err != nil {
		slog.Error("failed to encode configuration file", "error", err)
		return err
	}

	return nil
}

func writeConfigExample(format string, w io.Writer) error {
	type encoder interface {
		Encode(any) error
	}
	var enc encoder
	switch format {
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
		return fmt.Errorf("unsupported format: '%s'", format)
	}

	return enc.Encode(pipelines.NewDefaultConfig())
}
