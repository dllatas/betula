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
- **Library consumers**: start at `lib/view_instance.go` (Append, SpanBack, Filter, GroupBy) and `lib/store.go` (Register, Get)
- **HTTP API**: start at `cmd/server/main.go` for routes, then individual handler files (`append.go`, `span_back.go`, `views.go`, `delete.go`)
- **Persistence**: `lib/wal.go` for write-ahead log, `lib/store.go` for checkpoint save/load, `cmd/server/store.go` for startup recovery sequence

### Large Files
- `cmd/server/main.go` — server bootstrap, middleware, shutdown; treat as reference
- `lib/store.go` — store + WAL replay; read fully before modifying

## Deployment
Build and deploy the betula server container.

### Steps
1. Run `go test ./...` to validate
2. Run `go vet ./...` for correctness
3. Build image: `docker build -t betula:$(git rev-parse --short HEAD) .`
4. Tag for registry if needed: `docker tag betula:<sha> <registry>/betula:<sha>`
5. Push: `docker push <registry>/betula:<sha>`

### Notes
- The Dockerfile uses `COMMIT_SHA` build arg to embed version via `-ldflags`
- Server defaults: `PORT=6357`, `DATAPATH=./checkpoint.db`
- For in-memory mode set `DATAPATH=` (empty string)
