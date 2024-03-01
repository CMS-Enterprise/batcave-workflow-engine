package main

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"
	"workflow-engine/cmd/workflow-engine/cli"

	"github.com/lmittmann/tint"
)

const exitOK = 0
const exitCommandFailure = 1

var (
	cliVersion     = "[Not Provided]"
	buildDate      = "[Not Provided]"
	gitCommit      = "[Not Provided]"
	gitDescription = "[Not Provided]"
)

func main() {

	os.Exit(runCLI())
}

func runCLI() int {
	lvler := &slog.LevelVar{}
	lvler.Set(slog.LevelInfo)
	// Set up custom structured logging with colorized output
	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:      lvler,
		TimeFormat: time.TimeOnly,
	})))

	cli.AppMetadata = cli.ApplicationMetadata{
		CLIVersion:     cliVersion,
		GitCommit:      gitCommit,
		BuildDate:      buildDate,
		GitDescription: gitDescription,
		Platform:       fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		GoVersion:      runtime.Version(),
		Compiler:       runtime.Compiler,
	}

	cmd := cli.NewWorkflowEngineCommand(lvler)
	if err := cmd.Execute(); err != nil {
		slog.Error("command execution failure. See log for details")
		return exitCommandFailure
	}
	return exitOK
}
