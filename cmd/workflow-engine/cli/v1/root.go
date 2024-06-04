package cli

import (
	"log/slog"
	"workflow-engine/pkg/settings"

	"github.com/spf13/cobra"
)

var (
	AppLogLever *slog.LevelVar
	metaConfig  *settings.MetaConfig = settings.NewMetaConfig()
	config      *settings.Config     = settings.NewConfig()
)

func NewWorkflowEngineCommand() *cobra.Command {
	workflowEngineCmd.PersistentFlags().BoolP("verbose", "v", false, "set logging level to debug")
	workflowEngineCmd.PersistentFlags().BoolP("silent", "s", false, "set logging level to error")

	workflowEngineCmd.SilenceUsage = true

	workflowEngineCmd.AddCommand(newRunCommand())
	return workflowEngineCmd
}

var workflowEngineCmd = &cobra.Command{
	Use:   "workflow-engine",
	Short: "A portable, opinionated security pipeline",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		verboseFlag, _ := cmd.Flags().GetBool("verbose")
		silentFlag, _ := cmd.Flags().GetBool("silent")

		switch {
		case verboseFlag:
			AppLogLever.Set(slog.LevelDebug)
		case silentFlag:
			AppLogLever.Set(slog.LevelError)
		}
	},
}
