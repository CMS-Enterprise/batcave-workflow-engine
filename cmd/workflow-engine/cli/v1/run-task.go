package cli

import (
	"encoding/json"
	"os"
	"path"
	"workflow-engine/pkg/settings"
	"workflow-engine/pkg/tasks"

	"github.com/spf13/cobra"
)

var (
	flagAntivirusPull   = new(bool)
	flagExperimental    = new(bool)
	flagPodmanInterface = new(bool)
	flagSnyk            = new(bool)
)

func newRunTaskCommand() *cobra.Command {
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
	runAntivirusScanTask.Flags().BoolVar(flagAntivirusPull, "pull", false,
		"pull the image before saving if it's not locally loaded")

	// Image Build flags
	runImageBuildTask.Flags().BoolVar(flagPodmanInterface, "podman", false,
		"use the podman CLI interface for building images")

	settings.SetupCobra(&metaConfig.ImageTag, runImageBuildTask)
	settings.SetupCobra(&metaConfig.ImageBuildBuildDir, runImageBuildTask)
	settings.SetupCobra(&metaConfig.ImageBuildDockerfile, runImageBuildTask)
	settings.SetupCobra(&metaConfig.ImageBuildPlatform, runImageBuildTask)
	settings.SetupCobra(&metaConfig.ImageBuildTarget, runImageBuildTask)
	settings.SetupCobra(&metaConfig.ImageBuildCacheTo, runImageBuildTask)
	settings.SetupCobra(&metaConfig.ImageBuildCacheFrom, runImageBuildTask)
	settings.SetupCobra(&metaConfig.ImageBuildSquashLayers, runImageBuildTask)
	settings.SetupCobra(&metaConfig.ImageBuildArgs, runImageBuildTask)

	// Code Scan flags - gitleaks
	settings.SetupCobra(&metaConfig.CodeScanGitleaksFilename, runSecretsCodeScanTask)
	settings.SetupCobra(&metaConfig.CodeScanGitleaksSrcDir, runSecretsCodeScanTask)
	settings.SetupCobra(&metaConfig.ArtifactDir, runSecretsCodeScanTask)

	// Code Scan flags - semgrep
	settings.SetupCobra(&metaConfig.CodeScanSemgrepFilename, runSASTCodeScanTask)
	settings.SetupCobra(&metaConfig.CodeScanSemgrepRules, runSASTCodeScanTask)
	settings.SetupCobra(&metaConfig.ArtifactDir, runSASTCodeScanTask)
	runSASTCodeScanTask.Flags().BoolVar(flagExperimental, "experimental", false,
		"run using osemgrep, the statically compiled version of semgrep using OCAML")

	// Code Scan flags - snyk
	settings.SetupCobra(&metaConfig.CodeScanSnykFilename, runSASTCodeScanTask)
	settings.SetupCobra(&metaConfig.CodeScanSnykSrcDir, runSASTCodeScanTask)
	runSASTCodeScanTask.Flags().BoolVar(flagSnyk, "snyk", false, "use snyk for SAST scan")

	runTaskCmd.AddCommand(
		runImageScanTask,
		runAntivirusScanTask,
		runImageBuildTask,
		runSASTCodeScanTask,
		runSecretsCodeScanTask,
	)

	return runTaskCmd
}

func configPreRunE(cmd *cobra.Command, args []string) error {
	err := settings.Unmarshal(config, metaConfig)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")

	return enc.Encode(config)
}

var runImageBuildTask = &cobra.Command{
	Use:     "image-build",
	Short:   "build an image",
	PreRunE: configPreRunE,
	RunE: func(cmd *cobra.Command, args []string) error {
		cliInterfaceStr := "docker"
		if *flagPodmanInterface {
			cliInterfaceStr = "podman"
		}

		buildOpts := tasks.ImageBuildOptions{
			Context:      config.ImageBuild.BuildDir,
			Dockerfile:   config.ImageBuild.Dockerfile,
			Platform:     config.ImageBuild.Platform,
			Target:       config.ImageBuild.Target,
			CacheTo:      config.ImageBuild.CacheTo,
			CacheFrom:    config.ImageBuild.CacheFrom,
			SquashLayers: config.ImageBuild.SquashLayers,
			BuildArgs:    config.ImageBuild.Args,
		}
		opts := tasks.WithImageBuildOptions(buildOpts)
		task := tasks.NewImageBuildTask(cliInterfaceStr, opts)
		return task.Run(cmd.Context(), cmd.ErrOrStderr())
	},
}

var runImageScanTask = &cobra.Command{
	Use:     "image-vul-scan",
	Short:   "run security scans on an image",
	PreRunE: configPreRunE,
	RunE: func(cmd *cobra.Command, args []string) error {
		imageOptions := tasks.WithImgVulOptions(
			config.ImageTag,
			config.ImageScan.SyftFilename,
			config.ImageScan.GrypeFilename,
			config.ArtifactDir,
		)

		task := tasks.NewImageVulScanTask(tasks.GrypeTaskType, imageOptions, tasks.WithStdout(cmd.OutOrStdout()))

		return task.Run(cmd.Context(), cmd.ErrOrStderr())
	},
}

var runAntivirusScanTask = &cobra.Command{
	Use:     "image-antivirus-scan",
	Short:   "run an antivirus scan on an image or image archive",
	PreRunE: configPreRunE,
	RunE: func(cmd *cobra.Command, args []string) error {
		cliInterfaceStr := "docker"
		if *flagPodmanInterface {
			cliInterfaceStr = "podman"
		}

		f, err := os.CreateTemp(os.TempDir(), "*.container-image.tar")
		if err != nil {
			return err
		}
		imageTarFilename := f.Name()
		_ = f.Close()

		imageSaveOpts := tasks.WithImageSaveOptions(config.ImageTag, imageTarFilename, *flagAntivirusPull)
		imageSaveTask := tasks.NewImageSaveTask(cliInterfaceStr, imageSaveOpts)

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

var runSecretsCodeScanTask = &cobra.Command{
	Use:     "secrets-code-scan",
	Short:   "run secrets dectection in the code base",
	PreRunE: configPreRunE,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := tasks.CodeScanOptions{
			GitleaksFilename: path.Join(config.ArtifactDir, config.CodeScan.GitleaksFilename),
			GitleaksSrc:      config.CodeScan.GitleaksSrcDir,
		}

		task := new(tasks.GitleaksCodeScanTask)
		task.SetOptions(opts)
		task.SetDisplayWriter(cmd.OutOrStdout())
		return task.Run(cmd.Context(), cmd.ErrOrStderr())
	},
}

var runImagePushTask = &cobra.Command{
	Use:     "image-push",
	Short:   "push an image to an image registry",
	PreRunE: configPreRunE,
	RunE: func(cmd *cobra.Command, args []string) error {
		if *flagPodmanInterface {
			task := tasks.NewGenericImagePushTask("podman", config.ImageTag)
			return task.Run(cmd.Context(), cmd.ErrOrStderr())
		}
		task := tasks.NewGenericImagePushTask("docker", config.ImageTag)
		return task.Run(cmd.Context(), cmd.ErrOrStderr())
	},
}

var runSASTCodeScanTask = &cobra.Command{
	Use:     "sast-code-scan",
	Short:   "run static analysis security testing (SAST) on the code base",
	PreRunE: configPreRunE,
	RunE: func(cmd *cobra.Command, args []string) error {
		if *flagSnyk {
			return runSnyk(cmd, args)
		}
		return runSemgrep(cmd, args)
	},
}

func runSemgrep(cmd *cobra.Command, args []string) error {
	opts := tasks.CodeScanOptions{
		SemgrepRules:        config.CodeScan.SemgrepRules,
		SemgrepFilename:     path.Join(config.ArtifactDir, config.CodeScan.SemgrepFilename),
		SemgrepExperimental: *flagExperimental,
	}
	task := new(tasks.SemgrepCodeScanTask)
	task.SetOptions(opts)
	task.SetDisplayWriter(cmd.OutOrStdout())
	return task.Run(cmd.Context(), cmd.ErrOrStderr())
}

func runSnyk(cmd *cobra.Command, args []string) error {
	opts := tasks.CodeScanOptions{
		SnykCodeFilename: path.Join(config.ArtifactDir, config.CodeScan.SnykFilename),
		SnykSrcDir:       config.CodeScan.SnykSrcDir,
	}

	task := new(tasks.SnykCodeScanTask)
	task.SetOptions(opts)
	task.SetDisplayWriter(cmd.OutOrStdout())
	return task.Run(cmd.Context(), cmd.ErrOrStderr())
}
