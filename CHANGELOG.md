# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Version header format: `## [x.x.x] - yyyy-mm-dd`

## [UNRELEASED]

### Changed

- refactored, the directory structure, all pipelines will exist in pkg/pipelines
- updated Version commands to return Commands instead of just an error
- simplified Command Methods
- converted some Command to private to prevent auto-complete overload
- the way command dry running is called, uses builder pattern now

### Added

- debugging flag
- pkg/shell/commands Runner interface
- pkg/shell/commands Command struct
- pkg/shell for command wrappers
- grype version
- syft version
- podman version
- docker / podman support via CLI cmd interface
- docker info
- image build pipeline (info only)
- docker info, build, and push commands

## [0.0.1-rc.1] - 2024-01-01
### Added

- cmd/workflow-engine for cli
- pkg/environments
- pkg/jobs
- pkg/pipelines
- pkg/system
- initial project structure

