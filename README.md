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

Note: `(none)` means unset, left blank

| Env Variable                                  | Viper Key                             | Default Value                            | Description                                      |
| --------------------------------------------- | ------------------------------------- | ---------------------------------------- | ------------------------------------------------ |
| WFE_IMAGE_BUILD_ENABLED                       | imagebuild.enabled                    | 1                                        | Enables or disables the image build process.     |
| WFE_IMAGE_BUILD_DIR                           | imagebuild.builddir                   | "."                                      | The directory where the image build takes place. |
| WFE_IMAGE_BUILD_DOCKERFILE                    | imagebuild.dockerfile                 | "Dockerfile"                             | The name/path of the Dockerfile.                 |
| WFE_IMAGE_BUILD_TAG                           | imagebuild.tag                        | (none)                                   | The tag to be applied to the built image.        |
| WFE_BUILD_IMAGE_PLATFORM                      | imagebuild.platform                   | (none)                                   | The platform for the image build.                |
| WFE_IMAGE_BUILD_TARGET                        | imagebuild.target                     | (none)                                   | The target build stage in the Dockerfile.        |
| WFE_IMAGE_BUILD_CACHE_TO                      | imagebuild.cacheto                    | (none)                                   | Specifies where to store build cache.            |
| WFE_IMAGE_BUILD_CACHE_FROM                    | imagebuild.cachefrom                  | (none)                                   | Specifies where to load build cache from.        |
| WFE_IMAGE_BUILD_SQUASH_LAYERS                 | imagebuild.squashlayers               | (none)                                   | Enable or disable squashing of build layers.     |
| WFE_IMAGE_BUILD_SCAN_TARGET                   | imagebuild.scantarget                 | (none)                                   | The target for image scanning.                   |
| WFE_IMAGE_SCAN_ENABLED                        | imagescan.enabled                     | 1                                        | Enables or disables the image scanning process.  |
| WFE_IMAGE_SCAN_CLAMAV_FILENAME                | imagescan.clamavFilename              | "clamav-virus-report.txt"                | Filename for ClamAV scan report.                 |
| WFE_IMAGE_SCAN_SYFT_FILENAME                  | imagescan.syftFilename                | "syft-sbom-report.json"                  | Filename for Syft SBOM report.                   |
| WFE_IMAGE_SCAN_GRYPE_CONFIG_FILENAME          | imagescan.grypeConfigFilename         | (none)                                   | Configuration file for Grype.                    |
| WFE_IMAGE_SCAN_GRYPE_ACTIVE_FINDINGS_FILENAME | imagescan.grypeActiveFindingsFilename | "grype-vulnerability-report-active.json" | Filename for Grype active findings report.       |
| WFE_IMAGE_SCAN_GRYPE_ALL_FINDINGS_FILENAME    | imagescan.grypeAllFindingsFilename    | "grype-vulnerability-report-full.json"   | Filename for Grype full findings report.         |
| WFE_CODE_SCAN_ENABLED                         | codescan.enabled                      | 1                                        | Enables or disables the code scanning process.   |
| WFE_CODE_SCAN_GITLEAKS_FILENAME               | codescan.gitleaksFilename             | "gitleaks-secrets-report.json"           | Filename for Gitleaks secrets report.            |
| WFE_CODE_SCAN_GITLEAKS_SRC_DIR                | codescan.gitleaksSrcDir               | "."                                      | Source directory for Gitleaks scan.              |
| WFE_CODE_SCAN_SEMGREP_FILENAME                | codescan.semgrepFilename              | "semgrep-sast-report.json"               | Filename for Semgrep SAST report.                |
| WFE_CODE_SCAN_SEMGREP_RULES                   | codescan.semgrepRules                 | "p/default"                              | Rule set for Semgrep scan.                       |

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
