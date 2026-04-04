# Deploy

Build a Docker image for the betula server.

Arguments: $ARGUMENTS (optional image tag, defaults to git short SHA)

## Steps

1. Run `go test ./...` to validate before building
2. Determine tag: use `$ARGUMENTS` if provided, otherwise `git rev-parse --short HEAD`
3. Run `docker build --build-arg COMMIT_SHA=$(git rev-parse --short HEAD) -t betula:<tag> .`
4. Report the image name and tag
5. Ask whether to push to a registry
