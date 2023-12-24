FROM oven/bun:1 as base
WORKDIR /app

# install dependencies into temp directory
# this will cache them and speed up future builds
FROM base AS install
# install with --production (exclude devDependencies)
RUN mkdir -p /temp/prod
COPY package.json bun.lockb /temp/prod/
RUN cd /temp/prod && bun install --frozen-lockfile --production

# copy production dependencies and source code into final image
FROM base AS release

LABEL org.opencontainers.image.source="https://github.com/sportshead/powergrid"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.title="Bun Example Bot"
LABEL org.opencontainers.image.description="Example discord bot using Bun to demonstrate Powergrid"

COPY --from=install /temp/prod/node_modules node_modules
COPY bot.ts .
COPY package.json .

# run the app
USER bun
EXPOSE 3000/tcp
ENTRYPOINT [ "bun", "run", "bot.ts" ]
