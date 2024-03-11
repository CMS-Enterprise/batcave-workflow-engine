package shell

import (
	"fmt"
	"log/slog"
	"os/exec"
)

type DockerAlias int8

const (
	DockerAliasDocker DockerAlias = 0
	DockerAliasPodman DockerAlias = 1
)

// Save an image to a tar archive
//
// Requirements:
//   - optional WithDockerAlias option, defaults to DockerAliasDocker
//   - WithImage option
//
// Outputs: image tar archive to STDOUT
func DockerSave(options ...OptionFunc) error {
	o := newOptions(options...)
	switch o.dockerAlias {
	case DockerAliasDocker:
		cmd := exec.Command("docker", "save", o.imageTag)
		return run(cmd, o)
	case DockerAliasPodman:
		cmd := exec.Command("podman", "save", o.imageTag)
		return run(cmd, o)
	default:
		return fmt.Errorf("only docker/podman aliases are supported: %w", ErrBadParameters)
	}
}

// DockerBuild a container image, supports CLI aliases to docker build (ex. podman build)
//
// Requirements: WithImageBuildOptions, optional WithDockerAlias
//
// Outputs: debug to STDERR
func DockerBuild(optionFuncs ...OptionFunc) error {
	o := newOptions(optionFuncs...)
	// this parses argument values to determine the flags
	args := o.imageBuildOptions.args()
	switch o.dockerAlias {
	case DockerAliasDocker:
		cmd := exec.Command("docker", args...)
		return run(cmd, o)
	case DockerAliasPodman:
		cmd := exec.Command("podman", args...)
		return run(cmd, o)
	default:
		return fmt.Errorf("only docker/podman aliases are supported: %w", ErrBadParameters)
	}
}

// DockerPush a container image, supports CLI aliases to docker build (ex. podman build)
//
// Requirements: WithImageName
//
// Outputs: debug to STDERR
func DockerPush(optionFuncs ...OptionFunc) error {
	o := newOptions(optionFuncs...)
	// this parses argument values to determine the flags
	switch o.dockerAlias {
	case DockerAliasDocker:
		cmd := exec.Command("docker", "push", o.imageTag)
		return run(cmd, o)
	case DockerAliasPodman:
		cmd := exec.Command("podman", "push", o.imageTag)
		return run(cmd, o)
	default:
		return fmt.Errorf("only docker/podman aliases are supported: %w", ErrBadParameters)
	}
}

// DockerInfo print system configuration information
//
// Requirements: optional WithDockerAlias
//
// Outputs: debug to STDERR
func DockerInfo(optionFuncs ...OptionFunc) error {
	o := newOptions(optionFuncs...)
	// this parses argument values to determine the flags
	switch o.dockerAlias {
	case DockerAliasDocker:
		cmd := exec.Command("docker", "info")
		return run(cmd, o)
	case DockerAliasPodman:
		cmd := exec.Command("podman", "info")
		return run(cmd, o)
	default:
		return fmt.Errorf("only docker/podman aliases are supported: %w", ErrBadParameters)
	}
}

// ImageBuildOptions are specific to Docker builds
type ImageBuildOptions struct {
	Tag          string
	BuildDir     string
	Dockerfile   string
	Target       string
	Platform     string
	SquashLayers bool
	CacheTo      string
	CacheFrom    string
	BuildArgs    []string
}

func (o ImageBuildOptions) args() []string {
	args := []string{"build"}

	flags := map[string]string{
		"--file":       o.Dockerfile,
		"--target":     o.Target,
		"--platform":   o.Platform,
		"--cache-to":   o.CacheTo,
		"--cache-from": o.CacheFrom,
		"--tag":        o.Tag,
	}

	// append to the args list if not ""
	for flag, value := range flags {
		if value != "" {
			args = append(args, flag, value)
		}
	}

	// Special case for build arguments
	for _, arg := range o.BuildArgs {
		slog.Debug("docker build argument", "arg", arg)
		if arg != "" {
			args = append(args, "--build-arg", arg)
		}
	}

	// Squash is a bool flag
	if o.SquashLayers {
		args = append(args, "--squash-all")
	}

	args = append(args, o.BuildDir)

	return args
}
