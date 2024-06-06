package tasks

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
)

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
		err := fmt.Errorf("Clam Antivirus scan task pre-start error -> filename is required")
		errs = errors.Join(errs, err)
	}
	if strings.EqualFold(t.opts.ArtifactDir, "") {
		err := fmt.Errorf("Clam Antivirus scan task pre-start error -> artifact directory is required")
		errs = errors.Join(errs, err)
	}
	if strings.EqualFold(t.opts.ClamscanTarget, "") {
		err := fmt.Errorf("Clam Antivirus scan task pre-start error -> target is required")
		errs = errors.Join(errs, err)
	}

	if errs != nil {
		return errs
	}

	reportFilename := path.Join(t.opts.ArtifactDir, t.opts.ClamFilename)
	err = os.MkdirAll(t.opts.ArtifactDir, 0777)
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
		"--verbose",
		"--archive-verbose",
		"--scan-archive=yes",
		"--max-filesize=1000M", // files larger will be skipped and assumed clean
		"--max-scansize=1000M", // will only scan this much from each file
		"--stdout",
		t.opts.ClamscanTarget,
	}

	buf := new(bytes.Buffer)
	mw := io.MultiWriter(buf, t.clamReportFile)
	clamscanCmd := exec.Command("clamscan", clamscanArgs...)
	clamscanCmd.Stdout = mw

	err = StreamStderr(clamscanCmd, stderr, "clamscan")
	if err != nil {
		return err
	}
	_, err = buf.WriteTo(stderr)
	return err
}

func (t *ClamAntivirusScanTask) runFreshclam(ctx context.Context, stderr io.Writer) error {

	if t.opts.FreshclamDisabled {
		return nil
	}

	freshclamCmd := exec.Command("freshclam")

	return StreamStderr(freshclamCmd, stderr, "freshclam")
}
