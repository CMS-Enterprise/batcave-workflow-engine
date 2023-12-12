package cli

import (
	"context"
	"log/slog"

	"github.com/bep/simplecobra"
)

type Command struct {
	Version            string
	LogLevelController *slog.LevelVar
}

func NewCommand(lvlr *slog.LevelVar) *Command {
	return &Command{LogLevelController: lvlr}
}

func (_ *Command) Name() string {
	return "workflow-engine"
}

func (_ *Command) Commands() []simplecobra.Commander {
	return []simplecobra.Commander{}
}

func (c *Command) Init(cmdr *simplecobra.Commandeer) error {
	cmdr.CobraCommand.Version = c.Version
	cmdr.CobraCommand.Short = "Workflow Engine - A Portable Security Pipeline"
	cmdr.CobraCommand.PersistentFlags().BoolP("verbose", "v", false, "verbose logging output")
	cmdr.CobraCommand.InitDefaultVersionFlag()
	return nil
}

func (c *Command) PreRun(cmdr *simplecobra.Commandeer, _ *simplecobra.Commandeer) error {
	if verbose, _ := cmdr.CobraCommand.Flags().GetBool("verbose"); verbose {
		c.LogLevelController.Set(slog.LevelDebug)
	}
	return nil
}

func (_ *Command) Run(ctx context.Context, cmdr *simplecobra.Commandeer, args []string) error {
	return nil
}
