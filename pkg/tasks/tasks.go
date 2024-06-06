package tasks

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dustin/go-humanize"
)

type TaskType string

var (
	GrypeTaskType TaskType = "Grype Image Scan Task"
	ClamTaskType  TaskType = "Clam Antivirus Scan Task"
)

type taskOptions struct {
	ImageName            string
	SBOMFilename         string
	GrypeFilename        string
	FreshclamDisabled    bool
	ClamFilename         string
	ClamscanTarget       string
	ArtifactDir          string
	AntivirusScanEnabled bool
	DisplayStdout        io.Writer
	ImageTarFilename     string
	ImageTarCleanup      bool
	ImageSavePull        bool
}

func WithImageName(imageName string) taskOptionFunc {
	return func(o *taskOptions) {
		o.ImageName = imageName
	}
}

func WithStdout(w io.Writer) taskOptionFunc {
	return func(o *taskOptions) {
		o.DisplayStdout = w
	}
}

func WithImageOptions(imageName string, sbomFilename string, grypeFilename string, artifactDir string) taskOptionFunc {
	return func(o *taskOptions) {
		o.ImageName = imageName
		o.SBOMFilename = sbomFilename
		o.GrypeFilename = grypeFilename
		o.ArtifactDir = artifactDir
	}
}

func WithClamOptions(clamFilename string, clamscanTarget string, artifactDir string, freshclamDisabled bool) taskOptionFunc {
	return func(o *taskOptions) {
		o.ClamFilename = clamFilename
		o.ClamscanTarget = clamscanTarget
		o.ArtifactDir = artifactDir
		o.FreshclamDisabled = freshclamDisabled
	}
}

func WithImageSaveOptions(imageName string, imageTarFilename string, pull bool) taskOptionFunc {
	return func(o *taskOptions) {
		o.ImageName = imageName
		o.ImageTarFilename = imageTarFilename
		o.ImageSavePull = pull
	}
}

func newDefaultTaskOpts() *taskOptions {
	return &taskOptions{
		ImageName:         "my-image:latest",
		SBOMFilename:      "sbom.syft.json",
		GrypeFilename:     "image-scan.grype.json",
		ClamFilename:      "antivirus-scan.clamav.txt",
		ClamscanTarget:    "",
		ArtifactDir:       "artifacts",
		DisplayStdout:     os.Stdout,
		ImageTarFilename:  "image.tar",
		ImageTarCleanup:   true,
		ImageSavePull:     false,
		FreshclamDisabled: true,
	}
}

type taskOptionFunc func(*taskOptions)

func CopyWithPrefix(dst io.Writer, src io.Reader, prefix string) error {
	scanner := bufio.NewScanner(src)

	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintf(dst, "[%s] %s\n", prefix, line)
	}

	return nil
}

func StreamStdout(cmd *exec.Cmd, dstWriter io.Writer, prefix string) error {
	slog.Info("run", "command", cmd.String())

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdoutPipe)

	for scanner.Scan() {
		line := scanner.Text()
		template := "[%s] %s\n"
		if prefix == "" {
			template = "%s\n"
		}
		fmt.Fprintf(dstWriter, template, prefix, line)
	}

	return cmd.Wait()

}

func StreamStderr(cmd *exec.Cmd, dstWriter io.Writer, prefix string) error {
	slog.Info("run", "command", cmd.String())

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(stderrPipe)

	for scanner.Scan() {
		line := scanner.Text()
		template := "[%s] %s\n"
		if prefix == "" {
			template = "%s\n"
		}
		fmt.Fprintf(dstWriter, template, prefix, line)
	}

	return cmd.Wait()
}

type SizeMonitorWriter struct {
	prefix   string
	name     string
	writer   io.Writer
	buf      *bytes.Buffer
	Interval time.Duration
	mu       sync.Mutex
	n        atomic.Uint64
}

func NewSizeMonitorWriter(prefix string, name string, dst io.Writer) *SizeMonitorWriter {
	return &SizeMonitorWriter{
		prefix:   prefix,
		name:     name,
		writer:   dst,
		buf:      new(bytes.Buffer),
		Interval: time.Second,
		n:        atomic.Uint64{},
	}
}

func (w *SizeMonitorWriter) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			w.update()
			return
		case <-time.After(w.Interval):
			w.update()
		}
	}
}

func (w *SizeMonitorWriter) update() {
	size := w.n.Load()

	fmt.Fprintf(w.writer, "[%s] File %s: %s written\n", w.prefix, w.name, humanize.Bytes(size))
}

func (w *SizeMonitorWriter) Write(p []byte) (int, error) {
	n := len(p)
	w.n.Add(uint64(n))
	return n, nil
}
