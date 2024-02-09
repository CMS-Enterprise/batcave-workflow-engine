# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Version header format: `## [x.x.x] - yyyy-mm-dd`

## [UNRELEASED]

### Changed

- refactored CLI for readability and maintenance

### Added

- Configuration File template rendering with built-in values
- Configuration conversions
- Configuration init with the format option 

### Fixed

- Viper config key names
- Specified CLI command parameters for custom input and output for easier unit testing in the future

## [0.0.1-rc.1] - 2024-01-29

### Changed

- refactored, the directory structure, all pipelines will exist in pkg/pipelines
- updated Version commands to return Commands instead of just an error
- simplified Command Methods
- converted some Command to private to prevent auto-complete overload
- the way command dry running is called, uses builder pattern now
- fixed a bug where only the last docker build flag was being added to the final command
- remove args from wrapper functions in CLI
- fixed the debug-pipeline calling syft scan

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
- internal logger to image build pipeline
- json, yaml, toml meta tags for pipelines/config
- config parsing with viper
- grype scan sbom command
- image-scan pipeline
- syft scan image command
- syft to image-scan pipeline
- image scan pipeline to CLI
- image scan pipeline wiring in CLI for Viper config variables

### Added

- cmd/workflow-engine for cli
- pkg/environments
- pkg/jobs
- pkg/pipelines
- pkg/system
- initial project structure
