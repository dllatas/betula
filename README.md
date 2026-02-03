# Betula

A high-performance, in-memory time-partitioned tree engine for aggregating and querying hierarchical event data across user-defined views.

## Test Coverage

- Run coverage across the repo: `go test ./... -coverprofile=coverage.out`
- Inspect the HTML report locally: `go tool cover -html=coverage.out`
- Coverage result files (`coverage.out`, `cover.out`, etc.) are build artifacts; keep them out of version control.
