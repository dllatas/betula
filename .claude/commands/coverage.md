# Coverage

Run tests with coverage analysis and report results.

Arguments: $ARGUMENTS (optional package path, e.g. "./lib")

## Steps

1. If arguments provided: `go test $ARGUMENTS -coverprofile=coverage.out -v`
2. Otherwise: `go test ./... -coverprofile=coverage.out`
3. Run `go tool cover -func=coverage.out` to get per-function coverage
4. Report overall coverage percentage and any functions below 50% coverage
5. Mention `go tool cover -html=coverage.out` for visual inspection
