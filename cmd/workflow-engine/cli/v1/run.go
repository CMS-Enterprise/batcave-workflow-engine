package cli

import (
	"workflow-engine/pkg/settings"

	"github.com/spf13/cobra"
)

func newRunCommand() *cobra.Command {
	// Common Flags
	settings.SetupCobra(&metaConfig.ImageTag, runTaskCmd)
	settings.SetupCobra(&metaConfig.ArtifactDir, runTaskCmd)

	// Image scan flags
	settings.SetupCobra(&metaConfig.ImageScanSyftFilename, runTaskCmd)
	settings.SetupCobra(&metaConfig.ImageScanGrypeFilename, runTaskCmd)

	// Antivirus scan flags
	settings.SetupCobra(&metaConfig.ImageScanClamavFilename, runAntivirusScanTask)

	return runCmd
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run the full security and delivery workflow",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var runTaskCmd = &cobra.Command{
	Use:   "run-task",
	Short: "run an individual task",
}
