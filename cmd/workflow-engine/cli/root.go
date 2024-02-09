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

func NewWorkflowEngineCommand(logLeveler *slog.LevelVar) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workflow-engine",
		Short: "A portable, opinionate security pipeline",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			verboseFlag, _ := cmd.Flags().GetBool("verbose")
			silentFlag, _ := cmd.Flags().GetBool("silent")

			switch {
			case verboseFlag:
				logLeveler.Set(slog.LevelDebug)
			case silentFlag:
				logLeveler.Set(slog.LevelError)
			}

			viperKVs := []any{}
			for _, key := range viper.AllKeys() {
				viperKVs = append(viperKVs, key, viper.Get(key))
			}
			slog.Debug("config values", viperKVs...)

		},
	}

	// Create log leveling flags
	cmd.PersistentFlags().BoolP("verbose", "v", false, "verbose logging output")
	cmd.PersistentFlags().BoolP("silent", "q", false, "only log errors")
	cmd.MarkFlagsMutuallyExclusive("verbose", "silent")

	// Add Sub-commands
	cmd.AddCommand(newConfigCommand())

	return cmd
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

// ConfigFromViper sets the configuration values to a config object from env, flags, or config file
func ConfigFromViper(v *viper.Viper) pipelines.Config {
	viperKVs := []any{}
	for _, key := range v.AllKeys() {
		viperKVs = append(viperKVs, key, v.Get(key))
	}
	slog.Debug("viper config values", viperKVs...)

	return pipelines.Config{
		Image: pipelines.ImageConfig{
			BuildDir:        v.GetString("image.builddir"),
			BuildDockerfile: v.GetString("image.builddockerfile"),
			BuildTag:        v.GetString("image.buildtag"),
			BuildPlatform:   v.GetString("image.buildplatform"),
			BuildTarget:     v.GetString("image.buildtarget"),
			BuildCacheTo:    v.GetString("image.buildcacheto"),
			BuildCacheFrom:  v.GetString("image.buildcachefrom"),
			BuildArgs:       make([][2]string, 0),
			ScanTarget:      v.GetString("image.scantarget"),
		},
		Artifacts: pipelines.ArtifactConfig{
			Directory:     v.GetString("artifacts.directory"),
			SBOMFilename:  v.GetString("artifacts.sbomfilename"),
			GrypeFilename: v.GetString("artifacts.grypefilename"),
		},
	}
}

// Config checks for the `--config` value and hands off to viper for parsing
func Config(cmd *cobra.Command) (pipelines.Config, error) {
	configFilename, _ := cmd.Flags().GetString("config")

	if configFilename != "" {
		slog.Debug("set configuration from flag value", "config_filename", configFilename)
		viper.SetConfigFile(configFilename)
	}

	slog.Debug("attempt read configuration")
	if err := viper.ReadInConfig(); err != nil {
		return pipelines.Config{}, err
	}

	config := ConfigFromViper(viper.GetViper())
	return config, nil
}

// ParsedOutput splits the format and filename
//
// expects the `--output` argument in the <format>=<filename> format
func ParsedOutput(output string) (format, filename string) {
	switch {
	case strings.Contains(output, "="):
		parts := strings.Split(output, "=")
		return parts[0], parts[1]
	default:
		return output, ""
	}
}
