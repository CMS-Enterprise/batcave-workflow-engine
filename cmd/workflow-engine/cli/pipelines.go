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
	imageBuildCmd := newBasicCommand("image-build", "builds an image", runImageBuild)

	imageBuildCmd.Flags().StringP("cli-interface", "i", "docker", "[docker|podman] CLI interface to use for image building")

	imageBuildCmd.Flags().String("build-dir", ".", "image build context directory")
	_ = viper.BindPFlag("imagebuild.buildDir", imageBuildCmd.Flags().Lookup("build-dir"))

	imageBuildCmd.Flags().String("dockerfile", "Dockerfile", "image build custom Dockerfile")
	_ = viper.BindPFlag("image.dockerfile", imageBuildCmd.Flags().Lookup("dockerfile"))

	imageBuildCmd.Flags().StringToString("build-arg", map[string]string{}, "A build argument passed to the container build command")
	_ = viper.BindPFlag("imagebuild.args", imageBuildCmd.Flags().Lookup("build-arg"))

	imageBuildCmd.Flags().String("tag", "", "image build custom tag")
	_ = viper.BindPFlag("imagebuild.tag", imageBuildCmd.Flags().Lookup("tag"))

	imageBuildCmd.Flags().String("platform", "", "image build custom platform option")
	_ = viper.BindPFlag("imagebuild.platform", imageBuildCmd.Flags().Lookup("platform"))

	imageBuildCmd.Flags().String("target", "", "image build custom target option")
	_ = viper.BindPFlag("imagebuild.target", imageBuildCmd.Flags().Lookup("target"))

	imageBuildCmd.Flags().String("cache-to", "", "image build custom cache-to option")
	_ = viper.BindPFlag("imagebuild.cacheto", imageBuildCmd.Flags().Lookup("cache-to"))

	imageBuildCmd.Flags().String("cache-from", "", "image build custom cache-from option")
	_ = viper.BindPFlag("imagebuild.cachefrom", imageBuildCmd.Flags().Lookup("cache-from"))

	imageBuildCmd.Flags().Bool("squash-layers", true, "image build squash all layers into one option")
	_ = viper.BindPFlag("imagebuild.squashlayers", imageBuildCmd.Flags().Lookup("squash-layers"))

	// run image-scan
	imageScanCmd := newBasicCommand("image-scan", "run security scans on an image", runImageScan)

	imageScanCmd.Flags().String("artifact-directory", "", "the output directory for all artifacts generated in the pipeline")
	_ = viper.BindPFlag("artifactsdir", imageScanCmd.Flags().Lookup("artifact-directory"))

	imageScanCmd.Flags().String("sbom-filename", "", "the output filename for the syft SBOM")
	_ = viper.BindPFlag("imagescan.syftfilename", imageScanCmd.Flags().Lookup("sbom-filename"))

	imageScanCmd.Flags().String("grype-filename", "", "the output filename for the grype vulnerability report")
	_ = viper.BindPFlag("imagescan.grypefilename", imageScanCmd.Flags().Lookup("grype-filename"))

	imageScanCmd.Flags().String("scan-image-target", "", "scan a specific image")
	_ = viper.BindPFlag("imagescan.targetimage", imageScanCmd.Flags().Lookup("scan-image-target"))

	// run image-publish
	imagePublishCmd := newBasicCommand("image-publish", "publishes an image", runimagePublish)

	// run code-scan
	codeScanCmd := newBasicCommand("code-scan", "run Static Application Security Tests (SAST) scans", runCodeScan)

	codeScanCmd.Flags().String("gitleaks-filename", "", "the output filename for the gitleaks vulnerability report")
	_ = viper.BindPFlag("codescan.gitleaksfilename", codeScanCmd.Flags().Lookup("gitleaks-filename"))

	codeScanCmd.Flags().String("semgrep-filename", "", "the output filename for the semgrep vulnerability report")
	_ = viper.BindPFlag("codescan.semgrepfilename", codeScanCmd.Flags().Lookup("semgrep-filename"))

	codeScanCmd.Flags().String("semgrep-rules", "", "the rules semgrep will use for the scan")
	_ = viper.BindPFlag("codescan.semgreprules", codeScanCmd.Flags().Lookup("semgrep-rules"))

	codeScanCmd.Flags().Bool("semgrep-experimental", false, "use the osemgrep statically compiled binary")

	// run deploy
	deployCmd := newBasicCommand("deploy", "Beta Feature: VALIDATION ONLY - run gatecheck validate on artifacts from previous pipelines", runDeploy)

	// run
	cmd := &cobra.Command{Use: "run", Short: "run a pipeline"}

	// Persistent Flags, available on all sub commands
	cmd.PersistentFlags().BoolP("dry-run", "n", false, "log commands to debug but don't execute")
	cmd.PersistentFlags().StringP("config", "f", "", "workflow engine config file in json, yaml, or toml")

	// Flag marks
	_ = cmd.MarkFlagFilename("config", "json", "yaml", "yml", "toml")
	_ = cmd.MarkFlagDirname("artifact-directory")
	_ = cmd.MarkFlagDirname("build-dir")

	// Other settings
	cmd.SilenceUsage = true

	// Add sub commands
	cmd.AddCommand(debugCmd, imageBuildCmd, imageScanCmd, imagePublishCmd, codeScanCmd, deployCmd)

	return cmd
}

// Run Functions - Parsing flags and arguments at command runtime

func runDebug(cmd *cobra.Command, _ []string) error {
	dryRunEnabled, _ := cmd.Flags().GetBool("dry-run")
	return debugPipeline(cmd.OutOrStdout(), cmd.ErrOrStderr(), dryRunEnabled)
}

func runDeploy(cmd *cobra.Command, _ []string) error {
	dryRunEnabled, _ := cmd.Flags().GetBool("dry-run")
	configFilename, _ := cmd.Flags().GetString("config")

	config := new(pipelines.Config)
	if err := LoadOrDefault(configFilename, config, viper.GetViper()); err != nil {
		return err
	}

	return deployPipeline(cmd.OutOrStdout(), cmd.ErrOrStderr(), config, dryRunEnabled)
}

func runImageBuild(cmd *cobra.Command, _ []string) error {
	dryRunEnabled, _ := cmd.Flags().GetBool("dry-run")
	cliInterface, _ := cmd.Flags().GetString("cli-interface")
	configFilename, _ := cmd.Flags().GetString("config")

	config := new(pipelines.Config)
	if err := LoadOrDefault(configFilename, config, viper.GetViper()); err != nil {
		return err
	}

	return imageBuildPipeline(cmd.OutOrStdout(), cmd.ErrOrStderr(), config, dryRunEnabled, cliInterface)
}

func runImageScan(cmd *cobra.Command, _ []string) error {
	dryRunEnabled, _ := cmd.Flags().GetBool("dry-run")
	configFilename, _ := cmd.Flags().GetString("config")

	config := new(pipelines.Config)
	if err := LoadOrDefault(configFilename, config, viper.GetViper()); err != nil {
		return err
	}

	return imageScanPipeline(cmd.OutOrStdout(), cmd.ErrOrStderr(), config, dryRunEnabled)
}

func runimagePublish(cmd *cobra.Command, _ []string) error {
	dryRunEnabled, _ := cmd.Flags().GetBool("dry-run")
	configFilename, _ := cmd.Flags().GetString("config")

	config := new(pipelines.Config)
	if err := LoadOrDefault(configFilename, config, viper.GetViper()); err != nil {
		return err
	}
	return imagePublishPipeline(cmd.OutOrStdout(), cmd.ErrOrStderr(), config, dryRunEnabled)
}

func runCodeScan(cmd *cobra.Command, _ []string) error {
	dryRunEnabled, _ := cmd.Flags().GetBool("dry-run")
	configFilename, _ := cmd.Flags().GetString("config")
	semgrepExperimental, _ := cmd.Flags().GetBool("semgrep-experimental")

	config := new(pipelines.Config)
	if err := LoadOrDefault(configFilename, config, viper.GetViper()); err != nil {
		return err
	}

	return codeScanPipeline(cmd.OutOrStdout(), cmd.ErrOrStderr(), config, dryRunEnabled, semgrepExperimental)
}

// Execution functions - Logic for command execution

func imageBuildPipeline(stdout io.Writer, stderr io.Writer, config *pipelines.Config, dryRunEnabled bool, cliInterface string) error {
	pipeline := pipelines.NewImageBuild(stdout, stderr)
	pipeline.DryRunEnabled = dryRunEnabled
	if cliInterface == "podman" {
		pipeline = pipeline.WithPodman()
	}
	return pipeline.WithBuildConfig(config).Run()
}

func imageScanPipeline(stdout io.Writer, stderr io.Writer, config *pipelines.Config, dryRunEnabled bool) error {
	pipeline := pipelines.NewImageScan(stdout, stderr)
	pipeline.DryRunEnabled = dryRunEnabled

	return pipeline.WithConfig(config).Run()
}

func imagePublishPipeline(stdout io.Writer, stderr io.Writer, config *pipelines.Config, dryRunEnabled bool) error {
	pipeline := pipelines.NewimagePublish(stdout, stderr)
	pipeline.DryRunEnabled = dryRunEnabled

	return pipeline.WithConfig(config).Run()
}

func codeScanPipeline(stdout io.Writer, stderr io.Writer, config *pipelines.Config, dryRunEnabled bool, semgrepExperimental bool) error {
	pipeline := pipelines.NewCodeScan(stdout, stderr)
	pipeline.DryRunEnabled = dryRunEnabled
	pipeline.SemgrepExperimental = semgrepExperimental

	return pipeline.WithConfig(config).Run()
}

func deployPipeline(stdout io.Writer, stderr io.Writer, config *pipelines.Config, dryRunEnabled bool) error {
	pipeline := pipelines.NewDeploy(stdout, stderr)
	pipeline.DryRunEnabled = dryRunEnabled

	return pipeline.WithConfig(config).Run()
}

func debugPipeline(stdout io.Writer, stderr io.Writer, dryRunEnabled bool) error {
	pipeline := pipelines.NewDebug(stdout, stderr)
	pipeline.DryRunEnabled = dryRunEnabled

	return pipeline.Run()
}
