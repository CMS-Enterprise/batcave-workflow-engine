package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"workflow-engine/pkg/pipelines"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// ApplicationMetadata ...
type ApplicationMetadata struct {
	CLIVersion     string
	GitCommit      string
	BuildDate      string
	GitDescription string
	Platform       string
	GoVersion      string
	Compiler       string
}

func (m ApplicationMetadata) String() string {
	return fmt.Sprintf(`CLIVersion:     %s
GitCommit:      %s
Build Date:     %s
GitDescription: %s
Platform:       %s
GoVersion:      %s
Compiler:       %s
`,
		m.CLIVersion, m.GitCommit, m.BuildDate, m.GitDescription,
		m.Platform, m.GoVersion, m.Compiler)
}

func (m ApplicationMetadata) WriteTo(w io.Writer) (int64, error) {
	n, err := fmt.Fprintf(w, "%s\n", m)
	return int64(n), err
}

// newBasicCommand is a convienence function that includes the minimum fields for a command
//
// It simplifies the cobra interface which has a lot of useful fields for different use cases
// but in this CLI, we don't need all of those features.
func newBasicCommand(use string, short string, runE func(*cobra.Command, []string) error) *cobra.Command {
	return &cobra.Command{Use: use, Short: short, RunE: runE}
}

type AbstractEncoder struct {
	w   io.Writer
	obj any
}

func NewAbstractEncoder(w io.Writer, v any) *AbstractEncoder {
	return &AbstractEncoder{w: w, obj: v}
}

func (a *AbstractEncoder) EncodePrettyJSON() error {
	enc := json.NewEncoder(a.w)
	enc.SetIndent("", "  ")
	return enc.Encode(a.obj)
}

func (a *AbstractEncoder) EncodePrettyYAML() error {
	enc := yaml.NewEncoder(a.w)
	enc.SetIndent(4)
	return enc.Encode(a.obj)
}

func (a *AbstractEncoder) EncodeTOML() error {
	return toml.NewEncoder(a.w).Encode(a.obj)
}

func (a *AbstractEncoder) EncodeFormatedTable() error {
	var err error

	m, ok := a.obj.(map[string]string)
	if !ok {
		return errors.New("cannot encode object to table")
	}

	for key, value := range m {
		s := fmt.Sprintf("%-25s %s", key, value)
		_, err = fmt.Fprintln(a.w, s)
	}

	return err
}

func (a *AbstractEncoder) Encode(asFormat string) error {
	switch asFormat {
	case "json":
		return NewAbstractEncoder(a.w, a.obj).EncodePrettyJSON()
	case "yaml", "yml":
		return NewAbstractEncoder(a.w, a.obj).EncodePrettyYAML()
	case "toml":
		return NewAbstractEncoder(a.w, a.obj).EncodeTOML()
	default:
		return fmt.Errorf("unsupported format: '%s'", asFormat)
	}
}

// Helper Functions

func ListConfig(w io.Writer, v *viper.Viper) error {
	for _, key := range viper.AllKeys() {
		_, err := fmt.Fprintf(w, "%-45s %s\n", key, fmt.Sprintf("%v", viper.Get(key)))
		if err != nil {
			return err
		}
	}
	return nil
}

// LoadConfig will do a viper.ReadIn and unmarshal the values into a new config object
//
// This function assumes the caller already did v.SetFilename(configFilename)
// If the config file does not exist, this function will return an error.
// If the possibility of a non-existent configuration is expected, you don't need to
// use this function. you can call unmarshal directly to load default values
func LoadConfig(config *pipelines.Config, v *viper.Viper) error {
	if err := v.ReadInConfig(); err != nil {
		return err
	}
	if err := v.Unmarshal(config); err != nil {
		return err
	}
	return nil
}

// LoadOrDefault will use the default config if filename is blank
//
// Caller should pass in a new config object
func LoadOrDefault(filename string, config *pipelines.Config, v *viper.Viper) error {
	viper.SetConfigFile(filename)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
		slog.Warn("no configuration file provided with --config flag, using default configuration")
		_ = viper.Unmarshal(config)
	}
	return nil
}

// ParsedOutput splits the format and filename
//
// expects the `--output` argument format (<format>=<file>)
func ParsedOutput(output string) (format, filename string) {
	switch {
	case strings.Contains(output, "="):
		parts := strings.Split(output, "=")
		return parts[0], parts[1]
	default:
		return output, ""
	}
}
