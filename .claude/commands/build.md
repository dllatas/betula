# Build

Lint, test, and build the server binary.

Arguments: none

## Steps

1. Run `go fmt ./...` and report any files that were reformatted
2. Run `go vet ./...` and fix any issues found
3. Run `go test ./...` and stop if tests fail
4. Run `go build -o bin/server ./cmd/server`
5. Report build success with binary path
