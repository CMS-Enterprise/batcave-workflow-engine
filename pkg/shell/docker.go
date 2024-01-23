package shell

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

type ImageBuildOptions struct {
	buildFuncs   []func(*Executable)
	buildDirFunc func(*Executable)
}

func NewImageBuildOptions() *ImageBuildOptions {
	return &ImageBuildOptions{
		buildFuncs:   make([]func(*Executable), 0),
		buildDirFunc: func(*Executable) {},
	}
}

func (o *ImageBuildOptions) applyTo(e *Executable) {
	for _, f := range o.buildFuncs {
		f(e)
	}
	o.buildDirFunc(e)
}

func (o *ImageBuildOptions) append(f func(e *Executable)) *ImageBuildOptions {
	o.buildFuncs = append(o.buildFuncs, f)
	return o
}

func (o *ImageBuildOptions) WithTag(imageName string) *ImageBuildOptions {
	return o.append(func(e *Executable) {
		e = e.WithArgs("--tag", imageName)
	})

}

func (o *ImageBuildOptions) WithBuildDir(directory string) *ImageBuildOptions {
	return o.append(func(e *Executable) {
		e = e.WithArgs(directory)
	})
}

func (o *ImageBuildOptions) WithBuildFile(filename string) *ImageBuildOptions {
	return o.append(func(e *Executable) {
		e = e.WithArgs("--file", filename)
	})
}

func (o *ImageBuildOptions) WithBuildTarget(target string) *ImageBuildOptions {
	return o.append(func(e *Executable) {
		e = e.WithArgs("--target", target)
	})
}

func (o *ImageBuildOptions) WithBuildPlatform(platform string) *ImageBuildOptions {
	return o.append(func(e *Executable) {
		e = e.WithArgs("--platform", platform)
	})
}

func (o *ImageBuildOptions) WithSquashLayers() *ImageBuildOptions {
	return o.append(func(e *Executable) {
		e = e.WithArgs("--squash-all")
	})
}

func (o *ImageBuildOptions) WithCache(cacheTo string, cacheFrom string) *ImageBuildOptions {
	return o.append(func(e *Executable) {
		if cacheTo != "" {
			e = e.WithArgs("--cache-to", cacheTo)
		}
		if cacheFrom != "" {
			e = e.WithArgs("--cache-from", cacheFrom)
		}
	})
}

func (o *ImageBuildOptions) WithBuildArgs(args [][2]string) *ImageBuildOptions {
	return o.append(func(e *Executable) {
		for _, kv := range args {
			arg := fmt.Sprintf("%s=%s", kv[0], kv[1])
			e = e.WithArgs("--build-arg", arg)
		}

	})
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
	e := p.initCmd().WithArgs("build")
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
