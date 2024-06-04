package cli

import (
	"encoding/json"
	"workflow-engine/pkg/settings"
	"workflow-engine/pkg/tasks"

	"github.com/spf13/cobra"
)

func newRunCommand() *cobra.Command {
	settings.SetupCobra(&metaConfig.ImageTag, runImageScanTask)
	settings.SetupCobra(&metaConfig.ImageScanSyftFilename, runImageScanTask)
	settings.SetupCobra(&metaConfig.ImageScanGrypeFilename, runImageScanTask)
	settings.SetupCobra(&metaConfig.ArtifactDir, runImageScanTask)
	runCmd.AddCommand(runImageScanTask)
	return runCmd
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run a task",
}

var runImageScanTask = &cobra.Command{
	Use:   "image-scan",
	Short: "run security scans on an image",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := settings.Unmarshal(config, metaConfig)
		if err != nil {
			return err
		}

		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")

		return enc.Encode(config)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := tasks.WithOptions(
			config.ImageTag,
			config.ImageScan.SyftFilename,
			config.ImageScan.GrypeFilename,
			config.ArtifactDir,
		)

		task := tasks.NewImageScanTask(tasks.GrypeTaskType, opts)

		err := task.Start(cmd.Context())
		if err != nil {
			return err
		}

		return task.Stream(cmd.ErrOrStderr())
	},
}
