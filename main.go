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

const DefaultAlpineImage = "alpine:latest"
const DefaultOmnibusImage = "ghcr.io/nightwing-demo/omnibus:v1.0.0"

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

	if err := engine.DebugPipeline(); err != nil {
		slog.Error("execution failure", "error", err)
	}

	fmt.Printf("===== System Log =====\n%s\n===== Command Log =====\n%s\n", systemLogBuf.String(), commandLogBuf.String())
}
