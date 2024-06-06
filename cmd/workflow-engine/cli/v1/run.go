package cli

import (
	"encoding/json"
	"os"
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
	settings.SetupCobra(&metaConfig.ImageTag, runAntivirusScanTask)
	settings.SetupCobra(&metaConfig.ArtifactDir, runAntivirusScanTask)
	settings.SetupCobra(&metaConfig.ImageScanClamavFilename, runAntivirusScanTask)
	settings.SetupCobra(&metaConfig.ImageScanFreshclamDisabled, runAntivirusScanTask)
	runAntivirusScanTask.Flags().Bool("pull", false, "pull the image before saving if it's not locally loaded")

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
		cliInterface := "docker"
		imagePull, _ := cmd.Flags().GetBool("pull")

		if podmanEnabled, _ := cmd.Flags().GetBool("podman"); podmanEnabled {
			cliInterface = "podman"
		}
		f, err := os.CreateTemp(os.TempDir(), "*.container-image.tar")
		if err != nil {
			return err
		}
		imageTarFilename := f.Name()
		_ = f.Close()

		imageSaveOpts := tasks.WithImageSaveOptions(config.ImageTag, imageTarFilename, imagePull)
		imageSaveTask := tasks.NewImageSaveTask(cliInterface, imageSaveOpts)

		err = imageSaveTask.Run(cmd.Context(), cmd.ErrOrStderr())
		if err != nil {
			return err
		}

		clamOpts := tasks.WithClamOptions(
			config.ImageScan.ClamavFilename,
			imageTarFilename,
			config.ArtifactDir,
			config.ImageScan.FreshclamDisabled,
		)
		task := tasks.NewAntivirusScanTask(tasks.ClamTaskType, clamOpts)

		return task.Run(cmd.Context(), cmd.ErrOrStderr())
	},
}
