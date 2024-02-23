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

- Configuration via CLI flags
- Environment Variables
- Config File in JSON
- Config File in YAML
- Config File in TOML

Configuration Order-of-Precedence:

1. CLI Flag
2. Environment Variable
3. Config File Value
4. Default Value

| Variable Name             | Type               | Default              | CLI Flag             | Config Field Name            | Description                                           |
|---------------------------|--------------------|----------------------|----------------------|------------------------------|-------------------------------------------------------|
| `WFE_BUILD_DIR`           | string             | .                    | --build-dir          | `image.buildDir`             | The container build directory                         |
| `WFE_BUILD_DOCKERFILE`    | string             | Dockerfile           | --dockerfile         | `image.buildDockerfile`      | The name of the Dockerfile to build and scan          |
| `WFE_BUILD_ARGS`          | map\[string]string |                      | --build-arg          | `image.buildArgs`            | Build arguments passed to the container build command |
| `WFE_BUILD_TAG`           | string             |                      | --tag                | `image.buildTag`             | The container build tag to use for building an image  |
| `WFE_BUILD_PLATFORM`      | string             |                      | --platform           | `image.buildPlatform`        | The container build platform                          |
| `WFE_BUILD_TARGET`        | string             |                      | --target             | `image.buildTarget`          | The container build target                            |
| `WFE_BUILD_CACHE_TO`      | string             |                      | --cache-to           | `image.buildCacheTo`         | The container cache to directory                      |
| `WFE_BUILD_CACHE_FROM`    | string             |                      | --cache-from         | `image.buildCacheFrom`       | The container cache from directory                    |
| `WFE_BUILD_SQUASH_LAYERS` | bool               |                      | --squash-layers      | `image.buildSquashLayers`    | Flag to squash layers                                 |
| `WFE_SCAN_IMAGE_TARGET`   | string             |                      | --scan-image-target  | `image.scanTarget`           | The scan image tag name                               |
| `WFE_ARTIFACT_DIRECTORY`  | string             |                      | --artifact-directory | `artifacts.directory`        | The directory to store artifacts                      |
| `WFE_SBOM_FILENAME`       | string             | syft-sbom.json       | --sbom-filename      | `artifacts.sbomFilename`     | The SBOM file name                                    |
| `WFE_GRYPE_FILENAME`      | string             | grype-report.json    | --grype-filename     | `artifacts.grypeFilename`    | The Grype file name                                   |
| `WFE_GITLEAKS_FILENAME`   | string             | gitleaks-report.json | --gitleaks-filename  | `artifacts.gitleaksFilename` | The Gitleaks file name                                |
| `WFE_BUNDLE_DIRECTORY`    | string             |                      | --bundle-directory   | `artifacts.bundleDirectory`  | The Gatecheck bundle directory                        |
| `WFE_BUNDLE_FILENAME`     | string             |                      | --bundle-filename    | `artifacts.bundleFilename`   | The Gatecheck bundle filename                         |

## Running in Docker

When running workflow-engine in a docker container there are some pipelines that need to run docker commands.
In order for the docker CLI in the workflow-engine to connect to the docker daemon running on the host machine,
you must either mount the `/var/run/docker.sock` in the `workflow-engine` container, or provide configuration for
accessing the docker daemon remotely with the `DOCKER_HOST` environment variable.

If you don't have access to Artifactory to pull in the Omnibus base image, you can build the image manually which is
in `images/omnibus/Dockerfile`.

### Using `/var/run/docker.sock`

This approach assumes you have the docker daemon running on your host machine.

Example:

```
docker run -it --rm \
  `# Mount your Dockerfile and supporting files in the working directory: /app` \
  -v "$(pwd):/app:ro" \
  `# Mount docker.sock for use by the docker CLI running inside the container` \
  -v "/var/run/docker.sock:/var/run/docker.sock" \
  `# Run the workflow-engine container with the desired arguments` \
  workflow-engine run image-build
```

### Using a Remote Daemon

For more information see the
[Docker CLI](https://docs.docker.com/engine/reference/commandline/cli/#environment-variables) and
[Docker Daemon](https://docs.docker.com/config/daemon/remote-access/) documentation pages.

### Using Podman in Docker

In addition to building images with Docker it is also possible to build them with podman. When running podman in docker it is necessary to either launch the container in privileged mode, or to run as the `podman` user:

```bash
docker run --user podman -it --rm \
  `# Mount your Dockerfile and supporting files in the working directory: /app` \
  -v "$(pwd):/app:ro" \
  `# Run the workflow-engine container with the desired arguments` \
  workflow-engine:local run image-build -i podman
```

If root access is needed, the easiest solution for using podman inside a docker container is to run the container in "privileged" mode:

```bash
docker run -it --rm \
  `# Mount your Dockerfile and supporting files in the working directory: /app` \
  -v "$(pwd):/app:ro" \
  `# Run the container in privileged mode so that podman is fully functional` \
  --privileged \
  `# Run the workflow-engine container with the desired arguments` \
  workflow-engine run image-build -i podman
```

### Using Podman in Podman

To run the workflow-engine container using podman the process is quite similar, but there are a few additional security options required:

```bash
podman run --user podman  -it --rm \
  `# Mount your Dockerfile and supporting files in the working directory: /app` \
  -v "$(pwd):/app:ro" \
  `# Run the container with additional security options so that podman is fully functional` \
  --security-opt label=disable --device /dev/fuse \
  `# Run the workflow-engine container with the desired arguments` \
  workflow-engine:local run image-build -i podman
```
