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
	// Set up custom structured logging with colorized output
	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.TimeOnly,
	})))

	app := cli.NewApp()

	if err := app.Execute(); err != nil {
		slog.Error("command execution failure. See log for details")
		os.Exit(exitCommandFailure)
	}

}
