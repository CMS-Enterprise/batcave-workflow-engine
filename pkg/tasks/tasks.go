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
	DisplayStdout        io.Writer
	ImageName            string
	SBOMFilename         string
	GrypeFilename        string
	GitleaksFilename     string
	GitleaksTargetDir    string
	SemgrepFilename      string
	SemgrepRules         string
	SemgrepExperimental  bool
	FreshclamDisabled    bool
	ClamFilename         string
	ClamscanTarget       string
	ArtifactDir          string
	AntivirusScanEnabled bool
	ImageTarFilename     string
	ImageTarCleanup      bool
	ImageSavePull        bool
	ImageBuildBakefile   string
	ImageBuildBakeTarget string
	ImageBuildOpts       ImageBuildOptions
}

type ImageBuildOptions struct {
	Context      string
	Dockerfile   string
	Platform     string
	Target       string
	CacheTo      string
	CacheFrom    string
	SquashLayers bool
	BuildArgs    string
}

func WithImageBuildOptions(buildOpts ImageBuildOptions) taskOptionFunc {
	return func(taskOptions *taskOptions) {
		taskOptions.ImageBuildOpts = buildOpts
	}
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

func WithImgVulOptions(imageName string, sbomFilename string, grypeFilename string, artifactDir string) taskOptionFunc {
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
		GitleaksFilename:  "secrets-detection.gitleaks.json",
		SemgrepFilename:   "sast.semgrep.json",
		ClamscanTarget:    ".",
		ArtifactDir:       "artifacts",
		DisplayStdout:     os.Stdout,
		ImageTarFilename:  "image.tar",
		ImageTarCleanup:   true,
		ImageSavePull:     false,
		FreshclamDisabled: true,
		ImageBuildOpts:    ImageBuildOptions{},
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

func StreamElapsed(cmd *exec.Cmd, dstWriter io.Writer, interval time.Duration, prefix string) error {
	slog.Info("run", "command", cmd.String())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	start := time.Now()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(interval):
				ts := time.Since(start)
				fmt.Fprintf(dstWriter, "[%s] running... elapsed=%s\n", prefix, ts)
			}
		}
	}()

	return cmd.Run()
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
