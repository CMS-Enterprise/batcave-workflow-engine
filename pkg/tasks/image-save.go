package tasks

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
)

type ImageSaveTask interface {
	Run(context.Context, io.Writer) error
}

func NewImageSaveTask(t TaskType, opts ...taskOptionFunc) ImageSaveTask {
	o := newDefaultTaskOpts()
	for _, optFunc := range opts {
		optFunc(o)
	}
	switch t {
	case TaskType(DockerSaveType), TaskType(PodmanSaveType):
		task := new(GenericImageSaveTask)
		task.opts = o
		return task
	default:
		panic("Unsupported image save task type")
	}
}

type GenericImageSaveTask struct {
	opts *taskOptions
}

func (t *GenericImageSaveTask) Run(ctx context.Context, stderr io.Writer) error {

	var imageSaveCmd *exec.Cmd
	switch t.opts.ImageCLIInterface {
	case "docker":
		imageSaveCmd = exec.Command("docker", "save", t.opts.ImageName)
	case "podman":
		imageSaveCmd = exec.Command("podman", "save", t.opts.ImageName)
	default:
		return errors.New("image CLI interface is required")
	}

	imageSaveFile, err := os.OpenFile(t.opts.ImageTarFilename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	imageSaveCmd.Stdout = imageSaveFile

	return Stream(imageSaveCmd, stderr, "image save")
}
