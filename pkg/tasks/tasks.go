package tasks

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
)

type TaskType string

var (
	GrypeTaskType  TaskType = "Grype Image Scan Task"
	ClamTaskType   TaskType = "Clam Antivirus Scan Task"
	PodmanSaveType TaskType = "Podman Save Task"
	DockerSaveType TaskType = "Docker Save Task"
)

type taskOptions struct {
	ImageName            string
	SBOMFilename         string
	GrypeFilename        string
	ClamFilename         string
	ClamscanTarget       string
	ArtifactDir          string
	AntivirusScanEnabled bool
	DisplayStdout        io.Writer
	ImageTarFilename     string
	ImageTarCleanup      bool
	ImageCLIInterface    string
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

func WithClamOptions(clamFilename string, clamscanTarget string, artifactDir string) taskOptionFunc {
	return func(o *taskOptions) {
		o.ClamFilename = clamFilename
		o.ClamscanTarget = clamscanTarget
		o.ArtifactDir = artifactDir
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
		ImageCLIInterface: "docker",
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

func Stream(cmd *exec.Cmd, stderr io.Writer, prefix string) error {
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
		fmt.Fprintf(stderr, "[%s] %s\n", prefix, line)
	}

	return cmd.Wait()
}
