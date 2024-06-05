package cli

import (
	"encoding/json"
	"workflow-engine/pkg/settings"
	"workflow-engine/pkg/tasks"

	"github.com/spf13/cobra"
)

func newRunCommand() *cobra.Command {
	// Image scan flags
	settings.SetupCobra(&metaConfig.ImageTag, runImageScanTask)
	settings.SetupCobra(&metaConfig.ImageScanSyftFilename, runImageScanTask)
	settings.SetupCobra(&metaConfig.ImageScanGrypeFilename, runImageScanTask)
	settings.SetupCobra(&metaConfig.ArtifactDir, runImageScanTask)

	// Antivirus Scan flags
	settings.SetupCobra(&metaConfig.ArtifactDir, runAntivirusScanTask)
	settings.SetupCobra(&metaConfig.ImageScanClamavFilename, runAntivirusScanTask)

	runCmd.AddCommand(runImageScanTask, runAntivirusScanTask)
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
		imageOptions := tasks.WithImageOptions(
			config.ImageTag,
			config.ImageScan.SyftFilename,
			config.ImageScan.GrypeFilename,
			config.ArtifactDir,
		)

		task := tasks.NewImageScanTask(tasks.GrypeTaskType, imageOptions, tasks.WithStdout(cmd.OutOrStdout()))

		return task.Run(cmd.Context(), cmd.ErrOrStderr())
	},
}

var runAntivirusScanTask = &cobra.Command{
	Use:   "antivirus-scan",
	Short: "run an antivirus scan on an image or image archive",
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
		clamOpts := tasks.WithClamOptions(
			config.ImageScan.ClamavFilename,
			"image-tar",
			config.ArtifactDir,
		)
		task := tasks.NewAntivirusScanTask(tasks.ClamTaskType, clamOpts)

		return task.Run(cmd.Context(), cmd.ErrOrStderr())
	},
}
