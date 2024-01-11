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

The shell.Command wraps the exec.Cmd struct and adds some convenient methods for building a command.

Instead of writing something like

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

How to do it with the shell package

```go
cmd := shell.NewCommand("syft").WithArgs("version", "-o", "json").WithStdout(os.Stdout).WithStderr(os.Stderr)
err := cmd.Run()
```
This isn't to save lines of code.
It gives the developer the ability to define a command as syft and write the supported arguments as methods.

The resulting API is a lot neater for both the backend (where to command logic is contained) and the caller of the 
function which could eventually be any number of clients and the eventual pipeline where the command will be called.

```go
err := shell.SyftCommand(os.Stdout,os.Stderr).Version()
```

### Command Structure

#### `Command` Field

This gives custom commands access to the methods already defined in the `exec.Cmd` object, like `Run()` by using 
Go-syle Composition which is similar to inheritance in other languages.

#### `InitCmd` Field

There may be a need in the future to customize the binary that's being executed for some reason.
The `InitCmd` (which is a function that returns a command) is used to initialize the base command, `syft` as an example.
It shouldn't contain arguments since they will be attached when one of the public methods are called, like `Version()`.


### Pipelines

## Concepts

TODO: Additional things to note for developers

### Logging

### Passing Data Between Jobs

### Concurrency

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
