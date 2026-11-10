# Use Go Instead of Python for Tools and Scripts

- Never write Python for tools, helper scripts, automation, validation, or tests in the codebase.
- Always write developer utilities, verification scripts, and tests in Go.

# Repository Paths & Documentation Links

- Never use machine-specific absolute paths (such as `/Users/...` or `/home/...`) or `file:///`
  URLs inside repository files, documentation (`*.md`), or skill definitions.
- Reserve `file:///` links exclusively for chat responses and UI artifacts; never write them into
  git-tracked files.
- Always use relative paths in documentation and markdown files.

# Go Package Installation, Update, and Security Analysis Rules

## Approval Required Before Installing or Updating Go Packages

- Never install, add, or update any Go packages or module dependencies without asking first.
- Obtain explicit user approval before running commands like `go get` or modifying `go.mod` in a
  way that downloads a new package or updates an existing package to a newer version.

## Mandatory Security Analysis Workflow

If the user approves the installation or update of any Go package:
1. Install or update the approved package.
2. **Before running any code** (including builds, unit tests, or pipelines), perform a thorough
   static security analysis of the source code on this machine for every newly installed or updated
   direct and indirect dependency looking for suspicious code.
3. Check for suspicious patterns:
   - Unsolicited network connections, telemetry, or data exfiltration.
   - Command or shell execution (`os/exec`, `syscall.Exec`, `/bin/sh`).
   - Obfuscated payloads, eval mechanisms, or dynamic code loaders.
   - Unauthorized filesystem access or environment snooping (`os.Getenv`).
   - Embedded native binaries, assembly files, or CGO bindings.
4. Provide a structured report for each package containing:
   - **Package & Version**: Name and resolved version tag.
   - **Web URL**: Repository, pkg.go.dev, and documentation links.
   - **Origin & Developer**: Original author and maintaining organization.
   - **Maintainers & Contributors**: Count of maintainers and contributors.
   - **Go Team Affiliation**: Whether any author or maintainer is on the Google Go team.
   - **Purpose**: Summary of functionality.
   - **Safety & Malware**: License, implementation language (pure Go vs CGO), and audit findings.
   - **Local Location**: Absolute path in the local Go module cache (`$GOPATH/pkg/mod/...`).
5. Provide all of this information to the user and **stop and wait** for the user to review and
   approve before continuing anything else, especially anything that involves running that code.

# Separation of Concerns

`omega-core` (`module gradientcontrol.com/omega/core`) is the standalone GSL compiler, standard
library, spatial/interval evaluation, and code generation SDK:
- Must remain completely decoupled from `stirigma`, gRPC/dispatcher buses, cloud SDKs, and `omega`
  application services.
- Expose cohesive compiler and evaluation entry points (`parser`, `assembler`, `eval`, `program`,
  `stdlib`, `backend/*`) configured via `With...` functional options rather than leaking internal
  AST passes (`internal/*`).
