INSTALL_DIR := env('INSTALL_DIR', '/usr/local/bin')
WORKFLOW_ENGINE_IMAGE := "ghcr.io/cms-enterprise/batcave/workflow-engine"

# build workflow engine binary
build:
    mkdir -p bin
    go build -ldflags="-X 'main.cliVersion=$(git describe --tags)' -X 'main.gitCommit=$(git rev-parse HEAD)' -X 'main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)' -X 'main.gitDescription=$(git log -1 --pretty=%B)'" -o ./bin ./cmd/workflow-engine

# build and install binary
install: build
    cp ./bin/workflow-engine {{ INSTALL_DIR }}/workflow-engine

# golangci-lint view only
lint:
    golangci-lint run --fast

# golangci-lint fix linting errors and format if possible
fix:
    golangci-lint run --fast --fix
	
# Locally serve documentation
serve-docs:
	mdbook serve

gen-actions version-tag:
    mkdir -p .github/actions/code-scan/
    workflow-engine config gen code-scan-action -i {{ WORKFLOW_ENGINE_IMAGE }}:{{ version-tag }} > .github/actions/code-scan/action.yml
    mkdir -p .github/actions/image-build/
    workflow-engine config gen image-build-action -i {{ WORKFLOW_ENGINE_IMAGE }}:{{ version-tag }} > .github/actions/image-build/action.yml
    mkdir -p .github/actions/image-scan/
    workflow-engine config gen image-scan-action -i {{ WORKFLOW_ENGINE_IMAGE }}:{{ version-tag }} > .github/actions/image-scan/action.yml
    mkdir -p .github/actions/image-publish/
    workflow-engine config gen image-publish-action -i {{ WORKFLOW_ENGINE_IMAGE }}:{{ version-tag }} > .github/actions/image-publish/action.yml
    mkdir -p .github/actions/image-build-podman/

    workflow-engine config gen image-build-action --docker-alias podman -i {{ WORKFLOW_ENGINE_IMAGE }}:podman-{{ version-tag }} > .github/actions/image-build-podman/action.yml
    mkdir -p .github/actions/image-scan-podman/
    workflow-engine config gen image-scan-action --docker-alias podman -i {{ WORKFLOW_ENGINE_IMAGE }}:podman-{{ version-tag }} > .github/actions/image-scan-podman/action.yml
    mkdir -p .github/actions/image-publish-podman/
    workflow-engine config gen image-publish-action --docker-alias podman -i {{ WORKFLOW_ENGINE_IMAGE }}:podman-{{ version-tag }} > .github/actions/image-publish-podman/action.yml
