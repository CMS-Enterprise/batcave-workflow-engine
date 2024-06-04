package main

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"
	"workflow-engine/cmd/workflow-engine/cli/v0"
	cliv1 "workflow-engine/cmd/workflow-engine/cli/v1"

	"github.com/lmittmann/tint"
)

const (
	exitOK             = 0
	exitCommandFailure = 1
)

var (
	cliVersion        = "[Not Provided]"
	buildDate         = "[Not Provided]"
	gitCommit         = "[Not Provided]"
	gitDescription    = "[Not Provided]"
	experimentalCLIv1 = "0"
)

func main() {
	if experimentalCLIv1 == "1" {
		os.Exit(runCLIv1())
	}
	os.Exit(runCLIv0())
}

func runCLIv1() int {
	cliv1.AppLogLever = &slog.LevelVar{}
	cliv1.AppLogLever.Set(slog.LevelDebug)
	// Set up custom structured logging with colorized output
	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:      cliv1.AppLogLever,
		TimeFormat: time.TimeOnly,
	})))

	cmd := cliv1.NewWorkflowEngineCommand()

	start := time.Now()
	slog.Debug("execute command")
	defer func(t time.Time) {
	}(start)

	err := cmd.Execute()
	elapsed := time.Since(start)

	if err != nil {
		slog.Error("done", "elapsed", elapsed)
		return 1
	}
	slog.Info("done", "elapsed", elapsed)
	return 0
}

func runCLIv0() int {
	lvler := &slog.LevelVar{}
	lvler.Set(slog.LevelInfo)
	// Set up custom structured logging with colorized output
	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:      lvler,
		TimeFormat: time.TimeOnly,
	})))
	cli.AppLogLever = lvler
	cli.AppMetadata = cli.ApplicationMetadata{
		CLIVersion:     cliVersion,
		GitCommit:      gitCommit,
		BuildDate:      buildDate,
		GitDescription: gitDescription,
		Platform:       fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		GoVersion:      runtime.Version(),
		Compiler:       runtime.Compiler,
	}

	cmd := cli.NewWorkflowEngineCommand()
	start := time.Now()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "------------")
		slog.Error(fmt.Sprintf("%v", err), "elapsed", time.Since(start))
		return exitCommandFailure
	}
	fmt.Fprintln(os.Stderr, "------------")
	slog.Info("done", "elapsed", time.Since(start))
	return exitOK
}
