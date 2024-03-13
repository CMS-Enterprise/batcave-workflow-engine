INSTALL_DIR := env('INSTALL_DIR', '/usr/local/bin')


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

gen-actions workflow-engine-image workflow-engine-podman-image:
    mkdir -p .github/actions/code-scan/
    workflow-engine config gen code-scan-action -i {{ workflow-engine-image }} > .github/actions/code-scan/action.yml
    mkdir -p .github/actions/image-build/
    workflow-engine config gen image-build-action -i {{ workflow-engine-image }} > .github/actions/image-build/action.yml
    mkdir -p .github/actions/image-scan/
    workflow-engine config gen image-scan-action -i {{ workflow-engine-image }} > .github/actions/image-scan/action.yml
    mkdir -p .github/actions/image-publish/
    workflow-engine config gen image-publish-action -i {{ workflow-engine-image }} > .github/actions/image-publish/action.yml
    mkdir -p .github/actions/image-build-podman/

    workflow-engine config gen image-build-action -i {{ workflow-engine-podman-image }} > .github/actions/image-build/action.yml
    mkdir -p .github/actions/image-scan-podman/
    workflow-engine config gen image-scan-action -i {{ workflow-engine-podman-image }} > .github/actions/image-scan/action.yml
    mkdir -p .github/actions/image-publish-podman/
    workflow-engine config gen image-publish-action -i {{ workflow-engine-podman-image }} > .github/actions/image-publish/action.yml
