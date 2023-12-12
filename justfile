# Build workflow engine cli
build:
	mkdir -p bin
	go build -o bin/workflow-engine ./cmd/workflow-engine

# Run workflow engine
run:
	go run .

# Locally serve documentation
serve-docs:
	mdbook serve
