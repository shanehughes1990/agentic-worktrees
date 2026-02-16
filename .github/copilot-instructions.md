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
