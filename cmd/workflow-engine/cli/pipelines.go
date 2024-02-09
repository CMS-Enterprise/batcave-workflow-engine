package cli

import (
	"io"
	"workflow-engine/pkg/pipelines"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newRunCommand() *cobra.Command {
	// Create Flags, Bind Flags, Bind Environment Variables
	debugCmd := newBasicCommand("debug", "pipeline for smoke testing this application", runDebug)

	// run image-build
	imageBuildCmd := newBasicCommand("image-build", "", runImageBuild)

	imageBuildCmd.Flags().StringP("cli-interface", "i", "docker", "[docker|podman] CLI interface to use for image building")

	imageBuildCmd.Flags().String("build-dir", ".", "image build context directory")
	viper.BindPFlag("image.builddir", imageBuildCmd.Flags().Lookup("build-dir"))
	viper.MustBindEnv("image.builddir", "WFE_BUILD_DIR")

	imageBuildCmd.Flags().String("dockerfile", "Dockerfile", "image build custom Dockerfile")
	viper.BindPFlag("image.builddockerfile", imageBuildCmd.Flags().Lookup("dockerfile"))
	viper.MustBindEnv("image.builddockerfile", "WFE_BUILD_DOCKERFILE")

	imageBuildCmd.Flags().String("tag", "", "image build custom tag")
	viper.BindPFlag("image.buildtag", imageBuildCmd.Flags().Lookup("tag"))
	viper.MustBindEnv("image.buildtag", "WFE_BUILD_TAG")

	imageBuildCmd.Flags().String("platform", "", "image build custom platform option")
	viper.BindPFlag("image.buildplatform", imageBuildCmd.Flags().Lookup("platform"))
	viper.MustBindEnv("image.buildplatform", "WFE_BUILD_PLATFORM")

	imageBuildCmd.Flags().String("target", "", "image build custom target option")
	viper.BindPFlag("image.buildtarget", imageBuildCmd.Flags().Lookup("target"))
	viper.MustBindEnv("image.buildtarget", "WFE_BUILD_TARGET")

	imageBuildCmd.Flags().String("cache-to", "", "image build custom cache-to option")
	viper.BindPFlag("image.buildcacheto", imageBuildCmd.Flags().Lookup("cache-to"))
	viper.MustBindEnv("image.buildcacheto", "WFE_BUILD_CACHE_TO")

	imageBuildCmd.Flags().String("cache-from", "", "image build custom cache-from option")
	viper.BindPFlag("image.buildcachefrom", imageBuildCmd.Flags().Lookup("cache-from"))
	viper.MustBindEnv("image.buildcachefrom", "WFE_BUILD_CACHE_FROM")

	imageBuildCmd.Flags().Bool("squash-layers", true, "image build squash all layers into one option")
	viper.BindPFlag("image.buildsquashlayers", imageBuildCmd.Flags().Lookup("squash-layers"))
	viper.MustBindEnv("image.buildsquashlayers", "WFE_BUILD_SQUASH_LAYERS")

	// run image-scan
	imageScanCmd := newBasicCommand("image-scan", "", runImageScan)

	imageScanCmd.Flags().String("artifact-directory", "", "the output directory for all artifacts generated in the pipeline")
	viper.BindPFlag("artifacts.directory", imageScanCmd.Flags().Lookup("artifact-directory"))
	viper.MustBindEnv("artifacts.directory", "WFE_ARTIFACT_DIRECTORY")

	imageScanCmd.Flags().String("sbom-filename", "", "the output filename for the syft SBOM")
	viper.BindPFlag("artifacts.sbomfilename", imageScanCmd.Flags().Lookup("sbom-filename"))
	viper.MustBindEnv("artifacts.sbomfilename", "WFE_SBOM_FILENAME")

	imageScanCmd.Flags().String("grype-filename", "", "the output filename for the grype vulnerability report")
	viper.BindPFlag("artifacts.grypefilename", imageScanCmd.Flags().Lookup("grype-filename"))
	viper.MustBindEnv("artifacts.grypefilename", "WFE_GRYPE_FILENAME")

	imageScanCmd.Flags().String("scan-image-target", "", "scan a specific image")
	viper.BindPFlag("image.scantarget", imageScanCmd.Flags().Lookup("scan-image-target"))
	viper.MustBindEnv("image.scantarget", "WFE_SCAN_IMAGE_TARGET")

	// run
	cmd := &cobra.Command{Use: "run", Short: "run a pipeline"}

	// Persistent Flags, available on all sub commands
	cmd.PersistentFlags().BoolP("dry-run", "n", false, "log commands to debug but don't execute")
	cmd.PersistentFlags().StringP("config", "f", "", "workflow engine config file in json, yaml, or toml")

	// Flag marks
	cmd.MarkFlagFilename("config", "json", "yaml", "yml", "toml")
	cmd.MarkFlagDirname("artifact-directory")
	cmd.MarkFlagDirname("build-dir")

	// Other settings
	cmd.SilenceUsage = true

	// Add sub commands
	cmd.AddCommand(debugCmd, imageBuildCmd, imageScanCmd)

	return cmd
}

// Run Functions - Parsing flags and arguments at command runtime

func runDebug(cmd *cobra.Command, _ []string) error {
	dryRunEnabled, _ := cmd.Flags().GetBool("dry-run")
	return debugPipeline(cmd.OutOrStdout(), cmd.ErrOrStderr(), dryRunEnabled)
}

func runImageBuild(cmd *cobra.Command, _ []string) error {
	dryRunEnabled, _ := cmd.Flags().GetBool("dry-run")
	cliInterface, _ := cmd.Flags().GetString("cli-interface")
	config, err := Config(cmd)
	if err != nil {
		return err
	}

	return imageBuildPipeline(cmd.OutOrStdout(), cmd.ErrOrStderr(), config.Image, dryRunEnabled, cliInterface)
}

func runImageScan(cmd *cobra.Command, _ []string) error {
	dryRunEnabled, _ := cmd.Flags().GetBool("dry-run")
	config, err := Config(cmd)
	if err != nil {
		return err
	}
	return imageScanPipeline(cmd.OutOrStdout(), cmd.ErrOrStderr(), config.Artifacts, dryRunEnabled, config.Image.ScanTarget)
}

// Execution functions - Logic for command execution

func imageBuildPipeline(stdout io.Writer, stderr io.Writer, config pipelines.ImageConfig, dryRunEnabled bool, cliInterface string) error {
	pipeline := pipelines.NewImageBuild(stdout, stderr)
	pipeline.DryRunEnabled = dryRunEnabled
	if cliInterface == "podman" {
		pipeline = pipeline.WithPodman()
	}
	return pipeline.WithBuildConfig(config).Run()
}

func imageScanPipeline(stdout io.Writer, stderr io.Writer, config pipelines.ArtifactConfig, dryRunEnabled bool, imageName string) error {
	pipeline := pipelines.NewImageScan(stdout, stderr)
	pipeline.DryRunEnabled = dryRunEnabled

	return pipeline.WithArtifactConfig(config).WithImageName(imageName).Run()
}

func debugPipeline(stdout io.Writer, stderr io.Writer, dryRunEnabled bool) error {
	pipeline := pipelines.NewDebug(stdout, stderr)
	pipeline.DryRunEnabled = dryRunEnabled
	return pipeline.Run()
}
