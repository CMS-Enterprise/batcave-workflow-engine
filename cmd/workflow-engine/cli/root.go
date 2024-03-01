package cli

import (
	"log/slog"

	"github.com/spf13/cobra"
)

var AppMetadata ApplicationMetadata

func NewWorkflowEngineCommand(logLeveler *slog.LevelVar) *cobra.Command {
	versionCmd := newBasicCommand("version", "print version information", runVersion)
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
		},
	}

	// Create log leveling flags
	cmd.PersistentFlags().BoolP("verbose", "v", false, "verbose logging output")
	cmd.PersistentFlags().BoolP("silent", "q", false, "only log errors")
	cmd.MarkFlagsMutuallyExclusive("verbose", "silent")

	// Turn off usage after an error occurs which polutes the terminal
	cmd.SilenceUsage = true

	// Add Sub-commands
	cmd.AddCommand(newConfigCommand(), newRunCommand(), versionCmd)

	return cmd
}

// workflow-engine version
func runVersion(cmd *cobra.Command, args []string) error {
	versionFlag, _ := cmd.Flags().GetBool("version")
	switch {
	case versionFlag:
		cmd.Println(AppMetadata.CLIVersion)
		return nil
	default:
		_, err := AppMetadata.WriteTo(cmd.OutOrStdout())
		return err
	}
}
