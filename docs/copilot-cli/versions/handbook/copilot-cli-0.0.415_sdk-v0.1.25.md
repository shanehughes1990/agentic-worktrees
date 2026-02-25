# Copilot CLI/SDK Handbook

- CLI version: `GitHub Copilot CLI 0.0.415`
- SDK module: `github.com/github/copilot-sdk/go v0.1.25`
- Generated on: 2026-02-23 (local machine)

---

## 1) Terminal mapping (commands and observed outputs)

### 1.1 Version discovery

- Command: `copilot --version`
- Output:
  - `GitHub Copilot CLI 0.0.415.`
  - `Run 'copilot update' to check for updates.`

- Command: `go list -m github.com/github/copilot-sdk/go`
- Output:
  - `github.com/github/copilot-sdk/go v0.1.25`

### 1.2 Top-level CLI surface

- Command: `copilot --help`
- Output surface observed:
  - Usage: `copilot [options] [command]`
  - Main commands include: `login`, `help`, `init`, `update`, `version`, `plugin`
  - Help topics include: `config`, `commands`, `environment`, `logging`, `permissions`

### 1.3 Path-related global flags observed (from help output)

These path-sensitive flags appear on `copilot` and subcommand helps (`auth`, `setup`, `mcp`, `extension`, `feedback`):

- `--add-dir <directory>`
  - Adds allowed directories for file access; repeatable.
- `--additional-mcp-config <json>`
  - Additional MCP config via JSON string or file path using `@...` form.
- `--allow-all-paths`
  - Disables path verification.
- `--config-dir <directory>`
  - Overrides default config directory (`~/.copilot`).
- `--disallow-temp-dir`
  - Blocks automatic temp directory access.
- `--log-dir <directory>`
  - Overrides log directory (`~/.copilot/logs/`).
- `--log-level <level>`
  - Logging level control (`none`, `error`, `warning`, `info`, `debug`, `all`, `default`).

### 1.4 Path examples observed from help

- `copilot --add-dir /home/user/projects`
- `copilot --add-dir ~/workspace --add-dir /tmp`
- `copilot --allow-all-paths`

---

## 2) SDK path arguments (v0.1.25)

This section lists path arguments exposed by the SDK and/or transformed into RPC request payloads.

### 2.1 Client startup / process path args

From `client.go` (`NewClient`, `startCLIServer`):

- `ClientOptions.CLIPath`
  - Explicit CLI binary path; if empty, SDK checks embedded CLI, else falls back to `copilot` on `PATH`.
- `ClientOptions.CLIArgs`
  - User-provided extra CLI args appended before SDK-managed args.
- `ClientOptions.Cwd`
  - Process working directory (`exec.Cmd.Dir`) when spawning CLI.
- `COPILOT_CLI_PATH` (env override)
  - If set, overrides CLI path chosen in options.

SDK-managed process args appended by `startCLIServer`:

- Always:
  - `--headless`
  - `--no-auto-update`
  - `--log-level <value>`
- Transport:
  - `--stdio` (stdio mode), or
  - `--port <n>` (TCP mode)
- Auth-related:
  - `--auth-token-env COPILOT_SDK_AUTH_TOKEN` when token mode enabled
  - `--no-auto-login` when `UseLoggedInUser=false` or implied by token auth

### 2.2 Session and resume path args

From `types.go` + `client.go` request mapping:

- `SessionConfig.ConfigDir` -> `createSessionRequest.configDir`
- `SessionConfig.WorkingDirectory` -> `createSessionRequest.workingDirectory`
- `SessionConfig.SkillDirectories []string` -> `createSessionRequest.skillDirectories`
- `ResumeSessionConfig.WorkingDirectory` -> `resumeSessionRequest.workingDirectory`
- `ResumeSessionConfig.ConfigDir` -> `resumeSessionRequest.configDir`
- `ResumeSessionConfig.SkillDirectories []string` -> `resumeSessionRequest.skillDirectories`

### 2.3 MCP local server path arg

From `MCPLocalServerConfig`:

- `Cwd string` (`json:"cwd,omitempty"`)

### 2.4 Attachment path args

From generated session event type model (`generated_session_events.go`) and send request shape:

- `Attachment.Path *string`
- `Attachment.FilePath *string`
- `MessageOptions.Attachments []Attachment` -> serialized in `sessionSendRequest.attachments`

### 2.5 Context path outputs present in event models

From generated event structures:

- `Context.cwd` (serialized as `cwd`)
- `Context.gitRoot`
- Additional path-bearing fields in event data include optional `cwd`, `workingDirectory`, and attachment path fields where applicable.

---

## 3) SDK output contracts (v0.1.25)

### 3.1 Startup / lifecycle outputs and errors

From `client.go`:

- `Client.Start(ctx)`
  - Returns `nil` on connected + protocol-compatible startup.
  - Returns errors including:
    - `failed to start CLI server: <err>`
    - `timeout waiting for CLI server to start`
    - `failed to parse port: <err>`
    - `failed to connect to CLI server at <host:port>: <err>`
    - protocol mismatch errors from `verifyProtocolVersion`.

- `Client.Stop()` / `Client.ForceStop()`
  - Stop aggregates cleanup errors.
  - ForceStop suppresses errors and hard-clears process/connection state.

### 3.2 Status/auth outputs

From `types.go` + `client.go` methods:

- `Ping(ctx, message)` -> `PingResponse`
  - `message`, `timestamp`, optional `protocolVersion`
- `GetStatus(ctx)` -> `GetStatusResponse`
  - `version`, `protocolVersion`
- `GetAuthStatus(ctx)` -> `GetAuthStatusResponse`
  - `isAuthenticated`
  - optional: `authType`, `host`, `login`, `statusMessage`

### 3.3 Session send outputs

From `session.go`:

- `Session.Send(ctx, MessageOptions)` -> `messageId string`
  - RPC method: `session.send`
  - request includes `sessionId`, `prompt`, optional `attachments`, optional `mode`
- `Session.SendAndWait(ctx, MessageOptions)` -> `*SessionEvent`
  - Returns latest assistant message event when idle reached.
  - Returns timeout/session error when applicable.

### 3.4 Session create/resume outputs

From `types.go` + `client.go` unmarshalling:

- `session.create` -> `createSessionResponse`
  - `sessionId`, `workspacePath`
- `session.resume` -> `resumeSessionResponse`
  - `sessionId`, `workspacePath`

### 3.5 Tool and event output schema highlights

From `types.go` + `generated_session_events.go`:

- Tool result envelope:
  - `textResultForLlm`, optional binary payloads, `resultType`, optional `error`, optional telemetry/log fields.
- Session event content model includes:
  - `Result` with `content`, `contents[]`, optional `detailedContent`
  - `Content` with optional text/data/mime/cwd and optional exit code depending on content type.

---

## 4) Source mapping used for this handbook

Local SDK module path used during inspection:

- `/Users/shanehughes/go/pkg/mod/github.com/github/copilot-sdk/go@v0.1.25`

Primary inspected files:

- `client.go`
- `session.go`
- `types.go`
- `generated_session_events.go`
- `process_other.go`
- `embeddedcli/installer.go`

---

## 5) Notes for local integration test alignment

When local integration tests validate SDK startup/auth behavior against this version pair, keep these constraints pinned:

- CLI process startup args must include SDK-managed defaults (`--headless --no-auto-update --log-level ...`).
- Path behavior expectations must explicitly account for:
  - `CLIPath`/`COPILOT_CLI_PATH`
  - `Cwd`
  - session `ConfigDir` and `WorkingDirectory`
  - attachment path fields.
- Auth checks should use SDK `GetAuthStatus` output contract (`isAuthenticated` + optional detail fields).
