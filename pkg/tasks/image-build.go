package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"slices"

	"golang.org/x/exp/maps"
)

type ImageBuildTask interface {
	Run(ctx context.Context, stderrWriter io.Writer) error
}

func NewImageBuildTask(cliInterface string, opts ...taskOptionFunc) ImageBuildTask {
	o := newDefaultTaskOpts()
	for _, optFunc := range opts {
		optFunc(o)
	}
	var task ImageBuildTask

	switch cliInterface {
	case "bake":
		bakeTask := new(BakeImageBuildTask)
		bakeTask.opts = o
		task = bakeTask
	case "docker":
		dockerTask := NewGenericImageBuildTask("docker")
		dockerTask.opts = o
		task = dockerTask
	case "podman":
		podmanTask := NewGenericImageBuildTask("podman")
		podmanTask.opts = o
		task = podmanTask
	default:
		panic("Unsupported image build cli interface, must be docker or podman")
	}
	return task
}

func NewGenericImageBuildTask(cmdString string) *GenericImageBuildTask {
	return &GenericImageBuildTask{
		cmdString: cmdString,
		args:      make([]string, 0),
	}
}

type GenericImageBuildTask struct {
	opts      *taskOptions
	cmdString string
	args      []string
}

func (t *GenericImageBuildTask) preRun() error {
	buildOpts := t.opts.ImageBuildOpts
	buildArgs := map[string]string{}
	if buildOpts.BuildArgs != "" {
		err := json.Unmarshal([]byte(buildOpts.BuildArgs), &buildArgs)
		if err != nil {
			return fmt.Errorf("invalid build args format, must be JSON map as string: %w", err)
		}
	}
	// sort the keys so the order is deterministic
	keys := maps.Keys(buildArgs)
	slices.Sort(keys)

	t.args = []string{"build"}

	for _, key := range keys {
		t.args = append(t.args, "--build-arg", fmt.Sprintf("%s=%s", key, buildArgs[key]))
	}

	if buildOpts.Platform != "" {
		t.args = append(t.args, "--platform", buildOpts.Platform)
	}

	if buildOpts.Target != "" {
		t.args = append(t.args, "--target", buildOpts.Target)
	}

	if buildOpts.CacheTo != "" {
		t.args = append(t.args, "--cache-to", buildOpts.CacheTo)
	}

	if buildOpts.CacheFrom != "" {
		t.args = append(t.args, "--cache-to", buildOpts.CacheFrom)
	}

	if buildOpts.SquashLayers {
		t.args = append(t.args, "--squash-layers")
	}

	if buildOpts.Dockerfile == "" {
		return errors.New("build image Dockerfile required")
	}
	t.args = append(t.args, "--file", buildOpts.Dockerfile)

	if buildOpts.Context == "" {
		return errors.New("build image context / directory required")
	}
	t.args = append(t.args, buildOpts.Context)

	return nil
}

func (t *GenericImageBuildTask) Run(ctx context.Context, stderr io.Writer) error {
	if err := t.preRun(); err != nil {
		return err
	}

	buildCmd := exec.CommandContext(ctx, t.cmdString, t.args...)
	return StreamStderr(buildCmd, stderr, fmt.Sprintf("%s build", t.cmdString))
}

type BakeImageBuildTask struct {
	opts *taskOptions
}

func (t *BakeImageBuildTask) preRun() error {
	if t.opts.ImageBuildBakeTarget == "" {
		return errors.New("image build bake target required")
	}
	if t.opts.ImageBuildBakefile == "" {
		return errors.New("image build bake file required")
	}
	return nil
}

func (t *BakeImageBuildTask) Run(ctx context.Context, stderr io.Writer) error {
	if err := t.preRun(); err != nil {
		return err
	}

	args := []string{
		"buildx",
		"bake",
		"--file",
		t.opts.ImageBuildBakefile,
		t.opts.ImageBuildBakeTarget,
	}
	bakeCmd := exec.Command("docker", args...)

	return StreamStderr(bakeCmd, stderr, "image build")
}
