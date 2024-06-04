package cli

import (
	"io"
	"log/slog"
	"os"
	"workflow-engine/pkg/pipelines"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newConfigCommand() *cobra.Command {
	// config init
	initCmd := newBasicCommand("init <CONFIG FILE>.[json|yaml|yml|toml]", "write the default configuration file", runConfigInit)
	initCmd.Args = cobra.ExactArgs(1)

	// config info
	infoCmd := newBasicCommand("info <CONFIG FILE>.[json|yaml|yml|toml]", "print the loaded configuration values", runConfigInfo)
	infoCmd.Args = cobra.ExactArgs(1)

	// config vars
	varsCmd := newBasicCommand("vars", "list supported builtin variables that can be used in templates", runConfigVars)

	// config render
	renderCmd := newBasicCommand("render <TO CONFIG FILE>.[json|yaml|yml|toml] <FROM TEMPLATE>.[json|yaml|yml|toml]", "render a configuration template config using builtin variables", runConfigRender)
	renderCmd.Args = cobra.ExactArgs(2)

	// config convert
	convertCmd := newBasicCommand("convert <TO CONFIG FILE>.[json|yaml|yml|toml] <FROM CONFIG FILE>.[json|yaml|yml|toml]", "convert a configuration file", runConfigConvert)
	convertCmd.Args = cobra.ExactArgs(2)

	generateActionCmd := newBasicCommand("generate-action", "generate a single action for all pipelines", runGenAllAction)
	generateActionCmd.Flags().String("image", "Dockerfile", "The image or Dockerfile to use for the generated action")
	generateActionCmd.Flags().StringSlice("input", make([]string, 0), "Additional input(s) to make available as environment variables (format: \"input:variable:default:description\")")

	generateTableCmd := newBasicCommand("generate-table", "generate a markdown table with all of the keys, env variables, and defaults", runGenMarkdown)
	generateActionsTableCmd := newBasicCommand("generate-action-table", "generate a markdown table designed for github actions documentation", runGenActionMarkdown)
	generateActionsTableCmd.Flags().String("image", "Dockerfile", "The image or Dockerfile to use for the generated action")
	generateActionsTableCmd.Flags().StringSlice("input", make([]string, 0), "Additional input(s) to make available as environment variables (format: \"input:variable:default:description\")")

	// config
	cmd := &cobra.Command{Use: "config", Short: "manage the workflow engine config file"}

	// add sub commands
	cmd.AddCommand(infoCmd, initCmd, varsCmd, renderCmd, convertCmd, generateActionCmd, generateTableCmd, generateActionsTableCmd)

	return cmd
}

// Run Functions - Parsing flags and arguments at command runtime
func runGenAllAction(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")
	additionalInputs, _ := cmd.Flags().GetStringSlice("input")
	return pipelines.WriteGithubActionAll(cmd.OutOrStdout(), image, additionalInputs)
}

func runConfigInfo(cmd *cobra.Command, args []string) error {
	config := new(pipelines.Config)

	if err := LoadOrDefault(args[0], config, viper.GetViper()); err != nil {
		return err
	}
	return ListConfig(cmd.OutOrStdout(), viper.GetViper())
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	toConfigFilename := args[0]

	config := new(pipelines.Config)

	if err := LoadOrDefault(toConfigFilename, config, viper.GetViper()); err != nil {
		return err
	}

	return nil
}

func runGenMarkdown(cmd *cobra.Command, args []string) error {
	return pipelines.WriteConfigAsMarkdownTable(cmd.OutOrStdout())
}

func runGenActionMarkdown(cmd *cobra.Command, args []string) error {
	additionalInputs, _ := cmd.Flags().GetStringSlice("input")
	return pipelines.WriteConfigAsActionsTable(additionalInputs, cmd.OutOrStdout())
}

func runConfigVars(cmd *cobra.Command, _ []string) error {
	output, _ := cmd.Flags().GetString("output")

	return writeBuiltins(cmd.OutOrStdout(), output)
}

func runConfigRender(cmd *cobra.Command, _ []string) error {
	tmplFilename, _ := cmd.Flags().GetString("file")
	switch {
	case tmplFilename == "":
		return writeRenderConfigToFrom(cmd.OutOrStdout(), cmd.InOrStdin())
	default:
		return writeRenderedConfigTo(cmd.OutOrStdout(), tmplFilename)
	}
}

func runConfigConvert(cmd *cobra.Command, args []string) error {
	toConfigFilename := args[0]
	fromConfigFilename := args[1]

	slog.Debug("config convert", "to_config_filename", toConfigFilename, "from_config_filename", fromConfigFilename)

	// Let viper handle unmarshalling from the various file types without env or flag values
	tempViper := viper.New()

	tempViper.SetConfigFile(fromConfigFilename)
	if err := tempViper.ReadInConfig(); err != nil {
		return err
	}

	if err := tempViper.WriteConfigAs(toConfigFilename); err != nil {
		return err
	}

	return nil
}

// Execution functions - Logic for command execution

func writeRenderedConfigTo(w io.Writer, configTemplateFilename string) error {
	slog.Debug("open render src", "src_filename", configTemplateFilename)
	f, err := os.Open(configTemplateFilename)
	if err != nil {
		return err
	}
	return pipelines.RenderTemplate(w, f)
}

func writeRenderConfigToFrom(out io.Writer, in io.Reader) error {
	return pipelines.RenderTemplate(out, in)
}

func writeBuiltins(w io.Writer, asFormat string) error {
	builtins, err := pipelines.BuiltIns()
	if err != nil {
		return err
	}

	return NewAbstractEncoder(w, builtins).Encode(asFormat)
}
