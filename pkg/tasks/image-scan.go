package tasks

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

// ImageVulScanTask generates an SBOM and performs a vulnerability scan
//
// This can be used as an abstraction, independent of the underlying
// tool conducting the generation or scan.
type ImageVulScanTask interface {
	Run(context.Context, io.Writer) error
}

// NewImageVulScanTask create a new Image Scan Task of a specific type
func NewImageVulScanTask(t TaskType, opts ...taskOptionFunc) ImageVulScanTask {
	o := newDefaultTaskOpts()
	for _, optFunc := range opts {
		optFunc(o)
	}

	switch t {
	case TaskType(GrypeTaskType):
		task := new(GrypeImageVulScanTask)
		task.opts = o
		return task
	default:
		panic("Unsupported image scan type")
	}
}

// GrypeImageVulScanTask uses Syft for SBOM generation and Grype for vulnerability scan
type GrypeImageVulScanTask struct {
	opts *taskOptions
}

func (t *GrypeImageVulScanTask) preRun() error {
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

func (t *GrypeImageVulScanTask) Run(ctx context.Context, stderr io.Writer) error {
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

type AntivirusScanTask interface {
	Run(context.Context, io.Writer) error
}

func NewAntivirusScanTask(t TaskType, opts ...taskOptionFunc) AntivirusScanTask {
	o := newDefaultTaskOpts()
	for _, optFunc := range opts {
		optFunc(o)
	}

	switch t {
	case TaskType(ClamTaskType):
		task := new(ClamAntivirusScanTask)
		task.opts = o
		return task
	default:
		panic("Unsupported antivirus scan type")
	}
}

type ClamAntivirusScanTask struct {
	opts           *taskOptions
	clamReportFile *os.File
}

func (t *ClamAntivirusScanTask) preRun() error {
	var err, errs error

	if strings.EqualFold(t.opts.ClamFilename, "") {
		err := fmt.Errorf("Clam Antivirus scan task -> filename is required")
		errs = errors.Join(errs, err)
	}
	if strings.EqualFold(t.opts.ArtifactDir, "") {
		err := fmt.Errorf("Clam Antivirus scan task -> artifact directory is required")
		errs = errors.Join(errs, err)
	}
	if strings.EqualFold(t.opts.ClamscanTarget, "") {
		err := fmt.Errorf("Clam Antivirus scan task -> target is required")
		errs = errors.Join(errs, err)
	}

	if errs != nil {
		return errs
	}

	reportFilename := path.Join(t.opts.ArtifactDir, t.opts.ClamFilename)
	err = os.MkdirAll(t.opts.ArtifactDir, 0o777)
	if err != nil {
		return err
	}
	t.clamReportFile, err = os.Create(reportFilename)
	if err != nil {
		return err
	}

	return nil
}

func (t *ClamAntivirusScanTask) Run(ctx context.Context, stderr io.Writer) error {
	err := t.preRun()
	if err != nil {
		return err
	}

	err = t.runFreshclam(ctx, stderr)

	if err != nil {
		return err
	}

	clamscanArgs := []string{
		"--infected",
		"--recursive",
		"--scan-archive=yes",
		"--max-filesize=4095M", // files larger will be skipped and assumed clean
		"--max-scansize=4095M", // will only scan this much from each file
		"--stdout",
		t.opts.ClamscanTarget,
	}

	buf := new(bytes.Buffer)
	mw := io.MultiWriter(buf, t.clamReportFile)
	clamscanCmd := exec.Command("clamscan", clamscanArgs...)
	clamscanCmd.Stdout = mw

	// verbose output is really noisey so just count the lines and report progress
	err = StreamElapsed(clamscanCmd, stderr, time.Second*3, "clamscan")
	if err != nil {
		return err
	}

	// print the report to display stdout
	_, err = buf.WriteTo(t.opts.DisplayStdout)
	return err
}

func (t *ClamAntivirusScanTask) runFreshclam(ctx context.Context, stderr io.Writer) error {
	if t.opts.FreshclamDisabled {
		return nil
	}

	freshclamCmd := exec.Command("freshclam")

	return StreamStderr(freshclamCmd, stderr, "freshclam")
}
