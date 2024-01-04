# Workflow Engine

Workflow Engine is a security and delivery pipeline designed to orchestrate the process of building and scanning an 
application image for security vulnerabilities.
It solves the problem of having to configure a hardened-predefined security pipeline using traditional CI/CD.
Workflow Engine can be statically compiled as a binary and run with [Dagger](dagger.io) on virtually any platform, CI/CD
environment, or locally.

## Getting Started

Install Prerequisites:

- Container Engine
- Docker CLI
- Dagger
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
dagger run workflow-engine --pipeline debug
```

