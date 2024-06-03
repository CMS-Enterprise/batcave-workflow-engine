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

type TaskType string

var GrypeTaskType TaskType = "Grype Image Scan Task"

type taskOptions struct {
	ImageName     string
	SBOMFilename  string
	GrypeFilename string
	ArtifactDir   string
}

func WithImageName(imageName string) taskOptionFunc {
	return func(o *taskOptions) {
		o.ImageName = imageName
	}
}

func WithOptions(imageName string, sbomFilename string, grypeFilename string, artifactDir string) taskOptionFunc {
	return func(o *taskOptions) {
		o.ImageName = imageName
		o.SBOMFilename = sbomFilename
		o.GrypeFilename = grypeFilename
		o.ArtifactDir = artifactDir
	}
}

func withDefaultImageScan() taskOptionFunc {
	return func(o *taskOptions) {
		o.ImageName = "my-image:latest"
		o.SBOMFilename = "sbom.syft.json"
		o.GrypeFilename = "image-scan.grype.json"
		o.ArtifactDir = "artifacts"
	}
}

type taskOptionFunc func(*taskOptions)

type ImageScanTask interface {
	Start(ctx context.Context) error
	Stream(stderr io.Writer) error
	Apply(...taskOptionFunc)
}

func NewImageScanTask(t TaskType) ImageScanTask {
	switch t {
	case TaskType(GrypeTaskType):
		return new(GrypeImageScanTask)
	default:
		panic("Unsupported image scan type")

	}
}

type GrypeImageScanTask struct {
	opts        *taskOptions
	syftCmd     *exec.Cmd
	grypeCmd    *exec.Cmd
	syftStderr  io.ReadCloser
	grypeStderr io.ReadCloser
	logger      *slog.Logger
}

func (t *GrypeImageScanTask) Apply(opts ...taskOptionFunc) {
	o := new(taskOptions)
	withDefaultImageScan()(o)
	for _, optFunc := range opts {
		optFunc(o)
	}
}

func (t *GrypeImageScanTask) preStart() error {
	type option struct {
		name  string
		value string
	}
	opts := []option{
		{name: "image name", value: t.opts.ImageName},
		{name: "sbom filename", value: t.opts.SBOMFilename},
		{name: "grype filename", value: t.opts.GrypeFilename},
		{name: "artifact directory", value: t.opts.ArtifactDir},
	}

	var errs error

	for _, opt := range opts {
		if strings.EqualFold(opt.value, "") {
			err := fmt.Errorf("Grype image scan task pre-start error: %s is required", opt.name)
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

func (t *GrypeImageScanTask) Start(ctx context.Context) error {
	var err error

	t.logger = slog.Default().With("task_name", "grype image scan")

	t.logger.Debug("pre-start")
	err = t.preStart()
	if err != nil {
		return err
	}
	t.logger.Info("start task")

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

	t.syftCmd = exec.CommandContext(ctx, "syft", syftArgs...)

	t.syftStderr, err = t.syftCmd.StderrPipe()
	if err != nil {
		t.logger.Error("stderr pipe failure", "command", t.syftCmd.String())
		return err
	}

	t.logger.Info("start", "command", t.syftCmd.String())
	err = t.syftCmd.Start()
	if err != nil {
		t.logger.Error("start failure", "command", t.syftCmd.String())
		return err
	}

	grypeArgs := []string{
		fmt.Sprintf(
			"sbom:%s", fullSBOMPath),
		"-o",
		fmt.Sprintf("json=%s", fullGrypePath),
		"-vv",
	}

	// Start after syft is done in the Stream Function
	t.grypeCmd = exec.CommandContext(ctx, "grype", grypeArgs...)
	t.logger.Info("delayed start, wait for syft", "command", t.grypeCmd.String())

	t.grypeStderr, err = t.grypeCmd.StderrPipe()
	if err != nil {
		t.logger.Error("command stderr pipe failure", "command", t.grypeCmd.String())
		return err
	}

	return nil
}

func (t *GrypeImageScanTask) Stream(stderr io.Writer) error {
	var err error
	if t.logger == nil {
		return errors.New("task has not been started")
	}

	t.logger.Info("start stderr stream", "command", t.syftCmd.String())

	_, err = io.Copy(stderr, t.syftStderr)
	if err != nil {
		t.logger.Error("stderr write failure", "command", t.syftCmd.String())
		return err
	}

	err = t.syftCmd.Wait()
	if err != nil {
		t.logger.Error("run failure", "command", t.syftCmd.String())
		return err
	}

	err = t.grypeCmd.Start()
	if err != nil {
		t.logger.Error("run failure", "command", t.grypeCmd.String())
		return err
	}

	t.logger.Info("start stderr stream", "command", t.grypeCmd.String())
	_, err = io.Copy(stderr, t.grypeStderr)
	if err != nil {
		t.logger.Error("stderr write failure", "command", t.grypeCmd.String())
		return err
	}

	err = t.grypeCmd.Wait()
	if err != nil {
		t.logger.Error("run failure", "command", t.grypeCmd.String())
		return err
	}

	return nil
}
