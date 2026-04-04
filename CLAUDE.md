# Betula Agent Guide

## Purpose
Betula is an in-memory time-partitioned tree engine for aggregating and querying hierarchical event data. It ships as a reusable Go library (`lib/`) and a thin Echo HTTP server (`cmd/server/`).

## Stack
- Go 1.24, standard library first
- Echo v4 (HTTP), envconfig (config), Prometheus (metrics)
- gob encoding for WAL and checkpoint persistence
- Docker (multi-stage build, Alpine runner)

## Repo Map
- `lib/` — core data structures: views, shards, ranges, span-back windows, grouping, filtering, delete, store/WAL persistence
- `cmd/server/` — HTTP handlers, server bootstrap, middleware, config loading, API tests
- `.tekton/` — Pipelines-as-Code definitions (push-triggered CI/CD)
- `Dockerfile` — container build for the server binary
- `README.md` — library usage examples and coverage commands

## Commands
- `go fmt ./...` — required formatting pass before finishing
- `go vet ./...` — correctness check for touched packages
- `go test ./...` — full test suite (run before handing off work)
- `go test ./lib -run TestName` — target a single library test
- `go test ./cmd/server -run TestName` — target an API handler test
- `go test ./... -coverprofile=coverage.out` — coverage run
- `go tool cover -html=coverage.out` — HTML coverage report
- `go run ./cmd/server` — start API locally (default port 6357)
- `go build ./cmd/server` — build server binary
- `docker build -t betula .` — build container image

## Environment
- `PORT` — server listen port (default: `6357`)
- `DATAPATH` — checkpoint file path (default: `./checkpoint.db`); set empty for in-memory mode

## Runtime Behavior You Must Preserve
- When `DATAPATH` is empty, server runs fully in-memory — no file I/O
- WAL replay must handle empty, missing, or partially written files gracefully
- Store/WAL operations are mutex-guarded — never introduce lock-order changes
- Graceful shutdown: Echo shutdown with 20s timeout, then WAL close
- Duplicate `create-view` WAL entries are silently skipped if equal

## API Surface
- `GET /healthz` — health check (204)
- `GET /metrics` — Prometheus metrics
- `POST /v1/views` — create a new view
- `POST /v1/append` — append events to a view
- `POST /v1/delete` — delete events from a view
- `POST /v1/spanback` — query a time window with filter/group-by

## Editing Rules
- Keep new packages rare — most changes belong in `lib/` or `cmd/server/`
- Tests live next to source, use table-driven cases when behavior branches
- HTTP concerns stay in `cmd/server/`; time/window logic stays in `lib/`
- Prefer standard library before adding dependencies
- When changing exported APIs or routes, update `README.md` and tests in the same change

## CI/CD
- Tekton Pipelines-as-Code in `.tekton/`:
  - `betula-master.yaml` — push to master: clone → test → build+push
  - `betula-pr.yaml` — PRs against master: same pipeline, tagged `pr-<number>`
- Image: `harbor.harokilabs.com/staging/betula:<sha>` + `:<branch-tag>`
- Runs in namespace `ci`, service account `ci-services`
- Secrets: `docker-credentials` (Harbor push), `{{git_auth_secret}}` (clone auth)
- The build step passes `COMMIT_SHA` build arg to embed version in the binary

## High-Risk Areas
- `lib/store.go`, `lib/wal.go` — persistence and replay; preserve graceful degradation
- `lib/span_back.go`, `lib/range.go` — inclusive/exclusive window boundaries regress easily; extend tests around boundary timestamps
- `lib/` mutex-guarded paths — no lock-order changes or read/write races
- `cmd/server/main.go` — shutdown, WAL close, config defaults, middleware wiring

## Testing Expectations
- Run `go test ./...` before finishing any work
- Run `go vet ./...` when touching server bootstrap, persistence, or exported library behavior
- Add or update tests for time boundaries, WAL replay, delete behavior, HTTP validation

## Commit Style
- Short imperative lowercase subjects with optional scope: `server (cmd): add handler tests`
- Keep commits focused; explain why the behavior changed, not just what files moved
