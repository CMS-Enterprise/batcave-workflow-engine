package pipelines

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"os/exec"
)

type Debug struct {
	PipelineMode Mode
	Stdout       io.Writer
	Stderr       io.Writer
}

func NewDebug(m Mode) *Debug {
	return &Debug{PipelineMode: m, Stdout: os.Stdout, Stderr: os.Stderr}
}

func (d *Debug) Run() error {
	var errs error

	commands := []*exec.Cmd{
		exec.Command("grype", "version"),
		exec.Command("semgrep", "--version"),
		exec.Command("gitleaks", "version"),
		exec.Command("docker", "version"),
	}

	for _, cmd := range commands {
		slog.Info("run", "command", cmd.String())
		cmd.Stderr = d.Stderr
		cmd.Stdout = d.Stdout

		if d.PipelineMode == ModeRun {
			err := cmd.Run()
			errs = errors.Join(err)
		}
	}

	return errs
}
