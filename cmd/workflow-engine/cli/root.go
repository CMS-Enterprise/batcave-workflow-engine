package cli

import (
	"log/slog"

	"github.com/spf13/cobra"
)

var AppMetadata ApplicationMetadata
var AppLogLever *slog.LevelVar

func NewWorkflowEngineCommand() *cobra.Command {
	versionCmd := newBasicCommand("version", "print version information", runVersion)
	cmd := &cobra.Command{
		Use:              "workflow-engine",
		Short:            "A portable, opinionate security pipeline",
		PersistentPreRun: runCheckLoggingFlags,
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

func runCheckLoggingFlags(cmd *cobra.Command, _ []string) {
	verboseFlag, _ := cmd.Flags().GetBool("verbose")
	silentFlag, _ := cmd.Flags().GetBool("silent")

	switch {
	case verboseFlag:
		AppLogLever.Set(slog.LevelDebug)
	case silentFlag:
		AppLogLever.Set(slog.LevelError)
	}

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
