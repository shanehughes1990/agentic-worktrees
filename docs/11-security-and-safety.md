# 11 - Security and Safety Controls

## Security Objectives

- Protect repository integrity
- Protect credentials/tokens
- Prevent unsafe autonomous actions
- Maintain verifiable auditability

## Access and Identity

- least-privilege credentials for git and runtime APIs
- scoped tokens with rotation policy
- separate machine identity from human operator identity

## Command and Tool Safety

- allowlist tool/command execution per runtime policy
- deny destructive operations outside explicit policy scope
- enforce path boundaries to worktree and approved directories

## Git Safety Guards

- forbid force-push unless explicitly enabled by policy
- validate target branch and PR ownership before merge
- require clean conflict markers before merge

## Artifact and Log Safety

- redact secrets in logs/events
- classify sensitive outputs
- define retention and secure deletion policy

## Audit Requirements

- immutable append-only run ledger
- record who/what performed each critical action
- include policy decision context for escalations and overrides
- file-based audit trail enabled by default for all environments
- include step, path, command/action, and output/result details per critical transition
- enforce tamper-evident integrity controls for stored audit records

## Safety Defaults

- fail closed for unknown/unsafe actions
- escalate rather than guessing on high-risk merge/conflict actions
