package tasks

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"path"
	"strings"
)

type ImageScanTask interface {
	Run(context.Context, io.Writer) error
}

func NewImageScanTask(t TaskType, opts ...taskOptionFunc) ImageScanTask {
	o := newDefaultTaskOpts()
	for _, optFunc := range opts {
		optFunc(o)
	}

	switch t {
	case TaskType(GrypeTaskType):
		task := new(GrypeImageScanTask)
		task.opts = o
		return task
	default:
		panic("Unsupported image scan type")
	}
}

type GrypeImageScanTask struct {
	opts *taskOptions
}

func (t *GrypeImageScanTask) preRun() error {
	type setting struct {
		name  string
		value string
	}
	settings := []setting{
		{name: "image name", value: t.opts.ImageName},
		{name: "sbom filename", value: t.opts.SBOMFilename},
		{name: "grype filename", value: t.opts.GrypeFilename},
		{name: "artifact directory", value: t.opts.ArtifactDir},
	}

	var errs error

	for _, s := range settings {
		if strings.EqualFold(s.value, "") {
			err := fmt.Errorf("Grype image scan task pre-start error -> %s is required", s.name)
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

func (t *GrypeImageScanTask) Run(ctx context.Context, stderr io.Writer) error {
	var err error

	err = t.preRun()
	if err != nil {
		return err
	}

	slog.Info("start task")

	fullSBOMPath := path.Join(t.opts.ArtifactDir, t.opts.SBOMFilename)
	fullGrypePath := path.Join(t.opts.ArtifactDir, t.opts.GrypeFilename)

	syftArgs := []string{
		"scan",
		t.opts.ImageName,
		"--scope=squashed",
		"-o",
		fmt.Sprintf("syft-json=%s", fullSBOMPath),
		"-vv",
	}

	syftCmd := exec.CommandContext(ctx, "syft", syftArgs...)

	err = StreamStderr(syftCmd, stderr, "syft")
	if err != nil {
		return err
	}

	grypeArgs := []string{
		fmt.Sprintf("sbom:%s", fullSBOMPath),
		"-o",
		fmt.Sprintf("json=%s", fullGrypePath),
		"-vv",
	}

	grypeCmd := exec.CommandContext(ctx, "grype", grypeArgs...)
	slog.Info("run", "command", grypeCmd.String())

	err = StreamStderr(grypeCmd, stderr, "grype")
	if err != nil {
		return err
	}

	gatecheckCmd := exec.CommandContext(ctx, "gatecheck", "ls", "--verbose", "--epss", fullGrypePath)
	gatecheckCmd.Stdout = t.opts.DisplayStdout

	err = StreamStderr(gatecheckCmd, stderr, "gatecheck")
	if err != nil {
		return err
	}

	return nil
}
