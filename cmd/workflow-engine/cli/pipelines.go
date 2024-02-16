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
	_ = viper.BindPFlag("image.builddir", imageBuildCmd.Flags().Lookup("build-dir"))
	viper.MustBindEnv("image.builddir", "WFE_BUILD_DIR")

	imageBuildCmd.Flags().String("dockerfile", "Dockerfile", "image build custom Dockerfile")
	_ = viper.BindPFlag("image.builddockerfile", imageBuildCmd.Flags().Lookup("dockerfile"))
	viper.MustBindEnv("image.builddockerfile", "WFE_BUILD_DOCKERFILE")

	imageBuildCmd.Flags().StringToString("build-arg", map[string]string{}, "A build argument passed to the container build command")
	_ = viper.BindPFlag("image.buildargs", imageBuildCmd.Flags().Lookup("build-arg"))
	viper.MustBindEnv("image.buildargs", "WFE_BUILD_ARGS")

	imageBuildCmd.Flags().String("tag", "", "image build custom tag")
	_ = viper.BindPFlag("image.buildtag", imageBuildCmd.Flags().Lookup("tag"))
	viper.MustBindEnv("image.buildtag", "WFE_BUILD_TAG")

	imageBuildCmd.Flags().String("platform", "", "image build custom platform option")
	_ = viper.BindPFlag("image.buildplatform", imageBuildCmd.Flags().Lookup("platform"))
	viper.MustBindEnv("image.buildplatform", "WFE_BUILD_PLATFORM")

	imageBuildCmd.Flags().String("target", "", "image build custom target option")
	_ = viper.BindPFlag("image.buildtarget", imageBuildCmd.Flags().Lookup("target"))
	viper.MustBindEnv("image.buildtarget", "WFE_BUILD_TARGET")

	imageBuildCmd.Flags().String("cache-to", "", "image build custom cache-to option")
	_ = viper.BindPFlag("image.buildcacheto", imageBuildCmd.Flags().Lookup("cache-to"))
	viper.MustBindEnv("image.buildcacheto", "WFE_BUILD_CACHE_TO")

	imageBuildCmd.Flags().String("cache-from", "", "image build custom cache-from option")
	_ = viper.BindPFlag("image.buildcachefrom", imageBuildCmd.Flags().Lookup("cache-from"))
	viper.MustBindEnv("image.buildcachefrom", "WFE_BUILD_CACHE_FROM")

	imageBuildCmd.Flags().Bool("squash-layers", true, "image build squash all layers into one option")
	_ = viper.BindPFlag("image.buildsquashlayers", imageBuildCmd.Flags().Lookup("squash-layers"))
	viper.MustBindEnv("image.buildsquashlayers", "WFE_BUILD_SQUASH_LAYERS")

	// run image-scan
	imageScanCmd := newBasicCommand("image-scan", "run security scans on an image", runImageScan)

	imageScanCmd.Flags().String("artifact-directory", "", "the output directory for all artifacts generated in the pipeline")
	_ = viper.BindPFlag("artifacts.directory", imageScanCmd.Flags().Lookup("artifact-directory"))
	viper.MustBindEnv("artifacts.directory", "WFE_ARTIFACT_DIRECTORY")

	imageScanCmd.Flags().String("sbom-filename", "", "the output filename for the syft SBOM")
	_ = viper.BindPFlag("artifacts.sbomfilename", imageScanCmd.Flags().Lookup("sbom-filename"))
	viper.MustBindEnv("artifacts.sbomfilename", "WFE_SBOM_FILENAME")

	imageScanCmd.Flags().String("grype-filename", "", "the output filename for the grype vulnerability report")
	_ = viper.BindPFlag("artifacts.grypefilename", imageScanCmd.Flags().Lookup("grype-filename"))
	viper.MustBindEnv("artifacts.grypefilename", "WFE_GRYPE_FILENAME")

	imageScanCmd.Flags().String("scan-image-target", "", "scan a specific image")
	_ = viper.BindPFlag("image.scantarget", imageScanCmd.Flags().Lookup("scan-image-target"))
	viper.MustBindEnv("image.scantarget", "WFE_SCAN_IMAGE_TARGET")

	// run code-scan
	codeScanCmd := newBasicCommand("code-scan", "run Static Application Security Tests (SAST) scans", runCodeScan)

	codeScanCmd.Flags().String("gitleaks-filename", "", "the output filename for the gitleaks vulnerability report")
	_ = viper.BindPFlag("artifacts.gitleaksfilename", codeScanCmd.Flags().Lookup("gitleaks-filename"))
	viper.MustBindEnv("artifacts.gitleaksfilename", "WFE_GITLEAKS_FILENAME")

	codeScanCmd.Flags().String("semgrep-filename", "", "the output filename for the semgrep vulnerability report")
	_ = viper.BindPFlag("artifacts.semgrepfilename", codeScanCmd.Flags().Lookup("semgrep-filename"))
	viper.MustBindEnv("artifacts.semgrepfilename", "WFE_SEMGREP_FILENAME")

	codeScanCmd.Flags().Bool("semgrep-experimental", false, "use the osemgrep statically compiled binary")
	codeScanCmd.Flags().Bool("semgrep-error-on-findings", false, "exit code 1 if findings are detected by semgrep")

	codeScanCmd.Flags().String("semgrep-rules", "p/default", "the rules semgrep will use for the scan")
	_ = viper.BindPFlag("semgrep.rules", codeScanCmd.Flags().Lookup("semgrep-rules"))
	viper.MustBindEnv("semgrep.rules", "SEMGREP_RULES")

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
	cmd.AddCommand(debugCmd, imageBuildCmd, imageScanCmd, codeScanCmd)

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
	configFilename, _ := cmd.Flags().GetString("config")

	config, err := Config(configFilename)
	if err != nil {
		return err
	}

	return imageBuildPipeline(cmd.OutOrStdout(), cmd.ErrOrStderr(), config.Image, dryRunEnabled, cliInterface)
}

func runImageScan(cmd *cobra.Command, _ []string) error {
	dryRunEnabled, _ := cmd.Flags().GetBool("dry-run")
	configFilename, _ := cmd.Flags().GetString("config")

	config, err := Config(configFilename)
	if err != nil {
		return err
	}
	return imageScanPipeline(cmd.OutOrStdout(), cmd.ErrOrStderr(), config.Artifacts, dryRunEnabled, config.Image.ScanTarget)
}

func runCodeScan(cmd *cobra.Command, _ []string) error {
	dryRunEnabled, _ := cmd.Flags().GetBool("dry-run")
	configFilename, _ := cmd.Flags().GetString("config")
	semgrepErrorOnFindings, _ := cmd.Flags().GetBool("semgrep-error-on-findings")
	semgrepExperimental, _ := cmd.Flags().GetBool("semgrep-experimental")
	semgrepRules, _ := cmd.Flags().GetString("semgrep-rules")

	config, err := Config(configFilename)
	if err != nil {
		return err
	}

	return codeScanPipeline(cmd.OutOrStdout(), cmd.ErrOrStderr(), config.Artifacts, dryRunEnabled,
		semgrepErrorOnFindings, semgrepExperimental, semgrepRules)
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

func codeScanPipeline(stdout io.Writer, stderr io.Writer, config pipelines.ArtifactConfig, dryRunEnabled bool,
	semgrepErrorOnFindings bool, semgrepExperimental bool, semgrepRules string) error {

	pipeline := pipelines.NewCodeScan(stdout, stderr)
	pipeline.DryRunEnabled = dryRunEnabled
	pipeline.SemgrepErrorOnFindingsEnabled = semgrepErrorOnFindings
	pipeline.SemgrepExperimental = semgrepExperimental
	pipeline.SemgrepRules = semgrepRules

	return pipeline.WithArtifactConfig(config).Run()
}

func debugPipeline(stdout io.Writer, stderr io.Writer, dryRunEnabled bool) error {
	pipeline := pipelines.NewDebug(stdout, stderr)
	pipeline.DryRunEnabled = dryRunEnabled
	return pipeline.Run()
}
