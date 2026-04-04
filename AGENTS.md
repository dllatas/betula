# Agent Roles

## Reviewer
Review pull requests and code changes for betula.

### Checklist
- All exported API changes have corresponding test updates
- Time boundary logic (span-back, range) includes edge-case tests for inclusive/exclusive windows
- WAL and store persistence paths handle empty, missing, or partial files gracefully
- No new dependencies introduced without justification — standard library first
- HTTP concerns stay in `cmd/server/`, data/time logic stays in `lib/`
- `go vet ./...` and `go test ./...` pass cleanly
- Commit messages follow project convention: short imperative lowercase with optional scope

## Explorer
Navigate and explain the betula codebase.

### Entry Points
- **Library consumers**: start at `lib/view_instance.go` for the `ViewInstance` type, then see `lib/append.go`, `lib/span_back.go`, `lib/filter.go`, `lib/group_by.go` for the main operations, and `lib/store.go` for `Register`/`Get`
- **HTTP API**: start at `cmd/server/main.go` for routes, then individual handler files (`append.go`, `span_back.go`, `views.go`, `delete.go`)
- **Persistence**: `lib/wal.go` for write-ahead log, `lib/store.go` for checkpoint save/load, `cmd/server/store.go` for startup recovery sequence

### Large Files
- `cmd/server/main.go` — server bootstrap, middleware, shutdown; treat as reference
- `lib/store.go` — store + WAL replay; read fully before modifying

## Deployment
Build and deploy the betula server container.

### Production (Tekton PaC)
CI is handled by `.tekton/betula-master.yaml` (push to master) and `.tekton/betula-pr.yaml` (PRs). The pipeline runs clone → test → build+push automatically. Images land at `harbor.harokilabs.com/staging/betula:<sha>` and `:<branch-tag>`.

### Local Dev Build
1. Run `go test ./...` to validate
2. Run `go vet ./...` for correctness
3. Build image: `docker build --build-arg COMMIT_SHA=$(git rev-parse --short HEAD) -t betula:$(git rev-parse --short HEAD) .`

### Notes
- The Dockerfile uses `COMMIT_SHA` build arg to embed version via `-ldflags`
- Server defaults: `PORT=6357`, `DATAPATH=./checkpoint.db`
- For in-memory mode set `DATAPATH=` (empty string)
