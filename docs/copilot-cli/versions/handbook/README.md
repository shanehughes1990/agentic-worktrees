# Copilot CLI + SDK Versioned Handbook

This directory contains version-pinned operational handbooks built from:
- Local terminal command mapping
- Local Go module source inspection for `github.com/github/copilot-sdk/go`

## Current handbook files

- `copilot-cli-0.0.415_sdk-v0.1.25.md`

## How to add a new version

1. Capture local versions:
   - `copilot --version`
   - `go list -m github.com/github/copilot-sdk/go`
2. Capture CLI help and command surface:
   - `copilot --help`
   - `copilot <subcommand> --help`
3. Inspect SDK source at local module path from:
   - `go list -f '{{.Dir}}' -m github.com/github/copilot-sdk/go`
4. Record:
   - Every path-related CLI flag
   - Every SDK path argument field in public/session/request types
   - Every output contract used by startup, status/auth, send, and event streaming
5. Add a new file using naming pattern:
   - `copilot-cli-<cli-version>_sdk-<sdk-version>.md`
