# Code-scan

## Overview

lorem ipsum stuff


### Using Code-scan On CLI

    workflow-engine run code-scan [flags]

### Flags

|           Flags            |                        Definition                         |
|----------------------------|-----------------------------------------------------------|
| --gitleaks-filename string | the output filename for the gitleaks vulnerability report |
| -h, --help                 | help for code-scan                                        |
| --semgrep-experimental     | use the osemgrep statically compiled binary               |
| --semgrep-filename string  | the output filename for the semgrep vulnerability report  |
| --semgrep-rules string     | the rules semgrep will use for the scan                   |



## Code-scan Security Tools

- [Semgrep](./code_scan_tools/semgrep.md)