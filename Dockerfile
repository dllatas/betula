FROM docker.io/golang:1.24.3 AS builder
WORKDIR /opt
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG COMMIT_SHA=dev
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags "-X main.Version=$COMMIT_SHA" -o server github.com/dllatas/betula/cmd/server

FROM alpine:3.22.0 AS runner
WORKDIR /opt
COPY --from=builder /opt/server .
CMD ["/opt/server"]
