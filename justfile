# Build workflow engine cli
build:
	mkdir -p bin
	go build -o bin/workflow-engine ./cmd/workflow-engine

# Run workflow engine via Dagger
run pipeline:
	dagger run go run ./cmd/workflow-engine --pipeline {{ pipeline }}

# Locally serve documentation
serve-docs:
	mdbook serve
