# Workflow Engine

[![Build workflow-engine](https://github.com/CMS-Enterprise/batcave-workflow-engine/actions/workflows/delivery.yaml/badge.svg)](https://github.com/CMS-Enterprise/batcave-workflow-engine/actions/workflows/delivery.yaml)

![Workflow Engine Splash Logo](https://static.caffeineforcode.com/workflow-engine-splash-red.png)

Workflow Engine is a security and delivery pipeline designed to orchestrate the process of building and scanning an
application image for security vulnerabilities.
It solves the problem of having to configure a hardened-predefined security pipeline using traditional CI/CD.
Workflow Engine can be statically compiled as a binary and run on virtually any platform, CI/CD
environment, or locally.

## Getting Started

Install Prerequisites:

- Container Engine
- Docker or Podman CLI
- Golang >= v1.22.0
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
mkdir bin
go build -o bin/workflow-engine ./cmd/workflow-engine
```

Optionally, if you care to include metadata you use build arguments

```shell
go build -ldflags="-X 'main.cliVersion=$(git describe --tags)' -X 'main.gitCommit=$(git rev-parse HEAD)' -X 'main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)' -X 'main.gitDescription=$(git log -1 --pretty=%B)'" -o ./bin ./cmd/workflow-engine
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

| Config Key                        | Environment Variable                 | Default Value                        | Description                                                                        |
| --------------------------------- | ------------------------------------ | ------------------------------------ | ---------------------------------------------------------------------------------- |
| codescan.enabled                  | WFE_CODE_SCAN_ENABLED                | 1                                    | Enable/Disable the code scan pipeline                                              |
| codescan.gitleaksfilename         | WFE_CODE_SCAN_GITLEAKS_FILENAME      | gitleaks-secrets-report.json         | The filename for the gitleaks secret report - must contain 'gitleaks'              |
| codescan.gitleakssrcdir           | WFE_CODE_SCAN_GITLEAKS_SRC_DIR       | .                                    | The target directory for the gitleaks scan                                         |
| codescan.semgrepfilename          | WFE_CODE_SCAN_SEMGREP_FILENAME       | semgrep-sast-report.json             | The filename for the semgrep SAST report - must contain 'semgrep'                  |
| codescan.semgreprules             | WFE_CODE_SCAN_SEMGREP_RULES          | p/default                            | Semgrep ruleset manual override                                                    |
| deploy.enabled                    | WFE_IMAGE_PUBLISH_ENABLED            | 1                                    | Enable/Disable the deploy pipeline                                                 |
| deploy.gatecheckconfigfilename    | WFE_DEPLOY_GATECHECK_CONFIG_FILENAME | -                                    | The filename for the gatecheck config                                              |
| gatecheckbundlefilename           | WFE_GATECHECK_BUNDLE_FILENAME        | artifacts/gatecheck-bundle.tar.gz    | The filename for the gatecheck bundle, a validatable archive of security artifacts |
| imagebuild.args                   | WFE_IMAGE_BUILD_ARGS                 | -                                    | Comma seperated list of build time variables                                       |
| imagebuild.builddir               | WFE_IMAGE_BUILD_DIR                  | .                                    | The build directory to using during an image build                                 |
| imagebuild.cachefrom              | WFE_IMAGE_BUILD_CACHE_FROM           | -                                    | External cache sources (e.g., "user/app:cache", "type=local,src=path/to/dir")      |
| imagebuild.cacheto                | WFE_IMAGE_BUILD_CACHE_TO             | -                                    | Cache export destinations (e.g., "user/app:cache", "type=local,src=path/to/dir")   |
| imagebuild.dockerfile             | WFE_IMAGE_BUILD_DOCKERFILE           | Dockerfile                           | The Dockerfile/Containerfile to use during an image build                          |
| imagebuild.enabled                | WFE_IMAGE_BUILD_ENABLED              | 1                                    | Enable/Disable the image build pipeline                                            |
| imagebuild.platform               | WFE_IMAGE_BUILD_PLATFORM             | -                                    | The target platform for build                                                      |
| imagebuild.squashlayers           | WFE_IMAGE_BUILD_SQUASH_LAYERS        | 0                                    | squash image layers - Only Supported with Podman CLI                               |
| imagebuild.target                 | WFE_IMAGE_BUILD_TARGET               | -                                    | The target build stage to build (e.g., [linux/amd64])                              |
| imagepublish.bundlepublishenabled | WFE_IMAGE_BUNDLE_PUBLISH_ENABLED     | 1                                    | Enable/Disable gatecheck artifact bundle publish task                              |
| imagepublish.bundletag            | WFE_IMAGE_PUBLISH_BUNDLE_TAG         | my-app/artifact-bundle:latest        | The full image tag for the target gatecheck bundle image blob                      |
| imagepublish.enabled              | WFE_IMAGE_PUBLISH_ENABLED            | 1                                    | Enable/Disable the image publish pipeline                                          |
| imagescan.clamavfilename          | WFE_IMAGE_SCAN_CLAMAV_FILENAME       | clamav-virus-report.txt              | The filename for the clamscan virus report - must contain 'clamav'                 |
| imagescan.enabled                 | WFE_IMAGE_SCAN_ENABLED               | 1                                    | Enable/Disable the image scan pipeline                                             |
| imagescan.grypeconfigfilename     | WFE_IMAGE_SCAN_GRYPE_CONFIG_FILENAME | -                                    | The config filename for the grype vulnerability report                             |
| imagescan.grypefilename           | WFE_IMAGE_SCAN_GRYPE_FILENAME        | grype-vulnerability-report-full.json | The filename for the grype vulnerability report - must contain 'grype'             |
| imagescan.syftfilename            | WFE_IMAGE_SCAN_SYFT_FILENAME         | syft-sbom-report.json                | The filename for the syft SBOM report - must contain 'syft'                        |


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
