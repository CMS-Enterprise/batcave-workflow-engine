package main

import (
	"bytes"
	"fmt"
	"os"
	"time"
	"workflow-engine/pkg/system"

	"log/slog"

	"github.com/lmittmann/tint"
)

func main() {
	// Set up custom structured logging with colorized output
	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.TimeOnly,
	})))
	systemLogBuf := new(bytes.Buffer)
	commandLogBuf := new(bytes.Buffer)
	engine, err := system.NewEngine(system.WithSystemLogger(systemLogBuf), system.WithCommandLogger(commandLogBuf))
	if err != nil {
		panic(err)
	}

	err = engine.DebugPipeline()

	fmt.Fprintf(os.Stderr, "===== System Log =====\n%s\n===== Command Log =====\n", systemLogBuf.String())
	commandLogBuf.WriteTo(os.Stdout)
	if err != nil {
		slog.Error("execution failure", "error", err)
	}
}
