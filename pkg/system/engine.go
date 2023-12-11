package system

import (
	"context"
	"io"
	"log/slog"
	"time"
	"workflow-engine/pkg/pipelines"

	"dagger.io/dagger"
)

// Config is used to set arbitrary configuration settings that will be used in the engine
type Config struct {
	systemLogWriter  io.Writer // Dagger system logs
	commandLogWriter io.Writer // stdout from the container being run
	AlpineImage      string    `toml:"alpineImage"`
	OmnibusImage     string    `toml:"omnibusImage"`
}

// NewConfig can be used to generate a Config with default options set
func NewConfig() *Config {
	return &Config{
		systemLogWriter:  io.Discard,
		commandLogWriter: io.Discard,
	}
}

// ConfigOpt defines structs that set some configuration option
type ConfigOpt interface {
	setOpt(*Config)
}

// ConfigOptFunc is a specific function used to set configuration values
type configOptFunc func(cfg *Config)

func (fn configOptFunc) setOpt(cfg *Config) {
	fn(cfg)
}

// WithSystemLogger writes dagger system logs to the given writer
func WithSystemLogger(w io.Writer) ConfigOpt {
	return configOptFunc(func(cfg *Config) {
		cfg.systemLogWriter = w
	})
}

// WithCommandLogger writes command output to the given writer
func WithCommandLogger(w io.Writer) ConfigOpt {
	return configOptFunc(func(cfg *Config) {
		cfg.commandLogWriter = w
	})
}

// Engine contains the context and configuration for commands
type Engine struct {
	client        *dagger.Client
	config        *Config
	debugPipeline *pipelines.Debug
}

// NewEngine will create a new engine already connected to the client
func NewEngine(opts ...ConfigOpt) (*Engine, error) {
	config := NewConfig()
	for _, o := range opts {
		o.setOpt(config)
	}
	client, err := dagger.Connect(context.Background(), dagger.WithLogOutput(config.systemLogWriter))
	if err != nil {
		slog.Error("dagger could not connect to client", "err", err)
		return nil, err
	}
	engine := &Engine{
		client:        client,
		config:        config,
		debugPipeline: pipelines.NewDebugPipeline(client, config.commandLogWriter),
	}

	return engine, nil
}

// DebugPipeline for testing purposes
func (e *Engine) DebugPipeline() error {
	timeout := time.Second * 5
	doneChan := make(chan struct{}, 1)

	var debugPipelineError error = nil
	slog.Info("Running Debug Pipeline", "timeout", timeout)

	go func() {
		debugPipelineError = e.debugPipeline.Execute()
		doneChan <- struct{}{}
	}()

	select {
	case <-time.After(timeout):
		slog.Error("Debug Pipeline timed out")
		return debugPipelineError
	case <-doneChan:
		slog.Info("Debug Pipeline complete")
		return debugPipelineError
	}
}
