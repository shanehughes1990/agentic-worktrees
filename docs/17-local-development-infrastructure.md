# 17 - Local Development Infrastructure

## Objective

Provide a local environment that is production-like where needed, while remaining fast for day-to-day development.

## Components

- `redis` via Docker Compose (required for Asynq durability)
- optional `app` service via Docker Compose profile for production-like runtime
- `air` hot-reload flow for rapid local development
- `go-task` for local workflow automation
- `goreleaser` for snapshot/release packaging

## Files

- `docker-compose.yml`
- `Dockerfile`
- `.dockerignore`
- `.air.toml`
- `Taskfile.yml`
- `.goreleaser.yaml`

## Tooling Install

```bash
brew install go-task/tap/go-task goreleaser
go install github.com/air-verse/air@latest
```

## Default Local Workflow (Taskfile)

Run:

```bash
task
```

The default task runs:

1. `task infra` (starts Redis via Docker Compose)
2. `task air` (runs Air hot-reload on your host machine)

Additional tasks:

```bash
task infra
task infra:down
task air
task release:snapshot
```

## Docker Compose Profiles

### Default (`redis` only)

Use when developing app locally outside container but needing durable queue backend.

```bash
docker compose up -d redis
```

### Production-like app (`app` profile)

Use when validating containerized app behavior with Redis dependency.

```bash
docker compose --profile app up --build
```

### Dev container with Air (`dev` profile)

Use when you want code mounted into container and auto-rebuild/restart behavior.

```bash
docker compose --profile dev up --build
```

## Air Hot Reload

### Local host workflow

1. Start Redis:

```bash
docker compose up -d redis
```

2. Run Air:

```bash
air -c .air.toml
```

Equivalent Taskfile command:

```bash
task air
```

## GoReleaser Snapshot

Run a local, non-publishing release build:

```bash
task release:snapshot
```

## Notes on Current Repository State

The Air and GoReleaser build targets currently expect an app entrypoint package at:

- `./cmd/agentic-worktrees`

If that package is not present yet, Redis-only mode still works immediately, while app/dev/release tasks will fail until the entrypoint is added.

## Recommended Environment Variables

- `REDIS_ADDR` (`127.0.0.1:6379` for local host workflow)
- `APP_ENV` (`development` or `production`)

## Health Expectations

- `redis` must report healthy before app/app-dev startup
- app admission gate should remain blocked if preflight dependency/auth/schema checks fail

## Quick Verification

```bash
docker compose ps
docker compose logs -f redis
```

For app profile:

```bash
docker compose --profile app logs -f app
```

For dev profile:

```bash
docker compose --profile dev logs -f app-dev
```
