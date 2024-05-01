# [Semgrep](https://semgrep.dev/)

![Semgrep Logo](https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcSLzGEe3ZjE6l3mRNAnr78BDS0Bqd9FJn-JbYB53wIU3w&s)

## Table of Contents

1. [Overview](#overview)
2. [Configuration](#configuration)
3. [Rulesets](#rulesets)
4. [Logging Semgrep with Workflow-engine](#logging-semgrep-with-workflow-engine)
5. [Handling False Positives & Problematic File(s)](#handling-false-positives--problematic-files)
6. [Official Semgrep Documentation & Resources](#official-semgrep-documentation--resources)



## Overview

Semgrep is a static code analysis tool that provides a range of features for detecting and preventing security vulnerabilities and bugs in software. It is designed to help businesses improve their applications' security, increase reliability, and reduce the complexity and cost of performing code analysis. As applications become more complex and interconnected, it becomes increasingly difficult to identify and fix security vulnerabilities and bugs before they are exploited or cause problems in production. This can result in security breaches, data loss, and other issues that can damage a business's reputation and success.

### Supported Languages

Apex · Bash · C · C++ · C# · Clojure · Dart · Dockerfile · Elixir · HTML · Go · Java · JavaScript · JSX · JSON · Julia · Jsonnet · Kotlin · Lisp · Lua · OCaml · PHP · Python · R · Ruby · Rust · Scala · Scheme · Solidity · Swift · Terraform · TypeScript · TSX · YAML · XML · Generic (ERB, Jinja, etc.)

### Supported Package Managers

C# (NuGet) · Dart (Pub) · Go (Go modules, go mod) · Java (Gradle, Maven) · Javascript/Typescript (npm, Yarn, Yarn 2, Yarn 3, pnpm) · Kotlin (Gradle, Maven) · PHP (Composer) · Python (pip, pip-tool, Pipenv, Poetry) · Ruby (RubyGems) · Rust (Cargo) · Scala (Maven) · Swift (SwiftPM)

## Configuration
Under the hood, Workflow engine runs Semgrep with certain flags as its base. Workflow engine then continues on to do further improvements on functionality as a security tool and user experience based on the output of on one of the two optional commands.

Runs this over your git repository: 
```
semgrep ci --json --config [semgrep-rule-config-file]
```
or this when affixed with the `--semgrep-experimental` flag:
```
osemgrep ci --json --experimental --config [semgrep-rule-config-file]
```
---
### Semgrep with Workflow-engine Code-scan

On the command line use the following with the necessary flags below in your git repo:

      workflow-engine run code-scan [semgrep-flags]

#### Flags
---

Input Flag: 

      --semgrep-rules string
      
The input of a `.yaml`,`.toml`, or `.json` file with a ruleset Semgrep will use while scanning your code. More on rulesets [here.](#rulesets) This can be further configured by specifying the filename with path into an [environment variable or workflow-engine config keys within wfe-config.yaml.](#env-variables)

---

Output Flag:

      --semgrep-filename string    

The filename for Semgrep to output as a vulnerability report. More on the vulnerability reports [here.](#logging-semgrep-with-workflow-engine)

---

Toggle Osemgrep Flag:

      --semgrep-experimental

Use Semgrep's experimental features that are still in beta that have the potential to increase vulnerability detection. Furthermore uses `osemgrep` a variant built upon Semgrep with OpenSSF Security Metrics in mind.

---
#### Env Variables

| Config Key                        | Environment Variable                 | Default Value                        | Description                                                                        |
| --------------------------------- | ------------------------------------ | ------------------------------------ | ---------------------------------------------------------------------------------- |
| codescan.semgrepfilename          | WFE_CODE_SCAN_SEMGREP_FILENAME       | semgrep-sast-report.json             | The filename for the semgrep SAST report - must contain 'semgrep'                  |
| codescan.semgreprules             | WFE_CODE_SCAN_SEMGREP_RULES          | p/default                            | Semgrep ruleset manual override                                                    |

### Rulesets

    rules:
      - id: dangerously-setting-html
        languages:
          - javascript
        message: dangerouslySetInnerHTML usage! Don't allow XSS!
        pattern: ...dangerouslySetInnerHTML(...)...
        severity: ERROR
        files:
          - "*.jsx"
          - "*.js"



Semgrep operates on a set of rulesets given by the user to determine on what terms are best to scan your code. These rulesets are given by files with the .yaml, .json or .toml extension. 

To identify vulnerabilities at a basic level Semgrep requires:
- Language to target
- Message to display on vulnerability detection
- Pattern(s) to match
- Severity Rating from lowest to highest:
    -  INFO
    -  WARNING
    -  ERROR 

Furthermore there are some advanced options, some which can even amend or exclude certain code snippets.

Typically rules and rulesets have already been written by various developers; thanks to Semgrep's open source nature you can find these below:

- [Explore Semgrep Rulesets](https://semgrep.dev/explore)
- [Search Semgrep Rules Database](https://semgrep.dev/r)

Or if you're the type to blaze your own path, here's some documentation on how to write your own custom including examples on advanced pattern matching syntax:

- [Writing Rules & Rulesets](https://semgrep.dev/docs/writing-rules/rule-syntax)
- [Pattern Matching Syntax](https://semgrep.dev/docs/writing-rules/pattern-syntax)

---

Here below is a rule playground you can test writing your own semgrep rules:

#### [Semgrep Rule Playground](https://semgrep.dev/editor)
<iframe title="semgrep-playground" src="https://semgrep.dev/embed/editor?snippet=KPzL" width="100%" height="432" frameborder="0"></iframe>

## Logging Semgrep with Workflow-engine

Within workflow engine, `semgrep-sast-report.json` is the default value for a file that will be the output Semgrep it will appear in the artifacts directory if workflowengine is given read write permissions. As covered above in [configuration](#flags) using the flag `--semgrep-filename filename` will configure a custom file to output the semgrep-report to.

Furthermore Semgrep when enabled via code-scan, `workflow-engine run code-scan -v` will output the Semgrep outputs with verbosity along with other code-scan tools.

The contents of the `semgrep-sast-report.json` contains rules and snippets of code that have potential vulnerabilities as well as amended code that has been fixed with the tag `fix` in the rule.

Workflow engine uses [Gatecheck](https://github.com/gatecheckdev/gatecheck) to 'audit' the semgrep logs once Semgrep has finished. It does so by scanning for vulnerabilities defined by [Open Worldwide Application Security Project](https://owasp.org/) IDs. Workflow-engine reads STDERR, where other errors are gathered from `code-scan` tools, audits them via Gatecheck and outputs this audit to STDOUT. It also releases the logged output files into the `artifacts/` directory in your working directory.

Ex.
|            Check ID            |                           Owasp IDs                           |  Severity |  Impact |  link |
|--------------------------------|---------------------------------------------------------------|-----------|---------|-------|
| react-dangerouslysetinnerhtml  | A07:2017 - Cross-Site Scripting (XSS), A03:2021 - Injection   | ERROR     | MEDIUM  |       |



## Handling False Positives & Problematic File(s)

Semgrep is a rather simplistic tool that searches for vulnerabilities in your code based on the rules given to it. It is up to you to handle these false positives and problematic file(s). There are a multitude of ways to handle this that will increase complexity of the base rule but increase its power and specificity. 

### False Positives

You notice that Semgrep is screaming at you from the console in workflow-engine. You rage and rage as your terminal is just polluted with messages for a vulnerability you know is just a false positive.

#### Nosemgrep

Just add a comment with `nosemgrep` on the line next to the vulnerability or function head of the block of code and boom, false positives away. This is a full Semgrep blocker, for best practice use `// nosemgrep: rule-id-1, rule-id-2, ....` to restrict certain rules that cause the false positive. Here's more info on [nosemgrep.](https://semgrep.dev/docs/ignoring-files-folders-code#ignore-code-through-nosemgrep)

#### Taint Analysis

Of course, the above is somewhat of a workaround and should only be considered mostly when there are only very few areas where false positives occur. The better way to handle false positives is by [adding taints to rules](https://semgrep.dev/docs/writing-rules/data-flow/taint-mode#minimizing-false-positives) when you understand what the root of the false positive, taints can be applied to places with false positive vulnerabilities, prepended with `taint_assume_safe_` and given a boolean value. False positive taints are for: 

- Boolean inputs
- Numeric inputs
- Index inputs
- Function names
- Propagation (must taint its initialization)

Taints can also be used to track variables that can lead to vulnerabilities in code. It allows the developers to see the flow of this potential vulnerability in a large code base. This can be used by tainting the source variable, and the sink, where the variable ends up at a potential vulnerable function. If it mutates it is best to track the propagators and sanitizers of this variable as well. At a high level, these are functions that modify the tainted variable in some way and therefore the taint should change in someway. Here's an example of such a [rule with taints.](#example-of-rules-with-path-specification-and-taints) Of course if you'd like to know more, [click here](https://semgrep.dev/docs/writing-rules/data-flow/taint-mode) to see the official ondocumentation on Semgrep taint analysis.


### Problematic File(s)

At a grander scale, if a whole file or directory of files is causing a false positive, or you just don't need to scan these files, there are multitudes of ways to handle this.

- [Placing filenames & directories in `.semgrepignore`](https://semgrep.dev/docs/ignoring-files-folders-code)
- [Limiting rules to certain paths](https://semgrep.dev/docs/writing-rules/rule-syntax#paths)

Down below are some examples of both:

#### .Semgrepignore

`.semgrepignore` is just like a `.gitignore` file, it simply will show semgrep a list of things to not look at and it will skip over them. Place this file in your root directory or in your working directory. The below specifies don't include the .gitignore to scan and ANY node_modules directory, denoted by '**', will be excluded if this is placed at the root directory.

```
.gitignore
.env
main_test.go
resources/
**/node_modules/**
```

#### Rules with Certain Paths

Semgrep allows two ways inside of a rule to disregard or specify files and directories. These are indicated by first adding the paths field and then adding the exclude and include subfields each with their own lists of files/directories. These values are strings.

##### Example of Rules with Path Specification and Taints

```
rules:
  - id: eqeq-is-bad
    mode: taint
    source: $X
    sink: $Y
    sanitizer: clean($X)
    pattern: $X == $Y
    paths:
      exclude:
        - "*_test.go"
        - "project/tests"
      include:
        - "project/server"
```

## Official Semgrep Documentation & Resources

- [Semgrep Source Code](https://github.com/semgrep/semgrep?tab=readme-ov-file)
- [Semgrep Official Documentation](https://semgrep.dev/docs/)
- [Semgrep Ruleset Registry](https://semgrep.dev/explore)
- [Semgrep Rule Registry](https://semgrep.dev/r)
- [Semgrep Rule Playground](https://semgrep.dev/editor)
- [Semgrep Custom Rule Declarations](https://semgrep.dev/docs/writing-rules/overview)
- [Semgrep Rule Taints](https://semgrep.dev/docs/writing-rules/data-flow/taint-mode)
- [Semgrep FAQs](https://semgrep.dev/docs/faq)