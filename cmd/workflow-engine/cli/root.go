package cli

import (
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
