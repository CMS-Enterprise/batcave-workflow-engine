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

upgrade:
    git status --porcelain | grep -q . && echo "Repository is dirty, commit changes before upgrading." && exit 1 || exit 0
    go get -u ./...
    go mod tidy

# Locally serve documentation
serve-docs:
	mdbook serve
