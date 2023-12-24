FROM golang:1.21 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x && go mod verify

ARG GIT_HASH=dev
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=.,target=. \
    CGO_ENABLED=0 GOOS=linux go build -v -ldflags "-X main.BuildCommitHash=${GIT_HASH}" -o /coordinator .

FROM alpine:latest
LABEL org.opencontainers.image.source="https://github.com/sportshead/powergrid"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.title="Powergrid Coordinator"
LABEL org.opencontainers.image.description="Discord bot router"

WORKDIR /
RUN apk --no-cache add gcompat ca-certificates \
  && update-ca-certificates

COPY --from=builder /coordinator /coordinator

EXPOSE 80

ENTRYPOINT ["/coordinator"]
