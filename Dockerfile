# syntax=docker/dockerfile:1
FROM golang:1.21 AS builder

WORKDIR /app
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x && go mod verify

FROM builder AS coordinator-builder

ARG GIT_HASH=dev
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=bind,source=.,target=. \
    CGO_ENABLED=0 GOOS=linux go build -v -ldflags \
    "-s -w -X github.com/sportshead/powergrid/pkg/version.BuildCommitHash=${GIT_HASH}" -o /coordinator ./cmd/coordinator

FROM gcr.io/distroless/static-debian12:nonroot AS base

ARG GIT_HASH=dev
LABEL org.opencontainers.image.source="https://github.com/sportshead/powergrid"
LABEL org.opencontainers.image.url="https://github.com/sportshead/powergrid"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.base.name="gcr.io/distroless/static-debian12:nonroot"
LABEL org.opencontainers.image.revision="${GIT_HASH}"

FROM base AS coordinator
LABEL org.opencontainers.image.title="Powergrid Coordinator"
LABEL org.opencontainers.image.description="Like nginx, but for Discord bots"

COPY --from=coordinator-builder /coordinator /coordinator

EXPOSE 8000/tcp
ENTRYPOINT ["/coordinator"]
