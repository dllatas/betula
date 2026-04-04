# Test

Run the test suite with optional targeting.

Arguments: $ARGUMENTS (optional test name pattern, e.g. "TestSpanBack")

## Steps

1. If arguments are provided, detect the package:
   - Search `lib/` and `cmd/server/` for the test name
   - Run `go test ./<package> -run $ARGUMENTS -v`
2. If no arguments, run the full suite: `go test ./... -v`
3. Run `go vet ./...` to catch correctness issues
4. Report pass/fail summary with any failing test names and output
