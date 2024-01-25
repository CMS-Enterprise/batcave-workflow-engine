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
workflow-engine run debug
```

## Configuring a Pipeline

Configuration Options:

* Configuration via CLI flags
* Environment Variables
* Config File in JSON
* Config File in YAML
* Config File in TOML

Configuration Order-of-Precedence:

1. CLI Flag
2. Environment Variable
3. Config File Value
4. Default Value



| Variable Name             | Type   | Default | CLI Flag             | Config Field Name         | Description                                          |
| ------------------------- | ------ | ------- | -------------------- | ------------------------- | ---------------------------------------------------- |
| `WFE_BUILD_DIR`           | string |         | --build-dir          | `image.buildDir`          | The container build directory                        |
| `WFE_BUILD_DOCKERFILE`    | string |         | --dockerfile         | `image.buildDockerfile`   | The name of the Dockerfile to build and scan         |
| `WFE_BUILD_TAG`           | string |         | --tag                | `image.buildTag`          | The container build tag to use for building an image |
| `WFE_BUILD_PLATFORM`      | string |         | --platform           | `image.buildPlatform`     | The container build platform                         |
| `WFE_BUILD_TARGET`        | string |         | --target             | `image.buildTarget`       | The container build target                           |
| `WFE_BUILD_CACHE_TO`      | string |         | --cache-to           | `image.buildCacheTo`      | The container cache to directory                     |
| `WFE_BUILD_CACHE_FROM`    | string |         | --cache-from         | `image.buildCacheFrom`    | The container cache from directory                   |
| `WFE_BUILD_SQUASH_LAYERS` | bool   |         | --squash-layers      | `image.buildSquashLayers` | Flag to squash layers                                |
| `WFE_SCAN_IMAGE_TARGET`   | string |         | --scan-image-target  | `image.scanTarget`        | The scan image tag name                              |
| `WFE_ARTIFACT_DIRECTORY`  | string |         | --artifact-directory | `artifacts.directory`     | The directory to store artifacts                     |
| `WFE_SBOM_FILENAME`       | string |         | --sbom-filename      | `artifacts.sbomFilename`  | The SBOM file name                                   |
| `WFE_GRYPE_FILENAME`      | string |         | --grype-filename     | `artifacts.grypeFilename` | The Grype file name                                  |
