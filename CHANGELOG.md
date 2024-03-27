# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Version header format: `## [x.x.x] - yyyy-mm-dd`

## [UNRELEASED]

### Added

- GitHub action auth support

### Changed

- Fixed code scan stderr/stdout collision by moving the stdout dump to the end of the run function
- Fixed image scan stderr/stdout collision by moving the stdout dump to the end of the run function

## [v0.0.0-rc.12]

### Changed

- refactored CLI for readability and maintenance
- upgraded to go 1.22.0
- New Executable will default inputs and outputs to OS
- WithStdin, WithStdout, WithStderr all merged to WithIO
- Config syntax to correlated with pipelines
- Code Scan Run organization to use functions for simplicity
- Add multi writer for Gatecheck list
- async run execution for image scan pipeline
- moved shell package to legacy
- image scan execution flow
- docker build argument strategy for shell
- shell command errors instead of exit codes
- shell command rich errors
- async task wraps stderr for cleaner log output
- Limit the number of supported Github action fields

### Added

- Configuration File template rendering with built-in values
- Configuration conversions
- Configuration init with the format option 
- Semgrep, osemgrep, gitleaks shell commands
- Code Scan Pipeline
- Config Template auto rendering
- Version Command
- All commands will defer to viper for arguments and defaults
- no push flag to image publish
- Gatecheck Shell Command
- pipeline helper functions for common file operations
- Oras Command
- Deploy pipeline validation only (beta feature)
- clamscan & freshclam for virus scanning
- command run with context
- command run with IO
- grype CMD 
- async task object
- "Combo" pipelines for image-delivery and all pipelines
- GitHub Actions Code Generation

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
