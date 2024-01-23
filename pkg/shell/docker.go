package shell

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

type ImageBuildOptions struct {
	args     []string
	buildDir string
}

func NewImageBuildOptions() *ImageBuildOptions {
	return &ImageBuildOptions{
		args:     make([]string, 0),
		buildDir: "",
	}
}

func (o *ImageBuildOptions) applyTo(e *Executable) {
	allArgs := []string{"build"}
	allArgs = append(allArgs, o.args...)
	allArgs = append(allArgs, o.buildDir)
	e = e.WithArgs(allArgs...)
}

func (o *ImageBuildOptions) WithTag(imageName string) *ImageBuildOptions {
	if imageName != "" {
		o.args = append(o.args, "--tag", imageName)
	}
	return o
}

func (o *ImageBuildOptions) WithBuildDir(directory string) *ImageBuildOptions {
	o.buildDir = directory
	return o
}

func (o *ImageBuildOptions) WithBuildFile(filename string) *ImageBuildOptions {
	if filename != "" {
		o.args = append(o.args, "--file", filename)
	}
	return o
}

func (o *ImageBuildOptions) WithBuildTarget(target string) *ImageBuildOptions {
	if target != "" {
		o.args = append(o.args, "--target", target)
	}
	return o
}

func (o *ImageBuildOptions) WithBuildPlatform(platform string) *ImageBuildOptions {
	if platform != "" {
		o.args = append(o.args, "--platform", platform)
	}
	return o
}

func (o *ImageBuildOptions) WithSquashLayers(enabled bool) *ImageBuildOptions {
	if enabled {
		o.args = append(o.args, "--squash-all")
	}

	return o
}

func (o *ImageBuildOptions) WithCache(cacheTo string, cacheFrom string) *ImageBuildOptions {

	if cacheTo != "" {
		o.args = append(o.args, "--cache-to", cacheTo)
	}

	if cacheFrom != "" {
		o.args = append(o.args, "--cache-from", cacheFrom)
	}

	return o
}

func (o *ImageBuildOptions) WithBuildArgs(args [][2]string) *ImageBuildOptions {
	for _, kv := range args {
		arg := fmt.Sprintf("%s=%s", kv[0], kv[1])
		o.args = append(o.args, "--build-arg", arg)
	}
	return o
}

type dockerCLICmd struct {
	initCmd func() *Executable
}

// Version outputs the version of the CLI
//
// shell: `[docker|podman] version`
func (p *dockerCLICmd) Version() *Command {
	return NewCommand(p.initCmd().WithArgs("version"))
}

// Build will construct the docker build command with optional values.
//
// The caller is responsible for validation, strings will be directly inserted into the shell command
// shell: [docker|podman] build [args] <build directory>
func (p *dockerCLICmd) Build(opts *ImageBuildOptions) *Command {
	e := p.initCmd()
	opts.applyTo(e)
	return NewCommand(e)
}

// Push will push an image to a registry
//
// shell: [docker|podman] push <image>
func (p *dockerCLICmd) Push(imageName string) *Command {
	e := p.initCmd().WithArgs("push").WithArgs(imageName)
	return NewCommand(e)
}

// Info tests the connection to the container runtime daemon
//
// shell: `[docker|podman] info`
func (p *dockerCLICmd) Info() *Command {
	return NewCommand(p.initCmd().WithArgs("info"))
}

// PodmanCommand with custom stdout and stderr
func PodmanCommand(stdout io.Writer, stderr io.Writer) *dockerCLICmd {
	return &dockerCLICmd{
		initCmd: func() *Executable {
			return NewExecutable("podman").WithStdout(stdout).WithStderr(stderr)
		},
	}
}

// DockerCommand with custom stdout and stderr
func DockerCommand(stdout io.Writer, stderr io.Writer) *dockerCLICmd {
	return &dockerCLICmd{
		initCmd: func() *Executable {
			return NewExecutable("docker").WithStdout(stdout).WithStderr(stderr)
		},
	}
}

func ExampleDockerCommand() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
	buildOpts := NewImageBuildOptions().WithBuildDir("./some-dir").WithBuildFile("Dockerfile-custom")
	DockerCommand(os.Stdout, os.Stderr).Build(buildOpts).WithDryRun(true).RunLogError()
}
