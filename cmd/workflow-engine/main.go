package main

import (
	"flag"
	"log/slog"
	"os"
	"slices"
	"time"
	"workflow-engine/pkg/system"

	"github.com/lmittmann/tint"
)

const exitOk = 0
const exitUserInput = 1
const exitSystemFailure = 2
const exitPipelineFailure = 3

const pipelineTypeDebug = "debug"

func main() {
	pipelineFlag := flag.String("pipeline", "", "pipeline to run options: [debug]")
	flag.Parse()

	// Set up custom structured logging with colorized output
	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.TimeOnly,
	})))

	slog.SetDefault(slog.Default().With("pipeline_type", *pipelineFlag))

	if !slices.Contains([]string{pipelineTypeDebug}, *pipelineFlag) {
		slog.Error("pipeline type not supported")
		os.Exit(exitUserInput)
	}

	// Create a new engine to run pipelines
	slog.Debug("connecting to dagger engine")
	engine, err := system.NewEngine()
	if err != nil {
		slog.Error("system failure", "err", err)
		os.Exit(exitSystemFailure)
	}

	slog.Info("executing pipeline")

	var pipelineErr error

	// Select the pipeline to run based on the flag argument
	switch *pipelineFlag {
	case pipelineTypeDebug:
		pipelineErr = engine.DebugPipeline()
	}

	if pipelineErr != nil {
		slog.Error("pipeline execution error", "err", pipelineErr)
		os.Exit(exitPipelineFailure)
	}
	slog.Info("pipeline execution complete")
}
