FROM oven/bun:1-alpine as base

# install dependencies into temp directory
# this will cache them and speed up future builds
FROM base AS builder

RUN --mount=type=cache,target=/root/.bun/install/cache \
    --mount=type=bind,source=package.json,target=package.json \
    --mount=type=bind,source=bun.lockb,target=bun.lockb \
    bun install --frozen-lockfile --production

COPY . src/
RUN --mount=type=cache,target=/root/.bun/install/cache \
    --mount=type=bind,source=package.json,target=package.json \
    --mount=type=bind,source=bun.lockb,target=bun.lockb \
    bun build src/bot.ts --target=bun --outfile=/bot.js

# copy production dependencies and source code into final image
FROM base AS release

LABEL org.opencontainers.image.source="https://github.com/sportshead/powergrid"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.title="Bun Example Bot"
LABEL org.opencontainers.image.description="Example discord bot using Bun to demonstrate Powergrid"

WORKDIR /

RUN apk --no-cache add ca-certificates \
    && update-ca-certificates

COPY --from=builder /bot.js /bot.js

USER bun

EXPOSE 3000/tcp
ENTRYPOINT [ "bun", "run", "/bot.js" ]
