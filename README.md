# Workflow Engine

Workflow Engine is a security and delivery pipeline designed to orchestrate the process of building and scanning an
application image for security vulnerabilities.
It solves the problem of having to configure a hardened-predefined security pipeline using traditional CI/CD.
Workflow Engine can be statically compiled as a binary and run on virtually any platform, CI/CD
environment, or locally.

## Getting Started

Install Prerequisites:

- Container Engine
- Docker or Podman CLI
- Golang >= v1.21.5
- Just (optional)

## Compiling Workflow Engine

Running the just recipe will put the compiled-binary into `./bin`

```bash
just build
```

OR compile manually

```bash
git clone <this-repo> <target-dir>
cd <target-dir>
go build -o bin/workflow-engine ./cmd/workflow-engine
```

## Running A Pipeline

You can run the executable directory

```bash
dagger run workflow-engine --pipeline debug
```

## Configuring a Pipeline

| Variable Name           | Type   | Default Value | Description                                          |
| ----------------------- | ------ | ------------- | ---------------------------------------------------- |
| WFE_BUILD_DIR           | string |               | The container build directory                        |
| WFE_BUILD_DOCKERFILE    | string |               | The name of the Dockerfile to build and scan         |
| WFE_BUILD_TAG           | string |               | The container build tag to use for building an image |
|                         |        |               | This is passed to the --tag flag                     |
| WFE_BUILD_PLATFORM      | string |               | The container build platform                         |
|                         |        |               | This is passed to the --platform flag                |
| WFE_BUILD_TARGET        | string |               | The container build target                           |
|                         |        |               | This is passed to the --targe flag                   |
| WFE_BUILD_CACHE_TO      | string |               | The container cache to directory                     |
|                         |        |               | This is passed to the --cache-to flag                |
| WFE_BUILD_CACHE_FROM    | string |               | The container cache from directory                   |
|                         |        |               | This is passed to the --cache-from flag              |
| WFE_BUILD_SQUASH_LAYERS | bool   |               | Flag to squash layers                                |
|                         |        |               | Setting this to true enables the --squash-all flag   |
| WFE_ARTIFACT_DIRECTORY  | string |               | The directory to store artifacts                     |
| WFE_SBOM_FILENAME       | string |               | The SBOM file name                                   |
| WFE_GRYPE_FILENAME      | string |               | The Grype file name                                  |
| WFE_SCAN_IMAGE_TARGET   | string |               | The scan image tag name                              |
