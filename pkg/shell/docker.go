package shell

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

// ImageBuildOptions implements the builder pattern to handle
// construction of the `docker build` command. Support flags or arguments
// that are set to the default values are not included in the constructed command.
//
// Example:
//
//	If you want to generate a build command with possibly the existence of a build target,
//	`NewImageBuildOptions().WithBuildTarget("")`
//	generates:
//	`docker build .`
//
//	`NewImageBuildOptions().WithBuildTarget("debug")`
//	generates:
//	`docker build --target debug .`
type ImageBuildOptions struct {
	args     []string
	buildDir string
}

// NewImageBuildOptions inits the image build options to the default values
func NewImageBuildOptions() *ImageBuildOptions {
	return &ImageBuildOptions{
		args:     make([]string, 0),
		buildDir: "",
	}
}

// applyTo sets the arguments to an executable
func (o *ImageBuildOptions) applyTo(e *Executable) {
	allArgs := []string{"build"}
	allArgs = append(allArgs, o.args...)
	allArgs = append(allArgs, o.buildDir)
	e = e.WithArgs(allArgs...)
}

// WithTag sets the docker image name and tag ex. "alpine:latest"
func (o *ImageBuildOptions) WithTag(imageName string) *ImageBuildOptions {
	if imageName != "" {
		o.args = append(o.args, "--tag", imageName)
	}
	return o
}

// WithBuildDir sets the build context
func (o *ImageBuildOptions) WithBuildDir(directory string) *ImageBuildOptions {
	o.buildDir = directory
	return o
}

// WithBuildFile sets the target Dockerfile flag
func (o *ImageBuildOptions) WithBuildFile(filename string) *ImageBuildOptions {
	if filename != "" {
		o.args = append(o.args, "--file", filename)
	}
	return o
}

// WithBuildTarget sets a specific build target
func (o *ImageBuildOptions) WithBuildTarget(target string) *ImageBuildOptions {
	if target != "" {
		o.args = append(o.args, "--target", target)
	}
	return o
}

// WithBuildPlatform sets to build to a specific platform ex. "x86_64"
func (o *ImageBuildOptions) WithBuildPlatform(platform string) *ImageBuildOptions {
	if platform != "" {
		o.args = append(o.args, "--platform", platform)
	}
	return o
}

// WithSquashLayers sets the cli option to squash an image to a single layer
func (o *ImageBuildOptions) WithSquashLayers(enabled bool) *ImageBuildOptions {
	if enabled {
		o.args = append(o.args, "--squash-all")
	}

	return o
}

// WithCache sets caching to a registry
func (o *ImageBuildOptions) WithCache(cacheTo string, cacheFrom string) *ImageBuildOptions {

	if cacheTo != "" {
		o.args = append(o.args, "--cache-to", cacheTo)
	}

	if cacheFrom != "" {
		o.args = append(o.args, "--cache-from", cacheFrom)
	}

	return o
}

// WithBuildArgs defines specific build arguments for build time
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
func PodmanCommand(stdin io.Reader, stdout io.Writer, stderr io.Writer) *dockerCLICmd {
	return &dockerCLICmd{
		initCmd: func() *Executable {
			return NewExecutable("podman").WithIO(stdin, stdout, stderr)
		},
	}
}

// DockerCommand with custom stdout and stderr
func DockerCommand(stdin io.Reader, stdout io.Writer, stderr io.Writer) *dockerCLICmd {
	return &dockerCLICmd{
		initCmd: func() *Executable {
			return NewExecutable("docker").WithIO(stdin, stdout, stderr)
		},
	}
}

func ExampleDockerCommand() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
	buildOpts := NewImageBuildOptions().WithBuildDir("./some-dir").WithBuildFile("Dockerfile-custom")
	DockerCommand(nil, os.Stdout, os.Stderr).Build(buildOpts).WithDryRun(true).RunLogError()
}
