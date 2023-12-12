package main

import (
	"context"
	"os"
	"time"
	"workflow-engine/cmd/workflow-engine/cli"

	"log/slog"

	"github.com/bep/simplecobra"
	"github.com/lmittmann/tint"
)

const exitOk = 0
const exitSystemFailure = 1
const exitCommandFailure = 2

func main() {
	leveler := new(slog.LevelVar)
	// Set up custom structured logging with colorized output
	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:      leveler,
		TimeFormat: time.TimeOnly,
	})))

	command := cli.NewCommand(leveler)

	x, err := simplecobra.New(command)
	if err != nil {
		os.Exit(exitSystemFailure)
	}

	_, err = x.Execute(context.Background(), os.Args[1:])
	if err != nil {
		os.Exit(exitCommandFailure)
	}

	os.Exit(exitOk)
}
