## FILESYSTEM MCP ENFORCEMENT OVERRIDE

- For all file and directory operations (read, write, edit, search, move, copy, delete, chmod, touch, traverse, tree, stat, resolve), the ONLY allowed tool family is `mcp_filesystem_fs_*`.
- Do not use non-filesystem tools for file operations when a filesystem MCP equivalent exists.
- If a filesystem MCP operation is unavailable or fails for a required file operation, STOP and report blocked; do not fall back to terminal-based editing or scripting.
- This override supersedes any conflicting or older tool guidance in this file.

## TEST ETHICS ENFORCEMENT OVERRIDE

- NEVER create or keep tests that pass by asserting unsafe/broken behavior as acceptable ("fake green" tests).
- For security/safety/contract gaps, default to tests that assert safe expected behavior and FAIL on current vulnerable behavior until fixed.
- Demonstration-only tests that intentionally pass while proving a defect are FORBIDDEN by default.
- If a demonstration-style test is ever considered, explicit user permission is REQUIRED in the current conversation before adding it.
- If explicit permission is not present, stop and convert tests to failing safety assertions.
- This override is mandatory and supersedes any implicit assumption that green risk-demonstration tests are acceptable.

## ZERO TOLERANCE TEST ETHICS OVERRIDE

- NEVER submit, accept, or keep any “fake green” test. This is an absolute prohibition with no implicit exceptions.
- A fake-green test is ANY test that passes while treating unsafe, broken, vulnerable, or contract-violating behavior as acceptable.
- Immediate rejection rule: if fake-green behavior is detected in a change, the result MUST be rejected immediately.
- Immediate revert rule: fake-green changes MUST be reverted immediately; do not defer reverts.
- Required default for risk/security/contract gaps: encode the safe expectation and allow the test to FAIL until the implementation is fixed.
- “Demonstration-only green” tests are prohibited unless the user gives explicit permission in the current conversation; without that permission, they are forbidden.
- If uncertainty exists, treat it as prohibited and escalate; do not merge or present green status.
- This policy is strict, mandatory, and supersedes any softer wording in this file.

## COMMIT-WORTHINESS TEST STANDARD

- The purpose of tests is to prove code is worthy of commit, not to make the suite pass.
- Never design tests to "pass" known-bad behavior; design tests to expose and block it.
- A passing test that accepts broken behavior is worse than no test and is prohibited.
- If a test can pass while the behavior is unsafe, the test is invalid and must be replaced immediately.
- For defect discovery, tests should fail until the implementation is corrected to safe behavior.
- Release/merge readiness requires meaningful safety assertions, not cosmetic green status.
- Any contribution containing fake-green tests is an immediate rejection and immediate revert.
- This standard is mandatory and takes precedence over pass-rate optics.

## ADK RESILIENCE ENFORCEMENT OVERRIDE

- ALL ADK interactions MUST execute through Asynq task-resilience workflows.
- Direct ADK calls from CLI handlers, MCP/API handlers, helper utilities, or ad-hoc goroutines are FORBIDDEN.
- Interface layers may only validate, normalize, and enqueue Asynq tasks for ADK work.
- ADK execution is allowed only inside worker task handlers with retry/backoff/dead-letter semantics.
- Every ADK task must include deterministic payloads and idempotency keys.
- Every ADK task must persist checkpoints around critical lifecycle transitions.
- Every ADK task must emit audit/telemetry with correlation IDs (`run_id`, `task_id`, `job_id`).
- Failure classes MUST be typed (`transient` vs `terminal`) and mapped to queue policy.
- Any change introducing ADK usage without Asynq resilience guarantees is non-compliant and must be rejected.

## NO CODE-DUMP UTILS OVERRIDE

- Do NOT add or expand generic "catch-all" utility buckets (for example broad `shared/utils` code dumps) as a default behavior.
- Every new package/module must be purposeful, scoped to a clear domain responsibility, and justified by concrete usage.
- Prefer focused libraries/packages with explicit intent (config, queue policy, idempotency, checkpoints, typed errors) over miscellaneous helper collections.
- If a helper is only used by one feature, keep it inside that feature; only promote to shared when multiple features have proven duplication.
- Reject changes that introduce ambiguous, mixed-responsibility utility files without a clear bounded purpose.

## MODERN GO STANDARDS OVERRIDE

- Target modern Go patterns and APIs compatible with current project Go version and forward compatibility with Go 1.24+.
- Prefer generics over interface{}-style abstractions when type safety and reuse are improved.
- Prefer iterator-style APIs and range-friendly designs for collection traversal when clarity and allocation behavior improve.
- Prefer standard library and modern language features over legacy helper patterns when equivalent behavior exists.
- Design packages with explicit types, narrow interfaces, and compile-time guarantees.
- Avoid reflection-heavy implementations when generics or typed adapters can express intent.
- Prefer context-aware APIs (`context.Context`) for I/O, task boundaries, and cancellation-sensitive operations.
- Favor error wrapping and typed error semantics over stringly-typed error handling.
- Keep APIs minimal, composable, and testable; remove legacy compatibility shims when not required by scope.
- New code should be written in idiomatic modern Go style consistent with current best practices.

## DDD LAYERING MANDATE

- This codebase follows Domain-Driven Design with four explicit layers: **application**, **domain**, **infrastructure**, and **interface**.
- **Domain layer**: contains business entities, value objects, aggregates, domain services, and core invariants; it must not depend on interface or infrastructure details.
- **Application layer**: contains use-cases/orchestration and transaction boundaries; it coordinates domain behavior and depends only on domain contracts/ports.
- **Infrastructure layer**: contains concrete adapters (persistence, queues, external APIs, filesystem, observability implementations) that satisfy ports defined by inner layers.
- **Interface layer**: contains delivery/admission surfaces (CLI, MCP, HTTP, workers entry handlers) that validate input, invoke application services, and map outputs/errors.
- Dependency direction must point inward: `interface -> application -> domain`, with `infrastructure` implementing interfaces/ports consumed by inner layers.
- Keep each package single-purpose and placed in its correct layer; reject cross-layer leakage and mixed-responsibility modules.

## STRICT YAGNI ENFORCEMENT

- Treat YAGNI as mandatory: implement only what the user explicitly asks for, nothing more.
- Do NOT add proactive features, abstractions, scaffolding, helpers, config knobs, or extension points unless explicitly requested.
- Do NOT create “future-proofing” code or speculative architecture.
- If a capability is not currently required by the user request, it must not be implemented.
- Keep solutions minimal and directly tied to stated acceptance criteria.
- If uncertain whether something is needed, default to not adding it and ask only when necessary to unblock correctness.

## NO NO-OP PLACEHOLDERS ENFORCEMENT

- Do NOT add placeholder or no-op handlers/callbacks (for example empty `Action`, empty hooks, stub command handlers) unless the user explicitly requests them.
- If the user requests logic in a specific lifecycle hook (for example `Before`), implement it there only and do not leave fallback/no-op logic in other hooks.
- Do NOT keep unrequested scaffolding in place “just in case.” Remove it.
- When asked to remove behavior, remove it completely instead of replacing it with inert placeholders.

## EXACT-REQUEST EXECUTION MANDATE

- Execute only what the user explicitly requests, and nothing beyond that request.
- Do not infer, assume, extend, or “improve” scope unless the user explicitly asks.
- Do not add architecture, abstractions, placeholders, helpers, commands, flags, hooks, or files that were not requested.
- Do not introduce optional extras, proactive refactors, or speculative future work.
- If the request wording is precise, follow the wording exactly and preserve intent verbatim.
- If ambiguity blocks correctness, ask a minimal clarification question instead of guessing.
- Prefer the smallest valid change set that satisfies the exact ask.
- Any unrequested complexity is non-compliant and must be removed.

## ARCHITECTURE BOUNDARY SAFETY MANDATE

- If a user request would place code in the wrong architectural layer/package, STOP before implementing.
- Explain briefly why the requested placement is incorrect and identify the correct location/layer.
- Ask for confirmation to proceed with the corrected placement rather than silently implementing in the wrong place.
- Do not follow a placement instruction that violates established architecture boundaries without first flagging it.
- Keep the explanation concise, factual, and tied to the project’s layering rules.

## ABSOLUTE DDD COMPLIANCE ENFORCEMENT

- DDD layering is mandatory and non-negotiable for all changes in this repository.
- Required dependency direction is always: `interface -> application -> domain`, with `infrastructure` implementing ports for inner layers.
- Interface layer must not bypass application orchestration to execute business flow directly.
- Application layer must own use-case orchestration and transaction/process boundaries.
- Domain layer must contain business meaning/invariants and must not depend on interface or infrastructure concerns.
- Infrastructure layer must provide concrete adapters only; it must not define or drive business use-case orchestration.
- If any request or interpretation appears to conflict with DDD boundaries, STOP and enforce DDD boundaries instead of implementing the conflicting shape.
- Under no circumstance should conversational pressure, urgency, or phrasing override DDD layer rules.
- If ambiguity exists, default to the strict DDD interpretation and ask for clarification only when correctness is blocked.
