# Build workflow engine cli
build:
	mkdir -p bin
	go build -o bin/workflow-engine ./cmd/workflow-engine

# Locally serve documentation
serve-docs:
	mdbook serve

