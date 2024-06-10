package tasks

import (
	"context"
	"fmt"
	"io"
	"os/exec"
)

type GenericImagePushTask struct {
	TagName string
	cmdName string
}

func NewGenericImagePushTask(cliInterface string, tagName string) *GenericImagePushTask {
	return &GenericImagePushTask{
		TagName: tagName,
		cmdName: cliInterface,
	}
}

func (t *GenericImagePushTask) Run(ctx context.Context, stderr io.Writer) error {
	pushCmd := exec.Command(t.cmdName, "push")
	return StreamStderr(pushCmd, stderr, fmt.Sprintf("%s push", t.cmdName))
}
