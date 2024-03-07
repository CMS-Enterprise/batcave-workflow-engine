package shell

import "os/exec"

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
func DockerSave(optionFuncs ...OptionFunc) ExitCode {
	o := newOptions(optionFuncs...)
	switch o.dockerAlias {
	case DockerAliasDocker:
		cmd := exec.Command("docker", "save", o.imageName)
		return run(cmd, o)
	case DockerAliasPodman:
		cmd := exec.Command("podman", "save", o.imageName)
		return run(cmd, o)
	default:
		return ExitBadConfiguration
	}
}
