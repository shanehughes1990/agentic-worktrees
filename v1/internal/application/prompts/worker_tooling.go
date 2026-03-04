package prompts

const WorkerToolingBaseline = `Execution environment baseline (Docker worker):
- OS/runtime: Debian-based container with shell access.
- Core CLI: bash, sh, git, gh, copilot, curl, wget, tar, unzip, zip, make.
- Data/query tooling: jq, yq, sqlite3.
- Search/inspection: ripgrep (rg), fd, tree, less, findutils, coreutils, moreutils, file, xxd.
- Sync/automation: rsync, fzf, entr.
- Container tooling: docker CLI + dockerd (DinD capable with privileged runtime).
- Process/network diagnostics: procps (ps), lsof, netcat (nc), dnsutils (dig), iputils-ping (ping), tcpdump, traceroute, mtr, strace, ltrace (amd64 builds).
- Document/shell tooling: pandoc, shellcheck, shfmt, tini.
- Language runtimes/toolchains: python3 + pip3 + venv, node + npm, openjdk-17 (java/javac), ruby, php, go + gofmt, rustc + cargo, gcc + g++, cmake, pkg-config.
- Go developer tooling: gopls, goimports, dlv, staticcheck, golangci-lint, gotestsum, mockgen, govulncheck, task.

Use only tools available in this environment. Prefer deterministic, non-interactive commands, and keep file writes scoped to explicit project paths provided by the task.`
