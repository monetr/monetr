# vim: set ft=dockerfile

# Node is pulled out of the official image rather than installed from apt so we
# can pin an exact version. Debian trixie only ever ships Node 20, and the
# upstream image already verifies the nodejs.org tarball against the Node
# release GPG keys and its SHA256 before unpacking it into /usr/local. Copying
# that back out gives us the verified install without having to track the Node
# release key roster ourselves.
FROM node:24.20.0-trixie-slim@sha256:50c3b2f6988dfc307b86e5301d69611af31f4789bdf232863b07d3b02fe55ae0 AS node

FROM debian:13-slim@sha256:d7e12182ce18b85b93007c1dedf31f2d29e01ccf3182cc4017c709b6259bc132
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      # renovate: datasource=deb depName=ca-certificates versioning=deb
      ca-certificates=20250419 \
      # The ui service's compose healthcheck shells out to curl.
      # renovate: datasource=deb depName=curl versioning=deb
      curl=8.14.1-2+deb13u4 \
      # Node links against libatomic on arm64.
      # renovate: datasource=deb depName=libatomic1 versioning=deb
      libatomic1=14.2.0-19 && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Compose runs this container as ${UID:-1000}:${GID:-1000}. Mirror the user the
# upstream node image creates so that uid still has a real home directory.
RUN groupadd --gid 1000 node && \
    useradd --uid 1000 --gid node --shell /bin/bash --create-home node

COPY --from=node /usr/local/bin/node /usr/local/bin/node
COPY --from=node /usr/local/lib/node_modules /usr/local/lib/node_modules
RUN ln -s ../lib/node_modules/npm/bin/npm-cli.js /usr/local/bin/npm && \
    ln -s ../lib/node_modules/npm/bin/npx-cli.js /usr/local/bin/npx

# Pinned to the packageManager version in package.json. Left unpinned, npm
# installs whatever pnpm is latest, that pnpm then tries to self resolve down
# to our pin, and pnpm-workspace.yaml's trustPolicy: no-downgrade rejects it
# because pnpm dropped npm provenance attestations as of 10.34.0.
RUN npm install -g pnpm@10.34.5

USER node
