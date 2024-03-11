# Developer Guide

TODO: Project info, goals, etc.

## [Getting Started](./getting_started.md)

## Project Layout

TODO: Add the philosophy behind the project layout

### Shell

The Shell package (`pkg/shell`) is a library of commands and utilities used in workflow engine.
The standard library way to execute shell commands is by using the `os/exec` package which has a lot of features and
flexibility.
In our case, we want to restrict the ability to arbitrarily execute shell commands by carefully selecting a sub-set of 
features for each command.

For example, if you look at the Syft CLI reference, you'll see dozens of commands and configuration options.
This is all controlled by flag parsing the string of the command.
This is an opinionated security pipeline, so we don't need all the features Syft provides.
The user shouldn't care that we're using Syft to generate an SBOM which is then scanned by Grype for vulnerabilities.
The idea of Workflow Engine is that it's all abstracted to the Security Analysis pipeline.

In the Shell package, all necessary commands will be abstracted into a native Go object.
Only the used features for the given command will be written into this package.

The shell.Executable wraps the exec.Cmd struct and adds some convenient methods for building a command.

```shell
syft version -o json
```

How to execute regular std lib commands with `exec.Cmd`

```go
cmd := exec.Command("syft", "version", "-o","json")
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr
// some other options
err := cmd.Run()
```

There's also additional logic with the `os.exec` standard library command.
Since workflow engine is built around executing external binaries, there is an internal library called the `pkg/shell`
used to abstract a lot of the complexities involved with handling async patterns, possible interrupts, and parameters.

Commands can be represented as functions.

```go
func SyftVersion(options ...OptionFunc) error {
	o := newOptions(options...)
	cmd := exec.Command("syft", "version")
	return run(cmd, o)
}
```

The `OptionFunc` variadic parameter allows the caller to modify the behavior of the command with an arbitrary
number of `OptionFunc`(s).

`newOptions` generates the default `Options` structure and then applies all of passed in functions.
The `o` variable can now be used to apply parameters to the command before execution. 

Returning the `run` function handles off the execution phase of the command to another function which bootstraps
a lot of useful functionality without needing to write supported code for each new command.

For example, if you only want to output what the command would run but not actually run the command, 
```go
dryRun := false
SyftVersion(WithStdout(os.Stdout), WithDryRun(dryRun))
```

This would log the final output command without executing.

The motivation behind this architecture is to simply the Methods for all sub-commands on an executable.

Implementing a new sub command is trivial, just write a new function with the same pattern

```go
func SyftHelp(options ...OptionFunc) error {
	o := newOptions(options...)
	cmd := exec.Command("syft", "--help")
	return run(cmd, o)
}
```

If we wanted to build an optionFunc for version to optionally write JSON instead of plain text, it would go in the
`pkg/shell/shell.go` function.

Since there aren't many commands, they all share the same configuration object `Options`.

```go
func WithJSONOutput(enabled bool) OptionFunc {
	return func(o *Options) {
		o.JSONOutput = true
	}
}
```

Now, the version function can reference this field and change the shell command

```go
func SyftVersion(options ...OptionFunc) error {
	o := newOptions(options...)
	cmd := exec.Command("syft", "version")
  if o.JSONOutput {
    cmd = exec.Command("syft", "version", "-o", "json")
  }
	return run(cmd, o)
}
```

See `pkg/shell/docker.go` for a more complex example of a command with a lot of parameters.

### Pipelines

## Concepts

### Concurrency

[Workflow Engine PR #26](https://github.com/CMS-Enterprise/batcave-workflow-engine/pull/26)

This PR contains a detailed explanation of the concurrency pattern used in the pipeline definitions.

### Documentation

## Too Long; Might Read (TL;MR)

A collection of thoughts around design decisions made in Workflow Engine, mostly ramblings that some people may or may 
not find useful.

### Why CI/CD Flexible Configuration is Painful

In a traditional CI/CD environment, you would have to parse strings to build the exact command you want to execute.

Local Shell:
```bash
syft version
```

GitLab CI/CD Configuration let's use declare the execution environment by providing an image name
```yaml
syft-version:
  stage: scan
  image: anchore/syft:latest
  script:
    - syft version
```

What typically happens is configuration creep.
If you need to print the version information in JSON, (one of the many command options), you would have to provide 
multiple options in GitLab, only changing the script block, hiding each on behind an env variable

```yaml
.syft:
  stage: scan
  image: anchore/syft:latest

syft-version:text:
  extends: .syft
  script:
    - syft version
  rules:
    - if: $SYFT_VERSION_JSON != "true"

syft-version:json:
  extends: .syft
  script:
    - syft version -o json
  rules:
    - if: $SYFT_VERSION_JSON == "true"

```

The complexity increase exponentially in a GitLab CI/CD file for each configuration option you wish to support.
