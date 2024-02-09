package main

import (
	"log/slog"
	"os"
	"time"
	"workflow-engine/cmd/workflow-engine/cli"

	"github.com/lmittmann/tint"
)

const CLIVersion = "v0.0.0"
const exitOk = 0
const exitUserInput = 1
const exitSystemFailure = 2
const exitCommandFailure = 3

const pipelineTypeDebug = "debug"

func main() {
	lvler := &slog.LevelVar{}
	lvler.Set(slog.LevelInfo)
	// Set up custom structured logging with colorized output
	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:      lvler,
		TimeFormat: time.TimeOnly,
	})))

	cmd := cli.NewWorkflowEngineCommand(lvler)
	if err := cmd.Execute(); err != nil {
		slog.Error("command execution failure. See log for details")
		os.Exit(exitCommandFailure)
	}

}
